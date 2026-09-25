import {test,expect} from '@playwright/test';
import {existsSync,readFileSync,writeFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {authEnvironment,totp} from './support/auth-environment.mjs';
import {loadCorpus} from './support/catalogue.mjs';
const composerName='Message (Enter to send, Shift+Enter for newline)...';
async function source(info,id){await info.attach('gherkin',{body:loadCorpus().find(x=>x.id===id).steps.join('\n'),contentType:'text/plain'});}

// Browser bootstrap API prerequisites only: no setup controls or automatic
// enrolment. This does not earn the frozen first-owner Settings journey.
const setupAPI=(page,operation,body={})=>page.evaluate(async({operation,body})=>{const r=await fetch('/api/auth/setup/'+operation,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});return {status:r.status,body:await r.json(),cache:r.headers.get('cache-control')};},{operation,body});

test('Browser-bound setup atomically creates one owner and resists stale other-browser finish',async({page,context,browser},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});const otherContext=await browser.newContext();const other=await otherContext.newPage();
 try{
  await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);
  await page.goto(env.origin);await other.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();
  await page.locator('.compose-box textarea').fill('Bootstrap prerequisite retains draft Ω');
  const a=await setupAPI(page,'start'),b=await setupAPI(other,'start');expect(a.status).toBe(200);expect(b.status).toBe(200);expect(a.cache).toBe('private, no-store');expect(a.body.username).toBe('admin');expect(a.body.expires_in_seconds).toBe(600);expect(a.body.secret).not.toBe(b.body.secret);expect(existsSync(env.authPath)).toBe(false);
  const setup=(await context.cookies()).find(c=>c.name==='gi_setup');expect(setup).toMatchObject({httpOnly:true,sameSite:'Strict',path:'/api/auth/setup',secure:false});expect(setup.value).toHaveLength(64);
  expect(await page.evaluate(()=>document.cookie)).not.toContain('gi_setup');
  const otherCookie=(await otherContext.cookies()).find(c=>c.name==='gi_setup');expect(otherCookie.value).not.toBe(setup.value);
  const finish=await setupAPI(page,'finish',{code:totp(a.body.secret)});expect(finish).toEqual({status:200,body:{ok:true},cache:'private, no-store'});
  const cookies=await context.cookies();expect(cookies.find(c=>c.name==='gi_setup')).toBeUndefined();const owner=cookies.find(c=>c.name==='gi_session');expect(owner).toMatchObject({httpOnly:true,sameSite:'Strict',path:'/'});
  const saved=readFileSync(env.authPath,'utf8'),state=JSON.parse(saved);expect(state.sessions).toHaveLength(1);expect(state.sessions[0]).toMatchObject({token_hash:createHash('sha256').update(owner.value).digest('hex'),purpose:'browser-owner',auth_factor:'totp'});expect(state.totp_secret).toBe(a.body.secret);expect(saved).not.toContain(owner.value);expect(saved).not.toContain(setup.value);expect(saved).not.toContain(b.body.secret);
  expect(await page.evaluate(async()=>(await(await fetch('/api/auth/session/proof')).json()).reauth_required)).toBe(false);
  expect(await other.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(401);
  expect((await setupAPI(other,'finish',{code:totp(b.body.secret)})).status).toBe(409);expect((await otherContext.cookies()).find(c=>c.name==='gi_session')).toBeUndefined();expect(readFileSync(env.authPath,'utf8')).toBe(saved);
  await env.restart();await page.reload();await expect(page.locator('.compose-box textarea')).toHaveValue('Bootstrap prerequisite retains draft Ω');expect((await setupAPI(page,'start')).status).toBe(409);expect(readFileSync(env.authPath,'utf8')).toBe(saved);
  expect(await page.evaluate(({secret,token,binding})=>{const storage=JSON.stringify({local:{...localStorage},session:{...sessionStorage}});return [secret,token,binding].every(s=>!storage.includes(s));},{secret:a.body.secret,token:owner.value,binding:setup.value})).toBe(true);
 }finally{await otherContext.close();await env.close();}
});

