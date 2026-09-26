import {test,expect} from '@playwright/test';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';
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
 const modelButton=page.getByRole('button',{name:'Open model picker',exact:true});const menu=page.getByRole('listbox',{name:'Models',exact:true});
 const option=name=>menu.getByRole('option').filter({hasText:name});
 const switchTo=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();};
 return{main,child,input,state,turn,modelButton,menu,option,switchTo};
}

test('@ux-compaction-006 Check model context compatibility before switching',async({page,request},info)=>{
 const scenario=loadCorpus().find(x=>x.id==='@ux-compaction-006');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,state,modelButton,menu,option,switchTo}=await fixture(page,request,info);
 await input.fill('unsent retained');await page.locator('.compose-box input[type=file]').setInputFiles(file);
 let mutations=0;page.on('request',r=>{if(r.method()==='PATCH'&&r.url().endsWith(`/api/sessions/${main.id}/model`))mutations++;});
 await modelButton.click();const small=option('ux-local/small');await expect(small).toBeDisabled();await expect(small).toHaveAttribute('aria-disabled','true');await expect(small).toHaveAccessibleDescription('Context window is smaller than the latest measured request.');await expect(small).toHaveAttribute('title','Blocked: ux-local/small context window is smaller than latest measured request');
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
  // The pinned panel focuses its real search field. Search filters rows;
  // clear it before selection so the pending-write current-row guard is visible.
  const search=page.getByRole('combobox',{name:'Search models',exact:true});await expect(search).toBeFocused();
  await page.keyboard.type('large');await expect(menu.locator('.compose-model-catalogue-option.active')).toContainText('ux-local/large');
  await expect(option('ux-local/large')).toHaveAttribute('title','ux-local/large • 200 ctx');await expect(option('ux-local/gate')).toHaveCount(0);
  await search.fill('');await expect(option('ux-local/gate')).toBeVisible();
  if(method==='pointer')await option('ux-local/large').click();else{
   await option('ux-local/large').focus();await page.keyboard.press('Enter');
  }
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

test('@shared-25 Select one coherent session view despite late native responses',async({page,request},info)=>{
 const source=loadCorpus('shared').find(row=>row.id==='@shared-25');expect(source.name).toBe('Select one coherent session view');
 await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const token=`coherent-${info.project.name}-${Date.now()}`;
 const get=async(id,part)=>{const r=await request.get(`/api/sessions/${id}/${part}${part==='messages'?'?limit=50':''}`);expect(r.status()).toBe(200);return r.json();};
 const prompt=async(id,text,intent)=>{const r=await request.post(`/api/sessions/${id}/prompt`,{data:{prompt:text,...(intent?{intent}:{})}});expect(r.status()).toBe(202);return r.json();};
 const gates=[];
 const setup=async(name,model)=>{
  const created=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`${token}-${name}`}});expect(created.status()).toBe(201);
  const {id}=await created.json();
  const measured=await prompt(id,`${name} measured history ${token}`);
  await expect.poll(async()=>(await get(id,'turns')).turns.find(t=>t.id===measured.turn_id)?.status).toBe('completed');
  await expect.poll(async()=>(await get(id,'compaction')).reason).not.toBe('Session has active or queued work');
  expect((await get(id,'model')).context_usage).toMatchObject({tokens:100,source:'provider_request',measurement:{turn_id:measured.turn_id}});
  const patch=await request.patch(`/api/sessions/${id}/model`,{data:{model}});expect(patch.status()).toBe(200);
  const path=resolve('test-results/ux-parity/queue-gates',`${token}-${name}`);mkdirSync(resolve(path,'..'),{recursive:true});gates.push(path);
  const active=await prompt(id,`UX steer gate:${token}-${name}`);
  await expect.poll(async()=>(await get(id,'activity')).turn_id).toBe(active.turn_id);
  const queued=await prompt(id,`${name} durable queued follow-up ${token}`,'queue');
  const messages=await get(id,'messages');expect(messages.has_more).toBe(false);
  const queue=await get(id,'queue');expect(queue.items.map(t=>t.id)).toEqual([queued.turn_id]);expect(queue.active_turn_id).toBe(active.turn_id);
  return{id,name,model,active,queued,messages,queue,modelState:await get(id,'model'),draft:`${name} unsent draft`,media:`${name}-draft.txt`};
 };
 const held=[];let hold=true;
 try {
  const main=await setup('main','ux-local/gate'),research=await setup('research','ux-local/large');
  await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
  await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
  const trigger=page.getByRole('button',{name:/Manage sessions for/}).last();
  const picker=page.getByRole('menu',{name:'Sessions and agents',exact:true});
  const switchTo=async session=>{
   await trigger.focus();await trigger.press('Enter');
   const search=page.getByRole('searchbox',{name:'Search sessions',exact:true});await expect(search).toBeFocused();await search.fill(session.id);
   await expect(picker.getByRole('menuitem')).toHaveCount(1);
   await search.press('ArrowDown');await expect(picker.locator('[role="menuitem"].active')).toContainText(`gi:${session.id}`);await page.keyboard.press('Enter');
   await expect(picker).toHaveCount(0);await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(session.id);
  };
  const view=async(session,other)=>{
   // Positive research-only markers on every surface, never acceptance of an
   // empty transitional view. Timeline fixture is one native initial page.
   await expect.poll(()=>page.locator('.post').evaluateAll(nodes=>nodes.map(n=>n.id))).toEqual(session.messages.messages.map(m=>`post-${m.id}`));
   await expect(page.locator(`[data-queue-id="${session.queued.turn_id}"]`)).toBeVisible();
   await expect(page.locator(`[data-queue-id="${session.queued.turn_id}"]`)).toContainText(`${session.name} durable queued follow-up`);
   await expect(page.locator(`[data-queue-id="${other.queued.turn_id}"]`)).toHaveCount(0);
   await expect(page.getByRole('button',{name:'Open model picker',exact:true})).toHaveText(session.model);
   await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label',session===main?'Context: 100 / 32K tokens (0%)':'Context: 100 / 200 tokens (50%)');
   await expect(input).toHaveValue(session.draft);
   await expect(page.locator('.compose-file-pill')).toHaveCount(1);await expect(page.locator(`.compose-file-pill[title="${session.media}"]`)).toBeVisible();
  };
  const draft=async session=>{await input.fill(session.draft);await page.locator('.compose-box input[type=file]').setInputFiles({name:session.media,mimeType:'text/plain',buffer:Buffer.from(`${session.name} native draft bytes`)});};
  await draft(main);await view(main,research);await switchTo(research);await draft(research);await view(research,main);
  const parts=['messages','model','queue','activity','compaction'];
  const pattern=new RegExp(`/api/sessions/${main.id}/(${parts.join('|')})(?:\\?|$)`);
  await page.route(pattern,async route=>{
   if(!hold)return route.continue();
   const response=await route.fetch();expect(response.status()).toBe(200);
   const part=new URL(route.request().url()).pathname.split('/').pop(),body=await response.json();
   if(part==='messages'){expect(body).toEqual(main.messages);expect(body.has_more).toBe(false);expect(new URL(route.request().url()).searchParams.has('after')).toBe(false);}
   if(part==='model')expect(body).toEqual(main.modelState);
   if(part==='queue')expect(body).toEqual(main.queue);
   let release;const gate=new Promise(r=>release=r);const item={part,release,request:route.request(),body};held.push(item);
   await gate;await route.fulfill({response});
  });
  await switchTo(main);await expect.poll(()=>new Set(held.map(h=>h.part)).size).toBe(parts.length);
  await switchTo(research);await view(research,main);
  research.draft+=' typed while old reads are pending';await input.fill(research.draft);
  // Deliver each real old-session response, allowing response consumption and
  // paint between deliveries. Hold *all* matching requests, not just the first.
  let delivered=0;
  for(const part of parts){
   for(const item of held.filter(h=>h.part===part)){
    const finished=page.waitForEvent('requestfinished',{predicate:r=>r===item.request});item.release();await finished;delivered++;
    await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));await view(research,main);
   }
  }
  expect(delivered).toBe(held.length);expect(held.filter(h=>h.part==='messages')).toHaveLength(1);
  hold=false;await page.unroute(pattern);
  await page.reload();await view(research,main);await switchTo(main);await view(main,research);await switchTo(research);await view(research,main);
  for(const session of [main,research]){
   expect(await get(session.id,'messages')).toEqual(session.messages);expect(await get(session.id,'queue')).toEqual(session.queue);expect(await get(session.id,'model')).toEqual(session.modelState);
   expect((await get(session.id,'turns')).turns).toHaveLength(3);
  }
 }finally{hold=false;for(const item of held)item.release();for(const path of gates)writeFileSync(path,'release');}
});

