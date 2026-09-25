import {test,expect} from '@playwright/test';
import {authEnvironment,totp} from '../ux/support/auth-environment.mjs';

test('Native auth policy failure blocks bootstrap until Retry without changing sessions',async({page,request})=>{
 const before=await(await request.get('/api/sessions')).json();
 const policy=await(await request.get('/api/auth/status')).json();
 expect(policy.mode).toBe('single-user');expect(policy.enrolled).toBe(false);
 let release:()=>void;const gate=new Promise<void>(r=>release=r);let calls=0;
 await page.route('**/api/auth/status',async route=>{calls++;await gate;await route.fulfill({status:503,contentType:'application/json',body:'{"error":"unavailable"}'});});
 await page.goto('/');await expect(page.getByRole('status')).toHaveText('Loading sign-in options…');await expect(page.locator('input,textarea')).toHaveCount(0);
 release!();await expect(page.getByRole('alert')).toContainText('Cannot load sign-in options');expect(calls).toBe(1);
 await expect(page.getByRole('button',{name:'Sign in',exact:true})).toHaveCount(0);
 expect(await(await request.get('/api/sessions')).json()).toEqual(before);
 await page.unroute('**/api/auth/status');
 // Capture origin so existing sessions are selected, not an implicit fixture write.
 const sessions=before.sessions||before;
 if(sessions.length)await page.evaluate(id=>localStorage.setItem('gi_session_id',id),sessions[0].id);
 await page.getByRole('button',{name:'Retry',exact:true}).click();
 await expect(page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true})).toBeVisible();
 await expect(page.getByRole('heading',{name:'Sign in to Gi'})).toHaveCount(0);
});

test('Unenrolled browsing cannot acquire browser-owner proof authority',async({page,request})=>{
 const before=await(await request.get('/api/sessions')).json();
 if(before.sessions?.length)await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),before.sessions[0].id);
 await page.goto('/');await expect(page.locator('.compose-box textarea')).toBeVisible();
 const results=await page.evaluate(async()=>{
  const get=await fetch('/api/auth/session/proof');
  const post=await fetch('/api/auth/session/reauth/totp',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({code:'123456'})});
  return Promise.all([get,post].map(async r=>({status:r.status,cache:r.headers.get('cache-control'),body:await r.json()})));
 });
 expect(results).toEqual(Array(2).fill({status:401,cache:'private, no-store',body:{error:'browser owner sign-in required'}}));
 expect(await(await request.get('/api/auth/status')).json()).toMatchObject({enrolled:false});
 await expect(page.locator('.compose-box textarea')).toBeVisible();
});

test('Unconfigured passkey backend is unavailable without changing owner enrollment',async({request})=>{
 const before=await(await request.get('/api/auth/status')).json();
 expect((await request.get('/api/auth/passkeys')).status()).toBe(403);
 expect((await request.post('/api/auth/passkeys/login/start',{data:{}})).status()).toBe(403);
 expect(await(await request.get('/api/auth/status')).json()).toEqual(before);
});

test('Authentication settings explains unavailable enrollment without disturbing the composer',async({page})=>{
 await page.goto('/');const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();await input.fill('Authentication pane keeps draft');
 const writes:string[]=[];page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.startsWith('/api/auth/'))writes.push(r.url())});
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Settings',exact:true}).click();await page.getByRole('button',{name:'Authentication',exact:true}).click();
 await expect(page.getByText('Authentication must be configured first.',{exact:true})).toBeVisible();await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeDisabled();
 await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(input).toHaveValue('Authentication pane keeps draft');expect(writes).toEqual([]);
});

test('Authentication verification restores its replaced control without stealing focus outside the pane',async({page},info)=>{
 test.skip(!process.env.GI_UX_SERVER_BIN,'Requires the isolated auth fixture binary built by make test-ux-auth');
 const env=await authEnvironment(page,info);let release:()=>void=()=>{};
 try{
  await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();
  const composer=page.locator('.compose-box textarea');await expect(composer).toBeVisible();await composer.fill('Authentication focus draft');await page.keyboard.press('Control+,');await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
  const code=page.getByRole('textbox',{name:'Reauthentication code',exact:true}),verify=page.getByRole('button',{name:'Verify code',exact:true});
  await code.fill(String((Number(totp(env.secret))+1)%1000000).padStart(6,'0'));await verify.click();await expect(page.getByRole('alert')).toBeVisible();await expect(verify).toBeFocused();await expect(page.getByText('Authentication verified.',{exact:true})).toHaveCount(0);
  let held=false;const gate=new Promise<void>(r=>release=r);
  await page.route('**/api/auth/session/reauth/totp',async route=>{held=true;await gate;await route.continue();});
  await code.fill(totp(env.secret));await verify.click();await expect.poll(()=>held).toBe(true);
  const close=page.getByRole('button',{name:'Close settings',exact:true});await close.focus();release();await page.unrouteAll({behavior:'wait'});
  await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();await expect(close).toBeFocused();await close.click();await expect(composer).toHaveValue('Authentication focus draft');
 }finally{release();await env.close();}
});

