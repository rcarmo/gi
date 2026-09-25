import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';
const read=(page,path)=>page.evaluate(async path=>{const r=await fetch(path);if(!r.ok)throw Error(`HTTP ${r.status}`);return r.json();},path);
async function fixture(page,info){
 const env=await journeyEnvironment(info),posts=[];
 await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
 // Wait for actual catalogue readiness, not a timed delay or fabricated commands.
 await expect.poll(()=>page.evaluate(async()=>{const r=await fetch('/api/quick-actions');return r.status;})).toBe(200);
 await input.fill('/');await expect(page.locator('.slash-item')).toHaveCount(2);await input.press('Escape');await input.fill('');
 page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))posts.push(r.postDataJSON());});
 return {env,input,id,posts,popup:page.locator('.slash-autocomplete'),palette:page.locator('.timeline-quick-actions'),turns:async()=>(await read(page,`/api/sessions/${id}/turns`)).turns||[]};
}

test('Composer slash Tab completes without sending; Escape and Shift+Enter leave drafts owned by the composer',async({page},info)=>{
 const f=await fixture(page,info);
 try{
  await page.keyboard.type('/');await expect(f.popup).toBeVisible();await expect(f.palette).toHaveCount(0);await expect(page.locator('.slash-item')).toHaveCount(2);await page.keyboard.press('Escape');await expect(f.popup).toHaveCount(0);await expect(f.input).toHaveValue('/');await expect(f.input).toBeFocused();
  await f.input.fill('/m argument Ω');await expect(f.popup).toBeVisible();await f.input.press('Tab');await expect(f.input).toHaveValue('/model argument Ω');await expect(f.input).toBeFocused();expect(await f.input.evaluate(e=>[e.selectionStart,e.selectionEnd])).toEqual([17,17]);await expect(f.popup).toHaveCount(0);expect(f.posts).toEqual([]);
  await f.input.fill('/m');await expect(f.popup).toBeVisible();await f.input.press('Shift+Enter');await expect(f.input).toHaveValue('/m\n');await expect(f.popup).toHaveCount(0);expect(f.posts).toEqual([]);expect(await f.turns()).toEqual([]);
 }finally{await page.close();await f.env.close();}
});

test('Composer Enter on a bare slash fragment submits the selected native command once',async({page},info)=>{
 const f=await fixture(page,info);
 try{
  await f.input.fill('/m');await expect(page.locator('.slash-item.active .slash-name')).toHaveText('/model');
  const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${f.id}/prompt`));await f.input.press('Enter');const reply=await response;expect(reply.status()).toBe(200);expect(f.posts).toHaveLength(1);expect(f.posts[0]).toMatchObject({prompt:'/model',intent:'prompt'});await expect(f.input).toHaveValue('');await expect(f.popup).toHaveCount(0);await expect(f.palette).toHaveCount(0);expect(await f.turns()).toEqual([]);
 }finally{await page.close();await f.env.close();}
});

test('Slash fragment with arguments stays literal on Enter instead of silently replacing its command',async({page},info)=>{
 const f=await fixture(page,info);
 try{
  const text='/m exact argument Ω';await f.input.fill(text);await expect(f.popup).toBeVisible();const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${f.id}/prompt`));await f.input.press('Enter');const reply=await response;expect(reply.status()).toBe(202);const {turn_id}=await reply.json();expect(f.posts).toHaveLength(1);expect(f.posts[0].prompt).toBe(text);await expect.poll(async()=> (await f.turns()).map(t=>[t.id,t.status])).toEqual([[turn_id,'completed']]);expect((await read(page,`/api/sessions/${f.id}/messages`)).messages.filter(m=>m.role==='user').map(m=>m.content)).toEqual([text]);await expect(f.palette).toHaveCount(0);
 }finally{await page.close();await f.env.close();}
});

