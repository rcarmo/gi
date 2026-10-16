import {test,expect} from '@playwright/test';
import {journeyEnvironment} from '../ux/support/journey-environment.mjs';
async function setup(page,info){
 const env=await journeyEnvironment(info);
 await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();
 const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
 const read=async path=>(await(await page.request.get(env.origin+path)).json());
 const model=()=>read(`/api/sessions/${id}/model`);
 const open=async()=>{await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Settings',exact:true}).click();await page.getByRole('button',{name:'Models',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh models',exact:true})).toBeEnabled()};
 const close=async()=>{await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(input).toBeFocused();};
 return{...env,input,id,read,model,open,closeSettings:close};
}

test('@gi-settings-028 supported session thinking is explicit, durable and reaches native provider payloads',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('thinking draft Ω');await page.locator('.compose-box input[type=file]').setInputFiles({name:'thinking.txt',mimeType:'text/plain',buffer:Buffer.from('keep thinking media')});
  let prompts=0;page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))prompts++});
  await h.open();await expect(page.getByRole('combobox',{name:'Session thinking level',exact:true})).toHaveCount(0);
  await page.getByRole('listbox',{name:'Session model',exact:true}).selectOption('ux-local/reasoner');await page.getByRole('button',{name:'Apply model',exact:true}).click();
  const select=page.getByRole('combobox',{name:'Session thinking level',exact:true});await expect(select).toBeEnabled();expect(await select.locator('option').allTextContents()).toEqual(['Default (high)','low','high']);
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
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===next.turn_id)?.status).toBe('completed');// No session level: Pi's default (medium) clamped to the model's levels.
  expect((await h.model()).default_thinking_level).toBe('high');
  await expect(page.locator('.post').filter({hasText:'Provider model reasoner thinking high:'}).filter({hasText:'provider default next'})).toBeVisible();expect(prompts).toBe(2);
 }finally{await h.close()}
});

test('@gi-settings-029 model switching resets thinking without dispatching the existing draft',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('keep this model switch draft Ω');
  await h.open();const model=page.getByRole('listbox',{name:'Session model',exact:true});
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
  // The reset level falls back to Pi's default thinking level, clamped to the model.
  await expect(page.locator('.post').filter({hasText:'Provider model reasoner thinking high:'}).filter({hasText:'after model switch'})).toBeVisible();
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
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[id="compose-session-option-${encodeURIComponent('gi:'+other.id)}"]`).click();await h.input.fill('other thinking draft');release();await page.unrouteAll({behavior:'wait'});
  await h.open();await expect(page.getByTestId('settings-current-model')).toHaveText(before.current);await expect(select).toHaveCount(0);expect(await h.read(`/api/sessions/${other.id}/model`)).toEqual(before);expect((await h.model()).thinking_level).toBe('high');await h.closeSettings();await expect(h.input).toHaveValue('other thinking draft');
  expect((await h.read(`/api/sessions/${h.id}/turns`)).turns||[]).toHaveLength(0);expect((await h.read(`/api/sessions/${other.id}/turns`)).turns||[]).toHaveLength(0);
 }finally{release();await page.unrouteAll({behavior:'wait'});await h.close()}
});
