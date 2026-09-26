import {test,expect} from '@playwright/test';
import http from 'node:http';import{networkInterfaces}from'node:os';
import{journeyEnvironment}from'./support/journey-environment.mjs';

async function setup(page,info,startTurn=true){
 const fixture=await journeyEnvironment(info),address=Object.values(networkInterfaces()).flat().find(x=>x?.family==='IPv4'&&!x.internal)?.address;if(!address)throw Error('Nonloopback interface required');
 let blockSSE=false;const streams=new Set();
 const server=http.createServer((req,res)=>{
  const sse=req.url.startsWith('/sse/stream');if(sse&&blockSSE){res.writeHead(503);res.end();return}
  if(sse){streams.add(res);res.on('close',()=>streams.delete(res))}
  const up=http.request(new URL(req.url,fixture.origin),{method:req.method,headers:req.headers},r=>{res.writeHead(r.statusCode,r.headers);r.pipe(res)});up.on('error',()=>res.destroy());req.pipe(up);res.on('close',()=>up.destroy());
 });await new Promise(r=>server.listen(0,'0.0.0.0',r));const origin=`http://${address}:${server.address().port}`;
 const read=async path=>{const r=await page.request.get(origin+path);expect(r.status()).toBe(200);return r.json()};
 await page.goto(origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();expect(await page.evaluate(()=>isSecureContext)).toBe(false);const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
 const token=`basic-${info.project.name}-${Date.now()}`;let turn;
 if(startTurn){const accepted=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${id}/prompt`));await input.fill(`UX steer gate:${token}`);await input.press('Enter');const response=await accepted;expect(response.status()).toBe(202);turn=(await response.json()).turn_id;
 await expect.poll(async()=>(await read(`/api/sessions/${id}/turns`)).turns.find(x=>x.id===turn)?.status).toBe('running');await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeEnabled();}
 return{origin,input,id,turn,read,release:()=>fixture.release(token),drop(){blockSSE=true;for(const r of streams)r.destroy()},resume(){blockSSE=false},async close(){fixture.release(token);server.closeAllConnections();await new Promise(r=>server.close(r));await fixture.close()}};
}

test('HTTP Stop cancels the addressed running turn without submitting or clearing the next draft',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('next unsent draft 中文');const posts=[];page.on('request',r=>{if(r.method()==='POST')posts.push({path:new URL(r.url()).pathname,body:r.postDataJSON()})});
  const stopped=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/activity`));await page.getByRole('button',{name:'Stop response',exact:true}).click();expect((await stopped).status()).toBe(200);
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===h.turn)?.status).toBe('cancelled');
  await expect(h.input).toHaveValue('next unsent draft 中文');expect(posts).toEqual([{path:`/api/sessions/${h.id}/activity`,body:{turn_id:h.turn}}]);expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(1);
  await page.reload();await expect(h.input).toHaveValue('next unsent draft 中文');await expect(page.getByRole('button',{name:'Send message',exact:true})).toBeEnabled();
  const next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/prompt`));await h.input.press('Enter');const response=await next;expect(response.status()).toBe(202);const nextTurn=(await response.json()).turn_id;
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===nextTurn)?.status).toBe('completed');expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(2);await expect(h.input).toHaveValue('');
 }finally{await h.close()}
});

test('HTTP SSE disconnect and reconnect preserves draft and recovers native completion without resending',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('offline draft 中文');h.drop();await expect(page.getByText('Reconnecting',{exact:true})).toBeVisible();await expect(h.input).toHaveValue('offline draft 中文');
  h.release();await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===h.turn)?.status).toBe('completed');h.resume();await expect(page.getByText('Reconnecting',{exact:true})).toHaveCount(0);
  await expect(h.input).toHaveValue('offline draft 中文');await expect(page.locator('.post.agent-post')).toHaveCount(1);expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(1);
  await page.reload();await expect(h.input).toHaveValue('offline draft 中文');await expect(page.locator('.post.agent-post')).toHaveCount(1);
 }finally{await h.close()}
});

test('HTTP model selection governs actual provider requests after reload; rejection keeps last choice and draft',async({page},info)=>{
 const h=await setup(page,info,false);const errors=[];page.on('pageerror',e=>errors.push(e.message));
 try{
  const trigger=page.getByRole('button',{name:'Open model picker',exact:true});const option=id=>page.getByRole('listbox',{name:'Models',exact:true}).getByRole('option').filter({hasText:id});
  const draft=`model choice ${info.project.name} 中文🙂`;await h.input.fill(draft);
  const selected=page.waitForResponse(r=>r.request().method()==='PATCH'&&r.url().endsWith(`/api/sessions/${h.id}/model`));
  await trigger.click();await option('ux-local/alternate').click();expect((await selected).status()).toBe(200);await expect(trigger).toHaveText('ux-local/alternate');await expect(h.input).toHaveValue(draft);
  expect((await h.read(`/api/sessions/${h.id}/turns`)).turns||[]).toHaveLength(0);await page.reload();await expect(trigger).toHaveText('ux-local/alternate');await expect(h.input).toHaveValue(draft);
  async function sendAndAssert(text,model){
   await h.input.fill(text);const response=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/prompt`));await h.input.press('Enter');const res=await response;expect(res.status()).toBe(202);const id=(await res.json()).turn_id;
   await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===id)?.status).toBe('completed');
   const turn=(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===id);expect(turn.prompt).toBe(text);expect(turn.metadata.model).toBe(`ux-local/${model}`);
   const expected=`Provider model ${model}:`;await expect(page.locator('.post.agent-post .post-content').filter({hasText:expected}).filter({hasText:text})).toBeVisible();await expect(h.input).toHaveValue('');return id;
  }
  await sendAndAssert(draft,'alternate');
  const next='keep draft after model failure';await h.input.fill(next);
  await page.route(`**/api/sessions/${h.id}/model`,r=>r.request().method()==='PATCH'?r.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error:'Model selection unavailable'})}):r.continue());
  await trigger.click();await option('ux-local/gate').click();await expect(page.getByRole('alert')).toContainText('Model selection unavailable');await expect(trigger).toHaveText('ux-local/alternate');await expect(h.input).toHaveValue(next);expect((await h.read(`/api/sessions/${h.id}/model`)).current).toBe('ux-local/alternate');
  await page.unroute(`**/api/sessions/${h.id}/model`);await page.reload();await expect(trigger).toHaveText('ux-local/alternate');await expect(h.input).toHaveValue(next);await sendAndAssert(next,'alternate');
  await trigger.click();await option('ux-local/gate').click();await expect(trigger).toHaveText('ux-local/gate');await sendAndAssert('send after changing model again','gate');
  const turns=(await h.read(`/api/sessions/${h.id}/turns`)).turns;expect(turns).toHaveLength(3);expect(new Set(turns.map(t=>t.id)).size).toBe(3);expect(errors).toEqual([]);
 }finally{await h.close()}
});

