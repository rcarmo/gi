import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function setup(page,request,info){
 const token=`tabs-${info.project.name}-${Date.now()}`,paths=[`${token}-a.md`,`${token}-b.csv`,`${token}-c.md`];
 const bodies=['# First native tab\n\n**alpha** and `inline code`','<script>window.__tabUnsafe=true</script>\nplain β text','# Third native tab\n\ncharlie'];
 for(let i=0;i<paths.length;i++){const r=await request.post('/api/tools/execute',{data:{tool:'write',input:{path:paths[i],content:bodies[i]}}});expect(r.ok()).toBe(true);expect((await r.json()).error).toBeFalsy();}
 const created=await request.post('/api/sessions',{data:{agent_id:token,title:token}});const session=await created.json();await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('tabs keep this draft');await page.locator('.compose-box input[type=file]').setInputFiles({name:'tabs-unsent.txt',mimeType:'text/plain',buffer:Buffer.from('tabs retained bytes')});
 const stored=()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:Array.from(new Uint8Array(f.bytes))}))}}:null);};tx.onerror=()=>reject(tx.error);});
 },session.id);
 await expect.poll(stored).toMatchObject({draft:{text:'tabs keep this draft',media:[{name:'tabs-unsent.txt',bytes:Array.from(Buffer.from('tabs retained bytes'))}]},pending:[]});const initial=await stored();
 const preserved=async()=>{const current=await stored();expect({...current.draft,fileRefs:[]}).toEqual({...initial.draft,fileRefs:[]});expect(current.pending).toEqual([]);};
 const tabs=page.locator('.gi-readonly-tabs'),tab=path=>tabs.locator('.tab-item').filter({has:page.locator('.tab-label').filter({hasText:path})}),preview=page.getByRole('region',{name:/^Read-only preview:/});
 const open=async path=>{
  await page.getByTestId('hamburger').click();const show=page.getByRole('menuitem',{name:'Show workspace',exact:true});if(await show.count())await show.click();else await page.keyboard.press('Escape');
  await page.locator(`.workspace-row[data-path="${path}"] .workspace-label-text`).click();await expect(page.locator('.workspace-sidebar .workspace-preview-meta').first()).toContainText(path);
  await page.getByRole('button',{name:'Open read-only tab',exact:true}).click();await expect(tab(path)).toHaveClass(/active/);await expect(preview).toHaveAttribute('aria-label',`Read-only preview: ${path}`);
 };
 const settle=async()=>{if(!await page.locator('.app-shell').evaluate(el=>el.classList.contains('workspace-collapsed'))){await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Hide workspace',exact:true}).click();}};
 const untouched=async()=>{expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(session.id);expect((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns||[]).toEqual([]);};
 return{paths,bodies,session,input,tabs,tab,preview,open,settle,untouched,stored,preserved};
}

test('@ux-shell-007 Closing background native read-only tabs never activates them',async({page,request},info)=>{
 const source=loadCorpus().find(s=>s.id==='@ux-shell-007');await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const f=await setup(page,request,info),[a,b,c]=f.paths;
 for(const path of f.paths)await f.open(path);await f.settle();await expect(f.preview.getByRole('heading',{name:'Third native tab'})).toBeVisible();await f.preserved();
 await expect.poll(async()=>(await f.stored()).draft.fileRefs).toEqual(f.paths);const saved=await f.stored();
 const reads=[];page.on('request',r=>{const u=new URL(r.url());if(u.pathname==='/api/workspace/file')reads.push(u.searchParams.get('path'));});
 const activePaths=()=>page.locator('.gi-readonly-tabs .tab-item.active .tab-label').allTextContents();
 const close=async(path,keyboard)=>{if(keyboard){await f.tab(path).getByRole('button',{name:`Close ${path}`,exact:true}).focus();await page.keyboard.press('Enter');}else await f.tab(path).getByRole('button',{name:`Close ${path}`,exact:true}).click();await expect(f.tab(path)).toHaveCount(0);};
 await close(a,false);await expect.poll(activePaths).toEqual([c]);await expect(f.preview.getByRole('heading',{name:'Third native tab'})).toBeVisible();expect(reads).toEqual([]);
 await close(b,true);await expect.poll(activePaths).toEqual([c]);expect(reads).toEqual([]);await f.untouched();expect(await f.stored()).toEqual(saved);
 await page.screenshot({path:info.outputPath('readonly-workspace-tab.png')});
 await f.tab(c).click({button:'right'});await expect(page.locator('.tab-context-menu button')).toHaveText(['Close','Close Others','Close All','Pin']);await page.keyboard.press('Escape');await expect(page.locator('.tab-context-menu')).toHaveCount(0);
 await f.preview.getByRole('button',{name:'Close preview',exact:true}).click();await expect(f.tabs).toHaveCount(0);await expect(f.input).toBeVisible();await expect(f.input).toBeFocused();await expect(f.input).toHaveValue('tabs keep this draft');await expect(page.locator('.compose-file-pill[title="tabs-unsent.txt"]')).toHaveCount(1);await f.untouched();
 await page.reload();await expect(f.input).toHaveValue('tabs keep this draft');await expect(page.locator('.compose-file-pill[title="tabs-unsent.txt"]')).toHaveCount(1);expect(await f.stored()).toEqual(saved);
});

test('Read-only tab keyboard navigation activates one preview and restores context-menu focus',async({page,request},info)=>{
 const f=await setup(page,request,info),[a,b,c]=f.paths;for(const path of f.paths)await f.open(path);await f.settle();
 const current=f.tab(c);await expect(current).toHaveAttribute('tabindex','0');await expect(f.tab(a)).toHaveAttribute('tabindex','-1');await current.focus();await expect(current).toBeFocused();
 const reads=[];page.on('request',r=>{const u=new URL(r.url());if(u.pathname==='/api/workspace/file')reads.push(u.searchParams.get('path'));});
 for(const [key,path]of [['Home',a],['ArrowRight',b],['End',c],['ArrowRight',a],['ArrowLeft',c]]){
  const before=reads.length;await page.keyboard.press(key);await expect(f.tab(path)).toBeFocused();await expect(f.tab(path)).toHaveClass(/active/);await expect(f.preview).toHaveAttribute('aria-label',`Read-only preview: ${path}`);await expect.poll(()=>reads.slice(before)).toEqual([path]);
 }
 await page.keyboard.press('Shift+F10');const menu=page.locator('.tab-context-menu');await expect(menu).toBeVisible();await expect(menu.getByRole('button',{name:'Close',exact:true})).toBeFocused();await page.keyboard.press('Escape');await expect(menu).toHaveCount(0);await expect(f.tab(c)).toBeFocused();await f.preserved();await f.untouched();
});

test('Read-only preview can return to conversation and reopen retained tabs without re-reading',async({page,request},info)=>{
 const f=await setup(page,request,info);await f.open(f.paths[0]);await f.settle();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();
 const saved=await f.stored();let reads=0;page.on('request',r=>{if(new URL(r.url()).pathname==='/api/workspace/file')reads++;});
 await page.getByRole('button',{name:'Return to conversation',exact:true}).click();await expect(f.preview).toBeHidden();await expect(f.input).toBeFocused();await expect(f.input).toHaveValue('tabs keep this draft');expect(await f.stored()).toEqual(saved);
 await page.getByRole('button',{name:'Show read-only tabs',exact:true}).click();await expect(f.preview).toBeVisible();await expect(f.tab(f.paths[0])).toHaveClass(/active/);await expect(f.tab(f.paths[0])).toBeFocused();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();expect(reads).toBe(0);
 await page.getByRole('button',{name:'Return to conversation',exact:true}).click();await page.keyboard.press('Control+Tab');await expect(f.preview).toBeHidden();await expect(f.input).toBeFocused();await f.untouched();await f.preserved();
});

test('Hidden pending preview cannot take conversation or Settings focus when its read completes',async({page,request},info)=>{
 const f=await setup(page,request,info);let release=()=>{};const gate=new Promise(r=>release=r);let held=false;
 await page.route('**/api/workspace/file?*',async route=>{if(new URL(route.request().url()).searchParams.get('path')!==f.paths[0])return route.continue();const response=await route.fetch();held=true;await gate;try{await route.fulfill({response});}catch{}});
 try{
  // Open via the native tree, holding both sidebar preview and tab reads.
  await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();await page.locator(`.workspace-row[data-path="${f.paths[0]}"] .workspace-label-text`).click();await expect.poll(()=>held).toBe(true);
  // The open-tab action appears after sidebar content, so release that read,
  // then separately hold a Refresh of the mounted tab.
  release();await page.unrouteAll({behavior:'wait'});await page.getByRole('button',{name:'Open read-only tab',exact:true}).click();await f.settle();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();
  let finish;const pending=new Promise(r=>finish=r);release=finish;held=false;
  await page.route('**/api/workspace/file?*',async route=>{const response=await route.fetch();held=true;await pending;await route.fulfill({response});});
  await f.preview.getByRole('button',{name:'Refresh preview',exact:true}).click();await expect.poll(()=>held).toBe(true);await page.getByRole('button',{name:'Return to conversation',exact:true}).click();await expect(f.input).toBeFocused();await f.input.press('End');await page.keyboard.type(' newer');
  await page.keyboard.press('Control+,');const close=page.getByRole('button',{name:'Close settings',exact:true});await expect(close).toBeFocused();release();await page.unrouteAll({behavior:'wait'});await expect(close).toBeFocused();await expect(f.preview).toBeHidden();await page.keyboard.press('Escape');await expect(f.input).toBeFocused();await expect(f.input).toHaveValue('tabs keep this draft newer');
  await page.getByRole('button',{name:'Show read-only tabs',exact:true}).click();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();await expect(f.tab(f.paths[0])).toBeFocused();await f.untouched();
 }finally{release();await page.unrouteAll({behavior:'wait'});}
});

test('Returning to conversation clears an open tab menu without replay on reopen',async({page,request},info)=>{
 const f=await setup(page,request,info);await f.open(f.paths[0]);await f.settle();await f.tab(f.paths[0]).focus();await page.keyboard.press('Shift+F10');const menu=page.locator('.tab-context-menu');await expect(menu).toBeVisible();
 // Native focused-button keyboard activation, without relying on pointer
 // outside-dismiss to clear a retained pane's local context state.
 const back=page.getByRole('button',{name:'Return to conversation',exact:true});await back.focus();await back.press('Enter');await expect(f.preview).toBeHidden();await expect(menu).toHaveCount(0);await expect(f.input).toBeFocused();await page.getByRole('button',{name:'Show read-only tabs',exact:true}).click();await expect(f.preview).toBeVisible();await expect(menu).toHaveCount(0);await expect(f.tab(f.paths[0])).toBeFocused();await f.preserved();await f.untouched();
});

test('Keyboard context close focuses the surviving active tab without duplicate reads',async({page,request},info)=>{
 const f=await setup(page,request,info),[a,b]=f.paths;await f.open(a);await f.open(b);await f.settle();await f.tab(b).focus();await page.keyboard.press('Shift+F10');const menu=page.locator('.tab-context-menu');await expect(menu.getByRole('button',{name:'Close',exact:true})).toBeFocused();await page.keyboard.press('Enter');await expect(f.tab(b)).toHaveCount(0);await expect(f.tab(a)).toBeFocused();await expect(f.tab(a)).toHaveClass(/active/);await f.preserved();await f.untouched();
});

test.describe('Read-only workspace trusted touch transitions',()=>{
 test.use({hasTouch:true,viewport:{width:390,height:844}});
 test('Touch opens previews, closes a background tab and returns to retained conversation',async({page,request},info)=>{
  const f=await setup(page,request,info),[a,b]=f.paths;
  const openTouch=async path=>{
   await page.getByTestId('hamburger').tap();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).tap();await page.locator(`.workspace-row[data-path="${path}"] .workspace-label-text`).tap();await expect(page.locator('.workspace-sidebar .workspace-preview-meta').first()).toContainText(path);await page.getByRole('button',{name:'Open read-only tab',exact:true}).tap();await expect(f.tab(path)).toHaveClass(/active/);
  };
  await openTouch(a);await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();await openTouch(b);await expect(f.preview.locator('pre code')).toHaveText(f.bodies[1]);
  const saved=await f.stored(),writes=[],reads=[];page.on('request',r=>{const u=new URL(r.url());if(!['GET','HEAD','OPTIONS'].includes(r.method()))writes.push(u.pathname);if(u.pathname==='/api/workspace/file')reads.push(u.searchParams.get('path'));});
  const close=f.tab(a).getByRole('button',{name:`Close ${a}`,exact:true});const target=await close.boundingBox();expect(target.width).toBeGreaterThanOrEqual(44);expect(target.height).toBeGreaterThanOrEqual(44);await close.tap();await expect(f.tab(a)).toHaveCount(0);await expect(f.tab(b)).toHaveClass(/active/);expect(reads).toEqual([]);
  await page.getByRole('button',{name:'Return to conversation',exact:true}).tap();await expect(f.preview).toBeHidden();await expect(f.input).toBeFocused();await expect(f.input).toHaveValue('tabs keep this draft');expect(await f.stored()).toEqual(saved);
  await page.getByRole('button',{name:'Show read-only tabs',exact:true}).tap();await expect(f.preview).toBeVisible();await expect(f.tab(b)).toBeFocused();await f.preview.getByRole('button',{name:'Close preview',exact:true}).tap();await expect(f.tabs).toHaveCount(0);await expect(f.input).toBeFocused();expect(writes).toEqual([]);await f.untouched();
 });
});

test('Gi native preview tab lifecycle rejects late reads and exposes real retry without editing',async({page,request},info)=>{
 const f=await setup(page,request,info),[a,b,c]=f.paths;
 await f.open(a);await f.settle();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();await f.open(b);await f.settle();await expect(f.preview.locator('pre code')).toHaveText(f.bodies[1]);expect(await page.evaluate(()=>window.__tabUnsafe)).toBeUndefined();
 await f.tab(a).click();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();
 let release,held=false,delivered;const gate=new Promise(r=>release=r),done=new Promise(r=>delivered=r);
 const pattern='**/api/workspace/file?*';await page.route(pattern,async route=>{if(new URL(route.request().url()).searchParams.get('path')!==b||held)return route.continue();const response=await route.fetch();expect(response.status()).toBe(200);held=true;await gate;await route.fulfill({response});delivered();});
 try{
  await f.tab(b).click();await expect(f.preview.getByRole('status')).toHaveText('Loading preview…');await expect.poll(()=>held).toBe(true);
  await f.tab(b).getByRole('button',{name:`Close ${b}`,exact:true}).click();await expect(f.tab(a)).toHaveClass(/active/);await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();release();await done;await page.unroute(pattern);await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();await expect(f.preview.locator('pre code')).toHaveCount(0);
 }finally{release();}
 // Delete only the isolated source fixture; tab Refresh receives the native404.
 const root=(await(await request.get('/api/runtime/config')).json()).workspace_root;const quote=s=>"'"+s.replaceAll("'","'\\''")+"'";
 const removed=await request.post('/api/tools/execute',{data:{tool:'shell',input:{command:`rm -- ${quote(root+'/'+a)}`}}});expect((await removed.json()).error).toBeFalsy();await f.preview.getByRole('button',{name:'Refresh preview'}).click();await expect(f.preview.getByRole('alert')).toBeVisible();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toHaveCount(0);
 const restored=await request.post('/api/tools/execute',{data:{tool:'write',input:{path:a,content:'# Restored native tab'}}});expect((await restored.json()).error).toBeFalsy();await f.preview.getByRole('button',{name:'Retry preview'}).click();await expect(f.preview.getByRole('heading',{name:'Restored native tab'})).toBeVisible();
 await f.open(c);await f.settle();await expect(f.preview.getByRole('heading',{name:'Third native tab'})).toBeVisible();
 // Keyboard close only owns keys while the preview pane has focus.
 await f.preview.locator('.workspace-preview-render-host').focus();await page.keyboard.press('Control+w');await expect(f.tab(c)).toHaveCount(0);await expect(f.tab(a)).toHaveClass(/active/);await expect(f.preview.getByRole('heading',{name:'Restored native tab'})).toBeVisible();
 await f.preview.getByRole('button',{name:'Close preview'}).click();await expect(f.input).toHaveValue('tabs keep this draft');await expect(page.locator('.compose-file-pill[title="tabs-unsent.txt"]')).toHaveCount(1);await f.untouched();await f.preserved();
});

test('Gi last preview close cannot steal focus from Settings opened in the same frame',async({page,request},info)=>{
 const f=await setup(page,request,info);await f.open(f.paths[0]);await f.settle();await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();
 // Native shortcut events are delivered in one animation-frame callback. This
 // is a focus-race check, not the trusted activation evidence above.
 await f.preview.locator('.workspace-preview-render-host').focus();
 await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>{
  document.dispatchEvent(new KeyboardEvent('keydown',{key:'w',ctrlKey:true,bubbles:true,cancelable:true}));
  window.dispatchEvent(new KeyboardEvent('keydown',{key:',',ctrlKey:true,bubbles:true,cancelable:true}));
  requestAnimationFrame(()=>requestAnimationFrame(resolve));
 })));
 const settings=page.getByRole('dialog',{name:'Gi Settings',exact:true});await expect(settings).toBeVisible();await expect(settings.getByRole('button',{name:'Close settings',exact:true})).toBeFocused();await expect(f.tabs).toHaveCount(0);
 await page.keyboard.press('Escape');await expect(settings).toHaveCount(0);await expect(f.input).toBeFocused();await expect(f.input).toHaveValue('tabs keep this draft');await f.preserved();await f.untouched();
});

