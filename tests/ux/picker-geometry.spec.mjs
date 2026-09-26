import {test,expect} from '@playwright/test';
import reference from './fixtures/picker-geometry-reference.json' with {type:'json'};
import {journeyEnvironment} from './support/journey-environment.mjs';

const rules=reference.rules;
async function assertGeometry(page,kind){
 const g=await page.evaluate(()=>{
  const popup=document.querySelector('.compose-model-popup'),anchor=document.querySelector('.compose-input-main'),compose=document.querySelector('.compose-box');
  const rect=e=>{const r=e.getBoundingClientRect();return {x:r.x,y:r.y,w:r.width,h:r.height,bottom:r.bottom,right:r.right};};
  return {p:rect(popup),a:rect(anchor),c:rect(compose),padding:[getComputedStyle(compose).paddingLeft,getComputedStyle(compose).paddingRight],position:getComputedStyle(popup).position,overflow:popup.scrollWidth-popup.clientWidth,vw:innerWidth,vh:innerHeight};
 });
 expect(g.padding).toEqual([`${rules.composerInlinePadding}px`,`${rules.composerInlinePadding}px`]);
 if(g.vw<=rules.mobileMaxWidth){
  expect(g.position).toBe('fixed');expect(Math.abs(g.p.x-rules.mobileInset)).toBeLessThanOrEqual(1);expect(Math.abs(g.p.y-rules.mobileInset)).toBeLessThanOrEqual(1);expect(Math.abs(g.p.w-(g.vw-2*rules.mobileInset))).toBeLessThanOrEqual(1);expect(Math.abs(g.p.h-(g.vh-2*rules.mobileInset))).toBeLessThanOrEqual(1);
 }else{
  expect(g.position).toBe('absolute');expect(Math.abs(g.p.x-g.a.x)).toBeLessThanOrEqual(1);expect(Math.abs(g.p.bottom-(g.a.y-rules.desktopGap))).toBeLessThanOrEqual(1);
  const expected=kind==='session'?g.a.w:Math.min(rules.modelMaxWidth,g.vw-rules.modelViewportGutter,g.a.w);expect(Math.abs(g.p.w-expected)).toBeLessThanOrEqual(1);
 }
 expect(g.p.x).toBeGreaterThanOrEqual(0);expect(g.p.right).toBeLessThanOrEqual(g.vw+1);expect(g.overflow).toBeLessThanOrEqual(1);
}

for(const kind of ['session','model'])test(`Mobile ${kind} picker supports touch dismissal and keeps long lists within viewport`,async({browser},info)=>{
 const env=await journeyEnvironment(info),context=await browser.newContext({viewport:{width:390,height:600},hasTouch:true}),page=await context.newPage(),writes=[];
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
  // Only this overflow fixture adds inventory; the geometry journey below is
  // otherwise empty-store. No model inference or destructive action is used.
  if(kind==='session')for(let i=0;i<18;i++){const r=await page.request.post(env.origin+`/api/sessions/${id}/fork`,{data:{}});expect(r.status()).toBe(201);}
  await page.reload();const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();await input.fill('Touch picker draft');
  await page.route('**/*',r=>{if(!['GET','HEAD','OPTIONS'].includes(r.request().method())){writes.push(r.request().url());return r.abort();}return r.continue();});
  const trigger=kind==='session'?page.getByRole('button',{name:/Manage sessions for/}).last():page.locator('.compose-model-hint-btn');await trigger.tap();await expect(page.locator('.compose-model-popup')).toBeVisible();await assertGeometry(page,kind);
  if(kind==='session'){const menu=page.locator('.compose-session-popup .compose-model-popup-menu');await expect.poll(()=>menu.evaluate(e=>e.scrollHeight>e.clientHeight)).toBe(true);expect(await menu.evaluate(e=>getComputedStyle(e).overflowY)).toBe('auto');}
  const close=page.getByRole('button',{name:`Close ${kind} picker`,exact:true});await close.tap();await expect(page.locator('.compose-model-popup')).toHaveCount(0);await expect(trigger).toBeFocused();await expect(input).toHaveValue('Touch picker draft');expect(writes).toEqual([]);
 }finally{await context.close();await env.close();}
});

for(const kind of ['session','model'])test(`Pinned Classic ${kind} picker geometry preserves draft, search and dismissal across breakpoint`,async({page},info)=>{
 const env=await journeyEnvironment(info),writes=[];
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();await input.fill('Picker geometry retained draft Ω');const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));expect(id).toBeTruthy();
  // Include the deterministic initial chat but reject every subsequent mutation.
  await page.route('**/*',route=>{if(!['GET','HEAD','OPTIONS'].includes(route.request().method())){writes.push(new URL(route.request().url()).pathname);return route.abort();}return route.continue();});
  const trigger=kind==='session'?page.getByRole('button',{name:/Manage sessions for/}).last():page.locator('.compose-model-hint-btn');
  await trigger.click();const popup=page.locator('.compose-model-popup');await expect(popup).toBeVisible();
  const search=page.getByRole(kind==='session'?'searchbox':'combobox',{name:kind==='session'?'Search sessions':'Search models',exact:true});await expect(search).toBeVisible();
  await assertGeometry(page,kind);await search.fill('zz-no-match');await search.focus();
  for(const width of [639,640,390,820]){await page.setViewportSize({width,height:900});await assertGeometry(page,kind);await expect(search).toHaveValue('zz-no-match');await expect(search).toBeFocused();}
  await page.keyboard.press('Escape');await expect(popup).toHaveCount(0);await expect(trigger).toBeFocused();
  await page.setViewportSize({width:390,height:600});await trigger.click();await assertGeometry(page,kind);const close=page.getByRole('button',{name:`Close ${kind} picker`,exact:true});await expect(close).toBeVisible();const hit=await close.boundingBox();expect(hit.width).toBeGreaterThanOrEqual(44);expect(hit.height).toBeGreaterThanOrEqual(44);await close.focus();await page.keyboard.press('Enter');await expect(popup).toHaveCount(0);await expect(trigger).toBeFocused();
  await expect(input).toHaveValue('Picker geometry retained draft Ω');expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);expect(writes).toEqual([]);
 }finally{await page.close();await env.close();}
});
