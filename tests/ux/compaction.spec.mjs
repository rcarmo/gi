import {test,expect} from '@playwright/test';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function fixture(page,request,info,mode='complete'){
 const token=`compact-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-child',title:token+'-child'}})).json()).branch.chat_jid.slice(3);
 const activity=async()=>(await(await request.get(`/api/sessions/${main.id}/activity`)).json());
 const turns=async()=>(await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns;
 // Build context using native turns and provider responses, never SQL seeds.
 for(let i=0;i<3;i++){
  const res=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:`history ${i} preserve requirements and pending changes`,model:'ux-local/gate'}})).json();
  await expect.poll(async()=> (await turns()).find(x=>x.id===res.turn_id)?.status).toBe('completed');
  await expect.poll(async()=> (await activity()).status).toBe('idle');
 }
 const path=resolve('test-results/ux-parity/queue-gates',token);mkdirSync(resolve(path,'..'),{recursive:true});const release=()=>writeFileSync(path,'go');
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const sent=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${main.id}/prompt`)&&r.request().method()==='POST');
 await input.fill(`UX compact ${mode}:${token} UX meter tokens:180`);await input.press('Enter');const turn=await(await sent).json();
 await expect.poll(async()=> (await activity()).compaction?.active).toBe(true);
 await expect(page.locator('.compose-context-pie')).toHaveClass(/is-compacting/);
 await input.fill('preserved draft during compaction');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'keep.txt',mimeType:'text/plain',buffer:Buffer.from('preserved')});
 return{main,child,turn,input,activity,turns,release};
}
async function source(info,id){const scenario=loadCorpus().find(x=>x.id===id);await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});}
for(const id of ['@ux-compaction-001','@ux-compaction-002','@ux-compaction-004','@ux-context-004']){
 test(`${id} Native automatic compaction status and accepted completion`,async({page,request},info)=>{
  await source(info,id);const{main,turn,input,activity,turns,release}=await fixture(page,request,info);
  try{
   const pie=page.locator('.compose-context-pie');await expect(pie).toHaveAttribute('aria-label',/Compacting context/);await expect(pie).toHaveAttribute('data-tooltip',/Compacting context — \d+:\d\d/);
   await expect(page.locator('.gi-compaction-elapsed')).toHaveText(/\d+:\d\d/);await expect(page.getByRole('button',{name:'Compacting context — Stop response',exact:true})).toBeVisible();
   await expect(pie).toBeDisabled(); // manual compaction capability still absent
   const first=await page.locator('.gi-compaction-elapsed').textContent();await expect.poll(()=>page.locator('.gi-compaction-elapsed').textContent(),{timeout:4000}).not.toBe(first);
   release();await expect.poll(async()=> (await turns()).find(t=>t.id===turn.turn_id).status).toBe('completed');
   await expect(pie).not.toHaveClass(/is-compacting/);await expect(page.locator('.gi-compaction-elapsed')).toHaveCount(0);
   await expect(pie).toHaveAttribute('aria-label','Context: 180 / 32K tokens (1%)',{timeout:15000});
   await expect(input).toHaveValue('preserved draft during compaction');await expect(page.locator('.compose-file-pill[title="keep.txt"]')).toBeVisible();
   const messages=(await(await request.get(`/api/sessions/${main.id}/messages`)).json()).messages;
   expect(messages.some(m=>m.payload?.kind==='compaction'&&m.payload?.turn_id===turn.turn_id)).toBe(true);
   expect((await activity()).status).toBe('idle');
  }finally{release();}
 });
}
test('@ux-compaction-003 Stop compaction without clearing draft or targeting another run',async({page,request},info)=>{
 await source(info,'@ux-compaction-003');const{main,turn,input,turns,release}=await fixture(page,request,info);
 try{
  const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${main.id}/activity`)&&r.request().method()==='POST');
  await page.getByRole('button',{name:'Compacting context — Stop response',exact:true}).click();const result=await response;expect(result.status()).toBe(200);expect(result.request().postDataJSON()).toEqual({turn_id:turn.turn_id});
  await expect.poll(async()=> (await turns()).find(t=>t.id===turn.turn_id).status).toBe('cancelled');
  await expect(page.locator('.compose-context-pie')).not.toHaveClass(/is-compacting/);await expect(input).toHaveValue('preserved draft during compaction');await expect(page.locator('.compose-file-pill[title="keep.txt"]')).toBeVisible();
  expect(await turns()).toHaveLength(4);
 }finally{release();}
});
test('@ux-compaction-005 Suppression reports native detail and retains draft',async({page,request},info)=>{
 await source(info,'@ux-compaction-005');const{input,release}=await fixture(page,request,info,'suppress');
 try{release();await expect(page.locator('.compose-inline-status').filter({hasText:'Compaction temporarily suppressed'})).toBeVisible();await expect(page.locator('.compose-inline-status-detail')).toContainText('before-compact hook');await expect(input).toHaveValue('preserved draft during compaction');}finally{release();}
});

test('Gi compaction reload, late poll and session switch preserve ownership',async({page,request},info)=>{
 const{main,child,turn,input,activity,turns,release}=await fixture(page,request,info);
 let unblock,held=false;const gate=new Promise(r=>unblock=r);
 try{
  const before=(await activity()).compaction;
  await page.reload();await expect(page.locator('.compose-context-pie')).toHaveClass(/is-compacting/);expect((await activity()).compaction.timestamp).toBe(before.timestamp);await expect(input).toHaveValue('preserved draft during compaction');
  await page.route(`**/api/sessions/${main.id}/activity`,async route=>{if(route.request().method()!=='GET')return route.continue();const response=await route.fetch();held=true;await gate;await route.fulfill({response});});
  // A native appended queue record triggers a refresh while compaction is held.
  const queued=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'after compaction',intent:'queue',model:'ux-local/gate'}})).json();
  await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${child}"]`).getByRole('menuitem').click();await input.fill('target draft');
  release();await expect.poll(async()=> (await turns()).find(t=>t.id===queued.turn_id).status,{timeout:15000}).toBe('completed');unblock();
  await expect(page.locator('.compose-context-pie')).not.toHaveClass(/is-compacting/);await expect(input).toHaveValue('target draft');await expect(page.getByRole('alert')).toHaveCount(0);
  await page.unroute(`**/api/sessions/${main.id}/activity`);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();await expect(input).toHaveValue('preserved draft during compaction');await expect(page.locator('.compose-context-pie')).not.toHaveClass(/is-compacting/);
  expect((await request.post(`/api/sessions/${main.id}/activity`,{data:{turn_id:turn.turn_id}})).status()).toBe(409);
 }finally{unblock();release();}
});