test('Browser setup cancellation, replay and process restart leave no owner until a new finish',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();
  const a=await setupAPI(page,'start');expect(a.status).toBe(200);expect((await setupAPI(page,'cancel')).status).toBe(200);expect((await context.cookies()).find(c=>c.name==='gi_setup')).toBeUndefined();expect((await setupAPI(page,'finish',{code:totp(a.body.secret)})).status).toBe(401);expect(existsSync(env.authPath)).toBe(false);
  const b=await setupAPI(page,'start');expect(b.status).toBe(200);await env.restart();expect((await setupAPI(page,'finish',{code:totp(b.body.secret)})).status).toBe(400);expect(existsSync(env.authPath)).toBe(false);expect((await context.cookies()).find(c=>c.name==='gi_setup')).toBeUndefined();
  const c=await setupAPI(page,'start');expect(c.status).toBe(200);expect((await setupAPI(page,'finish',{code:totp(c.body.secret)})).status).toBe(200);const saved=readFileSync(env.authPath,'utf8');expect((await setupAPI(page,'finish',{code:totp(c.body.secret)})).status).toBe(401);expect(readFileSync(env.authPath,'utf8')).toBe(saved);
 }finally{await env.close();}
});

async function openSetup(page){await page.keyboard.press('Control+,');await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByRole('button',{name:'Set up owner',exact:true})).toBeEnabled();}
async function startSetupUI(page){await page.getByRole('button',{name:'Set up owner',exact:true}).click();const key=page.getByLabel('Authenticator setup key',{exact:true});await expect(key).toBeVisible();const secret=await key.innerText();expect(secret).toMatch(/^[A-Z2-7]{32}$/);return secret;}

test('Settings owner setup verifies before enabling authentication and hides local secrets on exit',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});let release=()=>{};
 try{
  await page.goto(env.origin);const composer=page.locator('.compose-box textarea');await expect(composer).toBeVisible();await composer.fill('Owner setup draft Ω');await page.locator('.compose-box input[type=file]').setInputFiles({name:'setup.txt',mimeType:'text/plain',buffer:Buffer.from('setup retained media')});await expect(page.locator('.compose-file-pill[title="setup.txt"]')).toBeVisible();
  await openSetup(page);const writes=[];page.on('request',r=>{if(r.method()==='POST'&&r.url().includes('/api/auth/setup/'))writes.push(new URL(r.url()).pathname);});
  const abandoned=await startSetupUI(page);await page.getByRole('button',{name:'General',exact:true}).click();await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);expect(writes).toEqual(['/api/auth/setup/start']);
  await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByRole('button',{name:'Set up owner',exact:true})).toBeEnabled();const secret=await startSetupUI(page);expect(secret).not.toBe(abandoned);expect(existsSync(env.authPath)).toBe(false);
  await expect(page.getByRole('textbox',{name:'Setup verification code',exact:true})).toBeFocused();
  const storage=await page.evaluate(()=>JSON.stringify({local:{...localStorage},session:{...sessionStorage}}));expect(storage).not.toContain(secret);expect(storage).not.toContain(abandoned);expect(await page.evaluate(()=>document.cookie)).not.toContain('gi_setup');
  let held=false;const gate=new Promise(r=>release=r);await page.route('**/api/auth/setup/finish',async route=>{held=true;await gate;await route.continue();});
  await page.getByRole('textbox',{name:'Setup verification code',exact:true}).fill(totp(secret));await page.getByRole('button',{name:'Verify and enable authentication',exact:true}).click();await expect.poll(()=>held).toBe(true);await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);expect(existsSync(env.authPath)).toBe(false);await expect(page.getByText('Owner authentication enabled.',{exact:true})).toHaveCount(0);
  release();await expect(page.getByText('Owner authentication enabled.',{exact:true})).toBeVisible();await expect(page.getByRole('heading',{name:'Set up instance owner',exact:true})).toHaveCount(0);expect((await context.cookies()).find(c=>c.name==='gi_session')).toBeDefined();expect(JSON.parse(readFileSync(env.authPath,'utf8')).totp_secret).toBe(secret);
  expect(writes).toEqual(['/api/auth/setup/start','/api/auth/setup/start','/api/auth/setup/finish']);await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(composer).toHaveValue('Owner setup draft Ω');await expect(page.locator('.compose-file-pill[title="setup.txt"]')).toBeVisible();await page.reload();await expect(composer).toHaveValue('Owner setup draft Ω');
 }finally{release();await page.unrouteAll({behavior:'wait'});await env.close();}
});

