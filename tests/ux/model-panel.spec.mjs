import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';

test('Native model panel search, metadata, clear and Models-settings handoff preserve draft and focus',async({page,request},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Panel handoff draft Ω');
  const id=await page.evaluate(()=>localStorage.getItem('gi_session_id')),before=await(await request.get(`${env.origin}/api/sessions/${id}/model`)).json();
  const writes=[];await page.route('**/api/**',route=>{if(!['GET','HEAD','OPTIONS'].includes(route.request().method())){writes.push(route.request().url());return route.abort();}return route.continue();});
  const trigger=page.getByRole('button',{name:'Open model picker',exact:true});await trigger.click();
  const panel=page.locator('.compose-model-catalogue'),search=panel.getByRole('combobox',{name:'Search models',exact:true});await expect(search).toBeFocused();
  await expect(panel.locator('.compose-model-catalogue-section-heading').first()).toContainText('Current');
  await expect(panel.locator('.current-model')).toHaveCount(1);await expect(panel.locator('.compose-model-catalogue-option-name').first()).not.toHaveText('');
  await search.fill('does-not-exist');await expect(panel.getByRole('listbox',{name:'Models',exact:true}).getByRole('option')).toHaveCount(0);await expect(panel.locator('.compose-model-catalogue-summary')).toContainText('0 models');
  await panel.getByRole('button',{name:'Clear model search',exact:true}).click();await expect(search).toBeFocused();await expect(panel.getByRole('listbox',{name:'Models',exact:true}).getByRole('option').first()).toBeVisible();
  await search.fill(before.current);await expect(panel.getByRole('listbox',{name:'Models',exact:true}).getByRole('option')).toHaveCount(1);await expect(panel.getByRole('listbox',{name:'Models',exact:true}).getByRole('option')).toContainText(before.current);
  await panel.getByRole('button',{name:'Open Models settings',exact:true}).click();await expect(panel).toHaveCount(0);
  const dialog=page.getByRole('dialog');await expect(dialog).toBeVisible();await expect(dialog.locator('.settings-nav-item.active')).toHaveText('Models');await expect(dialog.getByRole('searchbox',{name:'Filter models',exact:true})).toBeVisible();
  await page.keyboard.press('Escape');await expect(dialog).toBeHidden();await expect(trigger).toBeFocused();await expect(input).toHaveValue('Panel handoff draft Ω');
  await trigger.click();await expect(search).toBeFocused();await expect(panel.locator('.compose-model-catalogue-summary')).not.toContainText('Refreshing');
  const initial=await panel.locator('[data-model-index].active').getAttribute('data-model-index');
  await search.press('ArrowDown');await expect(search).toBeFocused();await expect(panel.locator('[data-model-index].active')).not.toHaveAttribute('data-model-index',initial);
  await page.keyboard.press('Escape');await expect(trigger).toBeFocused();expect(writes).toEqual([]);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);
 }finally{await env.close();}
});

test('Invalid Settings section requests use General and model panel dismissal stays modal-owned',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Settings fallback draft');
  await page.evaluate(()=>window.dispatchEvent(new CustomEvent('piclaw:open-settings',{detail:{section:'not-a-pane',opener:{}}})));
  const dialog=page.getByRole('dialog');await expect(dialog).toBeVisible();await expect(dialog.locator('.settings-nav-item.active')).toHaveText('General');await page.keyboard.press('Escape');await expect(input).toBeFocused();
  await page.getByRole('button',{name:'Open model picker',exact:true}).click();await page.keyboard.press('Control+,');await expect(dialog).toBeVisible();await expect(dialog.locator('.settings-nav-item.active')).toHaveText('General');
  await page.keyboard.press('Escape');await expect(dialog).toBeHidden();await expect(page.locator('.compose-model-catalogue')).toBeVisible();await page.keyboard.press('Escape');await expect(page.locator('.compose-model-catalogue')).toBeHidden();await expect(input).toHaveValue('Settings fallback draft');
 }finally{await env.close();}
});

