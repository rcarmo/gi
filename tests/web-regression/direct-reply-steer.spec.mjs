import {test,expect} from '@playwright/test';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';

const inputName='Message (Enter to send, Shift+Enter for newline)...';
const gates=resolve(process.env.GI_UX_QUEUE_GATES || `${process.env.GI_TEST_RUN_ROOT}/results/ux-parity/queue-gates`);

test('Native direct reply stays visible before queued steering continuation',async({page,request},info)=>{
 const token=`direct-${info.project.name}-${Date.now()}`;mkdirSync(gates,{recursive:true});
 const session=await request.post('/api/sessions',{data:{agent_id:token,title:token}});expect(session.status()).toBe(201);const {id}=await session.json();
 await page.addInitScript(sessionID=>localStorage.setItem('gi_session_id',sessionID),id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const first=await request.post(`/api/sessions/${id}/prompt`,{data:{prompt:`UX direct steer:${token}`,model:'ux-local/gate'}});expect(first.status()).toBe(202);const {turn_id}=await first.json();
 const activity=async()=>(await(await request.get(`/api/sessions/${id}/activity`)).json());
 const messages=async()=>(await(await request.get(`/api/sessions/${id}/messages`)).json()).messages||[];
 const gate=resolve(gates,token);
 try{
  await expect.poll(async()=> (await activity()).status).toBe('running');
  const queued=await request.post(`/api/sessions/${id}/prompt`,{data:{prompt:'steer after direct answer',intent:'queue',model:'ux-local/gate'}});expect(queued.status()).toBe(202);const {turn_id:queuedID}=await queued.json();
  const row=page.locator(`[data-queue-id="${queuedID}"]`);await expect(row).toBeVisible();
  const steered=await request.post(`/api/sessions/${id}/queue/${queuedID}/steer`,{data:{active_turn_id:turn_id}});expect(steered.status()).toBe(200);
  await input.fill('unsent draft stays here');writeFileSync(gate,'release');
  const firstText=`First direct answer ${token}`,secondText=`Second direct answer ${token}`;
  await expect(page.locator('.post-content').filter({hasText:firstText})).toHaveCount(1);
  await expect(page.locator('.post-content').filter({hasText:secondText})).toHaveCount(1);
  await expect.poll(async()=> (await activity()).status,{timeout:15000}).toBe('idle');
  const saved=(await messages()).filter(m=>m.content===firstText||m.content==='steer after direct answer'||m.content===secondText);
  expect(saved.map(m=>m.content)).toEqual([firstText,'steer after direct answer',secondText]);
  expect(saved[0].role).toBe('assistant');expect(saved[1].role).toBe('user');expect(saved[2].role).toBe('assistant');
  expect(saved[1].payload.turn_id).toBe(turn_id);expect(saved[1].payload.source_queue_id).toBe(queuedID);
  expect((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns.filter(t=>t.id===turn_id||t.id===queuedID)).toMatchObject([{id:turn_id,status:'completed'},{id:queuedID,status:'cancelled'}]);
  await expect(input).toHaveValue('unsent draft stays here');
  await page.reload();await expect(page.locator('.post-content').filter({hasText:firstText})).toHaveCount(1);await expect(page.locator('.post-content').filter({hasText:secondText})).toHaveCount(1);
  await expect(input).toHaveValue('unsent draft stays here');
 }finally{writeFileSync(gate,'release');}
});