test('Settings setup cancel and client expiry erase secrets without enrolment',async({page},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();await openSetup(page);const writes=[];page.on('request',r=>{if(r.method()==='POST'&&r.url().includes('/api/auth/setup/'))writes.push(new URL(r.url()).pathname);});
  await startSetupUI(page);await page.keyboard.press('Escape');await expect(page.getByText('Setup cancelled. No owner was created.',{exact:true})).toBeVisible();await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);await expect(page.locator('.settings-dialog')).toBeVisible();await expect(page.getByRole('button',{name:'Set up owner',exact:true})).toBeFocused();expect(existsSync(env.authPath)).toBe(false);expect(writes).toEqual(['/api/auth/setup/start','/api/auth/setup/cancel']);
  await page.clock.install();await startSetupUI(page);await page.clock.fastForward(600001);await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);await expect(page.getByRole('alert')).toContainText('Setup key expired');expect(writes).toEqual(['/api/auth/setup/start','/api/auth/setup/cancel','/api/auth/setup/start']);expect(existsSync(env.authPath)).toBe(false);
  await page.getByRole('button',{name:'Check setup status',exact:true}).click();await expect(page.getByRole('button',{name:'Set up owner',exact:true})).toBeEnabled();
 }finally{await env.close();}
});

for(const fault of ['invalid code','server error','false success','lost response','lost cookie'])test(`Settings setup ${fault} requires explicit status reconciliation`,async({page,context},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();await page.locator('.compose-box textarea').fill('Uncertain setup draft');await openSetup(page);const secret=await startSetupUI(page);let finishes=0;page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/api/auth/setup/finish'))finishes++;});
  if(fault!=='invalid code')await page.route('**/api/auth/setup/finish',async route=>{
   if(fault.startsWith('lost')){const result=await route.fetch();expect(result.status()).toBe(200);await route.abort('failed');}
   else await route.fulfill({status:fault==='server error'?503:200,contentType:'application/json',body:fault==='server error'?JSON.stringify({error:secret}):'{"ok":true}'});
  });
  const code=totp(secret);await page.getByRole('textbox',{name:'Setup verification code',exact:true}).fill(fault==='invalid code'?String((Number(code)+500000)%1000000).padStart(6,'0'):code);await page.getByRole('button',{name:'Verify and enable authentication',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Check setup status');await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);await expect(page.getByText('Owner authentication enabled.',{exact:true})).toHaveCount(0);expect(await page.locator('.settings-dialog').innerText()).not.toContain(secret);expect(finishes).toBe(1);
  await page.unrouteAll({behavior:'wait'});if(fault==='lost cookie')await context.clearCookies();
  await page.getByRole('button',{name:'Check setup status',exact:true}).click();
  if(fault==='lost response'){await expect(page.getByText('Owner authentication enabled.',{exact:true})).toBeVisible();}
  else if(fault==='lost cookie'){
   await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toBeVisible();await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('Uncertain setup draft');
  }else{
   expect(existsSync(env.authPath)).toBe(false);await expect(page.getByRole('button',{name:'Set up owner',exact:true})).toBeEnabled();const next=await startSetupUI(page);expect(next).not.toBe(secret);await page.getByRole('textbox',{name:'Setup verification code',exact:true}).fill(totp(next));await page.getByRole('button',{name:'Verify and enable authentication',exact:true}).click();await expect(page.getByText('Owner authentication enabled.',{exact:true})).toBeVisible();
  }
  expect(finishes).toBe(fault.startsWith('lost')?1:2);
 }finally{await env.close();}
});

for(const operation of ['start','finish'])test(`Settings setup ${operation} cancellation and late replies cannot replace another pane`,async({page},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});let release=()=>{};
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();await openSetup(page);
  let secret;if(operation==='finish')secret=await startSetupUI(page);
  let committed=false;const gate=new Promise(r=>release=r);let posts=0;
  page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/api/auth/setup/'+operation))posts++;});
  await page.route('**/api/auth/setup/'+operation,async route=>{const response=await route.fetch();expect(response.status()).toBe(200);committed=true;await gate;try{await route.fulfill({response});}catch{/* cancelled navigation/request */}});
  if(operation==='start')await page.getByRole('button',{name:'Set up owner',exact:true}).click();else{await page.getByRole('textbox',{name:'Setup verification code',exact:true}).fill(totp(secret));await page.getByRole('button',{name:'Verify and enable authentication',exact:true}).click();}
  await expect.poll(()=>committed).toBe(true);await page.getByRole('button',{name:'Cancel owner setup',exact:true}).click();await expect(page.getByRole('button',{name:'Check setup status',exact:true})).toBeEnabled();await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);
  await page.getByRole('button',{name:'General',exact:true}).click();release();await page.unrouteAll({behavior:'wait'});await expect(page.getByRole('button',{name:'General',exact:true})).toBeFocused();await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);expect(posts).toBe(1);
  await page.getByRole('button',{name:'Authentication',exact:true}).click();
  if(operation==='start'){await expect(page.getByRole('button',{name:'Set up owner',exact:true})).toBeEnabled();expect(existsSync(env.authPath)).toBe(false);}
  else{await expect(page.getByRole('button',{name:'Sign out this browser',exact:true})).toBeEnabled();expect(JSON.parse(readFileSync(env.authPath,'utf8')).totp_secret).toBe(secret);}
 }finally{release();await page.unrouteAll({behavior:'wait'});await env.close();}
});