test('Model combobox owns a stable listbox descendant without confusing active and selected options',async({page,request},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Accessible model draft Ω');
  const id=await page.evaluate(()=>localStorage.getItem('gi_session_id')),before=await(await request.get(`${env.origin}/api/sessions/${id}/model`)).json();
  const trigger=page.getByRole('button',{name:'Open model picker',exact:true});await expect(trigger).toHaveAttribute('aria-haspopup','listbox');await expect(trigger).toHaveAttribute('aria-expanded','false');await trigger.click();await expect(trigger).toHaveAttribute('aria-expanded','true');
  const box=page.getByRole('combobox',{name:'Search models',exact:true}),list=page.getByRole('listbox',{name:'Models',exact:true});await expect(box).toBeFocused();await expect(list).toHaveAttribute('aria-busy','false');await expect(box).toHaveAttribute('aria-expanded','true');await expect(box).toHaveAttribute('aria-autocomplete','list');
  const listId=await list.getAttribute('id');await expect(box).toHaveAttribute('aria-controls',listId);await expect(trigger).toHaveAttribute('aria-controls',listId);
  const ids=await list.getByRole('option').evaluateAll(nodes=>nodes.map(n=>n.id));expect(new Set(ids).size).toBe(ids.length);expect(ids.every(Boolean)).toBe(true);expect(ids.length).toBeGreaterThan(1);
  await expect(list.getByRole('option',{selected:true})).toHaveCount(1);await expect(list.getByRole('option',{selected:true})).toHaveAttribute('aria-label',new RegExp(before.current.replace(/[.*+?^${}()|[\]\\]/g,'\\$&')));
  const currentId=await list.getByRole('option',{selected:true}).getAttribute('id');await expect(box).toHaveAttribute('aria-activedescendant',currentId);
  await box.press('ArrowDown');await expect(box).toBeFocused();await expect(box).not.toHaveAttribute('aria-activedescendant',currentId);await expect(list.getByRole('option',{selected:true})).toHaveAttribute('id',currentId);
  const linked=await box.getAttribute('aria-activedescendant');expect(await list.getByRole('option').evaluateAll((nodes,id)=>nodes.some(n=>n.id===id&&n.getAttribute('aria-disabled')==='false'),linked)).toBe(true);
  await box.press('Tab');await expect(page.getByRole('button',{name:'Open Models settings',exact:true})).toBeFocused(); // options are not separate Tab stops
  await box.focus();await box.fill('no-matching-model');await expect(list.getByRole('option')).toHaveCount(0);await expect(box).not.toHaveAttribute('aria-activedescendant',/.+/);await box.press('Enter');await expect(list).toBeVisible();
  await page.getByRole('button',{name:'Clear model search',exact:true}).click();await expect(box).toBeFocused();expect(await list.getByRole('option').evaluateAll(nodes=>nodes.map(n=>n.id))).toEqual(ids);
  await box.press('Escape');await expect(trigger).toBeFocused();await expect(trigger).toHaveAttribute('aria-expanded','false');await expect(trigger).not.toHaveAttribute('aria-controls',/.+/);
  await trigger.click();await expect(list).toHaveAttribute('id',listId);await expect(box).toBeFocused();await box.press('Escape');await expect(input).toHaveValue('Accessible model draft Ω');
  expect((await(await request.get(`${env.origin}/api/sessions/${id}/model`)).json()).current).toBe(before.current);
 }finally{await env.close();}
});

test('Refreshing and switching models remove stale descendants and cannot activate hidden rows',async({page},info)=>{
 const env=await journeyEnvironment(info);let releaseGet,releasePatch;
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Loading model draft');
  const trigger=page.getByRole('button',{name:'Open model picker',exact:true}),box=page.getByRole('combobox',{name:'Search models',exact:true}),list=page.getByRole('listbox',{name:'Models',exact:true});
  await trigger.click();await expect(list).toHaveAttribute('aria-busy','false');const currentId=await list.getByRole('option',{selected:true}).getAttribute('id');await box.press('Escape');
  const getGate=new Promise(resolve=>{releaseGet=resolve;}),patchGate=new Promise(resolve=>{releasePatch=resolve;});let patches=0;
  await page.route('**/api/sessions/*/model',async route=>{
   if(route.request().method()==='GET'){await getGate;return route.continue();}
   if(route.request().method()==='PATCH'){patches++;await patchGate;return route.continue();}
   return route.continue();
  });
  await trigger.click();await expect(list).toHaveAttribute('aria-busy','true');await expect(list.getByRole('option')).toHaveCount(0);await expect(box).toBeFocused();await expect(box).not.toHaveAttribute('aria-activedescendant',/.+/);
  await box.press('ArrowDown');await box.press('Enter');await expect(list).toHaveAttribute('aria-busy','true');expect(patches).toBe(0);
  releaseGet();await expect(list).toHaveAttribute('aria-busy','false');await expect(box).toHaveAttribute('aria-activedescendant',currentId);
  await box.press('ArrowDown');const targetId=await box.getAttribute('aria-activedescendant');expect(targetId).not.toBe(currentId);
  const options=list.getByRole('option'),index=await options.evaluateAll((nodes,id)=>nodes.findIndex(n=>n.id===id),targetId);expect(index).toBeGreaterThanOrEqual(0);await options.nth(index).click();
  await expect.poll(()=>patches).toBe(1);await expect(box).toBeFocused();await expect(box).not.toHaveAttribute('aria-activedescendant',/.+/);
  expect(await list.getByRole('option').evaluateAll(nodes=>nodes.every(n=>n.disabled&&n.getAttribute('aria-disabled')==='true'))).toBe(true);
  await box.press('ArrowDown');await box.press('Enter');expect(patches).toBe(1);releasePatch();
  // The journey fixture lists bootstrap without credentials: rejection keeps the picker usable.
  await expect(page.getByRole('alert')).toContainText('unavailable or lacks credentials');await expect(box).toBeFocused();await expect(box).toHaveAttribute('aria-activedescendant',currentId);await expect(list.getByRole('option',{selected:true})).toBeEnabled();
  await list.getByRole('option',{selected:true}).click();await expect.poll(()=>patches).toBe(2);await expect(list).toHaveCount(0);await expect(trigger).toBeFocused();await expect(input).toHaveValue('Loading model draft');
 }finally{releaseGet?.();releasePatch?.();await env.close();}
});

