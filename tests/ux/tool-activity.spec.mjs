import {test,expect} from '@playwright/test';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {loadCorpus} from './support/catalogue.mjs';
const gates=resolve('test-results/ux-parity/queue-gates');
const inputName='Message (Enter to send, Shift+Enter for newline)...';
test('@ux-original-027 native tool activity has owned preview, elapsed and terminal timing',async({page,request},info)=>{
 const scenario=loadCorpus().find(s=>s.id==='@ux-original-027');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 test.setTimeout(45000);
 const token=`tools-${info.project.name}-${Date.now()}`;mkdirSync(gates,{recursive:true});
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();const other=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-other',title:token+'-other'}})).json()).branch.chat_jid.slice(3);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await input.fill('tool draft β');await page.locator('.compose-box input[type=file]').setInputFiles({name:'tool-ref.txt',mimeType:'text/plain',buffer:Buffer.from('retained tool bytes')});
 const submit=async prompt=>{const r=await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt,model:'test-model'}});expect(r.status()).toBe(202);return(await r.json()).turn_id;};
 const activity=async (id=main.id)=>(await(await request.get(`/api/sessions/${id}/activity`)).json());
 const region=page.locator('.gi-tool-activity');let id,release; 
 try{
  id=await submit(`UX queue gate:${token}`);await expect(region).toHaveAttribute('data-tool-state','running');await expect(region).toHaveAttribute('data-turn-id',id);const started=await activity();expect(started.tool.tool_call_id).toBeTruthy();expect(started.tool.preview).toContain("printf 'Gi received: %s'");await expect(region.locator('code')).toHaveText(started.tool.preview);
  const elapsed=region.getByLabel('Tool elapsed');const initial=await elapsed.textContent();await expect(elapsed).not.toHaveText(initial,{timeout:2500});
  await page.reload();await expect(input).toHaveValue('tool draft β');await expect(region).toHaveAttribute('data-tool-call-id',started.tool.tool_call_id);
  writeFileSync(resolve(gates,token),'release');await expect(region).toHaveAttribute('data-tool-state','completed');const done=await activity();expect(done.tool.duration_ms).toBeGreaterThanOrEqual(1000);await expect(region.getByLabel('Tool duration')).toHaveText(`${Math.floor(done.tool.duration_ms/1000)}s`);await expect(region.locator('.spinner')).toHaveCount(0);
  const duration=await region.getByLabel('Tool duration').textContent();await page.waitForTimeout(1100);await expect(region.getByLabel('Tool duration')).toHaveText(duration);
  // A held old activity read cannot replace a new occurrence after its invalidation.
  let held=false;const gate=new Promise(r=>release=r);
  await page.route(`**/api/sessions/${main.id}/activity`,async route=>{const response=await route.fetch();held=true;await gate;await route.fulfill({response});},{times:1});
  await page.evaluate(()=>window.dispatchEvent(new Event('focus')));
  // The regular native refresh interval provides the request if focus is throttled.
  await expect.poll(()=>held,{timeout:12000}).toBe(true);
  id=await submit('UX tool fail: native error');release();await expect(region).toHaveAttribute('data-tool-state','failed');await expect(region).toHaveAttribute('data-turn-id',id);await expect(region.getByLabel('shell: Failed')).toBeVisible();await expect(region.locator('.spinner')).toHaveCount(0);
  const failed=await activity();expect(failed.tool.duration_ms).toBeGreaterThanOrEqual(0);expect(failed.tool.tool_call_id).not.toBe(started.tool.tool_call_id);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${other}"]`).getByRole('menuitem').click();await expect(region).toHaveCount(0);await input.fill('other tool draft');expect((await activity(other)).tool).toBeUndefined();
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();await expect(region).toHaveAttribute('data-tool-state','failed');await expect(input).toHaveValue('tool draft β');await expect(page.locator('.compose-box')).toContainText('tool-ref.txt');
  await info.attach('tool-status',{body:await page.screenshot(),contentType:'image/png'});
 }finally{release?.();writeFileSync(resolve(gates,token),'release');await page.unrouteAll({behavior:'wait'});}
});

test('Gi tool cancellation has occurrence-bound terminal timing and reload reconstruction',async({page,request},info)=>{
 const token=`tool-cancel-${info.project.name}-${Date.now()}`;mkdirSync(gates,{recursive:true});
 const r=await request.post('/api/sessions',{data:{agent_id:token,title:token}});expect(r.ok()).toBe(true);const main=await r.json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('cancel tool draft β');
 const activity=async()=>(await(await request.get(`/api/sessions/${main.id}/activity`)).json());
 const region=page.locator('.gi-tool-activity');
 try{
  const submitted=await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:`UX queue gate:${token}`,model:'test-model'}});expect(submitted.status()).toBe(202);const {turn_id}=await submitted.json();
  await expect(region).toHaveAttribute('data-tool-state','running');const running=await activity();expect(running.tool.occurrence_id).toBeTruthy();
  await page.getByRole('button',{name:'Stop response',exact:true}).click();
  await expect(region).toHaveAttribute('data-tool-state','cancelled');await expect(region.getByLabel('shell: Cancelled')).toBeVisible();await expect(region.locator('.spinner')).toHaveCount(0);
  const stopped=await activity();expect(stopped.tool.turn_id).toBe(turn_id);expect(stopped.tool.occurrence_id).toBe(running.tool.occurrence_id);expect(stopped.tool.duration_ms).toBeGreaterThanOrEqual(0);
  const events=(await(await request.get(`/api/turns/${turn_id}/events`)).json()).events;
  expect(events.filter(e=>e.type==='tool.cancelled'&&e.payload.occurrence_id===running.tool.occurrence_id)).toHaveLength(1);
  expect(events.some(e=>e.type==='tool.finished'&&e.payload.occurrence_id===running.tool.occurrence_id)).toBe(false);
  await page.reload();await expect(input).toHaveValue('cancel tool draft β');await expect(region).toHaveAttribute('data-tool-state','cancelled');await expect(region.getByLabel('Tool duration')).toHaveText(`${Math.floor(stopped.tool.duration_ms/1000)}s`);
  expect((await activity()).tool).toEqual(stopped.tool);
 }finally{writeFileSync(resolve(gates,token),'release');}
});
