import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';
async function setup(page,info){
 const env=await journeyEnvironment(info);
 await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();
 const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
 const read=async path=>(await(await page.request.get(env.origin+path)).json());
 const model=()=>read(`/api/sessions/${id}/model`);
 const open=async()=>{await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Settings',exact:true}).click();await page.getByRole('button',{name:'Models',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh models',exact:true})).toBeEnabled()};
 const close=()=>page.getByRole('button',{name:'Close settings',exact:true}).click();
 return{...env,input,id,read,model,open,closeSettings:close};
}

test('@gi-settings-028 supported session thinking is explicit, durable and reaches native provider payloads',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('thinking draft Ω');await page.locator('.compose-box input[type=file]').setInputFiles({name:'thinking.txt',mimeType:'text/plain',buffer:Buffer.from('keep thinking media')});
  let prompts=0;page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))prompts++});
  await h.open();await expect(page.getByRole('combobox',{name:'Session thinking level',exact:true})).toHaveCount(0);
  await page.getByRole('combobox',{name:'Session model',exact:true}).selectOption('ux-local/reasoner');await page.getByRole('button',{name:'Apply model',exact:true}).click();
  const select=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(select).toBeEnabled();expect(await select.locator('option').allTextContents()).toEqual(['Provider default','low','high']);
  await select.selectOption('high');expect((await h.model()).thinking_level).toBe('');await page.getByRole('button',{name:'Apply thinking',exact:true}).click();
  await expect(page.getByText('Thinking applied to future turns in this session.',{exact:true})).toBeVisible();await expect(select).toHaveValue('high');expect(prompts).toBe(0);
  expect(await h.model()).toMatchObject({current:'ux-local/reasoner',thinking_level:'high',thinking_levels:['low','high'],supports_thinking:true});
  await h.closeSettings();await expect(h.input).toHaveValue('thinking draft Ω');await expect(page.locator('.compose-file-pill[title="thinking.txt"]')).toBeVisible();
  await page.reload();await expect(h.input).toHaveValue('thinking draft Ω');await h.open();await expect(select).toHaveValue('high');await h.closeSettings();
  const response=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/prompt`));await h.input.press('Enter');const admitted=await(await response).json();
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===admitted.turn_id)?.status).toBe('completed');
  const turn=(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===admitted.turn_id);expect(turn.metadata).toMatchObject({selected_thinking_level:'high',selected_thinking_model:'ux-local/reasoner'});
  await expect(page.locator('.post').filter({hasText:'Provider model reasoner thinking high:'})).toBeVisible();expect(prompts).toBe(1);
  await h.open();await select.selectOption('');await page.getByRole('button',{name:'Apply thinking',exact:true}).click();await expect(select).toHaveValue('');await expect(page.getByRole('button',{name:'Apply thinking',exact:true})).toBeDisabled();await h.closeSettings();
  await h.input.fill('provider default next');const second=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/prompt`));await h.input.press('Enter');const next=await(await second).json();
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===next.turn_id)?.status).toBe('completed');await expect(page.locator('.post').filter({hasText:'Provider model reasoner thinking absent:'})).toBeVisible();expect(prompts).toBe(2);
 }finally{await h.close()}
});

