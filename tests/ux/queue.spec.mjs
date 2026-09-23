import { test, expect } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { createServer, request as httpRequest } from 'node:http';
import { loadCorpus } from './support/catalogue.mjs';

const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function setup(page,request,info){
 const token=`${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:`@${token}`}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:`${token}-child`,title:`${token}-child`}})).json()).branch.chat_jid.slice(3);
 const gate=resolve('test-results/ux-parity/queue-gates',token);mkdirSync(resolve(gate,'..'),{recursive:true});
 const release=()=>writeFileSync(gate,'release');
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const starting=page.waitForResponse(res=>res.url().endsWith(`/api/sessions/${main.id}/prompt`)&&res.request().method()==='POST');
 await input.fill(`UX queue gate:${token}`);await input.press('Enter');
 const active=await(await starting).json();
 await expect.poll(async()=>{
  const state=await(await request.get(`/api/sessions/${main.id}`)).json();return state.state.status;
 }).toBe('running');
 await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeVisible({timeout:15000});
 const queue=async()=> (await(await request.get(`/api/sessions/${main.id}/queue`)).json()).items;
 const enqueue=async text=>{
  const accepted=page.waitForResponse(res=>res.url().endsWith(`/api/sessions/${main.id}/prompt`)&&res.request().method()==='POST');
  await input.fill(text);await input.press('Enter');
  const response=await accepted;expect(response.status()).toBe(202);
  const result=await response.json();expect(result.queued).toBe(true);
  await expect(page.locator('.compose-queue-stack-text').filter({hasText:text})).toBeVisible();
  return result.turn_id;
 };
 const switchTo=async id=>{
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();
  await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();
  await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);
 };
 return {main,child,input,active,queue,enqueue,release,switchTo};
}
const row=(page,id)=>page.locator(`[data-queue-id="${id}"]`);
const ids=page=>page.locator('[data-queue-id]').evaluateAll(elements=>elements.map(el=>el.dataset.queueId));

test('@ux-original-016 Display queued follow-ups during a busy turn',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-original-016');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const {main,input,queue,release}=await setup(page,request,info);
 let deliver,held=false;const gate=new Promise(resolve=>{deliver=resolve;});let accepted;
 let delivered;const delivery=new Promise(resolve=>{delivered=resolve;});
 try {
  await page.route(`**/api/sessions/${main.id}/prompt`,async route=>{
   const response=await route.fetch();accepted=await response.json();held=true;await gate;await route.fulfill({response});delivered();
  });
  await input.fill('same queued text');await input.press('Enter');
  await expect(page.locator('.compose-queue-stack-text').filter({hasText:'same queued text'})).toHaveCount(1);
  await expect.poll(()=>held).toBe(true);
  // SSE refresh arrives while the HTTP acknowledgement is held. Correlate by
  // client token, not text (another identical prompt may be legitimate).
  await expect.poll(()=>ids(page)).toEqual([accepted.turn_id]);
  await expect(page.locator('.compose-queue-stack-text').filter({hasText:'same queued text'})).toHaveCount(1);
  deliver();await delivery;await page.unroute(`**/api/sessions/${main.id}/prompt`);
  await expect(row(page,accepted.turn_id).getByRole('button',{name:'Cancel queued message'})).toBeEnabled();
  const external=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'same queued text',intent:'queue',model:'test-model'}})).json();
  await expect(row(page,external.turn_id)).toBeVisible({timeout:4000});
  await expect(page.locator('.compose-queue-stack-text').filter({hasText:'same queued text'})).toHaveCount(2);
  const deletion=await request.delete(`/api/sessions/${main.id}/queue/${accepted.turn_id}`);expect(deletion.status()).toBe(200);
  await expect.poll(()=>ids(page),{timeout:4000}).toEqual([external.turn_id]);
  expect((await queue()).map(item=>item.id)).toEqual([external.turn_id]);
 } finally {deliver();release();}
});

test('Gi rejected queued send removes its placeholder and restores only its origin draft',async({page,request},info)=>{
 const {main,input,release}=await setup(page,request,info);
 let unblock;const gate=new Promise(resolve=>{unblock=resolve;});let held=false;
 try {
  await page.route(`**/api/sessions/${main.id}/prompt`,async route=>{held=true;await gate;await route.abort('failed');});
  await input.fill('rejected queue text');await input.press('Enter');await expect.poll(()=>held).toBe(true);
  await expect(page.locator('[data-queue-id][aria-busy="true"]')).toHaveCount(1);
  await input.fill('new draft');unblock();
  await expect(input).toHaveValue('rejected queue text\n\nnew draft');
  await expect(page.locator('[data-queue-id]')).toHaveCount(0);
  await expect(page.getByRole('alert')).toContainText('Delivery is unknown');
 } finally {unblock();release();}
});

test('@ux-reconnect-001 Clear transient streaming state when the connection drops',async({page,request},info)=>{
 // Byte-for-byte transport proxy: sever real SSE sockets without injecting
 // timeline events. Browser offline emulation does not close established SSE.
 let blocked=false;const connections=new Set();
 const proxy=createServer((req,res)=>{
  res.setHeader('Access-Control-Allow-Origin','*');
  if(blocked){res.writeHead(503);res.end();return;}
  connections.add(res);res.on('close',()=>connections.delete(res));
  const upstream=httpRequest(new URL(req.url,process.env.GI_TEST_URL||'http://127.0.0.1:19091'),source=>{
   res.writeHead(source.statusCode,{'Content-Type':'text/event-stream','Cache-Control':'no-cache','Access-Control-Allow-Origin':'*'});
   source.pipe(res);
  });
  upstream.on('error',()=>res.destroy());res.on('close',()=>upstream.destroy());upstream.end();
 });
 await new Promise(resolve=>proxy.listen(0,'127.0.0.1',resolve));
 const proxyURL=`http://127.0.0.1:${proxy.address().port}`;
 await page.addInitScript(origin=>{
  const Native=window.EventSource;
  window.EventSource=class extends Native{constructor(url,options){const path=new URL(url,location.href);super(path.pathname==='/sse/stream'?origin+path.pathname+path.search:url,options);}};
 },proxyURL);
 const scenario=loadCorpus().find(row=>row.id==='@ux-reconnect-001');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const {main,input,enqueue,release}=await setup(page,request,info);
 try {
  const queued=await enqueue('surviving queue');
  await expect(page.locator('.agent-thinking').filter({hasText:'Queue gate streaming preview'})).toBeVisible();
  await input.fill('offline unsent draft');
  blocked=true;for(const response of connections)response.destroy();
  await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});
  await expect(page.getByRole('button',{name:'Stop response',exact:true})).toHaveCount(0);
  await expect(page.locator('.agent-thinking, .agent-status')).toHaveCount(0);
  await expect(input).toHaveValue('offline unsent draft');
  // The independent API actor changes persisted state during the SSE outage.
  await request.delete(`/api/sessions/${main.id}/queue/${queued}`);
  const replacement=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'offline queue update',intent:'queue',model:'test-model'}})).json();
  blocked=false;
  await expect(page.locator('.compose-connection-status')).toHaveCount(0,{timeout:15000});
  await expect.poll(()=>ids(page)).toEqual([replacement.turn_id]);
  await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeVisible();
  await expect(input).toHaveValue('offline unsent draft');
 } finally {blocked=false;release();for(const response of connections)response.destroy();await new Promise(resolve=>proxy.close(resolve));}
});

