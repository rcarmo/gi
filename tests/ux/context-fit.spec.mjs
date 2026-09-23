import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
const file={name:'draft.txt',mimeType:'text/plain',buffer:Buffer.from('keep these bytes')};
async function fixture(page,request,info){
 const token=`context-fit-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-child',title:token+'-child'}})).json()).branch.chat_jid.slice(3);
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const state=async (id=main.id)=>(await(await request.get(`/api/sessions/${id}/model`)).json());
 expect((await state()).context_usage.tokens).toBeNull();
 const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${main.id}/prompt`)&&r.request().method()==='POST');
 await input.fill('Measure a real local-provider request');await input.press('Enter');const turn=await(await response).json();
 await expect.poll(async()=> (await state()).context_usage.tokens).toBe(100);
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 32K tokens (0%)',{timeout:15000});
 // Measurement provenance is exposed by the real model API, never seeded in DB.
 expect((await state()).context_usage.measurement).toMatchObject({turn_id:turn.turn_id,model:'ux-local/gate',iteration:1});
 const modelButton=page.getByRole('button',{name:'Open model picker',exact:true});const menu=page.getByRole('menu',{name:'Model picker',exact:true});
 const option=name=>menu.getByRole('menuitem').filter({hasText:name});
 const switchTo=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();};
 return{main,child,input,state,turn,modelButton,menu,option,switchTo};
}

test('@ux-compaction-006 Check model context compatibility before switching',async({page,request},info)=>{
 const scenario=loadCorpus().find(x=>x.id==='@ux-compaction-006');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,state,modelButton,menu,option,switchTo}=await fixture(page,request,info);
 await input.fill('unsent retained');await page.locator('.compose-box input[type=file]').setInputFiles(file);
 let mutations=0;page.on('request',r=>{if(r.method()==='PATCH'&&r.url().endsWith(`/api/sessions/${main.id}/model`))mutations++;});
 await modelButton.click();const small=option('ux-local/small');await expect(small).toBeDisabled();await expect(small).toHaveAttribute('title','Blocked: ux-local/small context window is smaller than latest measured request');
 const box=await small.boundingBox();await page.mouse.click(box.x+box.width/2,box.y+box.height/2);expect(mutations).toBe(0);
 await page.keyboard.press('Escape');await expect(modelButton).toHaveText('ux-local/gate');await expect(input).toHaveValue('unsent retained');
 const denied=await request.patch(`/api/sessions/${main.id}/model`,{data:{model:'ux-local/small'}});expect(denied.status()).toBe(400);expect((await denied.json()).error).toContain('context window 80 is smaller than latest measured request (100 tokens)');
 expect((await state()).current).toBe('ux-local/gate');
 await modelButton.click();await expect(option('ux-local/equal')).toBeEnabled();await option('ux-local/equal').click();await expect(menu).toHaveCount(0);await expect(modelButton).toHaveText('ux-local/equal');
 expect((await state()).current).toBe('ux-local/equal');
 await page.reload();await expect(input).toHaveValue('unsent retained');await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);
 // Known capacity with unknown usage must not inherit another session's gate.
 await switchTo(child);expect((await state(child)).context_usage.tokens).toBeNull();await modelButton.click();await expect(option('ux-local/small')).toBeEnabled();await option('ux-local/small').click();
 await expect.poll(async()=>(await state(child)).current).toBe('ux-local/small');expect((await state()).current).toBe('ux-local/equal');
 await switchTo(main.id);await expect(input).toHaveValue('unsent retained');await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);
});

test('@ux-compaction-007 Refresh model information after an accepted switch',async({page,request},info)=>{
 const scenario=loadCorpus().find(x=>x.id==='@ux-compaction-007');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,state,turn,modelButton,menu,option,switchTo}=await fixture(page,request,info);
 await input.fill('retained after accepted switch');await page.locator('.compose-box input[type=file]').setInputFiles(file);
 await modelButton.click();await option('ux-local/equal').click();await expect(menu).toHaveCount(0);await expect(modelButton).toHaveText('ux-local/equal');
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 100 tokens (100%)');
 let selected=await state();expect(selected.current).toBe('ux-local/equal');expect(selected.context_usage).toMatchObject({tokens:100,contextWindow:100,percent:100,source:'provider_request'});
 await modelButton.click();await option('ux-local/large').click();await expect(modelButton).toHaveText('ux-local/large');await expect(menu).toHaveCount(0);
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 200 tokens (50%)');
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('data-tooltip','Context: 100 / 200 tokens (50%) — latest measured provider request — Compact context');
 selected=await state();expect(selected.context_usage.measurement.turn_id).toBe(turn.turn_id);expect(selected.context_usage.tokens).toBe(100);
 expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns).toHaveLength(1);
 await page.reload();await expect(modelButton).toHaveText('ux-local/large');await expect(input).toHaveValue('retained after accepted switch');await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);
 await switchTo(child);expect((await state(child)).context_usage.tokens).toBeNull();await switchTo(main.id);await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 200 tokens (50%)');
});