test('@gi-settings-029 model switching resets thinking without dispatching the existing draft',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('keep this model switch draft Ω');
  await h.open();const model=page.getByRole('combobox',{name:'Session model',exact:true});
  await model.selectOption('ux-local/reasoner');await page.getByRole('button',{name:'Apply model',exact:true}).click();
  const levels=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(levels).toBeEnabled();
  await levels.selectOption('high');await page.getByRole('button',{name:'Apply thinking',exact:true}).click();
  expect((await h.model()).thinking_level).toBe('high');
  await model.selectOption('ux-local/gate');await page.getByRole('button',{name:'Apply model',exact:true}).click();
  await expect(levels).toHaveCount(0);expect((await h.model()).thinking_level).toBe('');
  expect((await h.read(`/api/sessions/${h.id}/turns`)).turns||[]).toHaveLength(0);
  await h.closeSettings();await expect(h.input).toHaveValue('keep this model switch draft Ω');
  await page.reload();await expect(h.input).toHaveValue('keep this model switch draft Ω');
  await h.open();await model.selectOption('ux-local/reasoner');await page.getByRole('button',{name:'Apply model',exact:true}).click();
  await expect(levels).toHaveValue('');await h.closeSettings();
  await h.input.fill('after model switch');const response=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/prompt`));await h.input.press('Enter');const admitted=await(await response).json();
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===admitted.turn_id)?.status).toBe('completed');
  await expect(page.locator('.post').filter({hasText:'Provider model reasoner thinking absent:'})).toBeVisible();
 }finally{await h.close()}
});

test('stale or unsupported thinking changes fail without prompt, model or foreign-session changes',async({page},info)=>{
 const h=await setup(page,info);try{
  const other=await(await page.request.post(h.origin+'/api/sessions',{data:{title:'other'}})).json();const beforeOther=await h.read(`/api/sessions/${other.id}/model`);
  await page.request.patch(h.origin+`/api/sessions/${h.id}/model`,{data:{model:'ux-local/reasoner'}});await page.reload();await h.input.fill('unsent stale settings');await h.open();
  const select=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(select).toBeEnabled();const stale=await h.model();await select.selectOption('high');
  expect((await page.request.patch(h.origin+`/api/sessions/${h.id}/model`,{data:{model:stale.current,thinking_level:'medium',thinking_token:stale.thinking_token}})).status()).toBe(400);
  expect((await page.request.patch(h.origin+`/api/sessions/${other.id}/model`,{data:{model:stale.current,thinking_level:'high',thinking_token:stale.thinking_token}})).status()).toBe(409);
  await page.request.patch(h.origin+`/api/sessions/${h.id}/model`,{data:{model:'ux-local/gate'}});await page.getByRole('button',{name:'Apply thinking',exact:true}).click();await expect(page.getByRole('alert')).toBeVisible();await expect(select).toHaveCount(0);expect((await h.model()).current).toBe('ux-local/gate');
  const state=await h.model();expect((await page.request.patch(h.origin+`/api/sessions/${h.id}/model`,{data:{model:state.current,thinking_level:'high',thinking_token:state.thinking_token}})).status()).toBe(400);
  expect((await h.read(`/api/sessions/${h.id}/turns`)).turns||[]).toHaveLength(0);expect(await h.read(`/api/sessions/${other.id}/model`)).toEqual(beforeOther);await h.closeSettings();await expect(h.input).toHaveValue('unsent stale settings');
 }finally{await h.close()}
});

test('legacy bare default model uses its provider-bound thinking on normal send',async({page},info)=>{
 const previous=process.env.GI_UX_THINKING_DEFAULT;process.env.GI_UX_THINKING_DEFAULT='1';let h;
 try{
  h=await setup(page,info);const current=await h.model();expect(current.current).toBe('ux-local/reasoner');expect(current.thinking_level).toBe('');
  await h.open();const select=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(select).toBeEnabled();await select.selectOption('low');await page.getByRole('button',{name:'Apply thinking',exact:true}).click();await expect(page.getByText('Thinking applied to future turns in this session.',{exact:true})).toBeVisible();await h.closeSettings();
  await h.input.fill('bare default thinking');await h.input.press('Enter');await expect(page.locator('.post').filter({hasText:'Provider model reasoner thinking low:'})).toBeVisible();
  const rows=(await h.read(`/api/sessions/${h.id}/turns`)).turns;expect(rows).toHaveLength(1);expect(rows[0].metadata).toMatchObject({model:'ux-local/reasoner',selected_thinking_model:'ux-local/reasoner',selected_thinking_level:'low'});
 }finally{if(previous===undefined)delete process.env.GI_UX_THINKING_DEFAULT;else process.env.GI_UX_THINKING_DEFAULT=previous;await h?.close()}
});

test('lost thinking acknowledgement is not replayed and refresh reads native result',async({page},info)=>{
 const h=await setup(page,info);try{
  await page.request.patch(h.origin+`/api/sessions/${h.id}/model`,{data:{model:'ux-local/reasoner'}});await page.reload();await h.input.fill('lost thinking draft');await h.open();const select=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(select).toBeEnabled();await select.selectOption('high');
  let writes=0;await page.route(`**/api/sessions/${h.id}/model`,async route=>{if(route.request().method()!=='PATCH')return route.continue();writes++;const response=await route.fetch();expect(response.status()).toBe(200);await route.abort('failed')});
  await page.getByRole('button',{name:'Apply thinking',exact:true}).click();await expect(page.getByRole('alert')).toBeVisible();await expect(page.getByRole('button',{name:'Refresh models',exact:true})).toBeEnabled();await page.getByRole('button',{name:'Refresh models',exact:true}).click();await expect(select).toHaveValue('high');await expect(page.getByRole('button',{name:'Apply thinking',exact:true})).toBeDisabled();expect(writes).toBe(1);expect((await h.model()).thinking_level).toBe('high');
  await h.closeSettings();await expect(h.input).toHaveValue('lost thinking draft');expect((await h.read(`/api/sessions/${h.id}/turns`)).turns||[]).toHaveLength(0);
 }finally{await h.close()}
});

test('late origin thinking acknowledgement cannot update a different selected session',async({page},info)=>{
 const h=await setup(page,info);let release;const gate=new Promise(r=>release=r);let held=false;
 try{
  const other=await(await page.request.post(h.origin+'/api/sessions',{data:{title:'thinking other'}})).json();const before=await h.read(`/api/sessions/${other.id}/model`);
  await page.request.patch(h.origin+`/api/sessions/${h.id}/model`,{data:{model:'ux-local/reasoner'}});await page.reload();await h.input.fill('origin thinking draft');await h.open();const select=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(select).toBeEnabled();await select.selectOption('high');
  await page.route(`**/api/sessions/${h.id}/model`,async r=>{if(r.request().method()!=='PATCH')return r.continue();const response=await r.fetch();expect(response.status()).toBe(200);held=true;await gate;await r.fulfill({response})});
  await page.getByRole('button',{name:'Apply thinking',exact:true}).click();await expect.poll(()=>held).toBe(true);await h.closeSettings();
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${other.id}"]`).getByRole('menuitem').click();await h.input.fill('other thinking draft');release();await page.unrouteAll({behavior:'wait'});
  await h.open();await expect(page.getByTestId('settings-current-model')).toHaveText(before.current);await expect(select).toHaveCount(0);expect(await h.read(`/api/sessions/${other.id}/model`)).toEqual(before);expect((await h.model()).thinking_level).toBe('high');await h.closeSettings();await expect(h.input).toHaveValue('other thinking draft');
  expect((await h.read(`/api/sessions/${h.id}/turns`)).turns||[]).toHaveLength(0);expect((await h.read(`/api/sessions/${other.id}/turns`)).turns||[]).toHaveLength(0);
 }finally{release();await page.unrouteAll({behavior:'wait'});await h.close()}
});