test('@ux-original-018 Reorder and remove queued items with failure reconciliation',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-original-018');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const {main,input,queue,enqueue,release}=await setup(page,request,info);
 try {
  const first=await enqueue('first follow-up'),second=await enqueue('second follow-up'),third=await enqueue('third follow-up');
  await input.fill('keep this draft');
  let unblock,held=false;const gate=new Promise(resolve=>{unblock=resolve;});
  await page.route(`**/api/sessions/${main.id}/queue`,async route=>{
   if(route.request().method()!=='PATCH')return route.continue();
   held=true;await gate;await route.continue();
  });
  await row(page,third).getByRole('button',{name:'Move up in queue',exact:true}).click();
  await expect.poll(()=>held).toBe(true);
  await expect.poll(()=>ids(page)).toEqual([first,third,second]); // Optimistic, server still unchanged.
  expect((await queue()).map(t=>t.id)).toEqual([first,second,third]);
  unblock();
  await expect.poll(async()=> (await queue()).map(t=>t.id)).toEqual([first,third,second]);
  await expect(row(page,third).getByRole('button',{name:'Cancel queued message'})).toBeEnabled();
  await page.unroute(`**/api/sessions/${main.id}/queue`);
  await page.reload();await expect.poll(()=>ids(page)).toEqual([first,third,second]);await expect(input).toHaveValue('keep this draft');

  // Real conflict: add a native queued turn after the rendered snapshot but
  // before forwarding the reorder. The backend must reject the stale order.
  let added;
  await page.route(`**/api/sessions/${main.id}/queue`,async route=>{
   if(route.request().method()!=='PATCH')return route.continue();
   added=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'racing append',intent:'queue',model:'test-model'}})).json();
   const response=await route.fetch();expect(response.status()).toBe(409);await route.fulfill({response});
  });
  await row(page,third).getByRole('button',{name:'Move up in queue',exact:true}).click();
  await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();
  await expect.poll(()=>ids(page)).toEqual([first,third,second,added.turn_id]);
  await page.unroute(`**/api/sessions/${main.id}/queue`);

  let releaseFailure;const failureGate=new Promise(resolve=>{releaseFailure=resolve;});let cancelHeld=false;
  await page.route(`**/api/sessions/${main.id}/queue/${second}`,async route=>{
   cancelHeld=true;await failureGate;await route.abort('failed');
  });
  await row(page,second).getByRole('button',{name:'Cancel queued message'}).click();
  await expect.poll(()=>cancelHeld).toBe(true);await expect(row(page,second)).toHaveCount(0);
  expect((await queue()).map(t=>t.id)).toContain(second);
  releaseFailure();await expect(row(page,second)).toBeVisible();
  await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();
  await expect(row(page,second).getByRole('button',{name:'Cancel queued message'})).toBeEnabled();
  await page.unroute(`**/api/sessions/${main.id}/queue/${second}`);
  await row(page,second).getByRole('button',{name:'Cancel queued message'}).click();
  await expect.poll(async()=> (await queue()).map(t=>t.id)).toEqual([first,third,added.turn_id]);
  await expect(input).toHaveValue('keep this draft');
  release();
  await expect.poll(async()=>{
   const {turns}=await(await request.get(`/api/sessions/${main.id}/turns`)).json();return turns.filter(t=>[first,third,added.turn_id].includes(t.id)).every(t=>t.status==='completed');
  },{timeout:15000}).toBe(true);
  const {messages}=await(await request.get(`/api/sessions/${main.id}/messages`)).json();
  expect(messages.filter(m=>m.role==='user').map(m=>m.content).slice(1)).toEqual(['first follow-up','third follow-up','racing append']);
 } finally {release();}
});

