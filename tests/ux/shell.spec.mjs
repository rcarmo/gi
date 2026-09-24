import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function fixture(page,request,info,id){
 const scenario=loadCorpus().find(row=>row.id===id);expect(scenario).toBeTruthy();await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const token=`shell-${id.slice(-3)}-${info.project.name}-${Date.now()}`;
 const created=await request.post('/api/sessions',{data:{agent_id:token,title:token}});expect(created.status()).toBe(201);const session=await created.json();
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},session.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();const text='shell draft\nkeep native media';await input.fill(text);
 const bytes=Array.from(Buffer.from('exact shell attachment bytes'));await page.locator('.compose-box input[type=file]').setInputFiles({name:'shell-draft.txt',mimeType:'text/plain',buffer:Buffer.from(bytes)});
 const stored=()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:Array.from(new Uint8Array(f.bytes))}))}}:null);};tx.onerror=()=>reject(tx.error);});
 },session.id);
 await expect.poll(stored).toMatchObject({draft:{text,media:[{name:'shell-draft.txt',bytes}]},pending:[]});const before=await stored();
 const unchanged=async()=>{await expect(input).toHaveValue(text);await expect(page.locator('.compose-file-pill[title="shell-draft.txt"]')).toHaveCount(1);expect(await stored()).toEqual(before);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(session.id);expect((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns||[]).toEqual([]);};
 const menu=page.locator('.timeline-menu-dropdown'),hamburger=page.getByTestId('hamburger');
 const workspace=async(show)=>{await hamburger.click();await page.getByRole('menuitem',{name:show?'Show workspace':'Hide workspace',exact:true}).click();await expect(menu).toHaveCount(0);if(show)await expect(page.locator('.app-shell')).not.toHaveClass(/workspace-collapsed/);else await expect(page.locator('.app-shell')).toHaveClass(/workspace-collapsed/);};
 return{session,token,input,unchanged,menu,hamburger,workspace};
}

test('@ux-shell-002 Menu hidden toggle stores its native setting and dispatches one change event',async({page,request},info)=>{
 const f=await fixture(page,request,info,'@ux-shell-002');const path=`.hidden-${f.token}.txt`;
 const write=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'hidden native shell file'}}});expect(write.ok()).toBe(true);expect((await write.json()).error).toBeFalsy();
 await page.evaluate(()=>{window.__hiddenEvents=[];window.addEventListener('piclaw:toggle-hidden-files',e=>window.__hiddenEvents.push(e.detail));});
 await f.workspace(true);const row=page.locator(`.workspace-row[data-path="${path}"]`);await expect(row).toHaveCount(0);
 for(const [show,method]of [[true,'pointer'],[false,'keyboard'],[true,'keyboard']]){
  await f.hamburger.click();const item=page.getByRole('menuitem',{name:show?'Show hidden files':'Hide hidden files',exact:true});await expect(item).toBeEnabled();
  const before=await page.evaluate(()=>window.__hiddenEvents.length);
  const tree=page.waitForResponse(r=>{const u=new URL(r.url());return u.pathname==='/api/workspace/tree'&&u.searchParams.get('show_hidden')===String(show)&&u.searchParams.get('path')==='.';});
  if(method==='pointer')await item.click();else{await item.focus();await page.keyboard.press('Enter');}
  expect((await tree).status()).toBe(200);await expect(f.menu).toHaveCount(0);
  expect(await page.evaluate(()=>localStorage.getItem('workspaceShowHidden'))).toBe(String(show));
  expect(await page.evaluate(()=>window.__hiddenEvents.slice(-1))).toEqual([{showHidden:show}]);expect(await page.evaluate(()=>window.__hiddenEvents.length)).toBe(before+1);
  if(show)await expect(row).toBeVisible();else await expect(row).toHaveCount(0);await f.unchanged();
 }
 await page.reload();await f.unchanged();await f.workspace(true);await expect(row).toBeVisible();await f.hamburger.click();await expect(page.getByRole('menuitem',{name:'Hide hidden files',exact:true})).toBeEnabled();await page.keyboard.press('Escape');
});