test('@gi-settings-006 Settings model fit uses measured native context and keeps unknowns selectable',async({page,request},info)=>{
 const{input,state,child,switchTo}=await fixture(page,request,info);
 await input.fill('settings measured draft');
 await page.keyboard.press('Control+,');
 const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});
 await dialog.getByRole('button',{name:'Models',exact:true}).click();
 const choice=dialog.getByLabel('Session model',{exact:true});await expect(choice).toBeVisible();
 await dialog.getByLabel('Filter models',{exact:true}).fill('ux-local/small');
 await choice.selectOption('ux-local/small');
 await expect(dialog.getByRole('button',{name:'Apply model'})).toBeDisabled();
 await expect(dialog.getByRole('status')).toContainText('cannot fit the measured context');
 expect((await state()).current).toBe('ux-local/gate');
 await dialog.getByLabel('Filter models',{exact:true}).fill('ux-local/equal');
 await choice.selectOption('ux-local/equal');await dialog.getByRole('button',{name:'Apply model'}).click();
 await expect(dialog.getByTestId('settings-current-model')).toHaveText('ux-local/equal');
 await page.keyboard.press('Escape');await expect(input).toHaveValue('settings measured draft');
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 100 tokens (100%)');
 await switchTo(child);expect((await state(child)).context_usage.tokens).toBeNull();
 await page.keyboard.press('Control+,');await dialog.getByRole('button',{name:'Models',exact:true}).click();
 await dialog.getByLabel('Filter models',{exact:true}).fill('ux-local/small');
 await choice.selectOption('ux-local/small');await expect(dialog.getByRole('button',{name:'Apply model'})).toBeEnabled();
 await dialog.getByRole('button',{name:'Apply model'}).click();await expect(dialog.getByTestId('settings-current-model')).toHaveText('ux-local/small');
});

if(process.env.GI_UX_SETTINGS_CATALOGUE){
 test('@gi-settings-004 Settings caps a real native catalogue and filters beyond the first page',async({page,request},info)=>{
  const{input,state}=await fixture(page,request,info);
  expect((await state()).model_options.length).toBeGreaterThan(50);
  await input.fill('bounded catalogue draft');await page.keyboard.press('Control+,');
  const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});await dialog.getByRole('button',{name:'Models',exact:true}).click();
  const select=dialog.getByLabel('Session model',{exact:true});await expect(select).toBeVisible();
  await expect(select.locator('option:not([disabled])')).toHaveCount(50);await expect(dialog.getByText(/Refine the filter/)).toBeVisible();
  await dialog.getByLabel('Filter models',{exact:true}).fill('settings-59');await expect(select.locator('option:not([disabled])')).toHaveCount(1);
  await expect(dialog.getByRole('button',{name:'Apply model'})).toBeDisabled();
  await select.selectOption('ux-local/settings-59');await dialog.getByRole('button',{name:'Apply model'}).click();
  await expect(dialog.getByTestId('settings-current-model')).toHaveText('ux-local/settings-59');
  await page.keyboard.press('Escape');await expect(input).toHaveValue('bounded catalogue draft');
 });
}