test('Gi late queue poll cannot resurrect a cancelled durable row',async({page,request},info)=>{
 const {main,enqueue,release}=await setup(page,request,info);
 let deliver;const gate=new Promise(resolve=>{deliver=resolve;});let held=false;let delivered;const done=new Promise(resolve=>{delivered=resolve;});
 try {
  const id=await enqueue('held poll queued');
  await page.route(`**/api/sessions/${main.id}/queue`,async route=>{
   if(held||route.request().method()!=='GET')return route.continue();
   const response=await route.fetch();held=true;await gate;await route.fulfill({response});delivered();
  });
  await expect.poll(()=>held,{timeout:15000}).toBe(true); // Native periodic refresh.
  await row(page,id).getByRole('button',{name:'Cancel queued message'}).click();
  await expect.poll(async()=> (await(await request.get(`/api/sessions/${main.id}/queue`)).json()).items.length).toBe(0);
  await expect(row(page,id)).toHaveCount(0);
  deliver();await done;
  await expect(row(page,id)).toHaveCount(0);
 } finally {deliver();release();}
});

test('Gi queue responses and cancellation remain owned by their origin selection',async({page,request},info)=>{
 const {main,child,input,queue,enqueue,release,switchTo}=await setup(page,request,info);
 let deliver;const gate=new Promise(resolve=>{deliver=resolve;});let held=false;
 try {
  const first=await enqueue('origin queued');
  await page.route(`**/api/sessions/${main.id}/queue/${first}`,async route=>{
   const response=await route.fetch();held=true;await gate;await route.fulfill({response});
  });
  await row(page,first).getByRole('button',{name:'Cancel queued message'}).click();
  await expect.poll(()=>held).toBe(true);await switchTo(child);await input.fill('child draft');deliver();
  await expect.poll(async()=> (await queue()).length).toBe(0);
  await expect(page.locator('[data-queue-id]')).toHaveCount(0);await expect(input).toHaveValue('child draft');
  await expect(page.getByRole('alert')).toHaveCount(0);
 } finally {deliver();release();}
});