test('@ux-workspace-011 Native read-only tab MRU and bulk close preserve pinned tabs',async({page,request},info)=>{
 const source=loadCorpus().find(s=>s.id==='@ux-workspace-011');await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const f=await setup(page,request,info),[a,b,c]=f.paths;
 for(const path of f.paths)await f.open(path);await f.settle();
 const menu=page.locator('.tab-context-menu');
 const action=async(path,label)=>{await f.tab(path).click({button:'right'});await expect(menu).toBeVisible();await menu.getByRole('button',{name:label,exact:true}).click();await expect(menu).toHaveCount(0);};
 await action(a,'Pin');await expect(f.tab(a)).toHaveClass(/pinned/);
 // With a pinned, insertion order is a,b,c; MRU must still choose a, not c.
 await f.tab(a).click();await expect(f.tab(a)).toHaveClass(/active/);await f.tab(b).click();await expect(f.tab(b)).toHaveClass(/active/);
 await f.tab(b).getByRole('button',{name:`Close ${b}`,exact:true}).click();await expect(f.tab(a)).toHaveClass(/active/);await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();
 await f.open(b);await f.settle();
 // CSV would expose a standalone viewer action in the unadapted component.
 await f.tab(b).click({button:'right'});await expect(menu.getByRole('button')).toHaveText(['Close','Close Others','Close All','Pin']);await page.keyboard.press('Escape');await expect(menu).toHaveCount(0);
 await action(c,'Close Others');await expect(f.tab(a)).toHaveCount(1);await expect(f.tab(c)).toHaveCount(1);await expect(f.tab(b)).toHaveCount(0);await expect(f.tab(c)).toHaveClass(/active/);
 await action(c,'Close All');await expect(f.tab(a)).toHaveCount(1);await expect(f.tab(a)).toHaveClass(/active/);await expect(f.tab(c)).toHaveCount(0);await expect(f.preview.getByRole('heading',{name:'First native tab'})).toBeVisible();
 await f.tab(a).click({button:'right'});await expect(menu.getByRole('button')).toHaveText(['Close','Close Others','Close All','Unpin']);await page.keyboard.press('Escape');await expect(menu).toHaveCount(0);await f.preserved();await f.untouched();
 await page.screenshot({path:info.outputPath('pinned-readonly-tab.png')});
 // Pins protect bulk closes, not an explicit individual close.
 await f.tab(a).getByRole('button',{name:`Close ${a}`,exact:true}).click();await expect(f.tabs).toHaveCount(0);await expect(f.input).toBeFocused();await expect(f.input).toHaveValue('tabs keep this draft');await f.preserved();
});