if(process.env.GI_UX_MODEL_PICKER) test('@shared-34 Filter native models and navigate enabled results without duplicate activation',async({page,request},info)=>{
 const source=loadCorpus('shared').find(row=>row.id==='@shared-34');await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,state,modelButton,menu,option,switchTo}=await fixture(page,request,info);
 const draft='model filter durable draft';await input.fill(draft);await page.locator('.compose-box input[type=file]').setInputFiles(file);
 await page.locator('.post .post-time').first().click();
 await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);
 const storedDraft=()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts');r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:Array.from(new Uint8Array(f.bytes))}))}}:null);};tx.onerror=()=>reject(tx.error);});
 },main.id);
 await expect.poll(storedDraft).toMatchObject({draft:{text:draft,media:[{name:file.name,type:file.mimeType,bytes:Array.from(file.buffer)}]},pending:[]});const savedDraft=await storedDraft();expect(savedDraft.draft.messageRefs).toHaveLength(1);
 const before=await state(),catalogue=before.model_options;
 expect(catalogue.some(x=>x.name==='Forest pine')).toBe(true);
 let mutations=0;page.on('request',r=>{if(r.method()==='PATCH'&&r.url().endsWith(`/api/sessions/${main.id}/model`))mutations++;});
 await modelButton.click();const search=page.getByRole('combobox',{name:'Search models',exact:true});await expect(search).toBeVisible();
 const labels=()=>menu.getByRole('option').evaluateAll(nodes=>nodes.map(n=>n.title.startsWith('Blocked: ')?n.title.slice(9).split(' context window')[0]:n.title.split(' • ')[0]));
 const all=await labels();expect(all).toEqual(catalogue.map(x=>x.label).sort((a,b)=>a.localeCompare(b,undefined,{sensitivity:'base'})));
 for(const [query,expected] of [['ux-local/pine',all.filter(x=>x.includes('ux-local/pine'))],['Forest pine',all.filter(x=>x==='ux-local/pine'||x==='ux-local/pine-small')],['32K ctx',catalogue.filter(x=>x.context_window===32000).map(x=>x.label).sort((a,b)=>a.localeCompare(b,undefined,{sensitivity:'base'}))]]){
  await search.fill(query);await expect.poll(labels).toEqual(expected);await expect(search).toBeFocused();
 }
 await search.fill('no matching native model');await expect(menu.getByRole('option')).toHaveCount(0);await search.press('Enter');expect(mutations).toBe(0);
 await search.fill('Forest pine-small');await expect(menu.getByRole('option')).toHaveCount(1);await expect(menu.getByRole('option')).toBeDisabled();await expect(menu.getByRole('option')).toHaveAccessibleDescription('Context window is smaller than the latest measured request.');await expect(search).not.toHaveAttribute('aria-activedescendant',/.+/);await search.press('ArrowDown');await search.press('Enter');expect(mutations).toBe(0);
 await search.fill('Forest pine');await search.press('Home');await search.press('X');await expect(search).toHaveValue('XForest pine');await search.press('Backspace');await expect(search).toHaveValue('Forest pine');await search.press('End');await search.press(' ');await expect(search).toHaveValue('Forest pine ');
 await search.fill('');const substring=option('aux-local/ux-local/pine-shadow');await substring.focus();
 await page.keyboard.type('ux-local/pi',{delay:10});const pine=option('ux-local/pine •');await expect(pine).toHaveClass(/active/);await expect(pine).toBeFocused();await expect(search).toHaveValue('');
 await page.keyboard.type('per',{delay:10});const piper=option('ux-local/piper');await expect(piper).toBeFocused();await expect(piper).toHaveClass(/active/);
 const enabled=await menu.getByRole('option').evaluateAll(nodes=>nodes.filter(n=>!n.disabled).map(n=>n.textContent.trim()));
 const focused=()=>menu.getByRole('option').evaluateAll(nodes=>nodes.filter(n=>n===document.activeElement).map(n=>n.textContent.trim()));
 await page.keyboard.press('Home');await expect.poll(focused).toEqual([enabled[0]]);
 for(const [key,index] of [['PageDown',8],['PageUp',0],['End',enabled.length-1],['ArrowDown',0],['ArrowUp',enabled.length-1]]){await page.keyboard.press(key);await expect.poll(focused).toEqual([enabled[index]]);}
 await pine.focus();await page.keyboard.press('ArrowDown');await expect(piper).toBeFocused(); // skip the disabled pine-small row
 await page.keyboard.press('Escape');await expect(menu).toHaveCount(0);await expect(modelButton).toBeFocused();expect(mutations).toBe(0);expect((await state()).current).toBe(before.current);await expect(input).toHaveValue(draft);
 await modelButton.click();await expect(search).toHaveValue('');await search.fill('ux-local/piper');await search.press('Enter');await expect(menu).toHaveCount(0);await expect(modelButton).toHaveText('ux-local/piper');expect(mutations).toBe(1);expect((await state()).current).toBe('ux-local/piper');
 // Actual result buttons own Enter and Space even if the old highlight differs.
 for(const [label,key,count] of [['ux-local/pine •','Enter',2],['ux-local/large','Space',3]]){
  await modelButton.click();await option(label).focus();await page.keyboard.press(key);await expect(menu).toHaveCount(0);expect(mutations).toBe(count);
 }
 await expect.poll(storedDraft).toEqual(savedDraft);
 await page.reload();await expect(modelButton).toHaveText('ux-local/large');await expect(input).toHaveValue(draft);await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);expect(await storedDraft()).toEqual(savedDraft);
 await switchTo(child);expect((await state(child)).current).toBe(before.current);await switchTo(main.id);await expect(input).toHaveValue(draft);await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);expect(await storedDraft()).toEqual(savedDraft);
 expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns).toHaveLength(1);
});
