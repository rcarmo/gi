import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function fixture(page,request,info){
 const key=`quick-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:key,title:key}})).json();
 const child=await(await request.post('/api/sessions',{data:{agent_id:key+'-child',title:key+'-child'}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('untouched draft');
 await page.waitForTimeout(150); // Pinned component installs capture listeners in a passive effect.
 const palette=page.locator('.timeline-quick-actions'),query=page.locator('.timeline-quick-actions-input');
 const open=async key=>{await input.blur();await page.locator('.timeline').click({position:{x:160,y:180}});await page.keyboard.type(key);await expect(palette).toBeVisible();await expect(query).toBeFocused();await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));};
 const turns=async()=>((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns??[]);
 return {main,child,input,palette,query,open,turns};
}
async function evidence(info,id){const cases=loadCorpus().filter(c=>c.id===id);await info.attach('gherkin',{body:cases.map(c=>c.steps.join('\n')).join('\n\n'),contentType:'text/plain'});}

test('@ux-original-003 Timeline typing filters ordered native actions and keyboard selection wraps',async({page,request},info)=>{
 await evidence(info,'@ux-original-003');const {input,palette,query,open,turns}=await fixture(page,request,info);
 await open('m');await expect(query).toBeFocused();await expect(query).toHaveValue('m');
 await expect(page.locator('.timeline-quick-actions-item-title')).toContainText(['/model']);
 await query.fill('');await expect(page.locator('.timeline-quick-actions-section')).toHaveText(['Agents','Workspace','Slash commands']);
 const rows=palette.locator('.timeline-quick-actions-item');
 await expect(rows.first()).toHaveClass(/active/);await page.waitForTimeout(150);
 await query.press('ArrowUp');await expect(rows.last()).toHaveClass(/active/);await page.waitForTimeout(100);await query.press('ArrowDown');await expect(rows.first()).toHaveClass(/active/);
 await query.fill('model');await expect(page.locator('.timeline-quick-actions-item.active .timeline-quick-actions-item-title')).toHaveText('/model');
 await query.fill('comp');await expect(page.locator('.timeline-quick-actions-item.active .timeline-quick-actions-item-title')).toHaveText('/compact');
 await query.fill('not-found-anywhere');await expect(page.locator('.timeline-quick-actions-empty')).toBeVisible();
 await page.screenshot({path:info.outputPath('quick-no-results.png')});
 await query.press('Escape');await expect(palette).toHaveCount(0);await expect(input).toHaveValue('untouched draft');expect(await turns()).toEqual([]);
 await open('m');await query.fill('/model');await expect(page.locator('.timeline-quick-actions-item.active .timeline-quick-actions-item-title')).toHaveText('/model');await page.waitForTimeout(150);await query.press('Enter');
 await expect(palette).toHaveCount(0);await expect(input).toHaveValue('/model ');expect(await turns()).toEqual([]);
});

test('@ux-original-006 Escape and outside pointer dismiss without running an action and reopen with a fresh query',async({page,request},info)=>{
 await evidence(info,'@ux-original-006');const {input,palette,query,open,turns}=await fixture(page,request,info);
 await open('m');await query.press('Escape');await expect(palette).toHaveCount(0);
 await open('c');await expect(query).toHaveValue('c');await page.mouse.click(5,page.viewportSize().height-5);await expect(palette).toHaveCount(0);
 await open('m');await expect(query).toHaveValue('m');await query.press('Escape');await expect(input).toHaveValue('untouched draft');expect(await turns()).toEqual([]);
});

test('@ux-original-007 Native slash actions prefill the captured composer with a trailing space, focus and end cursor without sending',async({page,request},info)=>{
 await evidence(info,'@ux-original-007');const {main,child,input,palette,query,open,turns}=await fixture(page,request,info);
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'kept.txt',mimeType:'text/plain',buffer:Buffer.from('kept bytes')});
 await open('m');await query.fill('/model');await expect(palette.locator('.timeline-quick-actions-item-slash')).toHaveCount(1);await expect(page.locator('.timeline-quick-actions-item.active .timeline-quick-actions-item-title')).toHaveText('/model');await page.waitForTimeout(150);await query.press('Enter');
 await expect(palette).toHaveCount(0);await expect(input).toHaveValue('/model ');await expect(input).toBeFocused();expect(await input.evaluate(el=>[el.selectionStart,el.selectionEnd])).toEqual([7,7]);
 await expect(page.locator('.compose-file-pill').filter({hasText:'kept.txt'})).toBeVisible();expect(await turns()).toEqual([]);
 await page.screenshot({path:info.outputPath('quick-command-prefill.png')});
 await input.fill('newer text');await page.reload();await expect(input).toHaveValue('newer text');
 await open('c');await query.fill('/compact');await palette.locator('.timeline-quick-actions-item-slash').click();await expect(input).toHaveValue('/compact ');await expect(input).toBeFocused();expect(await turns()).toEqual([]);
 await page.reload();await expect(input).toHaveValue('/compact ');expect(await turns()).toEqual([]);
});

test('@ux-original-005 Repeated composing consumed whitespace and modifier events do not open Quick Actions',async({page,request},info)=>{
 await evidence(info,'@ux-original-005');const {input,palette,open,query,turns}=await fixture(page,request,info);
 await input.blur();
 for(const options of [{key:'m',repeat:true},{key:'m',isComposing:true},{key:' '},{key:'m',ctrlKey:true},{key:'m',metaKey:true},{key:'m',altKey:true}]){
  await page.locator('.timeline').evaluate((el,init)=>el.dispatchEvent(new KeyboardEvent('keydown',{bubbles:true,cancelable:true,...init})),options);await expect(palette).toHaveCount(0);
 }
 await page.locator('.timeline').evaluate(el=>{const event=new KeyboardEvent('keydown',{key:'m',bubbles:true,cancelable:true});event.preventDefault();el.dispatchEvent(event)});await expect(palette).toHaveCount(0);
 // A currently registered shortcut remains reserved, not swallowed as a query.
 await page.locator('.timeline').click({position:{x:160,y:180}});await page.keyboard.press('[');await expect(palette).toHaveCount(0);
 await open('M');await expect(query).toHaveValue('M');await query.press('Escape');expect(await turns()).toEqual([]);
});

test('Gi Quick Actions exclusions preserve native inputs and interactive controls',async({page,request},info)=>{
 const {input,palette,turns}=await fixture(page,request,info);
 await input.press('End');await input.press('q');await expect(input).toHaveValue('untouched draftq');await expect(palette).toHaveCount(0);
 await page.getByTestId('hamburger').focus();await page.keyboard.press('q');await expect(palette).toHaveCount(0);
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).focus();await page.keyboard.press('q');await expect(palette).toHaveCount(0);
 await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();await page.locator('.workspace-sidebar').evaluate(el=>el.dispatchEvent(new KeyboardEvent('keydown',{key:'q',bubbles:true})));await expect(palette).toHaveCount(0);
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Hide workspace',exact:true}).click();
 await page.getByRole('button',{name:/Manage sessions for/}).last().click();const search=page.getByRole('searchbox',{name:'Search sessions'});await search.fill('query');await expect(palette).toHaveCount(0);await search.press('Escape');
 // DOM-level guard fixtures cover selector classes absent from current Gi UI;
 // they earn no frozen original-004 credit for unimplemented Monaco/panels.
 for(const html of ['<input>','<textarea></textarea>','<select><option>x</option></select>','<div contenteditable="true"></div>','<a href="#">link</a>','<div role="button" tabindex="0">button</div>','<div class="monaco-editor" tabindex="0"></div>','<div class="terminal-pane" tabindex="0"></div>','<div class="post-reply" tabindex="0"></div>']){
  await page.evaluate(html=>{const wrap=document.createElement('div');wrap.id='exclusion-test';wrap.innerHTML=html;document.querySelector('.timeline').append(wrap);wrap.firstElementChild.dispatchEvent(new KeyboardEvent('keydown',{key:'m',bubbles:true}));},html);
  await expect(palette).toHaveCount(0);await page.evaluate(()=>document.getElementById('exclusion-test').remove());
 }
 expect(await turns()).toEqual([]);
});

test('Gi Quick Actions switches native sessions and gates unsupported actions',async({page,request},info)=>{
 const {main,child,input,palette,query,open,turns}=await fixture(page,request,info);
 await open('m');await query.fill('');
 await expect(page.locator('.timeline-quick-actions-item-workspace .timeline-quick-actions-item-title')).toHaveText(['Show workspace','Open explorer']);
 await expect(page.locator('.timeline-quick-actions-item-slash .timeline-quick-actions-item-title')).toHaveText(['/model','/compact']);
 await page.screenshot({path:info.outputPath('quick-actions-groups.png')});
 await query.fill('Open explorer');await expect(page.locator('.timeline-quick-actions-item.active .timeline-quick-actions-item-title')).toHaveText('Open explorer');await page.waitForTimeout(150);await query.press('Enter');await expect(page.locator('.workspace-sidebar')).toBeVisible();await expect(palette).toHaveCount(0);
 // The supplied palette rebinds its capture listener in a passive effect after
 // dismissal. Wait for that effect before the next distinct user interaction.
 await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
 await input.blur();await page.locator('.timeline').evaluate(el=>el.dispatchEvent(new KeyboardEvent('keydown',{key:'m',bubbles:true,cancelable:true})));await expect(palette).toBeVisible();await query.fill('Hide workspace');await expect(page.locator('.timeline-quick-actions-item.active .timeline-quick-actions-item-title')).toHaveText('Hide workspace');await page.waitForTimeout(150);await query.press('Enter');await expect(page.locator('.app-shell')).not.toHaveClass(/workspace-visible/);await expect(page.locator('.workspace-sidebar')).not.toHaveClass(/visible/);
 await open('m');await query.fill(child.title);await expect(palette.locator('.timeline-quick-actions-item-agent')).toHaveCount(1);await page.waitForTimeout(150);await query.press('Enter');
 await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(child.id);await expect(input).toHaveValue('');await input.fill('child unsent');
 await open('m');await query.fill(main.title);await palette.locator('.timeline-quick-actions-item-agent').filter({has:page.locator('.timeline-quick-actions-item-title',{hasText:new RegExp('^@'+main.title+'$')})}).click();await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);await expect(input).toHaveValue('untouched draft');
 expect(await turns()).toEqual([]);
});

test('Gi Quick Actions capability failure stays conservative and delayed responses cannot reopen another session',async({page,request},info)=>{
 const {main,child,input,palette,query,open,turns}=await fixture(page,request,info);
 let release,held=false,deny=false,inFlight=0,denied=0;const gate=new Promise(r=>release=r);
 // Keep one route registered across reload. Replacing routes while held
 // callbacks are draining can let a reload's requests bypass the new handler.
 await page.route('**/api/quick-actions',async route=>{
  if(deny){denied++;await route.abort('failed');return;}
  inFlight++;
  try{const response=await route.fetch();held=true;await gate;await route.fulfill({response});}finally{inFlight--;}
 });
 try{
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${child.id}"]`).getByRole('menuitem').click();await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();release();await expect(input).toHaveValue('untouched draft');await expect(palette).toHaveCount(0);
  await expect.poll(()=>inFlight).toBe(0);deny=true;
  await page.reload();await expect.poll(()=>denied).toBeGreaterThanOrEqual(2);await expect(input).toHaveValue('untouched draft');await open('m');await query.fill('');
  await expect(page.locator('.timeline-quick-actions-item-workspace')).toHaveCount(0);await expect(page.locator('.timeline-quick-actions-item-slash')).toHaveCount(0);await expect(page.locator('.timeline-quick-actions-item-agent')).not.toHaveCount(0);
  await query.press('Escape');expect(await turns()).toEqual([]);
 }finally{release()}
});