test('Gi read-only context actions clamp to viewport and Settings owns tab shortcuts',async({page,request},info)=>{
 const f=await setup(page,request,info),[a,b,c]=f.paths;for(const path of f.paths)await f.open(path);await f.settle();
 const menu=page.locator('.tab-context-menu');
 const size=page.viewportSize();
 // Coordinate stress complements the trusted right-click path in frozen011.
 await f.tab(c).dispatchEvent('contextmenu',{clientX:size.width-1,clientY:size.height-1});await expect(menu).toBeVisible();const box=await menu.boundingBox();expect(box.x).toBeGreaterThanOrEqual(0);expect(box.y).toBeGreaterThanOrEqual(0);expect(box.x+box.width).toBeLessThanOrEqual(size.width);expect(box.y+box.height).toBeLessThanOrEqual(size.height);
 // Focused keyboard activation also reaches the pin callback once.
 await menu.getByRole('button',{name:'Pin',exact:true}).focus();await page.keyboard.press('Enter');await expect(f.tab(c)).toHaveClass(/pinned/);await expect(menu).toHaveCount(0);
 await f.tab(c).click({button:'right'});await page.keyboard.press('Control+,');const settings=page.getByRole('dialog',{name:'Gi Settings',exact:true});await expect(settings).toBeVisible();
 await page.keyboard.press('Control+Tab');await expect(f.tab(c)).toHaveClass(/active/);await page.keyboard.press('Control+Shift+Tab');await expect(f.tab(c)).toHaveClass(/active/);await page.keyboard.press('Escape');await expect(settings).toHaveCount(0);await page.keyboard.press('Escape');await expect(menu).toHaveCount(0);
 await f.tab(c).click({button:'right'});await menu.getByRole('button',{name:'Unpin',exact:true}).click();await expect(f.tab(c)).not.toHaveClass(/pinned/);
 await f.tab(c).click({button:'right'});await menu.getByRole('button',{name:'Close All',exact:true}).click();await expect(f.tabs).toHaveCount(0);await expect(f.input).toBeFocused();await f.preserved();await f.untouched();
});