test('HTTP lost steering acknowledgement confirms the source receipt without restoring or resending follow-up',async({page},info)=>{
 const h=await setup(page,info);let release;const gate=new Promise(r=>release=r);let posts=0,admitted,payload;
 try{
  const text=`steered once ${info.project.name}`,newer='next unsent after steering';
  await page.route(`**/api/sessions/${h.id}/prompt`,async r=>{posts++;payload=r.request().postDataJSON();const response=await r.fetch();expect(response.status()).toBe(202);admitted=await response.json();await gate;await r.abort('failed')});
  // Ctrl+Enter is the supplied explicit steer shortcut, not queued Return.
  await h.input.fill(text);await h.input.press('Control+Enter');await expect.poll(()=>admitted?.turn_id).toBeTruthy();expect(admitted.turn_id).toBe(h.turn);expect(payload.intent).toBe('steer');await h.input.fill(newer);release();
  await expect(page.getByRole('status').filter({hasText:'Sending message'})).toHaveCount(0);await expect(h.input).toHaveValue(newer);expect(posts).toBe(1);
  const receipt=await h.read(`/api/sessions/${h.id}/send-receipt?client_request_id=${encodeURIComponent(payload.client_request_id)}`);expect(receipt.confirmed).toBe(true);expect(receipt.result.turn_id).toBe(h.turn);expect(receipt.result.session_id).toBe(h.id);
  h.release();await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===h.turn)?.status).toBe('completed');
  expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(1);await page.reload();await expect(h.input).toHaveValue(newer);expect(posts).toBe(1);
 }finally{release();await h.close()}
});