test('Gi Stop failure and delayed completion conflict retain composer data',async({page,request},info)=>{
 const{main,turn,input,turns,release}=await fixture(page,request,info);
 let unblock,held=false,calls=0;const gate=new Promise(r=>unblock=r);const stop=page.getByRole('button',{name:'Compacting context — Stop response',exact:true});
 await page.route(`**/api/sessions/${main.id}/activity`,async route=>{
  if(route.request().method()!=='POST')return route.continue();calls++;
  if(calls===1){await route.abort('failed');return;}
  held=true;await gate;await route.continue();
 });
 try{
  await stop.click();await expect(page.getByRole('alert').filter({hasText:'Stop failed:'})).toBeVisible();await expect(stop).toBeEnabled();await expect(input).toHaveValue('preserved draft during compaction');
  const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${main.id}/activity`)&&r.request().method()==='POST');
  await stop.click();await expect.poll(()=>held).toBe(true);await expect(stop).toBeDisabled();
  release();await expect.poll(async()=> (await turns()).find(t=>t.id===turn.turn_id).status).toBe('completed');unblock();expect((await response).status()).toBe(409);
  await expect(page.getByRole('button',{name:'Send message',exact:true})).toBeVisible();await expect(page.getByRole('alert')).toHaveCount(0);await expect(input).toHaveValue('preserved draft during compaction');await expect(page.locator('.compose-file-pill[title="keep.txt"]')).toBeVisible();expect(await turns()).toHaveLength(4);
 }finally{unblock();release();}
});

test('Gi durable compaction changes the next provider context without deleting timeline',async({page,request},info)=>{
 const{main,turn,input,turns,release}=await fixture(page,request,info);
 try{
  release();await expect.poll(async()=> (await turns()).find(t=>t.id===turn.turn_id).status).toBe('completed');
  const messages=async()=>(await(await request.get(`/api/sessions/${main.id}/messages`)).json()).messages;
  const before=await messages();const summary=before.find(m=>m.payload?.kind==='compaction'&&m.payload?.turn_id===turn.turn_id);expect(summary.payload.durable_context).toBe(true);
  await page.reload();await expect(input).toHaveValue('preserved draft during compaction');
  const next=await(await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'new request after durable boundary',model:'ux-local/gate'}})).json();
  await expect.poll(async()=> (await turns()).find(t=>t.id===next.turn_id).status).toBe('completed');
  const after=await messages();
  for(const message of before)expect(after.find(m=>m.id===message.id)).toEqual(message);
  // Fixture echoes actual provider user messages, not a fabricated result.
  const answer=after.find(m=>m.role==='assistant'&&m.payload?.turn_id===next.turn_id&&m.payload?.kind==='chat');
  expect(answer.content).toContain('Preserve user requirements and pending work.');expect(answer.content).toContain('new request after durable boundary');
  expect(answer.content).not.toContain('history 0 preserve requirements');
  expect(after.filter(m=>m.payload?.kind==='compaction')).toHaveLength(1);
  await page.reload();await expect(page.getByText('history 0 preserve requirements and pending changes',{exact:true})).toBeVisible();await expect(input).toHaveValue('preserved draft during compaction');
 }finally{release();}
});
