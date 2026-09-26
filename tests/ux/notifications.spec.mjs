import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';
import {authEnvironment,totp} from './support/auth-environment.mjs';
async function runtime(context){await context.addInitScript(()=>{
 window.__notices=[];window.__permissionCalls=0;window.__permission=localStorage.getItem('gi_notifications_v1')==='true'?'granted':'default';window.__visibility='visible';
 Object.defineProperty(document,'visibilityState',{configurable:true,get:()=>window.__visibility});
 window.__setVisibility=value=>{window.__visibility=value;document.dispatchEvent(new Event('visibilitychange'));};
 class LocalNotification{
  static get permission(){return window.__permission;}
  static async requestPermission(){window.__permissionCalls++;window.__permission='granted';return 'granted';}
  constructor(title,options){this.title=title;this.options=options;this.closed=false;window.__notices.push(this);}
  close(){this.closed=true;}
 }
 window.Notification=LocalNotification;
});}
async function completion(request,env,id,label){const r=await request.post(`${env.origin}/api/sessions/${id}/prompt`,{data:{prompt:`notification fixture ${label}`}});expect(r.status()).toBe(202);await expect.poll(async()=>{const x=await(await request.get(`${env.origin}/api/sessions/${id}/messages`)).json();return(x.messages||[]).filter(m=>m.role==='assistant').length;}).toBeGreaterThan(0);}

test('Explicit local opt-in notifies hidden native replies only and never posts Web Push endpoints',async({page,context,request},info)=>{
 const env=await journeyEnvironment(info);await runtime(context);const errors=[],writes=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()!=='GET')writes.push(new URL(r.url()).pathname);});
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Notification draft Ω');const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
  expect(await page.evaluate(()=>window.__permissionCalls)).toBe(0);await page.getByRole('button',{name:'Enable notifications',exact:true}).click();expect(await page.evaluate(()=>window.__permissionCalls)).toBe(1);await expect(page.getByRole('status').filter({hasText:'Matching chat tabs must remain open'})).toBeVisible();
  await completion(request,env,id,'visible');await page.waitForTimeout(150);expect(await page.evaluate(()=>window.__notices.length)).toBe(0);
  await page.evaluate(()=>window.__setVisibility('hidden'));await completion(request,env,id,'hidden');await expect.poll(()=>page.evaluate(()=>window.__notices.length)).toBe(1);
  expect(await page.evaluate(()=>({title:window.__notices[0].title,body:window.__notices[0].options.body}))).toEqual({title:'Gi',body:'An assistant reply is ready.'});
  await page.evaluate(()=>window.__setVisibility('visible'));await expect(input).toHaveValue('Notification draft Ω');await page.getByRole('button',{name:'Disable notifications',exact:true}).click();expect(await page.evaluate(()=>window.__notices[0].closed)).toBe(true);
  await page.evaluate(()=>window.__setVisibility('hidden'));await completion(request,env,id,'disabled');await page.waitForTimeout(150);expect(await page.evaluate(()=>window.__notices.length)).toBe(1);
  expect(writes.filter(path=>path.includes('/push')||path.startsWith('/agent/'))).toEqual([]);expect(errors).toEqual([]);
 }finally{await env.close();}
});