test('Gi read-only editor-pane inline code uses the code font across native appearance changes',async({page,request},info)=>{
 const f=await setup(page,request,info);await f.open(f.paths[0]);await f.settle();const code=f.preview.locator('.workspace-preview-text p code');await expect(code).toHaveText('inline code');
 const fonts=()=>code.evaluate(el=>{
  const host=el.closest('.editor-pane'),probe=document.createElement('span');probe.textContent='iiiWWW';probe.style.fontFamily='var(--font-family-mono)';host.append(probe);const expected=getComputedStyle(probe).fontFamily;probe.remove();
  const font=getComputedStyle(el).fontFamily,parent=getComputedStyle(el.parentElement).fontFamily;
  const canvas=document.createElement('canvas'),ctx=canvas.getContext('2d');ctx.font=`20px ${font}`;return{font,parent,expected,narrow:ctx.measureText('iiiiii').width,wide:ctx.measureText('WWWWWW').width};
 });
 const check=async()=>{const value=await fonts();expect(value.font).toBe(value.expected);expect(value.font).not.toBe(value.parent);expect(Math.abs(value.narrow-value.wide)).toBeLessThan(0.2);};
 await check();
 await page.keyboard.press('Control+,');const settings=page.getByRole('dialog',{name:'Gi Settings',exact:true});await expect(settings).toBeVisible();await settings.getByRole('button',{name:'Appearance',exact:true}).click();await settings.getByLabel('Theme preset',{exact:true}).selectOption('monokai');await settings.getByRole('button',{name:'Save appearance',exact:true}).click();await expect(page.locator('html')).toHaveAttribute('data-color-theme','monokai');await page.keyboard.press('Escape');await expect(settings).toHaveCount(0);await check();await f.preserved();await f.untouched();
 await page.reload();await expect(f.input).toHaveValue('tabs keep this draft');await f.open(f.paths[0]);await f.settle();await expect(page.locator('html')).toHaveAttribute('data-color-theme','monokai');await check();await f.preserved();await f.untouched();await page.screenshot({path:info.outputPath('preview-code-font.png')});
});