test('Owner passkey inventory rejects account selection and keeps its normal list available',async({page},info)=>{
 test.skip(!process.env.GI_UX_SERVER_BIN,'Requires the isolated auth fixture binary built by make test-ux-auth');
 const env=await authEnvironment(page,info,{passkeys:true});
 try{
  await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.locator('.compose-box textarea')).toBeVisible();
  const query=async path=>page.evaluate(async path=>{const response=await fetch(path);return {status:response.status,body:await response.json()};},path);
  const before=await query('/api/auth/passkeys');expect(before).toEqual({status:200,body:{passkeys:[]}});
  expect(await query('/api/auth/passkeys?account=another')).toEqual({status:400,body:{error:'Passkey inventory does not accept query parameters'}});expect(await query('/api/auth/passkeys')).toEqual(before);
 }finally{await env.close();}
});

test('Settings explains last-key refusal and the accepted fallback after native removal',async({page,context},info)=>{
 test.skip(!process.env.GI_UX_SERVER_BIN,'Requires the isolated auth fixture binary built by make test-ux-auth');
 const env=await authEnvironment(page,info,{passkeys:true});const cdp=await context.newCDPSession(page);
 try{
  await cdp.send('WebAuthn.enable');await cdp.send('WebAuthn.addVirtualAuthenticator',{options:{protocol:'ctap2',transport:'usb',hasResidentKey:true,hasUserVerification:true,isUserVerified:true,automaticPresenceSimulation:true}});
  await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();const composer=page.locator('.compose-box textarea');await expect(composer).toBeVisible();await composer.fill('Removal reason retains draft');
  await page.keyboard.press('Control+,');await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Laptop');await page.getByRole('button',{name:'Add passkey',exact:true}).click();await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();
  const policy=async value=>{await page.getByRole('combobox',{name:'Accepted sign-in methods',exact:true}).selectOption(value);await page.getByRole('button',{name:'Change sign-in policy',exact:true}).click();await page.getByRole('button',{name:'Confirm policy change',exact:true}).click();await expect(page.getByText('Sign-in policy saved. Existing login sessions are unchanged.',{exact:true})).toBeVisible();};
  await policy('passkey-only');await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();
  const remove=async()=>{await page.getByRole('button',{name:'Remove Laptop',exact:true}).click();await page.getByRole('button',{name:'Confirm removal',exact:true}).click();};
  await remove();await expect(page.getByRole('alert')).toContainText('Configured TOTP is not accepted in passkey-only mode.');await expect(page.locator('.gi-passkey-row')).toHaveCount(1);
  await page.getByRole('button',{name:'Cancel removal',exact:true}).click();await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();await policy('either');await remove();
  await expect(page.getByRole('status').filter({hasText:'TOTP remains available for sign-in.'})).toBeVisible();await expect(page.locator('.gi-passkey-row')).toHaveCount(0);
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();await expect(page.getByText('TOTP remains available for sign-in.',{exact:true})).toHaveCount(0);
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(composer).toHaveValue('Removal reason retains draft');
 }finally{await cdp.detach();await env.close();}
});

test('Explicit browser logout returns to sign-in without discarding the composer draft',async({page,context},info)=>{
 test.skip(!process.env.GI_UX_SERVER_BIN,'Requires the isolated auth fixture binary built by make test-ux-auth');const env=await authEnvironment(page,info);
 try{
  await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();const composer=page.locator('.compose-box textarea');await expect(composer).toBeVisible();await composer.fill('Functional logout draft');
  await page.keyboard.press('Control+,');await page.getByRole('button',{name:'Authentication',exact:true}).click();await page.getByRole('button',{name:'Sign out this browser',exact:true}).click();await page.getByRole('button',{name:'Confirm sign out',exact:true}).click();
  await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toBeVisible();expect((await context.cookies()).find(c=>c.name==='gi_session')).toBeUndefined();
  await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(composer).toHaveValue('Functional logout draft');
 }finally{await env.close();}
});

test('Unenrolled users cannot read or change authentication policy through owner routes',async({page,request})=>{
 const before=await(await request.get('/api/auth/status')).json();
 await page.goto('/');await expect(page.locator('.compose-box textarea')).toBeVisible();
 const results=await page.evaluate(async()=>{
  const read=await fetch('/api/auth/policy');const write=await fetch('/api/auth/policy',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({policy:'passkey-only',revision:'initial'})});
  return {read:read.status,write:write.status};
 });
 expect(results).toEqual({read:401,write:401});expect(await(await request.get('/api/auth/status')).json()).toEqual(before);
});