test('Settings setup cancels its server binding while the parent refresh is pending',async({page},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});let release=()=>{};
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();await openSetup(page);await startSetupUI(page);
  const gate=new Promise(r=>release=r);let held=false,cancels=0;await page.route('**/api/auth/status',async route=>{held=true;await gate;await route.continue();});page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/api/auth/setup/cancel'))cancels++;});
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect.poll(()=>held).toBe(true);await page.getByRole('button',{name:'Cancel owner setup',exact:true}).click();await expect.poll(()=>cancels).toBe(1);await expect(page.getByLabel('Authenticator setup key')).toHaveCount(0);release();await expect(page.getByText('Setup cancelled. No owner was created.',{exact:true})).toBeVisible();expect(existsSync(env.authPath)).toBe(false);
 }finally{release();await page.unrouteAll({behavior:'wait'});await env.close();}
});

test('Settings setup handles another tab replacing the shared setup cookie and owner creation elsewhere',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{enrolled:false});const second=await context.newPage();
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();await openSetup(page);const first=await startSetupUI(page);
  await second.goto(env.origin);await expect(second.locator('.compose-box textarea')).toBeVisible();await openSetup(second);let next=await startSetupUI(second);expect(next).not.toBe(first);
  // A shared cookie is replaced by the second explicit start. Do not let a
  // stale first tab submit a now-unbound secret successfully.
  await page.getByRole('textbox',{name:'Setup verification code',exact:true}).fill(totp(first));await page.getByRole('button',{name:'Verify and enable authentication',exact:true}).click();await expect(page.getByRole('button',{name:'Check setup status',exact:true})).toBeEnabled();expect(existsSync(env.authPath)).toBe(false);
  await second.getByRole('button',{name:'Cancel owner setup',exact:true}).click();await expect(second.getByRole('button',{name:'Check setup status',exact:true})).toBeEnabled();await second.getByRole('button',{name:'Check setup status',exact:true}).click();next=await startSetupUI(second);await second.getByRole('textbox',{name:'Setup verification code',exact:true}).fill(totp(next));await second.getByRole('button',{name:'Verify and enable authentication',exact:true}).click();await expect(second.getByText('Owner authentication enabled.',{exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Check setup status',exact:true}).click();await expect(page.getByText('Owner authentication enabled.',{exact:true})).toBeVisible();expect(JSON.parse(readFileSync(env.authPath,'utf8')).sessions).toHaveLength(1);
 }finally{await second.close();await env.close();}
});

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