for(const kind of ['repeat','consumed','composition','legacy composition'])test(`Composer declines ${kind} Enter while slash completion is open`,async({page},info)=>{
 const f=await fixture(page,info);
 try{
  await f.input.fill('/m');await expect(f.popup).toBeVisible();
  // These browser-event flags cannot all be produced by Playwright's key API.
  // Only event-boundary assertions use synthetic input; positive control is native.
  await f.input.evaluate((node,kind)=>{const event=new KeyboardEvent('keydown',{key:'Enter',code:'Enter',bubbles:true,cancelable:true,repeat:kind==='repeat',isComposing:kind==='composition'});if(kind==='consumed')event.preventDefault();if(kind==='legacy composition')Object.defineProperty(event,'keyCode',{value:229});node.dispatchEvent(event);},kind);
  await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));expect(f.posts).toEqual([]);await expect(f.input).toHaveValue('/m');await expect(f.popup).toBeVisible();await expect(f.palette).toHaveCount(0);expect(await f.turns()).toEqual([]);
  const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${f.id}/prompt`));await f.input.press('Enter');expect((await response).status()).toBe(200);expect(f.posts).toHaveLength(1);
 }finally{await page.close();await f.env.close();}
});

for(const fault of ['unavailable','malformed','delayed'])test(`Composer native catalogue ${fault} never exposes bundled fallback commands`,async({page},info)=>{
 const env=await journeyEnvironment(info);let release=()=>{};
 try{
  const gate=new Promise(r=>release=r);let calls=0;const obsolete=[];page.on('request',r=>{if(new URL(r.url()).pathname==='/agent/commands')obsolete.push(r.url());});
  await page.route('**/api/quick-actions',async route=>{calls++;if(fault==='delayed'){const response=await route.fetch();await gate;return route.fulfill({response});}await route.fulfill({status:fault==='unavailable'?503:200,contentType:'application/json',body:fault==='unavailable'?'{"error":"Catalogue unavailable"}':'{"commands":[{"name":7}]}'});});
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await expect.poll(()=>calls).toBeGreaterThan(0);await input.fill('/');await expect(page.locator('.slash-item')).toHaveCount(0);
  if(fault==='delayed'){
   await page.keyboard.press('Control+,');const close=page.getByRole('button',{name:'Close settings',exact:true});await expect(close).toBeFocused();release();await page.unrouteAll({behavior:'wait'});await expect(close).toBeFocused();await expect(page.locator('.slash-item')).toHaveCount(0);await close.click();await input.focus();await input.press('End');await input.press('m');await expect(page.locator('.slash-name')).toHaveText(['/model']);await input.press('Backspace');
  }else{
   await expect(page.getByRole('status').filter({hasText:'Command suggestions unavailable.'})).toBeVisible();await expect(page.locator('.slash-item')).toHaveCount(0);await page.unrouteAll({behavior:'wait'});await page.reload();await expect(input).toHaveValue('/');
  }
  await expect(page.locator('.slash-item')).toHaveCount(2);await expect(page.locator('.slash-name')).toHaveText(['/model','/compact']);expect(obsolete).toEqual([]);
 }finally{release();await page.unrouteAll({behavior:'wait'});await page.close();await env.close();}
});

test('Delayed command catalogue respects current search mode',async({page},info)=>{
 const env=await journeyEnvironment(info);let release=()=>{};
 try{
  const gate=new Promise(r=>release=r);let held=false;
  await page.route('**/api/quick-actions',async route=>{const response=await route.fetch();held=true;await gate;await route.fulfill({response});});
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('/');await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:'Search',exact:true}).click();await expect(input).toHaveAttribute('placeholder','Search (Enter to run)...');await input.fill('/m');release();await page.unrouteAll({behavior:'wait'});await expect(page.locator('.slash-item')).toHaveCount(0);await expect(input).toHaveValue('/m');
  await page.getByRole('button',{name:'Close search',exact:true}).click();await input.fill('/m');await expect(page.locator('.slash-name')).toHaveText(['/model']);
 }finally{release();await page.unrouteAll({behavior:'wait'});await page.close();await env.close();}
});

test('Quick Actions and composer slash have exclusive focus ownership around Settings and pickers',async({page},info)=>{
 const f=await fixture(page,info);
 try{
  await f.input.fill('retained draft');await f.input.blur();await page.locator('.timeline').click({position:{x:100,y:90}});await page.keyboard.type('/');await expect(f.palette).toBeVisible();const query=page.locator('.timeline-quick-actions-input');await expect(query).toBeFocused();await expect(f.popup).toHaveCount(0);await expect(f.input).toHaveValue('retained draft');await query.press('Escape');await expect(f.palette).toHaveCount(0);expect(f.posts).toEqual([]);
  await f.input.focus();await f.input.fill('/m');await expect(f.popup).toBeVisible();await page.keyboard.press('Control+,');await expect(page.getByRole('dialog',{name:'Gi Settings'})).toBeVisible();const close=page.getByRole('button',{name:'Close settings',exact:true});await expect(close).toBeFocused();await page.keyboard.type('/');await expect(f.palette).toHaveCount(0);expect(f.posts).toEqual([]);await close.click();await expect(f.input).toHaveValue('/m');await f.input.focus();await f.input.press('Escape');
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();const search=page.getByRole('searchbox',{name:'Search sessions',exact:true});await search.fill('/m');await expect(f.palette).toHaveCount(0);await search.press('Escape');await expect(f.input).toHaveValue('/m');expect(f.posts).toEqual([]);expect(await f.turns()).toEqual([]);
 }finally{await page.close();await f.env.close();}
});