test('@shared-29 Reorder only the captured queue and reconcile native removal of a consumed ID', async ({page,request},info)=>{
  const source=loadCorpus('shared').find(row=>row.id==='@shared-29');expect(source).toBeTruthy();await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
  const {main,child,input,queue,enqueue,release,switchTo}=await setup(page,request,info);
  const token=`shared29-${info.project.name}-${Date.now()}`;
  const childGate=resolve('test-results/ux-parity/queue-gates',token+'-child'),targetGate=resolve('test-results/ux-parity/queue-gates',token+'-target');mkdirSync(resolve(childGate,'..'),{recursive:true});
  let releaseDelete=()=>{},childTurn;
  const childQueue=async()=>(await(await request.get(`/api/sessions/${child}/queue`)).json()).items;
  try {
    const busy=await request.post(`/api/sessions/${child}/prompt`,{data:{prompt:`UX queue gate:${token}-child`,model:'test-model'}});expect(busy.status()).toBe(202);childTurn=(await busy.json()).turn_id;
    await expect.poll(async()=>(await(await request.get(`/api/sessions/${child}`)).json()).state.status).toBe('running');
    for(const text of ['research first','research second'])expect((await request.post(`/api/sessions/${child}/prompt`,{data:{prompt:text,intent:'queue',model:'test-model'}})).status()).toBe(202);
    const childBefore=await childQueue();expect(childBefore).toHaveLength(2);
    await switchTo(child);await input.fill('research draft stays');await switchTo(main.id);
    const first=await enqueue('main first'),target=await enqueue(`UX queue gate:${token}-target`),last=await enqueue('main last');
    const before=await queue();expect(before.map(t=>t.id)).toEqual([first,target,last]);
    await input.fill('main newer draft stays');
    const moved=page.waitForResponse(r=>r.request().method()==='PATCH'&&new URL(r.url()).pathname===`/api/sessions/${main.id}/queue`);
    await row(page,target).getByRole('button',{name:'Move up in queue',exact:true}).click();const move=await moved;expect(move.status()).toBe(200);
    expect(move.request().postDataJSON()).toEqual({expected:[first,target,last],order:[target,first,last]});
    await expect.poll(()=>ids(page)).toEqual([target,first,last]);
    expect(await queue()).toEqual([before[1],before[0],before[2]]);expect(await childQueue()).toEqual(childBefore);
    await page.reload();await expect.poll(()=>ids(page)).toEqual([target,first,last]);await expect(input).toHaveValue('main newer draft stays');
    // Hold the real user DELETE before admission. Let the native runner consume
    // precisely that ID, then submit the held request against an active turn.
    let held=false;const deletionGate=new Promise(r=>{releaseDelete=r;});
    await page.route(`**/api/sessions/${main.id}/queue/${target}`,async route=>{held=true;await deletionGate;await route.continue();});
    const rejected=page.waitForResponse(r=>r.request().method()==='DELETE'&&new URL(r.url()).pathname===`/api/sessions/${main.id}/queue/${target}`);
    await row(page,target).getByRole('button',{name:'Cancel queued message',exact:true}).click();await expect.poll(()=>held).toBe(true);
    release();
    await expect.poll(async()=>((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).find(t=>t.id===target)?.status,{timeout:15000}).toBe('running');
    releaseDelete();expect((await rejected).status()).toBe(409);
    await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();
    await expect.poll(()=>ids(page)).toEqual([first,last]);expect((await queue()).map(t=>t.id)).toEqual([first,last]);
    // Rejection cannot cancel the now-active ID, and other persisted work is
    // byte-for-byte unchanged while its independent provider stays gated.
    expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns.find(t=>t.id===target).status).toBe('running');
    expect(await childQueue()).toEqual(childBefore);await expect(input).toHaveValue('main newer draft stays');
    await switchTo(child);await expect(input).toHaveValue('research draft stays');await expect.poll(()=>ids(page)).toEqual(childBefore.map(t=>t.id));
    await switchTo(main.id);await expect(input).toHaveValue('main newer draft stays');await expect.poll(()=>ids(page)).toEqual([first,last]);
    writeFileSync(targetGate,'release');
    await expect.poll(async()=>((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).filter(t=>[target,first,last].includes(t.id)).every(t=>t.status==='completed'),{timeout:15000}).toBe(true);
    const messages=(await(await request.get(`/api/sessions/${main.id}/messages`)).json()).messages;
    expect(messages.filter(m=>m.role==='user').map(m=>m.content).slice(1)).toEqual([`UX queue gate:${token}-target`,'main first','main last']);
    expect(await childQueue()).toEqual(childBefore);
  } finally {
    releaseDelete();release();writeFileSync(targetGate,'release');writeFileSync(childGate,'release');
    if(childTurn)await expect.poll(async()=>((await(await request.get(`/api/sessions/${child}/turns`)).json()).turns||[]).every(t=>t.status==='completed'),{timeout:15000}).toBe(true);
  }
});

test('@shared-27 Queue two native composer follow-ups exactly once with media and references',async({page,request},info)=>{
 const source=loadCorpus('shared').find(r=>r.id==='@shared-27');expect(source.name).toBe('Queue two follow-ups exactly once');await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const {main,child,input,active,queue,release,switchTo}=await setup(page,request,info);
 const get=async(id,part)=>{const r=await request.get(`/api/sessions/${id}/${part}`);expect(r.status()).toBe(200);return r.json();};
 const token=`shared27-${info.project.name}-${Date.now()}`,folder=`${token}-folder`,path=`${folder}/reference.txt`;
 let deliver=()=>{};
 try {
  // Give research real history and a persisted local draft; an empty session
  // alone would be a weak isolation control.
  const other=await request.post(`/api/sessions/${child}/prompt`,{data:{prompt:`research history ${token}`,model:'test-model'}});expect(other.status()).toBe(202);const otherTurn=(await other.json()).turn_id;
  await expect.poll(async()=>(await get(child,'turns')).turns.find(t=>t.id===otherTurn)?.status).toBe('completed');
  const otherBefore={messages:await get(child,'messages'),turns:await get(child,'turns'),model:await get(child,'model'),queue:await get(child,'queue'),media:await get(child,'media')};
  await switchTo(child);await input.fill('research unsent draft');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'research-pending.txt',mimeType:'text/plain',buffer:Buffer.from('research unsent bytes')});
  await switchTo(main.id);await expect(input).toHaveValue('');
  const reference=page.locator('.post .post-time').first();const referenceId=(await reference.getAttribute('href')).replace(/^#msg-/,'');expect(referenceId).toBeTruthy();
  const written=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'native workspace reference bytes'}}});expect(written.ok()).toBe(true);expect((await written.json()).error).toBeFalsy();
  const sent=[],uploads=[];
  page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${main.id}/prompt`)sent.push(r.postDataJSON());});
  let accepted,held=false;
  const gate=new Promise(r=>deliver=r);
  await page.route(`**/api/sessions/${main.id}/prompt`,async route=>{
   const response=await route.fetch();expect(response.status()).toBe(202);
   if(!held){accepted=await response.json();expect(accepted.queued).toBe(true);held=true;await gate;}
   await route.fulfill({response});
  });
  const submitted=[];
  for(const n of [1,2]){
   expect((await get(main.id,'turns')).turns.find(t=>t.id===active.turn_id)?.status).toBe('running');
   expect((await get(main.id,'queue')).active_turn_id).toBe(active.turn_id);
   await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeVisible();
   await page.locator(`#post-${referenceId} .post-time`).click();
   await page.getByRole('button',{name:'Menu',exact:true}).click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
   await page.locator(`.workspace-row[data-path="${folder}"] .workspace-caret`).click();
   await page.getByRole('button',{name:'Reference selected folder',exact:true}).click();
   await page.locator(`.workspace-row[data-path="${path}"]`).click();
   // Restore the closed folder before the next composition. Use its caret
   // on reopening because clicking the already-selected label does not toggle.
   await page.locator(`.workspace-row[data-path="${folder}"]`).click();
   await expect(page.locator(`.workspace-row[data-path="${path}"]`)).toHaveCount(0);
   await page.locator('.workspace-toggle-tab.open').click();
   const filename=`follow-up-${n}.txt`,bytes=Buffer.from(`follow-up ${n} native media bytes\n第二行`);
   await page.locator('.compose-box input[type=file]').setInputFiles({name:filename,mimeType:'text/plain',buffer:bytes});
   await expect(page.locator('.compose-input-main .compose-file-pill')).toHaveCount(4);
   await input.fill(`canonical follow-up ${n}\nsecond line`);
   const uploaded=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${main.id}/media`);
   const response=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${main.id}/prompt`);
   await input.press('Enter');const upload=await uploaded;expect(upload.status()).toBe(201);const media=await upload.json();uploads.push(media);
   const expected=`canonical follow-up ${n}\nsecond line\n\nFiles:\n- ${folder}\n- ${path}\n\nReferenced messages:\n- message:${referenceId}\n\nAttachments:\n- attachment:${media.media.id} (${filename})`;
   if(n===1){
    await expect.poll(()=>held).toBe(true);
    // SSE supplies the durable ID while the POST acknowledgement is still
    // held. No optimistic ID may survive as a duplicate row.
    await expect.poll(()=>ids(page)).toEqual([accepted.turn_id]);expect((await queue()).map(t=>t.id)).toEqual([accepted.turn_id]);deliver();
   }
   const ack=await response;expect(ack.status()).toBe(202);const body=await ack.json();expect(body.queued).toBe(true);
   const posted=sent[n-1];expect(posted.prompt).toBe(expected);expect(posted.intent).toBe('queue');expect(posted.client_request_id).toBeTruthy();expect(posted.media).toEqual([{media_id:media.media.id,session_id:main.id}]);
   submitted.push({id:body.turn_id,prompt:expected,ref:media.ref,bytes});
   await expect.poll(()=>ids(page)).toEqual(submitted.map(t=>t.id));await expect(input).toHaveValue('');await expect(page.locator('.compose-input-main .compose-file-pill')).toHaveCount(0);
  }
  await page.unroute(`**/api/sessions/${main.id}/prompt`);
  expect(sent).toHaveLength(2);expect(new Set(sent.map(s=>s.client_request_id)).size).toBe(2);expect(new Set(submitted.map(t=>t.id)).size).toBe(2);
  const stored=await queue();expect(stored.map(t=>t.id)).toEqual(submitted.map(t=>t.id));
  for(const [i,t] of stored.entries()){
   expect(t.prompt).toBe(submitted[i].prompt);expect(t.metadata.client_request_id).toBe(sent[i].client_request_id);expect(t.metadata.media).toHaveLength(1);
   const {created_at,...ref}=t.metadata.media[0];expect(ref).toEqual(submitted[i].ref);expect(Number.isFinite(Date.parse(created_at))).toBe(true);
   await expect(row(page,t.id).locator('.compose-queue-stack-text')).toContainText(`canonical follow-up ${i+1}`);
   const raw=await request.get(`/api/sessions/${main.id}/media/${submitted[i].ref.media_id}`);expect(raw.status()).toBe(200);expect(await raw.body()).toEqual(submitted[i].bytes);
  }
  expect((await get(main.id,'media')).media.map(m=>m.id).sort((a,b)=>a-b)).toEqual(uploads.map(m=>m.media.id).sort((a,b)=>a-b));
  expect((await get(main.id,'turns')).turns).toHaveLength(3);
  await input.fill('main newer unsent draft');await page.reload();await expect.poll(()=>ids(page)).toEqual(submitted.map(t=>t.id));await expect(input).toHaveValue('main newer unsent draft');expect(await queue()).toEqual(stored);
  await switchTo(child);await expect(input).toHaveValue('research unsent draft');await expect(page.locator('.compose-input-main .compose-file-pill[title="research-pending.txt"]')).toHaveCount(1);await expect(page.locator('[data-queue-id]')).toHaveCount(0);
  await page.reload();await expect(input).toHaveValue('research unsent draft');await expect(page.locator('.compose-input-main .compose-file-pill[title="research-pending.txt"]')).toHaveCount(1);
  await switchTo(main.id);await expect(input).toHaveValue('main newer unsent draft');await expect(page.locator('.compose-input-main .compose-file-pill')).toHaveCount(0);await expect.poll(()=>ids(page)).toEqual(submitted.map(t=>t.id));
  release();await expect.poll(async()=>(await get(main.id,'turns')).turns.filter(t=>submitted.some(s=>s.id===t.id)).map(t=>t.status),{timeout:15000}).toEqual(['completed','completed']);
  await expect(page.locator('[data-queue-id]')).toHaveCount(0);
  const finalTurns=(await get(main.id,'turns')).turns;expect(finalTurns).toHaveLength(3);
  const users=(await get(main.id,'messages')).messages.filter(m=>m.role==='user');expect(users).toHaveLength(3);expect(users.slice(1).map(m=>m.content)).toEqual(submitted.map(t=>t.prompt));
  for(const [i,t] of submitted.entries())expect(finalTurns.find(row=>row.id===t.id)).toMatchObject({prompt:t.prompt,metadata:{client_request_id:sent[i].client_request_id,media:[t.ref]}});
  expect(await get(child,'messages')).toEqual(otherBefore.messages);expect(await get(child,'turns')).toEqual(otherBefore.turns);expect(await get(child,'model')).toEqual(otherBefore.model);expect(await get(child,'queue')).toEqual(otherBefore.queue);expect(await get(child,'media')).toEqual(otherBefore.media);
  expect(sent).toHaveLength(2);await expect(input).toHaveValue('main newer unsent draft');
 }finally{deliver();release();}
});