// Browser-origin API integration prerequisite only. There are no passkey
// controls/ceremonies yet and this does not map any additive passkey scenario.
test('Browser owner proof is isolated from other cookies and bearer tokens without changing drafts',async({page,browser,context},info)=>{
 const env=await authEnvironment(page,info);const otherContext=await browser.newContext(),other=await otherContext.newPage();
 const login=async p=>{await p.goto(env.origin);const code=p.getByRole('textbox',{name:'Authentication code',exact:true});await expect(code).toBeVisible();await code.fill(totp(env.secret));await p.getByRole('button',{name:'Sign in',exact:true}).click();await expect(p.getByRole('textbox',{name:composerName,exact:true})).toBeVisible();};
 const proof=p=>p.evaluate(async()=>{const r=await fetch('/api/auth/session/proof');return {status:r.status,body:await r.json(),cache:r.headers.get('cache-control')};});
 try{
  await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);
  await login(page);await login(other);
  const input=page.getByRole('textbox',{name:composerName,exact:true});await input.fill('Proof refresh keeps this draft Ω');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'proof.txt',mimeType:'text/plain',buffer:Buffer.from('proof bytes')});
  await expect(page.locator('.compose-file-pill[title="proof.txt"]')).toBeVisible();
  const cookieBefore=(await context.cookies()).find(c=>c.name==='gi_session');
  const state=JSON.parse(readFileSync(env.authPath,'utf8'));
  for(const session of state.sessions){expect(session.purpose).toBe('browser-owner');session.created_at=new Date(Date.now()-3600000).toISOString();session.authenticated_at=new Date(Date.now()-360000).toISOString();}
  writeFileSync(env.authPath,JSON.stringify(state));
  const staleBytes=readFileSync(env.authPath,'utf8');
  expect(await proof(page)).toMatchObject({status:200,cache:'private, no-store',body:{reauth_required:true}});
  expect(await proof(other)).toMatchObject({status:200,body:{reauth_required:true}});
  await page.reload();await expect(input).toHaveValue('Proof refresh keeps this draft Ω');
  expect(await proof(page)).toMatchObject({body:{reauth_required:true}});expect(readFileSync(env.authPath,'utf8')).toBe(staleBytes);
  const reauth=await page.evaluate(async code=>{const r=await fetch('/api/auth/session/reauth/totp',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({code})});return {status:r.status,body:await r.json()};},totp(env.secret));
  expect(reauth).toMatchObject({status:200,body:{reauth_required:false}});
  expect(await proof(other)).toMatchObject({status:200,body:{reauth_required:true}});
  expect((await context.cookies()).find(c=>c.name==='gi_session')).toEqual(cookieBefore);
  const renewed=JSON.parse(readFileSync(env.authPath,'utf8'));expect(renewed.sessions).toHaveLength(2);
  const beforeByHash=new Map(state.sessions.map(s=>[s.token_hash,s]));
  // Go normalises equivalent RFC3339 forms (for example .000Z to Z).
  // Compare instants and captured token identities, never array position/text.
  expect(renewed.sessions.filter(s=>Date.parse(s.authenticated_at)!==Date.parse(beforeByHash.get(s.token_hash).authenticated_at))).toHaveLength(1);
  for(const session of renewed.sessions){const before=beforeByHash.get(session.token_hash);expect(before).toBeTruthy();expect(Date.parse(session.expires_at)).toBe(Date.parse(before.expires_at));}
  // Native API tokens remain ordinary credentials even copied into a cookie.
  const r=await fetch(env.origin+'/api/auth/totp/verify',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({code:totp(env.secret)})});expect(r.status).toBe(200);const {token}=await r.json();
  expect(await page.evaluate(async token=>(await fetch('/api/auth/session/proof',{headers:{Authorization:`Bearer ${token}`}})).status,token)).toBe(403);
  await otherContext.addCookies([{name:'gi_session',value:token,url:env.origin,httpOnly:true,sameSite:'Strict'}]);
  expect(await other.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(200);
  expect(await proof(other)).toMatchObject({status:401});
  await page.reload();await expect(input).toHaveValue('Proof refresh keeps this draft Ω');await expect(page.locator('.compose-file-pill[title="proof.txt"]')).toBeVisible();
  expect(await proof(page)).toMatchObject({status:200,body:{reauth_required:false}});
 }finally{await otherContext.close();await env.close();}
});

test('Authentication pane explains unavailable passkeys and preserves drafts across narrow keyboard navigation',async({page},info)=>{
 const env=await authEnvironment(page,info);
 try{
  await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();
  const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();await input.fill('Unavailable passkeys draft');await input.focus();await page.keyboard.press('Control+,');
  await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByText('Passkeys are disabled by policy or are not configured for this origin.',{exact:true})).toBeVisible();await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeDisabled();
  await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();expect(await page.locator('.settings-dialog').evaluate(el=>el.scrollWidth-el.clientWidth)).toBeLessThanOrEqual(1);
  await page.keyboard.press('Escape');await expect(page.getByRole('dialog',{name:'Gi Settings',exact:true})).toHaveCount(0);await expect(input).toHaveValue('Unavailable passkeys draft');
 }finally{await env.close();}
});