test('@ux-shell-003 Hidden workspace disables every workspace mutation and visibility action',async({page,request},info)=>{
 const f=await fixture(page,request,info,'@ux-shell-003');
 await page.evaluate(()=>{window.__workspaceActions=[];for(const type of ['piclaw:workspace-action','piclaw:toggle-hidden-files'])window.addEventListener(type,e=>window.__workspaceActions.push({type,detail:e.detail}));});
 let writes=0;page.on('request',r=>{if(!['GET','HEAD'].includes(r.method())&&new URL(r.url()).pathname.startsWith('/api/workspace'))writes++;});
 const check=async()=>{
  const setting=await page.evaluate(()=>localStorage.getItem('workspaceShowHidden'));await f.hamburger.click();
  for(const name of ['New file','Refresh tree','Reindex workspace','Show hidden files']){
   const item=page.getByRole('menuitem',{name,exact:true});await expect(item).toBeDisabled();const box=await item.boundingBox();expect(box).toBeTruthy();await page.mouse.click(box.x+box.width/2,box.y+box.height/2);await expect(f.menu).toBeVisible();
   await item.evaluate(el=>el.focus());await expect(item).not.toBeFocused();
  }
  await f.hamburger.focus();await page.keyboard.press('Tab');const disabledNames=['New file','Refresh tree','Reindex workspace','Show hidden files'];
  for(let i=0;i<6;i++){expect(await page.evaluate(()=>document.activeElement?.textContent?.trim())).not.toMatch(new RegExp(`^(${disabledNames.join('|')})$`));await page.keyboard.press('Tab');}
  expect(await page.evaluate(()=>window.__workspaceActions)).toEqual([]);expect(await page.evaluate(()=>localStorage.getItem('workspaceShowHidden'))).toBe(setting);expect(writes).toBe(0);await page.keyboard.press('Escape');await f.unchanged();
 };
 await check();await f.workspace(true);await f.workspace(false);await check();
});

test('@ux-shell-005 Compose wrapper uses the available native chat-column width',async({page,request},info)=>{
 const f=await fixture(page,request,info,'@ux-shell-005');
 const geometry=()=>page.locator('.compose-box').evaluate(el=>{const column=el.closest('.container'),a=el.getBoundingClientRect(),b=column.getBoundingClientRect(),s=getComputedStyle(column),input=el.querySelector('.compose-input-wrapper').getBoundingClientRect(),own=getComputedStyle(el);return{width:a.width,expected:b.width-parseFloat(s.paddingLeft)-parseFloat(s.paddingRight)-parseFloat(s.borderLeftWidth)-parseFloat(s.borderRightWidth),left:a.left,expectedLeft:b.left+parseFloat(s.paddingLeft)+parseFloat(s.borderLeftWidth),inputWidth:input.width,expectedInput:a.width-parseFloat(own.paddingLeft)-parseFloat(own.paddingRight)-parseFloat(own.borderLeftWidth)-parseFloat(own.borderRightWidth),overflow:document.documentElement.scrollWidth-innerWidth};});
 const check=async()=>{await expect.poll(async()=>{const g=await geometry();return Math.abs(g.width-g.expected)}).toBeLessThanOrEqual(1);const g=await geometry();expect(Math.abs(g.left-g.expectedLeft)).toBeLessThanOrEqual(1);expect(Math.abs(g.inputWidth-g.expectedInput)).toBeLessThanOrEqual(1);expect(g.overflow).toBeLessThanOrEqual(1);await f.unchanged();};
 await check();await f.workspace(true);await check();await f.workspace(false);await check();
 const original=page.viewportSize();await page.setViewportSize({width:original.width+47,height:original.height-53});await check();await page.setViewportSize(original);await check();await page.reload();await check();
});