for(const [id,method] of [['@shared-31','pointer'],['@shared-32','keyboard']]) test(`${id} Search native model capabilities and accept a session-only change using ${method}`,async({page,request},info)=>{
 const source=loadCorpus('shared').find(row=>row.id===id);expect(source).toBeTruthy();expect(source.steps.join('\n')).toContain(method);await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,state,modelButton,menu,option,switchTo}=await fixture(page,request,info);
 await expect.poll(async()=>(await(await request.get(`/api/sessions/${main.id}`)).json()).state.status).toBe('idle');
 const initial=await state(),otherBefore=await state(child);
 const find=label=>initial.model_options.find(o=>o.label===label);
 expect(find('ux-local/gate')).toMatchObject({provider:'ux-local',context_window:32000,authenticated:true});
 expect(find('ux-local/large')).toMatchObject({provider:'ux-local',context_window:200,authenticated:true});
 const filename=`shared-model-${method}-${info.project.name}.txt`;
 const written=await request.post('/api/tools/execute',{data:{tool:'write',input:{path:filename,content:'retain file reference bytes'}}});expect(written.ok()).toBe(true);expect((await written.json()).error).toBeFalsy();
 await page.getByRole('button',{name:'Menu',exact:true}).click();
 const show=page.getByRole('menuitem',{name:'Show workspace',exact:true});if(await show.count())await show.click();else await page.keyboard.press('Escape');
 await page.locator(`.workspace-row[data-path="${filename}"]`).click();
 await page.locator('.workspace-toggle-tab.open').click();
 const referenceLink=page.locator('.post .post-time').first();
 const referenceId=(await referenceLink.getAttribute('href')).replace(/^#msg-/,'');expect(referenceId).toBeTruthy();
 await referenceLink.click();
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'shared-model-draft.txt',mimeType:'text/plain',buffer:Buffer.from('shared model retained media')});
 await input.fill('shared model unsent text');
 const draftLabels=await page.locator('.compose-file-pill').evaluateAll(nodes=>nodes.map(n=>n.getAttribute('title')));
 expect(draftLabels).toContain(filename);expect(draftLabels).toContain('shared-model-draft.txt');expect(draftLabels).toContain(`Message reference: ${referenceId}`);await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);
 let release,held=false;const gate=new Promise(r=>{release=r;});let patches=0;
 await page.route(`**/api/sessions/${main.id}/model`,async route=>{
  if(route.request().method()!=='PATCH')return route.continue();
  patches++;expect(route.request().postDataJSON()).toEqual({model:'ux-local/large'});
  const response=await route.fetch();expect(response.status()).toBe(200);expect(await response.json()).toMatchObject({current:'ux-local/large',context_window:200});held=true;await gate;await route.fulfill({response});
 });
 try {
  if(method==='pointer')await modelButton.click();else{await modelButton.focus();await modelButton.press('Enter');}
  await expect(menu).toBeVisible();await expect(option('ux-local/gate')).toBeVisible();await expect(option('ux-local/gate')).toHaveAttribute('title','ux-local/gate • 32K ctx');await expect(option('ux-local/large')).toBeVisible();
  // The supplied picker performs incremental native typeahead, not a fabricated
  // filter field. Nonmatching catalogue rows remain visible and unmodified.
  await page.keyboard.type('large');await expect(menu.locator('.compose-model-popup-item.active')).toContainText('ux-local/large');
  await expect(option('ux-local/large')).toHaveAttribute('title','ux-local/large • 200 ctx');
  if(method==='pointer')await option('ux-local/large').click();else await page.keyboard.press('Enter');
  await expect.poll(()=>held).toBe(true);expect(patches).toBe(1);
  await expect(modelButton).toHaveText('Switching…');await expect(modelButton).toBeDisabled();await expect(menu.locator('.current-model')).toContainText('ux-local/gate');await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 32K tokens (0%)');
  release();await expect(menu).toHaveCount(0);await expect(modelButton).toHaveText('ux-local/large');await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 200 tokens (50%)');
  expect((await state()).current).toBe('ux-local/large');expect((await state()).context_usage).toMatchObject({tokens:100,contextWindow:200});expect(await state(child)).toEqual(otherBefore);
  await expect(input).toHaveValue('shared model unsent text');expect(await page.locator('.compose-file-pill').evaluateAll(nodes=>nodes.map(n=>n.getAttribute('title')))).toEqual(draftLabels);await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);
  await page.unroute(`**/api/sessions/${main.id}/model`);await page.reload();await expect(modelButton).toHaveText('ux-local/large');await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 200 tokens (50%)');
  await expect(input).toHaveValue('shared model unsent text');expect(await page.locator('.compose-file-pill').evaluateAll(nodes=>nodes.map(n=>n.getAttribute('title')))).toEqual(draftLabels);await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);
  await switchTo(child);await expect(modelButton).toHaveText(otherBefore.current);await expect(input).toHaveValue('');await expect(page.locator('.compose-file-pill')).toHaveCount(0);
  await switchTo(main.id);await expect(input).toHaveValue('shared model unsent text');await expect(modelButton).toHaveText('ux-local/large');expect(await page.locator('.compose-file-pill').evaluateAll(nodes=>nodes.map(n=>n.getAttribute('title')))).toEqual(draftLabels);await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);
  expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns).toHaveLength(1);expect((await(await request.get(`/api/sessions/${child}/turns`)).json()).turns||[]).toHaveLength(0);
 } finally {release();}
});