test('Same-browser visible client suppresses delivery; hidden leader produces one notification and cleanup retires it',async({page,context,request},info)=>{
 const env=await journeyEnvironment(info);await runtime(context);let second;
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));await page.getByRole('button',{name:'Enable notifications',exact:true}).click();
  second=await context.newPage();await second.goto(env.origin);await expect(second.locator('.compose-box textarea')).toBeFocused();await expect(second.getByRole('button',{name:'Disable notifications',exact:true})).toBeVisible();
  await page.evaluate(()=>window.__setVisibility('hidden'));await completion(request,env,id,'suppressed');await page.waitForTimeout(150);expect(await page.evaluate(()=>window.__notices.length)+await second.evaluate(()=>window.__notices.length)).toBe(0);
  await second.evaluate(()=>window.__setVisibility('hidden'));await completion(request,env,id,'leader');await expect.poll(async()=>await page.evaluate(()=>window.__notices.length)+await second.evaluate(()=>window.__notices.length)).toBe(1);
  // Production logout dispatches this synchronously before its HTTP request.
  await page.evaluate(()=>window.dispatchEvent(new Event('gi-notification-cleanup')));await expect.poll(()=>second.evaluate(()=>localStorage.getItem('gi_notifications_v1'))).toBe('false');await expect.poll(async()=>await page.evaluate(()=>window.__notices.filter(n=>!n.closed).length)+await second.evaluate(()=>window.__notices.filter(n=>!n.closed).length)).toBe(0);
  await completion(request,env,id,'after-cleanup');await page.waitForTimeout(150);expect(await page.evaluate(()=>window.__notices.length)+await second.evaluate(()=>window.__notices.length)).toBe(1);
 }finally{await second?.close();await env.close();}
});

test('Pending permission, denied permission and unavailable API do not create idle delivery or alter drafts',async({page,context},info)=>{
 const env=await journeyEnvironment(info);await runtime(context);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Permission draft');
  await page.evaluate(()=>{Notification.requestPermission=()=>{window.__permissionCalls++;return new Promise(r=>window.__releasePermission=r);};});
  await page.getByRole('button',{name:'Enable notifications',exact:true}).click();await page.getByRole('button',{name:'Enable notifications',exact:true}).click();expect(await page.evaluate(()=>window.__permissionCalls)).toBe(1);
  await page.evaluate(()=>{window.dispatchEvent(new Event('gi-notification-cleanup'));window.__permission='granted';window.__releasePermission('granted');});await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_notifications_v1'))).toBe('false');await expect(input).toHaveValue('Permission draft');
  await page.evaluate(()=>{window.__permission='denied';window.dispatchEvent(new Event('focus'));});await expect(page.locator('.notification-btn')).toHaveCount(0);expect(await page.evaluate(()=>window.__notices.length)).toBe(0);
  await page.addInitScript(()=>{delete window.Notification;});await page.reload();await expect(page.locator('.notification-btn')).toHaveCount(0);await expect(input).toHaveValue('Permission draft');
 }finally{await env.close();}
});

test('Native browser logout drains notification delivery before revoking authority',async({page,context},info)=>{
 const env=await authEnvironment(page,info);await runtime(context);
 try{
  await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Logout notification draft');
  await page.getByRole('button',{name:'Enable notifications',exact:true}).click();await expect(page.getByRole('button',{name:'Disable notifications',exact:true})).toBeVisible();
  await page.evaluate(()=>{window.__lockHeld=false;window.__lockDone=navigator.locks.request('gi-notification-delivery',()=>new Promise(resolve=>{window.__lockHeld=true;window.__releaseLock=resolve;}));});await expect.poll(()=>page.evaluate(()=>window.__lockHeld)).toBe(true);
  let posts=0;page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname==='/api/auth/session/logout')posts++;});
  await page.keyboard.press('Control+,');const dialog=page.getByRole('dialog');await dialog.getByRole('button',{name:'Authentication',exact:true}).click();await dialog.getByRole('button',{name:'Sign out this browser',exact:true}).click();await dialog.getByRole('button',{name:'Confirm sign out',exact:true}).click();
  await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_notifications_v1'))).toBe('false');expect(posts).toBe(0);await page.evaluate(()=>window.__releaseLock());await page.getByRole('textbox',{name:'Authentication code',exact:true}).waitFor();expect(posts).toBe(1);
  expect(await page.evaluate(()=>Object.keys(localStorage).filter(k=>k.startsWith('piclaw.notifications.presence.')))).toHaveLength(0);expect(await page.evaluate(()=>window.__notices.length)).toBe(0);
 }finally{await page.evaluate(()=>window.__releaseLock?.()).catch(()=>{});await env.close();}
});