test('Gi Quick Actions linked agent tab selects the native destination without moving the original draft',async({page,request,context},info)=>{
 const {main,child,input,palette,query,open,turns}=await fixture(page,request,info);
 await open('m');await query.fill(child.title);await expect(palette.locator('.timeline-quick-actions-item-agent')).toHaveCount(1);await page.waitForTimeout(150);
 const newPage=context.waitForEvent('page');await query.press('Alt+Enter');const linked=await newPage;
 try{await expect(linked.getByRole('textbox',{name:inputName,exact:true})).toBeVisible();expect(new URL(linked.url()).searchParams.get('chat_jid')).toBe(`gi:${child.id}`);await expect.poll(()=>linked.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(child.id);
  await expect(input).toHaveValue('untouched draft');expect(await turns()).toEqual([]);
 }finally{await linked.close()}
});

async function sharedTypingFixture(page,request,info,id){
 const source=loadCorpus('shared').find(s=>s.id===id);expect(source).toBeTruthy();await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const f=await fixture(page,request,info);
 const posted=await request.post(`/api/sessions/${f.main.id}/prompt`,{data:{prompt:`native key guard history ${id}`,model:'test-model'}});expect(posted.status()).toBe(202);const turn=(await posted.json()).turn_id;
 await expect.poll(async()=>(await f.turns()).find(t=>t.id===turn)?.status).toBe('completed');
 await page.reload();await expect(f.input).toHaveValue('untouched draft');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'guard-unsent.txt',mimeType:'text/plain',buffer:Buffer.from('guard native bytes')});
 const history=page.locator('.post-content p').filter({hasText:`native key guard history ${id}`}).first();await expect(history).toBeVisible();
 const state=async()=>({turns:await f.turns(),messages:await(await request.get(`/api/sessions/${f.main.id}/messages`)).json(),model:await(await request.get(`/api/sessions/${f.main.id}/model`)).json()});
 const before=await state();let mutations=0;page.on('request',r=>{if(!['GET','HEAD'].includes(r.method())&&new URL(r.url()).pathname.startsWith('/api/sessions'))mutations++;});
 const frames=()=>page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
 const open=async()=>{await history.click();await page.keyboard.press('m');await expect(f.palette).toHaveCount(1);await expect(f.query).toBeFocused();await expect(f.query).toHaveValue('m');};
 const close=async()=>{await f.query.press('Escape');await expect(f.palette).toHaveCount(0);await frames();};
 const unchanged=async(text='untouched draft')=>{await expect(f.input).toHaveValue(text);await expect(page.locator('.compose-input-main .compose-file-pill[title="guard-unsent.txt"]')).toHaveCount(1);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(f.main.id);expect(await state()).toEqual(before);expect(mutations).toBe(0);};
 // Prove the native palette is loaded and capable of opening before checking
 // that another surface or key prevents it. Never pass on an unready listener.
 await open();await close();await unchanged();
 return{...f,history,open,close,frames,unchanged};
}

