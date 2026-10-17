import {test,expect} from '@playwright/test';
import {attachGiDeviation} from '../ux/support/gi-deviations.mjs';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
const file=name=>({name,mimeType:'text/plain',buffer:Buffer.from(`bytes:${name}`)});
async function fixture(page,request,info){
 const token=`return-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-child',title:token+'-child'}})).json()).branch.chat_jid.slice(3);
 const path=resolve(process.env.GI_UX_QUEUE_GATES || `${process.env.GI_TEST_RUN_ROOT}/results/ux-parity/queue-gates`,token);mkdirSync(resolve(path,'..'),{recursive:true});const release=()=>writeFileSync(path,'ok');
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 await input.fill(`UX queue gate:${token}`);await input.press('Enter');await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeVisible({timeout:15000});
 const media=await(await request.post(`/api/sessions/${main.id}/media`,{multipart:{file:file('queued.txt')}})).json();
 const prompt=`queued origin\n\nFiles:\n- folder/\n- queued-file.txt\n\nReferenced messages:\n- message:source-message\n\nAttachments:\n- attachment:${media.media.id} (queued.txt)`;
 const queued=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt,intent:'queue',model:'test-model',media:[media.ref]}})).json();
 const row=page.getByTestId('queue-item').filter({hasText:'queued origin'});await expect(row).toBeVisible();
 const stored=async()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const req=indexedDB.open('gi-session-drafts');req.onsuccess=()=>resolve(req.result);req.onerror=()=>reject(req.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly');const req=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const row=req.result;resolve({text:row?.draft.text,media:row?.draft.media.map(x=>x.name),fileRefs:row?.draft.fileRefs,messageRefs:row?.draft.messageRefs,recovery:row?.queueReturns});};tx.onerror=()=>reject(tx.error);});
 },main.id);
 return{main,child,queued,media,input,row,stored,release};
}

test('Return replaces the editor and fences recovery-time edits without duplicate recovery',async({page,request},info)=>{
 await attachGiDeviation(info,'@gi-ux-004');
 const{main,queued,media,input,row,stored,release}=await fixture(page,request,info);
 let unblock,held=false,delivered;const gate=new Promise(resolve=>{unblock=resolve;});const done=new Promise(resolve=>{delivered=resolve;});
 const mediaURL=`**/api/sessions/${main.id}/media/${media.media.id}`;
 const deleteURL=`**/api/sessions/${main.id}/queue/${queued.turn_id}`;
 let deletes=0;let failDelete=true;
 await page.route(mediaURL,async route=>{const response=await route.fetch();held=true;await gate;await route.fulfill({response});delivered();});
 await page.route(deleteURL,async route=>{
  deletes++;
  // Inspect committed IndexedDB before letting the actual DELETE run.
  const snapshot=await stored();expect(snapshot.recovery[queued.turn_id].state).toBe('prepared');
  expect(snapshot.text).toBe('queued origin');
  expect(snapshot.media).toEqual(['queued.txt']);expect(snapshot.fileRefs).toContain('folder/');expect(snapshot.messageRefs).toContain('source-message');
  if(failDelete){await route.abort('failed');return;}await route.continue();
 });
 try{
  await input.fill('earlier draft');await row.getByRole('button',{name:'Return queued message to editor'}).click();await expect.poll(()=>held).toBe(true);
  await input.fill('concurrent latest');await page.locator('.compose-box input[type=file]').setInputFiles(file('new.txt'));
  // Storage failure only for recovery writes; ordinary native edits still persist.
  await page.evaluate(()=>{window.__put=IDBObjectStore.prototype.put;IDBObjectStore.prototype.put=function(value,...args){if(this.name==='drafts'&&Object.keys(value.queueReturns||{}).length)throw new DOMException('Recovery quota failure','QuotaExceededError');return window.__put.call(this,value,...args);};});
  unblock();await done;
  await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();expect(deletes).toBe(0);
  await expect(page.getByRole('alert').filter({hasText:'Draft changed while returning'})).toBeVisible();
  await expect(row).toBeVisible();await expect(input).toHaveValue('concurrent latest');
  // Retry explicitly authorises replacement, but failed persistence keeps queue.
  await row.getByRole('button',{name:'Return queued message to editor'}).click();
  await expect(input).toHaveValue('queued origin');await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();expect(deletes).toBe(0);
  await page.evaluate(()=>{IDBObjectStore.prototype.put=window.__put;});await page.unroute(mediaURL);
  // Retry persists the existing replacement, then an injected transport error retains
  // the server row. A reload must preserve both draft and idempotency record.
  await row.getByRole('button',{name:'Return queued message to editor'}).click();
  await expect.poll(()=>deletes).toBe(1);await expect(row.getByRole('button',{name:'Return queued message to editor'})).toBeEnabled();
  await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();
  await page.reload();await expect(row).toBeVisible();await expect(input).toHaveValue('queued origin');
  failDelete=false;await row.getByRole('button',{name:'Return queued message to editor'}).click();await expect.poll(()=>deletes).toBe(2);await expect(row).toHaveCount(0);
  await expect(input).toHaveValue('queued origin');await expect(input).toBeFocused();
  await expect.poll(()=>input.evaluate(el=>el.selectionStart)).toBe('queued origin'.length);
  await expect.poll(async()=> (await stored()).recovery[queued.turn_id].state).toBe('removed');
  await expect(page.locator('.compose-file-pill[title="queued.txt"]')).toHaveCount(1);await expect(page.locator('.compose-file-pill[title="new.txt"]')).toHaveCount(0);
  const{turns}=await(await request.get(`/api/sessions/${main.id}/turns`)).json();expect(turns.find(t=>t.id===queued.turn_id).status).toBe('cancelled');
 }finally{unblock();release();}
});

test('Gi consumed queue return retains recovered data and reports an incomplete return',async({page,request},info)=>{
 const{main,queued,input,row,stored,release}=await fixture(page,request,info);
 let unblock,held=false;const gate=new Promise(resolve=>{unblock=resolve;});
 await page.route(`**/api/sessions/${main.id}/queue/${queued.turn_id}`,async route=>{
  held=true;await gate;const response=await route.fetch();expect(response.status()).toBe(409);await route.fulfill({response});
 });
 try{
  await input.fill('new unsent draft');await row.getByRole('button',{name:'Return queued message to editor'}).click();await expect.poll(()=>held).toBe(true);
  expect((await stored()).recovery[queued.turn_id].state).toBe('prepared');
  await input.fill('queued origin edited after replacement');
  release();await expect.poll(async()=>{const{turns}=await(await request.get(`/api/sessions/${main.id}/turns`)).json();return turns.find(t=>t.id===queued.turn_id).status;},{timeout:15000}).toBe('completed');
  unblock();await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();
  await expect(input).toHaveValue('queued origin edited after replacement');await expect(row).toHaveCount(0);
  await page.reload();await expect(page.getByRole('alert').filter({hasText:'Queue return incomplete'})).toBeVisible();
  await expect(input).toHaveValue('queued origin edited after replacement');
 }finally{unblock();release();}
});

test('Gi queue return after selection changes recovers only the origin',async({page,request},info)=>{
 const{main,child,queued,media,input,row,stored,release}=await fixture(page,request,info);
 let unblock,held=false;const gate=new Promise(resolve=>{unblock=resolve;});
 await page.route(`**/api/sessions/${main.id}/media/${media.media.id}`,async route=>{const response=await route.fetch();held=true;await gate;await route.fulfill({response});});
 try{
  await input.fill('origin draft');await row.getByRole('button',{name:'Return queued message to editor'}).click();await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[id="compose-session-option-gi%3A${child}"]`).click();
  await input.fill('target draft');unblock();
  await expect.poll(async()=> (await stored()).recovery?.[queued.turn_id]?.state).toBe('removed');
  await expect(input).toHaveValue('target draft');await expect(page.getByRole('alert')).toHaveCount(0);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[id="compose-session-option-gi%3A${main.id}"]`).click();
  await expect(input).toHaveValue('queued origin');await expect(page.locator('.compose-file-pill[title="queued.txt"]')).toBeVisible();
 }finally{unblock();release();}
});
