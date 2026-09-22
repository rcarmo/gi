import {test,expect} from '@playwright/test';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
const file=name=>({name,mimeType:'text/plain',buffer:Buffer.from(`bytes:${name}`)});
async function fixture(page,request,info){
 const token=`return-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-child',title:token+'-child'}})).json()).branch.chat_jid.slice(3);
 const path=resolve('test-results/ux-parity/queue-gates',token);mkdirSync(resolve(path,'..'),{recursive:true});const release=()=>writeFileSync(path,'ok');
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 await input.fill(`UX queue gate:${token}`);await input.press('Enter');await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeVisible({timeout:15000});
 const media=await(await request.post(`/api/sessions/${main.id}/media`,{multipart:{file:file('queued.txt')}})).json();
 const prompt=`queued origin\n\nFiles:\n- folder/\n- queued-file.txt\n\nReferenced messages:\n- message:source-message\n\nAttachments:\n- attachment:${media.media.id} (queued.txt)`;
 const queued=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt,intent:'queue',model:'test-model',media:[media.ref]}})).json();
 const row=page.locator(`[data-queue-id="${queued.turn_id}"]`);await expect(row).toBeVisible();
 const stored=async()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const req=indexedDB.open('gi-session-drafts',1);req.onsuccess=()=>resolve(req.result);req.onerror=()=>reject(req.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly');const req=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const row=req.result;resolve({text:row?.draft.text,media:row?.draft.media.map(x=>x.name),fileRefs:row?.draft.fileRefs,messageRefs:row?.draft.messageRefs,recovery:row?.queueReturns});};tx.onerror=()=>reject(tx.error);});
 },main.id);
 return{main,child,queued,media,input,row,stored,release};
}

test('@shared-28 Return a queued item to the latest editor draft',async({page,request},info)=>{
 const scenario=loadCorpus('shared').find(row=>row.id==='@shared-28');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
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
  expect(snapshot.text).toContain('queued origin');expect(snapshot.text).toContain('concurrent latest');
  expect(snapshot.media).toEqual(['queued.txt','new.txt']);expect(snapshot.fileRefs).toContain('folder/');expect(snapshot.messageRefs).toContain('source-message');
  if(failDelete){await route.abort('failed');return;}await route.continue();
 });
 try{
  await input.fill('earlier draft');await row.getByRole('button',{name:'Return queued message to editor'}).click();await expect.poll(()=>held).toBe(true);
  await input.fill('concurrent latest');await page.locator('.compose-box input[type=file]').setInputFiles(file('new.txt'));
  // Storage failure only for recovery writes; ordinary native edits still persist.
  await page.evaluate(()=>{window.__put=IDBObjectStore.prototype.put;IDBObjectStore.prototype.put=function(value,...args){if(this.name==='drafts'&&Object.keys(value.queueReturns||{}).length)throw new DOMException('Recovery quota failure','QuotaExceededError');return window.__put.call(this,value,...args);};});
  unblock();await done;
  await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();expect(deletes).toBe(0);
  await expect(row).toBeVisible();await expect(input).toHaveValue('queued origin\n\nconcurrent latest');
  await page.evaluate(()=>{IDBObjectStore.prototype.put=window.__put;});await page.unroute(mediaURL);
  // Retry persists the existing merge, then an injected transport error retains
  // the server row. A reload must preserve both draft and idempotency record.
  await row.getByRole('button',{name:'Return queued message to editor'}).click();
  await expect.poll(()=>deletes).toBe(1);await expect(row.getByRole('button',{name:'Return queued message to editor'})).toBeEnabled();
  await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();
  await page.reload();await expect(row).toBeVisible();await expect(input).toHaveValue('queued origin\n\nconcurrent latest');
  failDelete=false;await row.getByRole('button',{name:'Return queued message to editor'}).click();await expect.poll(()=>deletes).toBe(2);await expect(row).toHaveCount(0);
  await expect(input).toHaveValue('queued origin\n\nconcurrent latest');await expect(input).toBeFocused();
  await expect.poll(()=>input.evaluate(el=>el.selectionStart)).toBe('queued origin\n\nconcurrent latest'.length);
  await expect.poll(async()=> (await stored()).recovery[queued.turn_id].state).toBe('removed');
  await expect(page.locator('.compose-file-pill[title="queued.txt"]')).toHaveCount(1);await expect(page.locator('.compose-file-pill[title="new.txt"]')).toHaveCount(1);
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
  release();await expect.poll(async()=>{const{turns}=await(await request.get(`/api/sessions/${main.id}/turns`)).json();return turns.find(t=>t.id===queued.turn_id).status;},{timeout:15000}).toBe('completed');
  unblock();await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();
  await expect(input).toHaveValue('queued origin\n\nnew unsent draft');await expect(row).toHaveCount(0);
  await page.reload();await expect(page.getByRole('alert').filter({hasText:'Queue return incomplete'})).toBeVisible();
  await expect(input).toHaveValue('queued origin\n\nnew unsent draft');
 }finally{unblock();release();}
});

test('Gi queue return after selection changes recovers only the origin',async({page,request},info)=>{
 const{main,child,queued,media,input,row,stored,release}=await fixture(page,request,info);
 let unblock,held=false;const gate=new Promise(resolve=>{unblock=resolve;});
 await page.route(`**/api/sessions/${main.id}/media/${media.media.id}`,async route=>{const response=await route.fetch();held=true;await gate;await route.fulfill({response});});
 try{
  await input.fill('origin draft');await row.getByRole('button',{name:'Return queued message to editor'}).click();await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${child}"]`).getByRole('menuitem').click();
  await input.fill('target draft');unblock();
  await expect.poll(async()=> (await stored()).recovery?.[queued.turn_id]?.state).toBe('removed');
  await expect(input).toHaveValue('target draft');await expect(page.getByRole('alert')).toHaveCount(0);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();
  await expect(input).toHaveValue('queued origin\n\norigin draft');await expect(page.locator('.compose-file-pill[title="queued.txt"]')).toBeVisible();
 }finally{unblock();release();}
});