for(const [id,surface] of [['@shared-5','composer textarea'],['@shared-6','input or select'],['@shared-11','session or model picker']])test(`${id} Native ${surface} keeps its typing without Quick Actions`,async({page,request},info)=>{
 const f=await sharedTypingFixture(page,request,info,id);let text='untouched draft';
 if(id==='@shared-5'){
  await f.input.focus();await f.input.press('End');await f.input.pressSequentially('é文');await f.frames();
  await expect(f.input).toHaveValue(text+'é文');expect(await f.input.evaluate(el=>[el.selectionStart,el.selectionEnd])).toEqual([text.length+2,text.length+2]);await expect(f.palette).toHaveCount(0);
  await f.input.press('Backspace');await f.input.pressSequentially('ß');text+='éß';await expect(f.input).toHaveValue(text);await expect(f.input).toBeFocused();
 }else if(id==='@shared-6'){
  await page.keyboard.press('Control+,');const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});await dialog.getByRole('button',{name:'Models',exact:true}).click();
  const filter=dialog.getByLabel('Filter models',{exact:true}),select=dialog.getByLabel('Session model',{exact:true});await expect(select).toBeVisible();await expect(filter).toBeFocused();
  await filter.pressSequentially('boot');await expect(filter).toHaveValue('boot');await filter.pressSequentially('strap');await expect(filter).toHaveValue('bootstrap');await expect(filter).toBeFocused();
  await expect(select.locator('option:not([disabled])')).toHaveCount(1);await expect(select.locator('option:not([disabled])')).toHaveText(/test\/bootstrap/);await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/test-model');await expect(dialog.getByRole('button',{name:'Apply model',exact:true})).toBeDisabled();await expect(f.palette).toHaveCount(0);
  await page.keyboard.press('Escape');await expect(dialog).toHaveCount(0);
 }else{
  const trigger=page.getByRole('button',{name:/Manage sessions for/}).last();await trigger.click();const search=page.getByRole('searchbox',{name:'Search sessions',exact:true});await expect(search).toBeFocused();await search.pressSequentially(f.child.id);
  await expect(search).toHaveValue(f.child.id);await expect(search).toBeFocused();const picker=page.getByRole('menu',{name:'Sessions and agents',exact:true});await expect(picker.getByRole('menuitem')).toHaveCount(1);await expect(picker.getByRole('menuitem')).toContainText(`gi:${f.child.id}`);await expect(f.palette).toHaveCount(0);
  await search.press('Escape');await expect(picker).toHaveCount(0);await expect(trigger).toBeFocused();
 }
 await f.frames();await expect(f.palette).toHaveCount(0);await f.unchanged(text);
 await f.open();await f.close();await f.unchanged(text);
 expect((await(await request.get(`/api/sessions/${f.child.id}/turns`)).json()).turns||[]).toEqual([]);
});

