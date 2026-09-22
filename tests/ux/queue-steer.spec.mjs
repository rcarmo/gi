import {test,expect} from '@playwright/test';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {loadCorpus} from './support/catalogue.mjs';
import {createServer, request as httpRequest} from 'node:http';
async function sseProxy(page){
 let blocked=false;const connections=new Set();
 const proxy=createServer((req,res)=>{
  res.setHeader('Access-Control-Allow-Origin','*');if(blocked){res.writeHead(503);res.end();return;}
  connections.add(res);res.on('close',()=>connections.delete(res));
  const upstream=httpRequest(new URL(req.url,process.env.GI_TEST_URL),source=>{res.writeHead(source.statusCode,{'Content-Type':'text/event-stream','Cache-Control':'no-cache','Access-Control-Allow-Origin':'*'});source.pipe(res);});
  upstream.on('error',()=>res.destroy());res.on('close',()=>upstream.destroy());upstream.end();
 });
 await new Promise(resolve=>proxy.listen(0,'127.0.0.1',resolve));
 await page.addInitScript(origin=>{const Native=window.EventSource;window.EventSource=class extends Native{constructor(url,options){const path=new URL(url,location.href);super(path.pathname==='/sse/stream'?origin+path.pathname+path.search:url,options);}};},`http://127.0.0.1:${proxy.address().port}`);
 return{drop(){blocked=true;for(const res of connections)res.destroy();},resume(){blocked=false;},async close(){blocked=false;for(const res of connections)res.destroy();await new Promise(resolve=>proxy.close(resolve));}};
}
const inputName='Message (Enter to send, Shift+Enter for newline)...';
const steerName='Inject queued follow-up as steer';
async function fixture(page,request,info){
 const token=`steer-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-child',title:token+'-child'}})).json()).branch.chat_jid.slice(3);
 const path=resolve('test-results/ux-parity/queue-gates',token);mkdirSync(resolve(path,'..'),{recursive:true});const release=()=>writeFileSync(path,'ok');
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 await input.fill(`UX steer gate:${token}`);await input.press('Enter');
 await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeVisible({timeout:15000});
 const turns=async()=> (await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns;
 let active;await expect.poll(async()=>{active=(await turns()).find(t=>t.status==='running');return active?.id;}).toBeTruthy();
 const queuedResponse=await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:`steered content ${token}`,intent:'queue',model:'ux-local/gate'}});expect(queuedResponse.status()).toBe(202);
 const queued=await queuedResponse.json();const row=page.locator(`[data-queue-id="${queued.turn_id}"]`);await expect(row).toBeVisible();
 const button=row.getByRole('button',{name:steerName});await expect(button).toBeEnabled();
 return{main,child,active,queued,row,button,input,turns,release,token};
}

test('@shared-30 Steer only a matching active run',async({page,request},info)=>{
 const scenario=loadCorpus('shared').find(x=>x.id==='@shared-30');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const proxy=await sseProxy(page);
 const f=await fixture(page,request,info);const{main,child,active,queued,row,button,input,turns,release,token}=f;
 const url=`/api/sessions/${main.id}/queue/${queued.turn_id}/steer`;let calls=0,unblock,held=false;const gate=new Promise(r=>unblock=r);
 try{
  await input.fill('unsent retained');
  proxy.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});await expect(button).toBeDisabled();
  let unknownCalls=0;const count=r=>{if(r.url().endsWith(url))unknownCalls++;};page.on('request',count);
  const disabledBox=await button.boundingBox();await page.mouse.click(disabledBox.x+disabledBox.width/2,disabledBox.y+disabledBox.height/2);expect(unknownCalls).toBe(0);page.off('request',count);
  proxy.resume();await expect(button).toBeEnabled({timeout:15000});
  // Native stale/foreign ownership rejection leaves the exact queue ID untouched.
  for(const [session,run] of [[main.id,'obsolete'],[child,active.id]]){
   const res=await request.post(`/api/sessions/${session}/queue/${queued.turn_id}/steer`,{data:{active_turn_id:run}});expect(res.status()).toBe(409);
  }
  await page.route(`**${url}`,async route=>{calls++;if(calls===1){await route.abort('failed');return;}held=true;await gate;await route.continue();});
  await button.click();await expect(page.getByRole('alert').filter({hasText:'Queue action failed'})).toBeVisible();await expect(button).toBeEnabled();await expect(row).toBeVisible();
  expect((await turns()).find(t=>t.id===queued.turn_id).status).toBe('queued');
  await button.focus();await page.keyboard.press('Enter');await expect.poll(()=>held).toBe(true);await expect(button).toBeDisabled();
  // Repeat physical activation while request is pending: no second browser request.
  await page.keyboard.press('Enter');await expect.poll(()=>calls).toBe(2);
  unblock();await expect(row).toHaveCount(0);await expect(input).toHaveValue('unsent retained');
  expect((await request.post(url,{data:{active_turn_id:active.id}})).status()).toBe(409);
  release();await expect.poll(async()=> (await turns()).find(t=>t.id===active.id).status,{timeout:15000}).toBe('completed');
  const records=await turns();expect(records.find(t=>t.id===queued.turn_id)).toMatchObject({status:'cancelled',phase:'steered'});expect(records).toHaveLength(2);
  const messages=(await(await request.get(`/api/sessions/${main.id}/messages`)).json()).messages;
  const delivered=messages.filter(m=>m.role==='user'&&m.content===`steered content ${token}`);expect(delivered).toHaveLength(1);expect(delivered[0].payload).toMatchObject({turn_id:active.id,active_turn_id:active.id,source_queue_id:queued.turn_id});
  expect(messages.filter(m=>m.role==='assistant').at(-1).content).toContain(`steered content ${token}`);
  expect((await(await request.get(`/api/sessions/${child}/messages`)).json()).messages || []).toHaveLength(0);
  // A real run that cannot consume steering returns a visible, held queue row.
  // This gives native idle and initial-unknown activity, not fabricated SSE.
  const shellToken=`idle-${token}`;const shellPath=resolve('test-results/ux-parity/queue-gates',shellToken);
  const shell=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:`UX queue gate:${shellToken}`,model:'test-model'}})).json();
  const idleQueued=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'unconsumed idle row',intent:'queue',model:'test-model'}})).json();
  expect((await request.post(`/api/sessions/${main.id}/queue/${idleQueued.turn_id}/steer`,{data:{active_turn_id:shell.turn_id}})).status()).toBe(200);
  writeFileSync(shellPath,'ok');
  const idleRow=page.locator(`[data-queue-id="${idleQueued.turn_id}"]`);await expect(idleRow).toBeVisible({timeout:15000});await expect(idleRow.getByRole('button',{name:steerName})).toBeDisabled();
  const idleURL=`/api/sessions/${main.id}/queue/${idleQueued.turn_id}/steer`;let idleCalls=0;page.on('request',r=>{if(r.url().endsWith(idleURL))idleCalls++;});
  const box=await idleRow.getByRole('button',{name:steerName}).boundingBox();await page.mouse.click(box.x+box.width/2,box.y+box.height/2);expect(idleCalls).toBe(0);
  expect((await request.post(idleURL,{data:{active_turn_id:shell.turn_id}})).status()).toBe(409);
  await page.reload();await expect(idleRow.getByRole('button',{name:steerName})).toBeDisabled();expect(idleCalls).toBe(0);
  await expect(page.getByRole('alert').filter({hasText:'will not auto-send'})).toBeVisible();
  await page.getByRole('button',{name:'Open model picker',exact:true}).click();
  await page.getByRole('menu',{name:'Model picker',exact:true}).getByRole('menuitem').filter({hasText:'ux-local/gate'}).click();
  const newToken=`resume-${token}`;await input.fill(`UX steer gate:${newToken}`);await input.press('Enter');
  const retry=idleRow.getByRole('button',{name:steerName});await expect(retry).toBeEnabled();
  const freshRun=(await turns()).find(t=>t.status==='running');expect(freshRun.id).not.toBe(shell.turn_id);
  await retry.click();await expect(idleRow).toHaveCount(0);writeFileSync(resolve('test-results/ux-parity/queue-gates',newToken),'ok');
  await expect.poll(async()=> (await turns()).find(t=>t.id===freshRun.id).status,{timeout:15000}).toBe('completed');
  const resumed=(await(await request.get(`/api/sessions/${main.id}/messages`)).json()).messages.filter(m=>m.role==='user'&&m.content==='unconsumed idle row');
  expect(resumed).toHaveLength(1);expect(resumed[0].payload.turn_id).toBe(freshRun.id);
 }finally{unblock();release();await proxy.close();}
});

test('Gi stale Steer reply after session switch cannot change target draft',async({page,request},info)=>{
 const{main,child,active,queued,row,button,input,turns,release}=await fixture(page,request,info);
 let unblock,held=false;const gate=new Promise(r=>unblock=r);
 await page.route(`**/api/sessions/${main.id}/queue/${queued.turn_id}/steer`,async route=>{held=true;await gate;await route.continue();});
 try{
  await input.fill('origin draft');await button.click();await expect.poll(()=>held).toBe(true);
  // End the original run; queued source is allowed to run normally before the
  // delayed request reaches the server. It must not steer any newer run.
  release();await expect.poll(async()=> (await turns()).find(t=>t.id===queued.turn_id).status,{timeout:15000}).toBe('completed');
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${child}"]`).getByRole('menuitem').click();await input.fill('target draft');
  const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${main.id}/queue/${queued.turn_id}/steer`));
  unblock();expect((await response).status()).toBe(409);await expect(input).toHaveValue('target draft');await expect(page.getByRole('alert')).toHaveCount(0);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();await expect(input).toHaveValue('origin draft');await expect(row).toHaveCount(0);
  expect((await turns()).find(t=>t.id===active.id).status).toBe('completed');
 }finally{unblock();release();}
});