test('HTTP captured Stop preserves FIFO until explicit fenced Resume queue, including reload',async({page},info)=>{
 const h=await setup(page,info);try{
  const queued=[];
  for(const prompt of ['queued first','queued second']){const response=await page.request.post(h.origin+`/api/sessions/${h.id}/prompt`,{data:{prompt,model:'test-model',intent:'queue'}});expect(response.status()).toBe(202);queued.push(await response.json());}
  const before=(await h.read(`/api/sessions/${h.id}/queue`)).items;
  await h.input.fill('unsent after stop Ω');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'resume-keep.txt',mimeType:'text/plain',buffer:Buffer.from('resume preserves attachment')});
  const stop=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/activity`));await page.getByRole('button',{name:'Stop response',exact:true}).click();expect((await stop).status()).toBe(200);
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/activity`)).status).toBe('idle');
  expect((await h.read(`/api/sessions/${h.id}/activity`)).queue_hold_turn_id).toBe(h.turn);
  expect((await h.read(`/api/sessions/${h.id}/queue`)).items).toEqual(before);
  await expect(page.getByRole('button',{name:'Resume queue',exact:true})).toBeEnabled();await expect(h.input).toHaveValue('unsent after stop Ω');await expect(page.locator('.compose-file-pill[title="resume-keep.txt"]')).toBeVisible();
  await page.reload();await expect(page.getByRole('button',{name:'Resume queue',exact:true})).toBeEnabled();await expect(h.input).toHaveValue('unsent after stop Ω');
  expect((await page.request.post(h.origin+`/api/sessions/${h.id}/resume-queue`,{data:{stop_turn_id:'stale'}})).status()).toBe(409);
  expect((await page.request.post(h.origin+`/api/sessions/${h.id}/continue`,{data:{}})).status()).toBe(400);
  expect((await h.read(`/api/sessions/${h.id}/queue`)).items).toEqual(before);
  const resumed=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/resume-queue`));await page.getByRole('button',{name:'Resume queue',exact:true}).click();const response=await resumed;expect(response.status()).toBe(200);expect(response.request().postDataJSON()).toEqual({stop_turn_id:h.turn});
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.filter(t=>queued.some(q=>q.turn_id===t.id)&&t.status==='completed').length).toBe(2);
  await expect(page.getByRole('button',{name:'Resume queue',exact:true})).toHaveCount(0);
  await expect(h.input).toHaveValue('unsent after stop Ω');await expect(page.locator('.compose-file-pill[title="resume-keep.txt"]')).toBeVisible();
  expect((await page.request.post(h.origin+`/api/sessions/${h.id}/resume-queue`,{data:{stop_turn_id:h.turn}})).status()).toBe(409);
  expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(3);
 }finally{await h.close()}
});

test('HTTP lost Resume acknowledgement reconciles without replay or clearing next draft',async({page},info)=>{
 const h=await setup(page,info);try{
  const response=await page.request.post(h.origin+`/api/sessions/${h.id}/prompt`,{data:{prompt:'resume once',intent:'queue',model:'test-model'}});expect(response.status()).toBe(202);const next=await response.json();
  await h.input.fill('keep after lost resume');await page.getByRole('button',{name:'Stop response',exact:true}).click();await expect(page.getByRole('button',{name:'Resume queue',exact:true})).toBeEnabled();
  let posts=0;await page.route(`**/api/sessions/${h.id}/resume-queue`,async r=>{posts++;const accepted=await r.fetch();expect(accepted.status()).toBe(200);await r.abort('failed')});
  await page.getByRole('button',{name:'Resume queue',exact:true}).click();await expect(page.getByText(/Resume not confirmed/)).toBeVisible();
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===next.turn_id)?.status).toBe('completed');await expect(page.getByRole('button',{name:'Resume queue',exact:true})).toHaveCount(0);await expect(h.input).toHaveValue('keep after lost resume');expect(posts).toBe(1);
  await page.reload();await expect(h.input).toHaveValue('keep after lost resume');expect(posts).toBe(1);expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(2);
 }finally{await h.close()}
});