test('@shared-12 Consumed modified repeated and composing timeline keys cannot activate Quick Actions',async({page,request},info)=>{
 const f=await sharedTypingFixture(page,request,info,'@shared-12');
 for(const options of [{key:'m',prevented:true},{key:'m',repeat:true},{key:'m',isComposing:true},{key:' '},{key:'m',ctrlKey:true},{key:'m',metaKey:true},{key:'m',altKey:true}]){
  await f.history.click();
  // Native content remains the target. Event metadata fixtures exercise IME,
  // consumed and repeat states without inventing DOM or claiming physical IME.
  const observed=await f.history.evaluate((el,options)=>{const {prevented,...init}=options;const event=new KeyboardEvent('keydown',{bubbles:true,cancelable:true,...init});if(prevented)event.preventDefault();el.dispatchEvent(event);return{key:event.key,repeat:event.repeat,isComposing:event.isComposing,ctrlKey:event.ctrlKey,metaKey:event.metaKey,altKey:event.altKey,defaultPrevented:event.defaultPrevented};},options);
  expect(observed.key).toBe(options.key);for(const key of ['repeat','isComposing','ctrlKey','metaKey','altKey'])expect(observed[key]).toBe(Boolean(options[key]));if(options.prevented)expect(observed.defaultPrevented).toBe(true);
  await f.frames();await expect(f.palette).toHaveCount(0);await f.unchanged();
  // Positive real-key control after *each* rejected event, not just once at
  // setup: the guard must not pass by leaving the palette listener detached.
  await f.open();await f.close();await f.unchanged();
 }
});