test('TOTP browser sign-out waits for native status and preserves drafts and other sessions',async({page,browser,context},info)=>{
 const env=await authEnvironment(page,info);const otherContext=await browser.newContext();const other=await otherContext.newPage();let release;const gate=new Promise(r=>release=r);
 const login=async p=>{await p.goto(env.origin);await p.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await p.getByRole('button',{name:'Sign in',exact:true}).click();await expect(p.locator('.compose-box textarea')).toBeVisible();};
 try{
  await login(page);await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);await login(other);const otherCookies=await otherContext.cookies();
  await page.locator('.compose-box textarea').fill('Logout keeps draft and selection Ω');await page.locator('.compose-box input[type=file]').setInputFiles({name:'signout.txt',mimeType:'text/plain',buffer:Buffer.from('Stored logout attachment')});await expect(page.locator('.compose-file-pill[title="signout.txt"]')).toBeVisible();
  await page.keyboard.press('Control+,');await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByRole('button',{name:'Sign out this browser',exact:true})).toBeEnabled();
  let posts=0,held=false;page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/api/auth/session/logout'))posts++;});
  await page.getByRole('button',{name:'Sign out this browser',exact:true}).click();await page.keyboard.press('Escape');await expect(page.getByRole('dialog',{name:'Gi Settings',exact:true})).toBeVisible();await expect(page.getByRole('button',{name:'Sign out this browser',exact:true})).toBeFocused();expect(posts).toBe(0);
  await page.route('**/api/auth/status',async route=>{held=true;await gate;await route.continue();});
  await page.getByRole('button',{name:'Sign out this browser',exact:true}).click();await page.getByRole('button',{name:'Confirm sign out',exact:true}).click();await expect.poll(()=>held).toBe(true);
  await expect(page.locator('.settings-dialog')).toBeVisible();await expect(page.getByRole('heading',{name:'Sign in to Gi',exact:true})).toHaveCount(0);expect(posts).toBe(1);expect((await context.cookies()).find(c=>c.name==='gi_session')).toBeUndefined();
  release();await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toBeVisible();expect(await otherContext.cookies()).toEqual(otherCookies);expect(await other.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(200);expect(await page.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(401);
  await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('Logout keeps draft and selection Ω');await expect(page.locator('.compose-file-pill[title="signout.txt"]')).toBeVisible();expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(env.main.id);
 }finally{release();await page.unrouteAll({behavior:'wait'});await otherContext.close();await env.close();}
});

test('TOTP owner changes policy without passkey configuration and a lost policy response requires refresh',async({page},info)=>{
 const env=await authEnvironment(page,info);
 try{
  await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.locator('.compose-box textarea')).toBeVisible();await page.keyboard.press('Control+,');await page.getByRole('button',{name:'Authentication',exact:true}).click();
  const choice=page.getByRole('combobox',{name:'Accepted sign-in methods',exact:true});await expect(choice).toHaveValue('either');
  await choice.selectOption('passkey-only');await expect(page.getByRole('button',{name:'Change sign-in policy',exact:true})).toBeDisabled();
  await choice.selectOption('totp-only');await page.getByRole('button',{name:'Change sign-in policy',exact:true}).click();
  let posts=0;await page.route('**/api/auth/policy',async route=>{if(route.request().method()!=='POST')return route.continue();posts++;const response=await route.fetch();expect(response.status()).toBe(200);await route.abort('failed');});
  await page.getByRole('button',{name:'Confirm policy change',exact:true}).click();await expect(page.getByRole('alert')).toBeVisible();expect(posts).toBe(1);expect(JSON.parse(readFileSync(env.authPath,'utf8')).login_policy).toBe('totp-only');
  await page.unroute('**/api/auth/policy');await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(choice).toHaveValue('totp-only');expect(posts).toBe(1);await expect(page.getByRole('button',{name:'Verify code',exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await page.reload();await expect(page.locator('.compose-box textarea')).toBeVisible();
 }finally{await env.close();}
});
