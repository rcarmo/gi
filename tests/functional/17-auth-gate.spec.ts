import {test,expect} from '@playwright/test';

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