for(const [id,surface] of [['@shared-7','button or link'],['@shared-9','workspace sidebar'],['@shared-10','open modal dialog']])test(`${id} Native ${surface} receives printable keys without palette interception`,async({page,request},info)=>{
 const f=await sharedTypingFixture(page,request,info,id);
 const receives=async target=>{
  await target.focus();await expect(target).toBeFocused();
  await target.evaluate(el=>{window.__nativeGuardKeys=[];el.addEventListener('keydown',e=>window.__nativeGuardKeys.push({key:e.key,prevented:e.defaultPrevented,trusted:e.isTrusted}));});
  await target.press('q');await f.frames();await expect(f.palette).toHaveCount(0);await expect(target).toBeFocused();
  expect(await page.evaluate(()=>window.__nativeGuardKeys)).toEqual([{key:'q',prevented:false,trusted:true}]);
 };
 if(id==='@shared-7'){
  // Both controls already exist in the native timeline: no dummy interactive
  // element and no DOM-only palette suppression can satisfy event delivery.
  await receives(page.locator('.post').first().getByRole('button',{name:'Copy message',exact:true}));
  await receives(page.locator('.post .post-time').first());await f.unchanged();
 }else if(id==='@shared-9'){
  await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
  const tree=page.locator('.workspace-tree-list');await expect(tree).toBeVisible();const rows=await tree.locator('.workspace-row').evaluateAll(nodes=>nodes.map(n=>({path:n.dataset.path,selected:n.classList.contains('selected')})));
  await receives(tree);expect(await tree.locator('.workspace-row').evaluateAll(nodes=>nodes.map(n=>({path:n.dataset.path,selected:n.classList.contains('selected')})))).toEqual(rows);
  await page.locator('.workspace-toggle-tab.open').click();
 }else{
  await page.keyboard.press('Control+,');const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});await expect(dialog).toBeVisible();await expect(dialog).toHaveAttribute('aria-modal','true');
  const name=dialog.getByLabel('Assistant display name',{exact:true}),value=await name.inputValue();await name.focus();await name.press('End');await receives(name);await expect(name).toHaveValue(value+'q');
  await page.keyboard.press('Escape');await expect(dialog).toHaveCount(0);await page.keyboard.press('Control+,');await expect(name).toHaveValue(value);await page.keyboard.press('Escape');
 }
 await f.unchanged();await f.open();await f.close();await f.unchanged();
});

