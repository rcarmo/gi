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
 let release,held=false;const gate=new Promise(r=>release=r);
 await page.route('**/api/quick-actions',async route=>{const response=await route.fetch();held=true;await gate;await route.fulfill({response});});
 try{
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${child.id}"]`).getByRole('menuitem').click();await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();release();await expect(input).toHaveValue('untouched draft');await expect(palette).toHaveCount(0);
  await page.unrouteAll({behavior:'wait'});
  await page.route('**/api/quick-actions',route=>route.abort('failed'));
  await page.reload();await expect(input).toHaveValue('untouched draft');await open('m');await query.fill('');
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