for(const owner of ['composer','Settings'])test(`Accepted model response does not steal focus from ${owner}`,async({page},info)=>{
 const env=await journeyEnvironment(info);let release;
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Keep response focus');
  await page.getByRole('button',{name:'Open model picker',exact:true}).click();const list=page.getByRole('listbox',{name:'Models',exact:true});await expect(list).toHaveAttribute('aria-busy','false');
  const gate=new Promise(resolve=>{release=resolve;});let pending=false;
  await page.route('**/api/sessions/*/model',async route=>{if(route.request().method()==='PATCH'){pending=true;await gate;}await route.continue();});
  await list.getByRole('option',{selected:true}).click();await expect.poll(()=>pending).toBe(true);
  if(owner==='Settings'){await page.keyboard.press('Control+,');await expect(page.getByRole('dialog')).toBeVisible();}
  else await input.focus();
  const focused=await page.evaluateHandle(()=>document.activeElement);release();await expect(list).toHaveCount(0);
  await expect.poll(()=>focused.evaluate(element=>document.activeElement===element)).toBe(true);
  if(owner==='Settings'){await expect(page.getByRole('dialog')).toBeVisible();await page.keyboard.press('Escape');}
  else await expect(input).toBeFocused();
  await expect(input).toHaveValue('Keep response focus');
 }finally{release?.();await env.close();}
});

test('Overflowing model list stays out of Tab order beside read-only thinking',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  const writes=[];
  await page.route('**/api/sessions/*/model',async route=>{
   if(route.request().method()!=='GET'){writes.push(route.request().method());return route.abort();}
   const response=await route.fetch(),state=await response.json();
   const extra=Array.from({length:40},(_,i)=>({label:`ux-local/scroll-${i}`,id:`scroll-${i}`,name:`Scroll model ${i}`,provider:'ux-local',context_window:32000}));
   await route.fulfill({response,json:{...state,supports_thinking:true,thinking_level:'high',model_options:[...state.model_options,...extra]}});
  });
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Long catalogue draft');
  await page.getByRole('button',{name:'Open model picker',exact:true}).click();const panel=page.locator('.compose-model-catalogue'),box=page.getByRole('combobox',{name:'Search models',exact:true}),list=page.getByRole('listbox',{name:'Models',exact:true}),settings=panel.getByRole('button',{name:'Open Models settings',exact:true});
  await expect(list).toHaveAttribute('aria-busy','false');await expect(panel.getByLabel('Thinking level (read-only)',{exact:true})).toBeDisabled();await expect.poll(()=>list.evaluate(e=>e.scrollHeight>e.clientHeight)).toBe(true);
  await expect(box).toBeFocused();await box.press('ArrowDown');await expect(box).toBeFocused();await box.press('Tab');await expect(settings).toBeFocused();await settings.press('Shift+Tab');await expect(box).toBeFocused();
  await box.fill('no-such-model');await expect(list.getByRole('option')).toHaveCount(0);await expect(box).not.toHaveAttribute('aria-activedescendant',/.+/);await box.press('Tab');await expect(panel.getByRole('button',{name:'Clear model search',exact:true})).toBeFocused();await page.keyboard.press('Tab');await expect(settings).toBeFocused();
  await panel.getByRole('button',{name:'Clear model search',exact:true}).click();await expect(box).toBeFocused();await expect(list.getByRole('option')).toHaveCount(43);await box.press('Tab');await expect(settings).toBeFocused();await page.keyboard.press('Escape');await expect(input).toHaveValue('Long catalogue draft');expect(writes).toEqual([]);
 }finally{await env.close();}
});