test('Gi Settings delivers target keys while native background popups remain suspended',async({page,request},info)=>{
 const f=await sharedTypingFixture(page,request,info,'@shared-10');
 const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});
 for(const background of ['model','session','palette','menu']){
  let popup,open;
  if(background==='model'){popup=page.locator('.compose-model-popup-menu[aria-label="Model picker"]');open=()=>page.getByRole('button',{name:'Open model picker',exact:true}).click();}
  if(background==='session'){popup=page.getByRole('menu',{name:'Sessions and agents',exact:true});open=()=>page.getByRole('button',{name:/Manage sessions for/}).last().click();}
  if(background==='palette'){popup=f.palette;open=f.open;}
  if(background==='menu'){popup=page.locator('.timeline-menu-dropdown');open=()=>page.getByTestId('hamburger').click();}
  await open();await expect(popup).toBeVisible();
  if(background==='model')await expect(popup.locator('.current-model')).toContainText('test/test-model');
  if(background!=='menu')await expect(popup.locator('.active')).toHaveCount(1);
  const active=()=>popup.locator('.active').evaluateAll(nodes=>nodes.map(n=>n.textContent));const before=await active();
  await page.keyboard.press('Control+,');await expect(dialog).toBeVisible();
  const input=dialog.getByLabel('Assistant display name',{exact:true});await input.click();await input.press('End');const original=await input.inputValue();
  await input.evaluate(el=>{window.__modalTargetKeys=[];el.addEventListener('keydown',e=>window.__modalTargetKeys.push({key:e.key,prevented:e.defaultPrevented}));});
  await input.press('q');await input.press('ArrowDown');await input.press('Home');await f.frames();await expect(input).toHaveValue(original+'q');
  expect(await page.evaluate(()=>window.__modalTargetKeys)).toEqual([{key:'q',prevented:false},{key:'ArrowDown',prevented:false},{key:'Home',prevented:false}]);expect(await active()).toEqual(before);
  await input.dispatchEvent('keydown',{key:'Escape',isComposing:true});await expect(dialog).toBeVisible();await expect(input).toBeFocused();expect(await page.evaluate(()=>window.__modalTargetKeys.at(-1))).toEqual({key:'Escape',prevented:false});
  const close=dialog.getByRole('button',{name:'Close settings',exact:true}),last=dialog.getByRole('button',{name:'Reload saved names',exact:true});await close.focus();await page.keyboard.press('Shift+Tab');await expect(last).toBeFocused();await page.keyboard.press('Tab');await expect(close).toBeFocused();expect(await active()).toEqual(before);
  await page.keyboard.press('Escape');await expect(dialog).toHaveCount(0);await expect(popup).toBeVisible();expect(await active()).toEqual(before);
  await page.keyboard.press('Escape');await expect(popup).toHaveCount(0);await f.frames();await f.unchanged();
 }
});

