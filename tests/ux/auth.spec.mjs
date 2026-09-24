import {test,expect} from '@playwright/test';
import {readFileSync,writeFileSync} from 'node:fs';
import {authEnvironment,totp} from './support/auth-environment.mjs';
import {loadCorpus} from './support/catalogue.mjs';
const composerName='Message (Enter to send, Shift+Enter for newline)...';
async function source(info,id){await info.attach('gherkin',{body:loadCorpus().find(x=>x.id===id).steps.join('\n'),contentType:'text/plain'});}

test('@ux-auth-002 Single-user code-only native sign-in opens cookie-authenticated application',async({page,context},info)=>{
 await source(info,'@ux-auth-002');const env=await authEnvironment(page,info);
 const writes=[],errors=[];page.on('request',r=>{if(r.method()==='POST')writes.push(r);});page.on('pageerror',e=>errors.push(e.message));
 try {
  expect((await fetch(env.origin+'/api/sessions')).status).toBe(401);
  expect((await fetch(env.origin+'/sse/stream')).status).toBe(401);
  await page.goto(env.origin);
  const code=page.getByRole('textbox',{name:'Authentication code',exact:true}),signIn=page.getByRole('button',{name:'Sign in',exact:true});
  await expect(code).toBeVisible();await expect(code).toBeFocused();await expect(page.getByRole('textbox',{name:/username/i})).toHaveCount(0);
  await expect(page.getByRole('button',{name:/passkey/i})).toHaveCount(0);await expect(page.getByRole('textbox',{name:composerName,exact:true})).toHaveCount(0);
  expect(writes).toHaveLength(0);
  await code.fill('123');await expect(signIn).toBeDisabled();
  const valid=totp(env.secret);await code.fill(String((Number(valid)+1)%1000000).padStart(6,'0'));await signIn.click();
  await expect(page.getByRole('alert')).toContainText('Invalid authentication code');await expect(signIn).toBeEnabled();expect(await context.cookies()).toHaveLength(0);
  await page.route('**/api/auth/session',route=>route.abort('failed'));await signIn.click();
  await expect(page.getByRole('alert')).toContainText('Unable to reach the server');await page.unroute('**/api/auth/session');
  let release,held=false;const gate=new Promise(r=>release=r);
  await page.route('**/api/auth/session',async route=>{held=true;await gate;await route.continue();});
  const sentCode=totp(env.secret);await code.fill(sentCode);await signIn.click();await expect.poll(()=>held).toBe(true);await expect(code).toBeDisabled();await expect(page.getByRole('button',{name:'Signing in…'})).toBeDisabled();
  release();await page.unrouteAll({behavior:'wait'});
  const input=page.getByRole('textbox',{name:composerName,exact:true});await expect(input).toBeVisible();await expect.poll(()=>page.evaluate(()=>window.__authConnected)).toBeGreaterThan(0);
  const login=writes.filter(r=>new URL(r.url()).pathname==='/api/auth/session').at(-1);expect(login.postDataJSON()).toEqual({code:sentCode});
  const cookie=(await context.cookies()).find(c=>c.name==='gi_session');expect(cookie).toMatchObject({httpOnly:true,sameSite:'Strict',path:'/'});
  expect(await page.evaluate(()=>document.cookie)).not.toContain('gi_session');
  const storage=await page.evaluate(()=>JSON.stringify({local:{...localStorage},session:{...sessionStorage}}));expect(storage).not.toContain(cookie.value);expect(storage).not.toContain(env.secret);
  await input.fill('Auth draft Ω survives reload');await page.locator('.compose-box input[type=file]').setInputFiles({name:'auth.txt',mimeType:'text/plain',buffer:Buffer.from('authenticated media bytes')});await expect(page.locator('.compose-file-pill[title="auth.txt"]')).toBeVisible();
  await page.reload();await expect(input).toHaveValue('Auth draft Ω survives reload');await expect(page.locator('.compose-file-pill[title="auth.txt"]')).toBeVisible();await expect.poll(()=>page.evaluate(()=>window.__authConnected)).toBeGreaterThan(0);
  expect(await page.evaluate(async()=> (await fetch('/api/sessions')).status)).toBe(200);
  await page.screenshot({path:info.outputPath('authenticated.png')});
  // Expiry is revalidated on each new request and bootstrap, without erasing drafts.
  const state=JSON.parse(readFileSync(env.authPath,'utf8'));for(const session of state.sessions)session.expires_at='2000-01-01T00:00:00Z';writeFileSync(env.authPath,JSON.stringify(state));
  expect(await page.evaluate(async()=> (await fetch('/api/sessions')).status)).toBe(401);
  await page.reload();await expect(code).toBeVisible();await code.fill(totp(env.secret));await signIn.click();await expect(input).toHaveValue('Auth draft Ω survives reload');await expect(page.locator('.compose-file-pill[title="auth.txt"]')).toBeVisible();
  expect(errors).toEqual([]);
 }finally{await env.close();}
});