test('@shared-35 Configure thinking and context controls only from advertised capabilities',async({page},info)=>{
 const previous=process.env.GI_UX_COMPACTION;process.env.GI_UX_COMPACTION='1';let h;
 try{
  h=await setup(page,info);const pie=page.locator('.compose-context-pie');const initial=await h.model();
  expect(initial.context_usage.tokens).toBeNull();expect(initial.context_usage.percent).toBeNull();expect(initial.context_usage.source).toBe('unavailable');
  await expect(pie).toHaveAttribute('aria-label',/Context: \? .*\(\?%\)/);await expect(pie).toHaveAttribute('data-tooltip',/usage unavailable/);
  const unavailable=await h.read(`/api/sessions/${h.id}/compaction`);expect(unavailable.available).toBe(false);expect(unavailable.reason).toBe('Not enough eligible context');expect(unavailable.token).toBeTruthy();expect(unavailable.policy).toBeTruthy();await expect(pie).toBeDisabled();
  let compactions=0;page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/compaction'))compactions++});await pie.dispatchEvent('click');expect(compactions).toBe(0);
  await h.open();await expect(page.getByRole('combobox',{name:'Session thinking level',exact:true})).toHaveCount(0);expect(initial.supports_thinking).toBe(false);
  await page.getByRole('combobox',{name:'Session model',exact:true}).selectOption('ux-local/reasoner');await page.getByRole('button',{name:'Apply model',exact:true}).click();
  const levels=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(levels).toBeEnabled();const capable=await h.model();expect(capable.supports_thinking).toBe(true);expect(capable.thinking_levels).toEqual(['low','high']);await levels.selectOption('low');await page.getByRole('button',{name:'Apply thinking',exact:true}).click();await expect(page.getByText('Thinking applied to future turns in this session.',{exact:true})).toBeVisible();await h.closeSettings();
  for(let i=0;i<2;i++){
   const response=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/prompt`));await h.input.fill(`shared35 native history ${i}`);await h.input.press('Enter');const turn=await(await response).json();
   await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===turn.turn_id)?.status).toBe('completed');
   await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/compaction`)).reason).not.toBe('Session has active or queued work');
  }
  await expect(page.locator('.post').filter({hasText:'Provider model reasoner thinking low:'})).toHaveCount(2);
  const measured=(await h.model()).context_usage;expect(measured.source).toBe('provider_request');expect(measured.tokens).toBeGreaterThan(0);
  const native=await h.read(`/api/sessions/${h.id}/compaction`);expect(native.available).toBe(true);expect(native.reason).toBe('');expect(native.token).toBeTruthy();expect(native.policy).toEqual(unavailable.policy);await expect(pie).toBeEnabled();
  await h.input.fill('Shared35 unsent draft Ω');await page.locator('.compose-box input[type=file]').setInputFiles({name:'shared35.txt',mimeType:'text/plain',buffer:Buffer.from('keep context attachment')});
  const compact=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/compaction`));await pie.click();const response=await compact;expect(response.status()).toBe(202);expect(response.request().postDataJSON()).toEqual({token:native.token});const turn=await response.json();
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/activity`)).compaction?.active).toBe(true);const event=(await h.read(`/api/sessions/${h.id}/activity`)).compaction;
  expect(event.tokens_source).toBe('estimate');expect(Number.isSafeInteger(event.tokens_before)).toBe(true);expect(event.tokens_before).toBeGreaterThan(0);
  await expect(pie).toHaveAttribute('data-tooltip',new RegExp(`estimated history: ${event.tokens_before} tokens`));await expect(pie).toHaveAttribute('aria-label',/estimated history:/);await expect(pie).toHaveAttribute('data-tooltip',/latest measured provider request/);
  expect((await h.model()).context_usage).toEqual(measured);await expect(pie).toBeDisabled();await pie.dispatchEvent('click');expect(compactions).toBe(1);
  // Older activity payloads can omit estimate provenance; the real count must
  // then disappear from the estimate label without changing measured usage.
  await page.route(`**/api/sessions/${h.id}/activity`,async route=>{
   if(route.request().method()!=='GET')return route.continue();const response=await route.fetch(),data=await response.json();if(data.compaction)delete data.compaction.tokens_source;await route.fulfill({response,json:data});
  });
  await page.reload();await expect(pie).toHaveClass(/is-compacting/);await expect(pie).not.toHaveAttribute('data-tooltip',/estimated history:/);await expect(pie).not.toHaveAttribute('aria-label',/estimated history:/);await expect(pie).toHaveAttribute('data-tooltip',/latest measured provider request/);expect((await h.model()).context_usage).toEqual(measured);await expect(h.input).toHaveValue('Shared35 unsent draft Ω');
  await page.unrouteAll({behavior:'wait'});
  h.release('manual-'+h.id);await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===turn.turn_id)?.status).toBe('completed');
  await expect(pie).not.toHaveAttribute('data-tooltip',/estimated history:/);expect((await h.model()).context_usage).toEqual(measured);await expect(h.input).toHaveValue('Shared35 unsent draft Ω');await expect(page.locator('.compose-file-pill[title="shared35.txt"]')).toBeVisible();
 }finally{if(previous===undefined)delete process.env.GI_UX_COMPACTION;else process.env.GI_UX_COMPACTION=previous;if(h){h.release('manual-'+h.id);await h.close()}}
});