for(const [id,dismissal] of [['@shared-13','Escape'],['@shared-14','outside pointer']])test(`${id} Quick Actions ${dismissal} returns to the native Conversation trigger without changing draft references`,async({page,request},info)=>{
 const f=await sharedTypingFixture(page,request,info,id);
 const filename=`dismiss-${info.project.name}-${id.slice(1)}.txt`;
 const write=await request.post('/api/tools/execute',{data:{tool:'write',input:{path:filename,content:'retained workspace reference'}}});expect((await write.json()).error).toBeFalsy();
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();await page.locator(`.workspace-row[data-path="${filename}"]`).click();await page.locator('.workspace-toggle-tab.open').click();
 const link=page.locator('.post .post-time').first(),messageId=(await link.getAttribute('href')).replace(/^#msg-/,'');await link.click();
 const titles=()=>page.locator('.compose-input-main .compose-file-pill').evaluateAll(nodes=>nodes.map(n=>n.title));
 const labels=[`Message reference: ${messageId}`,filename,'guard-unsent.txt'];await expect.poll(titles).toEqual(labels);
 const stored=()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:Array.from(new Uint8Array(f.bytes))}))}}:null);};tx.onerror=()=>reject(tx.error);});
 },f.main.id);
 await expect.poll(stored).toMatchObject({draft:{text:'untouched draft',fileRefs:[filename],messageRefs:[messageId],media:[{name:'guard-unsent.txt',bytes:Array.from(Buffer.from('guard native bytes'))}]},pending:[]});const before=await stored();
 const conversation=page.getByRole('region',{name:'Conversation',exact:true});await expect(conversation).toBeVisible();await conversation.focus();await expect(conversation).toBeFocused();await conversation.press('m');await expect(f.query).toBeFocused();await expect(f.palette).toHaveCount(1);await expect(f.query).toHaveValue('m');await f.query.fill('no-action-chosen');
 await page.locator('.compose-box').evaluate(el=>{window.__dismissComposerClicks=0;el.addEventListener('click',()=>window.__dismissComposerClicks++);});
 if(dismissal==='Escape')await f.query.press('Escape');else{
  const send=page.getByRole('button',{name:'Send message',exact:true}),box=await send.boundingBox(),point={x:box.x+box.width/2,y:box.y+box.height/2};
  // The supplied overlay is pointer-transparent. The enabled Send control is
  // the actual hit target; dismissal must consume its whole gesture.
  expect(await send.evaluate((el,p)=>el.contains(document.elementFromPoint(p.x,p.y)),point)).toBe(true);await page.mouse.click(point.x,point.y);
 }
 await expect(f.palette).toHaveCount(0);await expect(conversation).toBeFocused();await f.frames();await expect(conversation).toBeFocused();expect(await page.evaluate(()=>window.__dismissComposerClicks)).toBe(0);await f.unchanged();await expect.poll(titles).toEqual(labels);expect(await stored()).toEqual(before);
 // Reopen from the restored trigger with native typing and dismiss again;
 // no stale focus callback or old query may survive the previous occurrence.
 await page.keyboard.press('q');await expect(f.query).toBeFocused();await expect(f.query).toHaveValue('q');await f.query.press('Escape');await expect(conversation).toBeFocused();
 await page.reload();await f.unchanged();await expect.poll(titles).toEqual(labels);expect(await stored()).toEqual(before);expect((await(await request.get(`/api/sessions/${f.child.id}/turns`)).json()).turns||[]).toEqual([]);
});

test.describe('Gi Quick Actions trusted touch dismissal',()=>{
 test.use({hasTouch:true});
 test('Gi palette touch dismissal consumes Send and restores Conversation without stale gesture capture',async({page,request},info)=>{
  const f=await sharedTypingFixture(page,request,info,'@shared-14'),conversation=page.getByRole('region',{name:'Conversation',exact:true});
  await conversation.focus();await conversation.press('m');await expect(f.query).toBeFocused();const send=page.getByRole('button',{name:'Send message',exact:true}),box=await send.boundingBox();
  await page.touchscreen.tap(box.x+box.width/2,box.y+box.height/2);await expect(f.palette).toHaveCount(0);await expect(conversation).toBeFocused();await f.unchanged();await f.input.tap();await expect(f.input).toBeFocused();
  await conversation.focus();await conversation.press('q');await expect(f.query).toBeFocused();await f.query.dispatchEvent('keydown',{key:'Escape',isComposing:true});await expect(f.palette).toBeVisible();await f.query.press('Escape');await expect(conversation).toBeFocused();await f.unchanged();
 });
});