test('@ux-auth-004 Native policy failure hides credentials and Retry restores loaded code-only policy',async({page},info)=>{
 await source(info,'@ux-auth-004');const env=await authEnvironment(page,info),saved=readFileSync(env.authPath);const writes=[];
 page.on('request',r=>{if(!['GET','HEAD'].includes(r.method()))writes.push(r.url());});
 let release;const gate=new Promise(r=>release=r);
 try{
  writeFileSync(env.authPath,'{corrupt');
  await page.route('**/api/auth/status',async route=>{await gate;await route.continue();});
  await page.goto(env.origin);await expect(page.getByRole('status')).toHaveText('Loading sign-in options…');await expect(page.locator('input')).toHaveCount(0);await expect(page.getByRole('button',{name:'Sign in',exact:true})).toHaveCount(0);
  release();await page.unrouteAll({behavior:'wait'});await expect(page.getByRole('alert')).toContainText('Cannot load sign-in options');await expect(page.locator('input')).toHaveCount(0);await expect(page.getByRole('button',{name:/passkey/i})).toHaveCount(0);
  await page.screenshot({path:info.outputPath('policy-error.png')});
  writeFileSync(env.authPath,saved);await page.getByRole('button',{name:'Retry',exact:true}).click();await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toBeVisible();await expect(page.getByRole('textbox',{name:/username/i})).toHaveCount(0);expect(writes).toEqual([]);
  await page.screenshot({path:info.outputPath('sign-in.png')});
 }finally{release();writeFileSync(env.authPath,saved);await env.close();}
});

test('Browser gate rejects malformed policy and false-success login without mounting the app',async({page},info)=>{
 const env=await authEnvironment(page,info);
 try{
  await page.route('**/api/auth/status',route=>route.fulfill({status:200,contentType:'application/json',body:JSON.stringify({enrolled:false,mode:'family'})}));
  await page.goto(env.origin);await expect(page.getByRole('alert')).toContainText('Cannot load sign-in options');await expect(page.locator('input')).toHaveCount(0);
  await page.unroute('**/api/auth/status');await page.getByRole('button',{name:'Retry',exact:true}).click();
  const code=page.getByRole('textbox',{name:'Authentication code',exact:true});await expect(code).toBeVisible();
  await page.route('**/api/auth/session',route=>route.fulfill({status:200,contentType:'application/json',body:'{"ok":true}'}));
  await code.fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();
  await expect(page.getByRole('alert')).toContainText('Sign-in could not be confirmed');await expect(page.getByRole('textbox',{name:composerName,exact:true})).toHaveCount(0);
  await page.unroute('**/api/auth/session');await code.fill(totp(env.secret));await code.press('Enter');await expect(page.getByRole('textbox',{name:composerName,exact:true})).toBeVisible();
 }finally{await env.close();}
});
