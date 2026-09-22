import { test, expect } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
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
