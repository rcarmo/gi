import {test,expect} from '@playwright/test';
import {readFileSync,writeFileSync} from 'node:fs';
import {createHash,createPrivateKey,sign} from 'node:crypto';
import {authEnvironment,totp} from './support/auth-environment.mjs';

// Real browser credential API and signatures, native Go HTTP/crypto/storage.
// The first four tests use direct API integration; later tests drive Classic
// Settings. localhost/CDP coverage is not Visual or native physical prompt proof.
async function post(page,path,body={}){return page.evaluate(async({path,body})=>{const r=await fetch('/api/auth/passkeys'+path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});return {status:r.status,body:await r.json()};},{path,body});}
async function ceremony(page,operation,authenticator,name){
 const start=await post(page,`/${operation}/start`,name?{name}:{});expect(start.status).toBe(200);
 const credential=await page.evaluate(async({operation,options})=>{
  const from=s=>Uint8Array.from(atob(s.replace(/-/g,'+').replace(/_/g,'/')),c=>c.charCodeAt(0));
  const pub=options.publicKey;pub.challenge=from(pub.challenge);
  if(operation==='register'){pub.user.id=from(pub.user.id);pub.excludeCredentials=(pub.excludeCredentials||[]).map(c=>({...c,id:from(c.id)}));}
  else pub.allowCredentials=(pub.allowCredentials||[]).map(c=>({...c,id:from(c.id)}));
  const credential=await (operation==='register'?navigator.credentials.create({publicKey:pub}):navigator.credentials.get({publicKey:pub}));
  return credential.toJSON();
 },{operation,options:start.body.options});
 return {start,credential,finish:()=>post(page,`/${operation}/finish`,{ceremony_id:start.body.ceremony_id,credential})};
}
async function authenticator(page){const cdp=await page.context().newCDPSession(page);await cdp.send('WebAuthn.enable');const {authenticatorId}=await cdp.send('WebAuthn.addVirtualAuthenticator',{options:{protocol:'ctap2',transport:'usb',hasResidentKey:true,hasUserVerification:true,isUserVerified:true,automaticPresenceSimulation:true}});return {cdp,id:authenticatorId};}
async function list(page){return page.evaluate(async()=>{const r=await fetch('/api/auth/passkeys');return {status:r.status,body:await r.json()};});}
async function loginTOTP(page,env){await page.goto(env.origin);await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.locator('.compose-box textarea')).toBeVisible();}

test('native WebAuthn keeps two independent keys across restart and permits passkey-only further enrollment',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);const one=await ceremony(page,'register',auth,'Laptop');expect((await one.finish()).status).toBe(200);
  expect((await one.finish()).status).toBe(400); // consumed challenge
  const laptop=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials;expect(laptop).toHaveLength(1);
  const first=(await list(page)).body.passkeys;expect(first).toHaveLength(1);expect(first[0].name).toBe('Laptop');
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  const two=await ceremony(page,'register',auth,'Backup key');expect(two.start.body.options.publicKey.excludeCredentials.map(c=>c.id)).toContain(first[0].id);expect((await two.finish()).status).toBe(200);
  const backup=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials;expect(backup).toHaveLength(1);expect(backup[0].credentialId).not.toBe(laptop[0].credentialId);
  const keys=(await list(page)).body.passkeys;expect(keys).toHaveLength(2);expect(keys[0]).toEqual(first[0]);
  const responseText=JSON.stringify(await list(page));expect(responseText).not.toContain('public_key');expect(responseText).not.toContain('totp_secret');
  // Disable/remove TOTP in the disposable store to exercise a passkey-only
  // account. No production policy UI or policy-selection parity is claimed.
  const state=JSON.parse(readFileSync(env.authPath,'utf8'));state.totp_enabled=false;state.totp_secret='';state.login_policy='passkey-only';writeFileSync(env.authPath,JSON.stringify(state));
  await env.restart();
  let previous=keys;
  for(const [index,key]of [[0,laptop[0]],[1,backup[0]]]){
   await context.clearCookies();await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:key});
   const login=await ceremony(page,'login',auth);expect((await login.finish()).status).toBe(200);expect((await login.finish()).status).toBe(400);
   const after=(await list(page)).body.passkeys;expect(after.map(k=>k.name)).toEqual(['Laptop','Backup key']);expect(Date.parse(after[index].last_used_at)).toBeGreaterThan(Date.parse(previous[index].last_used_at)||0);expect(after[1-index]).toEqual(previous[1-index]);previous=after;
   expect((await context.cookies()).find(c=>c.name==='gi_session')).toMatchObject({httpOnly:true,sameSite:'Strict',path:'/'});
   expect(await page.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(200);
   const advanced=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
   if(index===0)laptop[0]=advanced;else backup[0]=advanced;
  }
  const beforeRename=JSON.parse(readFileSync(env.authPath,'utf8')).passkeys;
  expect((await post(page,'/rename',{id:keys[0].id,name:'Laptop Ω'})).status).toBe(200);
  const afterRename=JSON.parse(readFileSync(env.authPath,'utf8')).passkeys;expect(afterRename[0].credential).toEqual(beforeRename[0].credential);expect(afterRename[1]).toEqual(beforeRename[1]);
  expect((await post(page,'/remove',{id:keys[0].id})).status).toBe(200);
  expect((await list(page)).body.passkeys.map(k=>k.name)).toEqual(['Backup key']);
  expect((await post(page,'/remove',{id:keys[1].id})).status).toBe(409);
  // Generate a genuine assertion with the removed key by overriding only the
  // browser allow-list. Server validation still uses its persisted allow-list.
  await context.clearCookies();await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:laptop[0]});
  const rejectStart=await post(page,'/login/start');expect(rejectStart.status).toBe(200);
  const rejected=await page.evaluate(async({options,id})=>{const from=s=>Uint8Array.from(atob(s.replace(/-/g,'+').replace(/_/g,'/')),c=>c.charCodeAt(0));options.publicKey.challenge=from(options.publicKey.challenge);options.publicKey.allowCredentials=[{type:'public-key',id:from(id)}];return (await navigator.credentials.get(options)).toJSON();},{options:rejectStart.body.options,id:keys[0].id});
  expect((await post(page,'/login/finish',{ceremony_id:rejectStart.body.ceremony_id,credential:rejected})).status).toBe(400);
  expect(await page.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(401);
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:backup[0]});
  const remaining=await ceremony(page,'login',auth);expect((await remaining.finish()).status).toBe(200);
  // Recent passkey proof (no TOTP) authorises a third distinct key.
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});const third=await ceremony(page,'register',auth,'Third');expect((await third.finish()).status).toBe(200);expect((await list(page)).body.passkeys).toHaveLength(2);
 }finally{await auth.cdp.detach();await env.close();}
});

test('native WebAuthn rejects cross-session/revoked registration and consumes invalid proof',async({page,browser},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);const otherContext=await browser.newContext();const other=await otherContext.newPage();
 try{
  await loginTOTP(page,env);await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);await loginTOTP(other,env);
  const reg=await ceremony(page,'register',auth,'Laptop');
  expect((await post(other,'/register/finish',{ceremony_id:reg.start.body.ceremony_id,credential:reg.credential})).status).toBe(400);
  expect((await reg.finish()).status).toBe(200);
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  const bad=await ceremony(page,'register',auth,'Bad proof');const changed=structuredClone(bad.credential);changed.response.attestationObject='AAAA';
  expect((await post(page,'/register/finish',{ceremony_id:bad.start.body.ceremony_id,credential:changed})).status).toBe(400);expect((await bad.finish()).status).toBe(400);
  const revoked=await ceremony(page,'register',auth,'Revoked');const state=JSON.parse(readFileSync(env.authPath,'utf8'));state.sessions=[];writeFileSync(env.authPath,JSON.stringify(state));
  expect((await revoked.finish()).status).toBe(401);
  const saved=JSON.parse(readFileSync(env.authPath,'utf8'));expect(saved.passkeys).toHaveLength(1);expect(saved.passkeys[0].name).toBe('Laptop');
 }finally{await otherContext.close();await auth.cdp.detach();await env.close();}
});

test('native WebAuthn enforces expiry origin RP and recent passkey proof with deliberate cancellation',async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);
  const initial=await ceremony(page,'register',auth,'Laptop');expect((await initial.finish()).status).toBe(200);
  // Sign an assertion with the real credential key but a different RP hash.
  // This distinguishes server RP validation from malformed-CBOR rejection.
  const wrongRP=await ceremony(page,'reauth',auth);const signed=structuredClone(wrongRP.credential);
  const authData=Buffer.from(signed.response.authenticatorData,'base64url');createHash('sha256').update('evil.example').digest().copy(authData,0);
  const nativeKey=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  const key=createPrivateKey({key:Buffer.from(nativeKey.privateKey,'base64'),format:'der',type:'pkcs8'});
  signed.response.authenticatorData=authData.toString('base64url');
  signed.response.signature=sign('sha256',Buffer.concat([authData,createHash('sha256').update(Buffer.from(signed.response.clientDataJSON,'base64url')).digest()]),key).toString('base64url');
  expect((await post(page,'/reauth/finish',{ceremony_id:wrongRP.start.body.ceremony_id,credential:signed})).status).toBe(400);
  expect((await wrongRP.finish()).status).toBe(400);
  let state=JSON.parse(readFileSync(env.authPath,'utf8'));for(const s of state.sessions){s.created_at=new Date(Date.now()-3600000).toISOString();s.authenticated_at=new Date(Date.now()-360000).toISOString()};writeFileSync(env.authPath,JSON.stringify(state));
  expect((await post(page,'/register/start',{name:'stale'})).status).toBe(403);
  const reauth=await ceremony(page,'reauth',auth);expect((await reauth.finish()).status).toBe(200);
  expect(await page.evaluate(async()=>(await(await fetch('/api/auth/session/proof')).json()).reauth_required)).toBe(false);
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  const cancelled=await post(page,'/register/start',{name:'Cancelled'});expect(cancelled.status).toBe(200);
  await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:false});
  const cancellation=await page.evaluate(async options=>{
   const from=s=>Uint8Array.from(atob(s.replace(/-/g,'+').replace(/_/g,'/')),c=>c.charCodeAt(0));options.publicKey.challenge=from(options.publicKey.challenge);options.publicKey.user.id=from(options.publicKey.user.id);options.publicKey.excludeCredentials=[];
   const controller=new AbortController();const pending=navigator.credentials.create({...options,signal:controller.signal});setTimeout(()=>controller.abort(),50);
   try{await pending;return 'unexpected success'}catch(e){return e.name}
  },cancelled.body.options);expect(cancellation).toBe('AbortError');expect((await list(page)).body.passkeys).toHaveLength(1);
  await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:true});
  for(const invalid of ['origin','rp','expired']){
   const reg=await ceremony(page,'register',auth,invalid);const modified=structuredClone(reg.credential);
   if(invalid==='origin'){const client=JSON.parse(Buffer.from(modified.response.clientDataJSON,'base64url'));client.origin='https://evil.example';modified.response.clientDataJSON=Buffer.from(JSON.stringify(client)).toString('base64url');}
   if(invalid==='rp'){
    // Re-run with browser signing for another RP under the allowed localhost
    // parent; native verifier must reject the resulting authData RP hash.
    const alt=structuredClone(reg.start.body.options);alt.publicKey.rp.id='invalid.localhost';
    // Browser rejects an RP outside the origin's scope before signing.
    const rejected=await page.evaluate(async options=>{const from=s=>Uint8Array.from(atob(s.replace(/-/g,'+').replace(/_/g,'/')),c=>c.charCodeAt(0));options.publicKey.challenge=from(options.publicKey.challenge);options.publicKey.user.id=from(options.publicKey.user.id);options.publicKey.excludeCredentials=[];try{await navigator.credentials.create(options);return 'success'}catch(e){return e.name}},alt);expect(rejected).toBe('SecurityError');
    // Corrupt RP-bound attestation bytes on the server path as well.
    modified.response.attestationObject='AAAA';
   }
   if(invalid==='expired'){state=JSON.parse(readFileSync(env.authPath,'utf8'));state.webauthn_ceremonies.find(c=>c.id===reg.start.body.ceremony_id).expires_at='2000-01-01T00:00:00Z';writeFileSync(env.authPath,JSON.stringify(state));}
   expect((await post(page,'/register/finish',{ceremony_id:reg.start.body.ceremony_id,credential:modified})).status).toBe(400);
   expect((await reg.finish()).status).toBe(400);expect((await list(page)).body.passkeys).toHaveLength(1);
   await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  }
 }finally{await auth.cdp.detach();await env.close();}
});

test('native duplicate registration does not replace keys and lost finish response reconciles once',async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);const first=await ceremony(page,'register',auth,'Laptop');expect((await first.finish()).status).toBe(200);
  const saved=JSON.parse(readFileSync(env.authPath,'utf8')).passkeys[0];
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  const second=await ceremony(page,'register',auth,'Second');
  // Corrupt duplicate-ID state is seeded only to test the server's independent
  // duplicate guard, beyond WebAuthn excludeCredentials client behaviour.
  let state=JSON.parse(readFileSync(env.authPath,'utf8'));const collision=structuredClone(saved);collision.name='Existing collision';collision.credential.id=Buffer.from(second.credential.rawId,'base64url').toString('base64');state.passkeys.push(collision);writeFileSync(env.authPath,JSON.stringify(state));
  const before=JSON.parse(readFileSync(env.authPath,'utf8')).passkeys;
  expect((await second.finish()).status).toBe(409);expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys).toEqual(before);
  state=JSON.parse(readFileSync(env.authPath,'utf8'));state.passkeys=[saved];writeFileSync(env.authPath,JSON.stringify(state));
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});const uncertain=await ceremony(page,'register',auth,'Uncertain');
  await page.route('**/api/auth/passkeys/register/finish',async route=>{const response=await route.fetch();expect(response.status()).toBe(200);await route.abort('failed');});
  await expect(uncertain.finish()).rejects.toThrow();await page.unrouteAll({behavior:'wait'});
  expect((await list(page)).body.passkeys.map(k=>k.name)).toEqual(['Laptop','Uncertain']);
  expect((await uncertain.finish()).status).toBe(400);expect((await list(page)).body.passkeys).toHaveLength(2);
 }finally{await auth.cdp.detach();await env.close();}
});

async function openAuthentication(page){
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Settings',exact:true}).click();
 await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByRole('heading',{name:'Passkeys',exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
}
async function addFromSettings(page,name){
 await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill(name);
 await page.getByRole('button',{name:'Add passkey',exact:true}).click();
 await expect(page.locator('.gi-passkey-row').filter({has:page.locator('strong',{hasText:name})})).toHaveCount(1);
 await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
}
test('Settings enrolls two passkeys and fresh login accepts each after restart without replacing drafts',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await page.locator('.compose-box textarea').fill('Passkey settings preserve draft Ω');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'keep.txt',mimeType:'text/plain',buffer:Buffer.from('passkey settings bytes')});
  await expect(page.locator('.compose-file-pill[title="keep.txt"]')).toBeVisible();
  await openAuthentication(page);await addFromSettings(page,'Laptop');
  const laptop=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  const original=JSON.parse(readFileSync(env.authPath,'utf8')).passkeys[0];
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Backup key');
  const backup=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys[0]).toEqual(original);
  expect(await page.locator('.settings-dialog').evaluate(el=>el.scrollWidth-el.clientWidth)).toBeLessThanOrEqual(1);
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await env.restart();
  for(const key of [laptop,backup]){
   await context.clearCookies();await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:key});
   await page.reload();await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();
   await expect(page.locator('.compose-box textarea')).toHaveValue('Passkey settings preserve draft Ω');await expect(page.locator('.compose-file-pill[title="keep.txt"]')).toBeVisible();
   await openAuthentication(page);await expect(page.locator('.gi-passkey-row')).toHaveCount(2);await page.getByRole('button',{name:'Close settings',exact:true}).click();
  }
  await openAuthentication(page);
  await page.getByRole('button',{name:'Rename Laptop',exact:true}).click();await page.getByRole('textbox',{name:'Rename passkey',exact:true}).fill('Laptop Ω');await page.getByRole('button',{name:'Save passkey name',exact:true}).click();
  await expect(page.getByRole('button',{name:'Remove Laptop Ω',exact:true})).toBeEnabled();
  await page.getByRole('button',{name:'Remove Laptop Ω',exact:true}).click();await expect(page.getByRole('group',{name:'Confirm passkey removal'})).toContainText('Existing login sessions are not signed out');
  await page.getByRole('button',{name:'Cancel removal',exact:true}).click();await expect(page.getByRole('button',{name:'Remove Laptop Ω',exact:true})).toBeFocused();expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys).toHaveLength(2);
  await page.getByRole('button',{name:'Remove Laptop Ω',exact:true}).click();await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect(page.locator('.gi-passkey-row')).toHaveCount(1);
  // Owner can add using recent passkey proof with TOTP disabled and no code UI.
  const state=JSON.parse(readFileSync(env.authPath,'utf8'));state.login_policy='passkey-only';state.totp_enabled=false;state.totp_secret='';for(const s of state.sessions){s.created_at=new Date(Date.now()-3600000).toISOString();s.authenticated_at=new Date(Date.now()-360000).toISOString()};writeFileSync(env.authPath,JSON.stringify(state));
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Verify code',exact:true})).toHaveCount(0);
  await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Third key');await expect(page.locator('.gi-passkey-row')).toHaveCount(2);
  await page.screenshot({path:info.outputPath('passkey-settings.png')});
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('Passkey settings preserve draft Ω');
  await context.clearCookies();await page.reload();await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toHaveCount(0);
  await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:false});await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();await expect(page.getByRole('button',{name:'Cancel passkey prompt',exact:true})).toBeVisible();await page.waitForTimeout(100);await page.getByRole('button',{name:'Cancel passkey prompt',exact:true}).click();await expect(page.getByRole('alert')).toContainText('cancelled');await expect(page.getByRole('button',{name:'Sign in with passkey',exact:true})).toBeFocused();
  await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:true});await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('Passkey settings preserve draft Ω');
 }finally{await auth.cdp.detach();await env.close();}
});

test('Settings owns cancellation, reauth and failed or uncertain writes without replacing confirmed rows',async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);const errors=[];page.on('pageerror',e=>errors.push(e.message));
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  const first=JSON.parse(readFileSync(env.authPath,'utf8')).passkeys[0];
  // Stale proof requires explicit factor verification in this pane.
  let state=JSON.parse(readFileSync(env.authPath,'utf8'));for(const s of state.sessions){s.created_at=new Date(Date.now()-3600000).toISOString();s.authenticated_at=new Date(Date.now()-360000).toISOString()};writeFileSync(env.authPath,JSON.stringify(state));
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Cancelled');await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeDisabled();
  await page.getByRole('textbox',{name:'Reauthentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Verify code',exact:true}).click();await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeEnabled();
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:false});
  let starts=0,finishes=0;page.on('request',r=>{if(r.url().endsWith('/register/start'))starts++;if(r.url().endsWith('/register/finish'))finishes++;});
  await page.getByRole('button',{name:'Add passkey',exact:true}).click();await expect.poll(()=>starts).toBe(1);await page.waitForTimeout(100);
  // Native prompts may blur the page. This synthetic notification tests only
  // app blur ownership; physical OS prompt focus remains manual evidence.
  await page.evaluate(()=>window.dispatchEvent(new Event('blur')));await expect(page.getByRole('button',{name:'Cancel pending operation',exact:true})).toBeVisible();expect(finishes).toBe(0);
  await page.keyboard.press('Escape');await expect(page.getByRole('dialog',{name:'Gi Settings',exact:true})).toBeVisible();await expect(page.getByRole('alert')).toContainText('cancelled');expect(finishes).toBe(0);expect(starts).toBe(1);
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeEnabled();
  await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:true});await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('After cancellation');await page.getByRole('button',{name:'Add passkey',exact:true}).click();await expect(page.locator('.gi-passkey-row')).toHaveCount(2);expect(starts).toBe(2);
  await page.getByRole('button',{name:'Rename Laptop',exact:true}).click();await page.getByRole('textbox',{name:'Rename passkey',exact:true}).fill('Bad network rename');
  await page.route('**/api/auth/passkeys/rename',route=>route.fulfill({status:503,contentType:'application/json',body:'{"error":"Rename unavailable"}'}));await page.getByRole('button',{name:'Save passkey name',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Rename unavailable');await expect(page.locator('.gi-passkey-row strong',{hasText:'Laptop'})).toHaveCount(1);await page.unroute('**/api/auth/passkeys/rename');
  await page.getByRole('button',{name:'Cancel rename',exact:true}).click();await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();
  await page.getByRole('button',{name:'Remove Laptop',exact:true}).click();await page.route('**/api/auth/passkeys/remove',route=>route.fulfill({status:503,contentType:'application/json',body:'{"error":"Removal unavailable"}'}));await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Removal unavailable');await expect(page.locator('.gi-passkey-row')).toHaveCount(2);await page.unroute('**/api/auth/passkeys/remove');await page.getByRole('button',{name:'Cancel removal',exact:true}).click();
  await page.route('**/api/auth/passkeys',route=>route.fulfill({status:503,contentType:'application/json',body:'{"error":"List unavailable"}'}));await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('alert')).toContainText('List unavailable');await expect(page.locator('.gi-passkey-row')).toHaveCount(2);await expect(page.getByText('No passkeys registered.',{exact:true})).toHaveCount(0);await page.unroute('**/api/auth/passkeys');
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  await page.route('**/api/auth/passkeys/register/finish',async route=>{const r=await route.fetch();expect(r.status()).toBe(200);await route.abort('failed');});
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Uncertain');await page.getByRole('button',{name:'Add passkey',exact:true}).click();await expect(page.getByRole('alert')).toContainText('could not be confirmed');await expect(page.getByRole('alert')).toContainText('local credential may remain');const count=finishes;await page.waitForTimeout(200);expect(finishes).toBe(count);
  await page.unroute('**/api/auth/passkeys/register/finish');await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.locator('.gi-passkey-row')).toHaveCount(3);expect(finishes).toBe(count);
  expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys[0]).toEqual(first);expect(errors).toEqual([]);
 }finally{await auth.cdp.detach();await env.close();}
});

test('Settings explains rejected local credentials and closing pending ceremony cannot update another pane',async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);
  await page.route('**/api/auth/passkeys/register/finish',route=>route.fulfill({status:400,contentType:'application/json',body:'{"error":"Passkey verification failed"}'}));
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Not registered');await page.getByRole('button',{name:'Add passkey',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Not registered on the server');await expect(page.getByRole('alert')).toContainText('Gi has not removed it');
  expect((await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials).toHaveLength(1);expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys||[]).toHaveLength(0);
  await page.unroute('**/api/auth/passkeys/register/finish');await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:false});
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Close during prompt');await page.getByRole('button',{name:'Add passkey',exact:true}).click();await expect(page.getByRole('button',{name:'Cancel pending operation',exact:true})).toBeVisible();
  await page.getByRole('button',{name:'General',exact:true}).click();await expect(page.getByRole('heading',{name:'General',exact:true})).toBeVisible();
  await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:true});await page.waitForTimeout(150);expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys||[]).toHaveLength(0);
  await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByText('No passkeys registered.',{exact:true})).toBeVisible();await expect(page.getByRole('alert')).toHaveCount(0);
  await addFromSettings(page,'Fresh attempt');await expect(page.locator('.gi-passkey-row')).toHaveCount(1);
 }finally{await auth.cdp.detach();await env.close();}
});

// Bounded evidence for additive 006/007/009/026, not full frozen mappings:
// these use Classic on localhost, not both skins on the pinned HTTPS origin.
const savedAuth=env=>JSON.parse(readFileSync(env.authPath,'utf8'));
const credentialRow=(page,id)=>page.locator('.gi-passkey-row').filter({has:page.locator('small',{hasText:`Identifier: ${id}`})});

test('Settings renames one of two unnamed credentials without changing material and retains it after reload',async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Backup key');
  // Simulate legacy unnamed credentials only in this case's disposable store.
  const state=savedAuth(env);for(const key of state.passkeys)key.name='';writeFileSync(env.authPath,JSON.stringify(state));
  const before=savedAuth(env).passkeys;const keys=(await list(page)).body.passkeys;
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Unnamed passkey','Unnamed passkey']);
  const row=credentialRow(page,keys[0].id);await expect(row).toHaveCount(1);
  await row.getByRole('button',{name:'Rename passkey',exact:true}).click();await row.getByRole('textbox',{name:'Rename passkey',exact:true}).fill('Tablet');
  await row.getByRole('button',{name:'Save passkey name',exact:true}).click();await expect(page.getByText('Name saved.',{exact:true})).toBeVisible();
  expect(savedAuth(env).passkeys).toEqual([{...before[0],name:'Tablet'},before[1]]);
  await page.reload();await expect(page.locator('.compose-box textarea')).toBeVisible();await openAuthentication(page);
  await expect(credentialRow(page,keys[0].id).locator('strong')).toHaveText('Tablet');await expect(credentialRow(page,keys[1].id).locator('strong')).toHaveText('Unnamed passkey');
  await expect(page.locator('.gi-passkey-row')).toHaveCount(2);expect(savedAuth(env).passkeys).toEqual([{...before[0],name:'Tablet'},before[1]]);
 }finally{await auth.cdp.detach();await env.close();}
});

for(const example of [
 {label:'blank',name:'',valid:false},
 {label:'whitespace',name:'   ',valid:false},
 {label:'81 Unicode characters',name:'🔑'.repeat(81),valid:false},
 {label:'control character',name:'Laptop\u0007',valid:false},
 {label:'literal HTML',name:'<img src=x onerror="window.__passkeyMarkupRan=true">',valid:true},
])test(`Settings rename validates ${example.label} and the same credential still signs in`,async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  const before=savedAuth(env).passkeys;const id=(await list(page)).body.passkeys[0].id;const row=credentialRow(page,id);
  await row.getByRole('button',{name:'Rename Laptop',exact:true}).click();await row.getByRole('textbox',{name:'Rename passkey',exact:true}).fill(example.name);
  await expect(row.getByRole('textbox',{name:'Rename passkey',exact:true})).toHaveValue(example.name);
  const response=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/rename')&&r.request().method()==='POST');
  await row.getByRole('button',{name:'Save passkey name',exact:true}).click();const renamed=await response;
  expect(renamed.request().postDataJSON()).toEqual({id,name:example.name});expect(renamed.status()).toBe(example.valid?200:400);
  const expectedName=example.valid?example.name:'Laptop';
  if(example.valid){
   await expect(page.getByText('Name saved.',{exact:true})).toBeVisible();await expect(row.locator('img')).toHaveCount(0);expect(await page.evaluate(()=>window.__passkeyMarkupRan)).toBeUndefined();
  }else{
   await expect(page.getByRole('alert')).toContainText('1 to 80 characters without controls');await expect(page.getByText('Name saved.',{exact:true})).toHaveCount(0);
  }
  await expect(row.locator('strong')).toHaveText(expectedName);
  expect(savedAuth(env).passkeys).toEqual([{...before[0],name:expectedName}]);
  await context.clearCookies();await page.reload();await expect(page.getByRole('button',{name:'Sign in with passkey',exact:true})).toBeVisible();
  const login=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/login/finish')&&r.request().method()==='POST');
  await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();expect((await login).status()).toBe(200);await expect(page.locator('.compose-box textarea')).toBeVisible();
  await openAuthentication(page);await expect(credentialRow(page,id).locator('strong')).toHaveText(expectedName);await expect(page.locator('.gi-passkey-row')).toHaveCount(1);
 }finally{await auth.cdp.detach();await env.close();}
});

test('Settings removal cancellation sends no request and restores the exact identified control',async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Backup key');
  const before=savedAuth(env).passkeys;const keys=(await list(page)).body.passkeys;const row=credentialRow(page,keys[0].id);const remove=row.getByRole('button',{name:'Remove Laptop',exact:true});
  const writes=[];page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname==='/api/auth/passkeys/remove')writes.push(r.postDataJSON());});
  for(const method of ['button','Escape']){
   await remove.click();const confirmation=page.getByRole('group',{name:'Confirm passkey removal'});
   await expect(confirmation).toContainText(`Laptop (${keys[0].id})`);expect(writes).toEqual([]);
   if(method==='button')await confirmation.getByRole('button',{name:'Cancel removal',exact:true}).click();else await page.keyboard.press('Escape');
   await expect(confirmation).toHaveCount(0);await expect(remove).toBeFocused();await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop','Backup key']);
   expect(writes).toEqual([]);expect(savedAuth(env).passkeys).toEqual(before);
  }
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(remove).toBeEnabled();expect(writes).toEqual([]);expect((await list(page)).body.passkeys).toEqual(keys);
  // Positive control: the same observer sees exactly one explicitly confirmed
  // removal, so a broken URL matcher cannot make cancellation pass vacuously.
  await remove.click();await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect(page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toBeVisible();
  expect(writes).toEqual([{id:keys[0].id}]);expect(savedAuth(env).passkeys).toEqual([before[1]]);
 }finally{await auth.cdp.detach();await env.close();}
});

test('Settings proof in one browser leaves another stale for add rename and remove',async({page,browser},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);const otherContext=await browser.newContext();const other=await otherContext.newPage();
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);await loginTOTP(other,env);await openAuthentication(other);
  const cookie=(await page.context().cookies()).find(c=>c.name==='gi_session');const otherCookie=(await otherContext.cookies()).find(c=>c.name==='gi_session');expect(cookie.value).not.toBe(otherCookie.value);
  const state=savedAuth(env);expect(state.sessions).toHaveLength(2);
  for(const s of state.sessions){s.created_at=new Date(Date.now()-3600000).toISOString();s.authenticated_at=new Date(Date.now()-360000).toISOString();}writeFileSync(env.authPath,JSON.stringify(state));
  const proof=p=>p.evaluate(async()=>(await(await fetch('/api/auth/session/proof')).json()));
  const staleProof=await proof(other);expect(staleProof.reauth_required).toBe(true);
  for(const p of [page,other]){
   await p.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(p.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
   await p.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Not authorised yet');
   for(const action of ['Add passkey','Rename Laptop','Remove Laptop'])await expect(p.getByRole('button',{name:action,exact:true})).toBeDisabled();
  }
  await page.getByRole('textbox',{name:'Reauthentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Verify code',exact:true}).click();await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();expect((await proof(page)).reauth_required).toBe(false);
  await page.getByRole('button',{name:'Rename Laptop',exact:true}).click();await page.getByRole('textbox',{name:'Rename passkey',exact:true}).fill('Verified laptop');await page.getByRole('button',{name:'Save passkey name',exact:true}).click();await expect(page.getByText('Name saved.',{exact:true})).toBeVisible();
  // A reload and a list/proof refresh in the second browser are not proof.
  await other.reload();await expect(other.locator('.compose-box textarea')).toBeVisible();await openAuthentication(other);
  await other.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Stale browser');
  for(const action of ['Add passkey','Rename Verified laptop','Remove Verified laptop'])await expect(other.getByRole('button',{name:action,exact:true})).toBeDisabled();
  expect(await proof(other)).toEqual(staleProof);
  const before=savedAuth(env);const id=(await list(other)).body.passkeys[0].id;
  for(const [path,body]of [['/register/start',{name:'Stale browser'}],['/rename',{id,name:'Unauthorised rename'}],['/remove',{id}]]){
   expect(await post(other,path,body)).toEqual({status:403,body:{error:'recent authentication required'}});expect(savedAuth(env)).toEqual(before);
  }
  // This browser can act only after its own accepted proof succeeds.
  await other.getByRole('textbox',{name:'Reauthentication code',exact:true}).fill(totp(env.secret));await other.getByRole('button',{name:'Verify code',exact:true}).click();await expect(other.getByText('Authentication verified.',{exact:true})).toBeVisible();
  for(const action of ['Add passkey','Rename Verified laptop','Remove Verified laptop'])await expect(other.getByRole('button',{name:action,exact:true})).toBeEnabled();
  await other.getByRole('button',{name:'Rename Verified laptop',exact:true}).click();await other.getByRole('textbox',{name:'Rename passkey',exact:true}).fill('Own proof laptop');await other.getByRole('button',{name:'Save passkey name',exact:true}).click();await expect(other.getByText('Name saved.',{exact:true})).toBeVisible();
  expect(savedAuth(env).passkeys).toEqual([{...before.passkeys[0],name:'Own proof laptop'}]);
 }finally{await otherContext.close();await auth.cdp.detach();await env.close();}
});

for(const operation of ['list','rename','remove'])for(const fault of ['server','network'])test(`Settings ${operation} ${fault} failure retains confirmed state until explicit native retry`,async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  const before=savedAuth(env);const id=(await list(page)).body.passkeys[0].id;const row=credentialRow(page,id);
  const path='/api/auth/passkeys'+(operation==='list'?'':`/${operation}`);const method=operation==='list'?'GET':'POST';let attempts=0;
  page.on('request',r=>{if(new URL(r.url()).pathname===path&&r.method()===method)attempts++;});
  // Both faults prevent delivery to the server. Lost responses after a commit
  // are a separate uncertainty case, not evidence from this fixture.
  await page.route(`**${path}`,route=>fault==='server'?route.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error:`${operation} unavailable`})}):route.abort('failed'));
  if(operation==='list')await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();
  if(operation==='rename'){
   await row.getByRole('button',{name:'Rename Laptop',exact:true}).click();await row.getByRole('textbox',{name:'Rename passkey',exact:true}).fill('Retry laptop');await row.getByRole('button',{name:'Save passkey name',exact:true}).click();
  }
  if(operation==='remove'){await row.getByRole('button',{name:'Remove Laptop',exact:true}).click();await page.getByRole('button',{name:'Confirm removal',exact:true}).click();}
  const alert=page.getByRole('alert');await expect(alert).toBeVisible();await expect(alert).not.toBeEmpty();if(fault==='server')await expect(alert).toHaveText(`${operation} unavailable`);
  for(const notice of ['Passkey registered.','Name saved.','Passkey removed. Existing login sessions are not signed out.'])await expect(page.getByText(notice,{exact:true})).toHaveCount(0);
  await expect(page.getByText('The displayed list is the last confirmed snapshot. Refresh before making changes.',{exact:true})).toBeVisible();await expect(page.getByText('No passkeys registered.',{exact:true})).toHaveCount(0);
  await expect(page.locator('.gi-passkey-row')).toHaveCount(1);await expect(row.locator('strong')).toHaveText('Laptop');await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Not authorised from stale list');
  for(const action of ['Add passkey','Rename Laptop','Remove Laptop'])await expect(page.getByRole('button',{name:action,exact:true})).toBeDisabled();
  await page.waitForTimeout(200);expect(attempts).toBe(1);expect(savedAuth(env)).toEqual(before);
  await page.unroute(`**${path}`);await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(row.getByRole('button',{name:'Rename Laptop',exact:true})).toBeEnabled();await expect(alert).toHaveCount(0);
  await expect(page.getByText('The displayed list is the last confirmed snapshot. Refresh before making changes.',{exact:true})).toHaveCount(0);expect(savedAuth(env)).toEqual(before);
  if(operation==='list'){expect(attempts).toBe(2);await expect(row.locator('strong')).toHaveText('Laptop');}
  if(operation==='rename'){
   expect(attempts).toBe(1);await expect(row.getByRole('textbox',{name:'Rename passkey',exact:true})).toHaveValue('Retry laptop');
   await row.getByRole('button',{name:'Save passkey name',exact:true}).click();await expect(page.getByText('Name saved.',{exact:true})).toBeVisible();await expect(row.locator('strong')).toHaveText('Retry laptop');
   expect(attempts).toBe(2);expect(savedAuth(env).passkeys).toEqual([{...before.passkeys[0],name:'Retry laptop'}]);
  }
  if(operation==='remove'){
   expect(attempts).toBe(1);await page.getByRole('button',{name:'Cancel removal',exact:true}).click();await row.getByRole('button',{name:'Remove Laptop',exact:true}).click();
   await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect(page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toBeVisible();await expect(page.locator('.gi-passkey-row')).toHaveCount(0);
   expect(attempts).toBe(2);expect(savedAuth(env).passkeys||[]).toEqual([]);
  }
  expect(savedAuth(env).sessions).toEqual(before.sessions);
 }finally{await auth.cdp.detach();await env.close();}
});

for(const operation of ['add','rename','remove'])test(`Settings cancelled reauthentication cannot authorise ${operation} before fresh proof`,async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await page.locator('.compose-box textarea').fill(`Keep ${operation} draft Ω`);await openAuthentication(page);await addFromSettings(page,'Laptop');
  // Whole-second fixture times retain exact representation through Go's
  // RFC3339 serialisation; equality must still catch real proof/expiry changes.
  const state=savedAuth(env);for(const s of state.sessions){s.created_at=new Date(Math.floor(Date.now()/1000)*1000-3600000).toISOString().replace('.000Z','Z');s.authenticated_at=new Date(Math.floor(Date.now()/1000)*1000-360000).toISOString().replace('.000Z','Z');}writeFileSync(env.authPath,JSON.stringify(state));
  const proof=()=>page.evaluate(async()=>(await(await fetch('/api/auth/session/proof')).json()));const stale=await proof();expect(stale.reauth_required).toBe(true);
  await page.reload();await expect(page.locator('.compose-box textarea')).toHaveValue(`Keep ${operation} draft Ω`);await openAuthentication(page);await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('After proof');const action=operation==='add'?'Add passkey':`${operation==='rename'?'Rename':'Remove'} Laptop`;
  await expect(page.getByRole('button',{name:action,exact:true})).toBeDisabled();expect(await proof()).toEqual(stale);
  const id=(await list(page)).body.passkeys[0].id;const path=operation==='add'?'/register/start':`/${operation}`;const body=operation==='add'?{name:'Before proof'}:{id,...(operation==='rename'?{name:'Before proof'}:{})};
  expect(await post(page,path,body)).toEqual({status:403,body:{error:'recent authentication required'}});expect(savedAuth(env)).toEqual(state);
  const changes=[];let starts=0,finishes=0;page.on('request',r=>{
   if(r.method()!=='POST')return;const path=new URL(r.url()).pathname;
   if(path==='/api/auth/passkeys/reauth/start')starts++;if(path==='/api/auth/passkeys/reauth/finish')finishes++;
   if(['/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish','/api/auth/passkeys/rename','/api/auth/passkeys/remove'].includes(path))changes.push(path);
  });
  // Observe entry without replacing the native credential promise/signature.
  await page.evaluate(()=>{const get=navigator.credentials.get.bind(navigator.credentials);window.__reauthGets=0;navigator.credentials.get=(...args)=>{window.__reauthGets++;return get(...args);};});
  const ceremonies=[];
  for(const [index,cancel]of ['button','Escape'].entries()){
   await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:false});
   const response=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/reauth/start')&&r.request().method()==='POST');await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();const started=await response;expect(started.status()).toBe(200);ceremonies.push((await started.json()).ceremony_id);
   await expect.poll(()=>page.evaluate(()=>window.__reauthGets)).toBe(index+1);await expect(page.getByRole('button',{name:'Cancel pending operation',exact:true})).toBeVisible();
   if(cancel==='button')await page.getByRole('button',{name:'Cancel pending operation',exact:true}).click();else await page.keyboard.press('Escape');
   await expect(page.getByRole('alert')).toContainText('cancelled');await expect(page.getByRole('dialog',{name:'Gi Settings',exact:true})).toBeVisible();await expect(page.getByText('Authentication verified.',{exact:true})).toHaveCount(0);await expect(page.getByRole('button',{name:'Verify with passkey',exact:true})).toBeFocused();
   // Presence after cancellation cannot complete the abandoned ceremony.
   await auth.cdp.send('WebAuthn.setAutomaticPresenceSimulation',{authenticatorId:auth.id,enabled:true});await page.waitForTimeout(150);
   expect(starts).toBe(index+1);expect(finishes).toBe(0);expect(changes).toEqual([]);expect(await proof()).toEqual(stale);expect(savedAuth(env).passkeys).toEqual(state.passkeys);expect(savedAuth(env).sessions).toEqual(state.sessions);
   await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();await expect(page.getByRole('button',{name:action,exact:true})).toBeDisabled();
   expect(await post(page,path,body)).toEqual({status:403,body:{error:'recent authentication required'}});changes.length=0;
  }
  const response=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/reauth/start')&&r.request().method()==='POST');await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();const started=await response;expect(started.status()).toBe(200);ceremonies.push((await started.json()).ceremony_id);
  await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();expect(new Set(ceremonies).size).toBe(3);expect(starts).toBe(3);expect(finishes).toBe(1);expect(changes).toEqual([]);await expect(page.getByRole('button',{name:action,exact:true})).toBeEnabled();
  expect((await proof()).reauth_required).toBe(false);expect((await proof()).expires_at).toBe(stale.expires_at);
  if(operation==='add'){await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'After proof');await expect(page.locator('.gi-passkey-row')).toHaveCount(2);expect(changes).toEqual(['/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish']);}
  if(operation==='rename'){await page.getByRole('button',{name:action,exact:true}).click();await page.getByRole('textbox',{name:'Rename passkey',exact:true}).fill('After proof');await page.getByRole('button',{name:'Save passkey name',exact:true}).click();await expect(page.getByText('Name saved.',{exact:true})).toBeVisible();expect(savedAuth(env).passkeys[0].name).toBe('After proof');expect(changes).toEqual(['/api/auth/passkeys/rename']);}
  if(operation==='remove'){await page.getByRole('button',{name:action,exact:true}).click();await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect(page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toBeVisible();expect(savedAuth(env).passkeys||[]).toEqual([]);expect(changes).toEqual(['/api/auth/passkeys/remove']);}
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue(`Keep ${operation} draft Ω`);
 }finally{await auth.cdp.detach();await env.close();}
});

test('Classic Settings shows native credential metadata without exposing authentication material',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Backup key');
  const stored=savedAuth(env);const inventory=await list(page);expect(inventory.status).toBe(200);expect(inventory.body.passkeys).toHaveLength(2);
  for(const [index,key]of inventory.body.passkeys.entries()){
   const row=credentialRow(page,key.id);await expect(row).toHaveCount(1);await expect(row.locator('strong')).toHaveText(key.name);
   expect(key.id).toBe(Buffer.from(stored.passkeys[index].credential.id,'base64').toString('base64url'));expect(Date.parse(key.created_at)).toBe(Date.parse(stored.passkeys[index].created_at));
   const date=await page.evaluate(value=>new Date(value).toLocaleString(),key.created_at);await expect(row.locator('small')).toContainText([`Identifier: ${key.id}`,`Created: ${date}`]);
   if(index===0){expect(Date.parse(key.last_used_at)).toBeGreaterThan(0);expect(Date.parse(key.last_used_at)).toBe(Date.parse(stored.passkeys[index].last_used_at));await expect(row.getByText(`Last used: ${await page.evaluate(value=>new Date(value).toLocaleString(),key.last_used_at)}`,{exact:true})).toBeVisible();}
   else{expect(!key.last_used_at||key.last_used_at.startsWith('0001-')).toBe(true);await expect(row.getByText('Last used: Never used',{exact:true})).toBeVisible();}
   await expect(row.getByRole('button',{name:`Rename ${key.name}`,exact:true})).toBeEnabled();await expect(row.getByRole('button',{name:`Remove ${key.name}`,exact:true})).toBeEnabled();
  }
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Visible add control');await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeEnabled();
  const cookie=(await context.cookies()).find(c=>c.name==='gi_session');expect(cookie.httpOnly).toBe(true);
  const snapshot=await page.evaluate(()=>JSON.stringify({html:document.documentElement.outerHTML,local:{...localStorage},session:{...sessionStorage},url:location.href,cookie:document.cookie}));
  // Actual nonempty canaries from this disposable owner and its credentials;
  // public credential identifiers intentionally remain visible.
  for(const secret of [env.secret,cookie.value,stored.webauthn_user_id,...stored.passkeys.map(k=>k.credential.publicKey),...stored.sessions.map(s=>s.token_hash)]){
   expect(typeof secret).toBe('string');expect(secret.length).toBeGreaterThan(10);expect(snapshot).not.toContain(secret);expect(JSON.stringify(inventory.body)).not.toContain(secret);
  }
  expect(new URL(page.url()).search).toBe('');expect(new URL(page.url()).hash).toBe('');await page.screenshot({path:info.outputPath('passkey-metadata.png')});
 }finally{await auth.cdp.detach();await env.close();}
});

for(const inputMode of ['keyboard','touch'])test(`Classic narrow passkeys avoid horizontal clipping and retain ${inputMode} cancellation ownership`,async({browser},info)=>{
 const context=await browser.newContext({viewport:{width:390,height:844},hasTouch:true});const page=await context.newPage();const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 const name='Long name '+'界'.repeat(70);expect([...name]).toHaveLength(80);
 const reach=async target=>{
  if(inputMode==='touch'){await target.tap();return;}
  // Use only Tab and Enter from the currently focused control. Do not focus
  // the target directly: this proves real reachability through the focus trap.
  for(let i=0;i<45;i++){if(await target.evaluate(el=>document.activeElement===el)){await page.keyboard.press('Enter');return;}await page.keyboard.press('Tab');}
  throw new Error('Control not reachable by Tab');
 };
 let release;const gate=new Promise(r=>release=r);
 try{
  await loginTOTP(page,env);await page.locator('.compose-box textarea').fill('Narrow authentication draft Ω');await openAuthentication(page);await addFromSettings(page,name);
  const before=savedAuth(env),key=(await list(page)).body.passkeys[0],row=credentialRow(page,key.id);const rename=row.getByRole('button',{name:`Rename ${name}`,exact:true}),remove=row.getByRole('button',{name:`Remove ${name}`,exact:true});
  const geometry=async()=>{
   expect(await page.evaluate(()=>document.documentElement.scrollWidth-innerWidth)).toBeLessThanOrEqual(1);expect(await page.locator('.settings-dialog').evaluate(el=>el.scrollWidth-el.clientWidth)).toBeLessThanOrEqual(1);
   for(const el of [row.locator('strong'),row.locator('small').nth(0),row.locator('small').nth(1),row.locator('small').nth(2),rename,remove]){
    const box=await el.boundingBox();expect(box.width).toBeGreaterThan(0);expect(box.x).toBeGreaterThanOrEqual(0);expect(box.x+box.width).toBeLessThanOrEqual(391);expect(await el.evaluate(e=>e.scrollWidth-e.clientWidth)).toBeLessThanOrEqual(1);
   }
  };
  await geometry();await expect(row.locator('strong')).toHaveText(name);
  // Inputs have actual labels; all enabled buttons have an accessible name.
  for(const [accessible,label]of [['New passkey name','Passkey name'],['Reauthentication code','Authentication code']]){
   const input=page.getByRole('textbox',{name:accessible,exact:true});await expect(input).toBeVisible();expect(await input.evaluate(el=>Array.from(el.labels||[],l=>l.textContent))).toEqual([label]);
  }
  for(const button of await page.locator('.gi-authentication-pane button').all())await expect(button).toHaveAccessibleName(/\S/);
  const writes=[];page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.startsWith('/api/auth/passkeys/'))writes.push(new URL(r.url()).pathname);});
  await reach(rename);await expect(page.getByRole('textbox',{name:'Rename passkey',exact:true})).toBeVisible();expect(await page.getByRole('textbox',{name:'Rename passkey',exact:true}).evaluate(el=>Array.from(el.labels||[],l=>l.textContent))).toEqual(['New name']);await reach(page.getByRole('button',{name:'Cancel rename',exact:true}));await expect(rename).toBeFocused();
  await reach(remove);await expect(page.getByRole('group',{name:'Confirm passkey removal'})).toContainText(key.id);await reach(page.getByRole('button',{name:'Cancel removal',exact:true}));await expect(remove).toBeFocused();expect(writes).toEqual([]);expect(savedAuth(env)).toEqual(before);
  // Status and alert semantics are observable; physical screen-reader speech
  // is not. Hold a read, move focus outside the pane, then surface a long error.
  let held=false;const error='Passkey inventory unavailable. '.repeat(5);
  await page.route('**/api/auth/passkeys',async route=>{held=true;await gate;await route.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error})});});
  await reach(page.getByRole('button',{name:'Refresh passkeys',exact:true}));await expect.poll(()=>held).toBe(true);await expect(page.getByRole('status',{exact:true}).filter({hasText:'Refreshing passkeys…'})).toBeVisible();
  const close=page.getByRole('button',{name:'Close settings',exact:true});
  if(inputMode==='keyboard'){for(let i=0;i<45&&!await close.evaluate(el=>document.activeElement===el);i++)await page.keyboard.press('Tab');await expect(close).toBeFocused();}else await close.focus();
  release();const alert=page.getByRole('alert');await expect(alert).toHaveText(error.trim());await expect(close).toBeFocused();await geometry();
  const errorBox=await alert.boundingBox();expect(errorBox.x).toBeGreaterThanOrEqual(0);expect(errorBox.x+errorBox.width).toBeLessThanOrEqual(391);expect(await alert.evaluate(el=>el.scrollWidth-el.clientWidth)).toBeLessThanOrEqual(1);
  expect(writes).toEqual([]);expect(savedAuth(env)).toEqual(before);
  await page.screenshot({path:info.outputPath(`passkey-narrow-${inputMode}.png`)});await reach(close);await expect(page.locator('.compose-box textarea')).toHaveValue('Narrow authentication draft Ω');
 }finally{release();await auth.cdp.detach();await env.close();await context.close();}
});

for(const missing of ['PublicKeyCredential','credentials','create','get','totp-only'])test(`Settings unavailable ${missing} never starts a credential operation and recovers explicitly`,async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Existing laptop');
  if(missing==='totp-only')await changePolicy(page,'totp-only');
  await page.getByRole('button',{name:'Close settings',exact:true}).click();
  const before=savedAuth(env);const writes=[];
  page.on('request',r=>{if(!['GET','HEAD','OPTIONS'].includes(r.method())&&new URL(r.url()).pathname.startsWith('/api/auth/'))writes.push(new URL(r.url()).pathname);});
  await page.evaluate(missing=>{
   const container=navigator.credentials,constructor=window.PublicKeyCredential;
   const ownContainer=Object.getOwnPropertyDescriptor(navigator,'credentials');
   const create=container.create.bind(container),get=container.get.bind(container);
   window.__capabilityCalls=[];
   const wrappedCreate=(...args)=>{window.__capabilityCalls.push('create');return create(...args);};
   const wrappedGet=(...args)=>{window.__capabilityCalls.push('get');return get(...args);};
   container.create=wrappedCreate;container.get=wrappedGet;
   window.__restoreCredentialAPI=()=>{
    window.PublicKeyCredential=constructor;
    if(ownContainer)Object.defineProperty(navigator,'credentials',ownContainer);else delete navigator.credentials;
    container.create=wrappedCreate;container.get=wrappedGet;
   };
   if(missing==='PublicKeyCredential')window.PublicKeyCredential=undefined;
   if(missing==='credentials')Object.defineProperty(navigator,'credentials',{configurable:true,value:undefined});
   if(missing==='create')container.create=undefined;
   if(missing==='get')container.get=undefined;
  },missing);
  await openAuthentication(page);
  const explanation=missing==='totp-only'?'Passkeys are disabled by policy or are not configured for this origin.':'This browser cannot use passkeys.';
  await expect(page.getByText(explanation,{exact:true})).toBeVisible();await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeDisabled();await expect(page.getByRole('textbox',{name:'New passkey name',exact:true})).toBeDisabled();
  await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Existing laptop']);
  for(const action of ['Rename Existing laptop','Remove Existing laptop'])await expect(page.getByRole('button',{name:action,exact:true})).toBeDisabled();
  if(missing==='totp-only')await expect(page.getByRole('button',{name:'Verify with passkey',exact:true})).toHaveCount(0);else await expect(page.getByRole('button',{name:'Verify with passkey',exact:true})).toBeDisabled();
  // Read-only refresh/re-entry must not turn a capability failure into a prompt
  // or drop an existing credential. Native policy is not mocked here.
  await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
  await page.getByRole('button',{name:'General',exact:true}).click();await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
  await expect(page.getByText(explanation,{exact:true})).toBeVisible();await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Existing laptop']);
  expect(await page.evaluate(()=>window.__capabilityCalls)).toEqual([]);expect(writes).toEqual([]);expect(savedAuth(env)).toEqual(before);
  await page.evaluate(()=>window.__restoreCredentialAPI());
  if(missing==='totp-only')await changePolicy(page,'either');else{await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();}
  await expect(page.getByText(explanation,{exact:true})).toHaveCount(0);
  expect(writes).toEqual(missing==='totp-only'?['/api/auth/policy']:[]);writes.length=0;
  // Same call observer now sees a genuine assertion and creation, using the
  // restored native APIs. This is not a mock success/always-zero counter.
  await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Recovered API key');
  expect(await page.evaluate(()=>window.__capabilityCalls)).toEqual(['get','create']);expect(writes).toEqual(['/api/auth/passkeys/reauth/start','/api/auth/passkeys/reauth/finish','/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish']);
  await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Existing laptop','Recovered API key']);
  expect(savedAuth(env).passkeys[0].credential.id).toBe(before.passkeys[0].credential.id);expect(savedAuth(env).passkeys[0].credential.publicKey).toBe(before.passkeys[0].credential.publicKey);
 }finally{await auth.cdp.detach();await env.close();}
});

for(const fault of ['expired','consumed','other-session','origin','rp','proof','revoked'])test(`Settings native registration ${fault} failure hides proof and retains usable credentials`,async({page,context,browser},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);let otherContext,otherAuth;
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Retained laptop');
  const original=savedAuth(env);const laptop=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  const cookie=(await context.cookies()).find(c=>c.name==='gi_session');const canaries=[env.secret,cookie.value,original.webauthn_user_id,original.passkeys[0].credential.publicKey];
  let foreign;
  if(fault==='other-session'){
   otherContext=await browser.newContext();const other=await otherContext.newPage();await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);await loginTOTP(other,env);
   otherAuth=await authenticator(other);foreign=await ceremony(other,'register',otherAuth,'Other browser');
   canaries.push(foreign.start.body.ceremony_id,foreign.start.body.options.publicKey.challenge,foreign.credential.response.attestationObject,foreign.credential.response.clientDataJSON);
  }
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  if(fault==='revoked'){
   // Hold delivery of a genuine created credential before the app can POST
   // finish. This is not a claim about a physical prompt remaining open.
   await page.evaluate(()=>{const create=navigator.credentials.create.bind(navigator.credentials);window.__registrationCreated=false;const held=new Promise(resolve=>window.__releaseRegistration=resolve);navigator.credentials.create=async(...args)=>{const credential=await create(...args);window.__registrationCreated=true;await held;return credential;};});
  }
  let nativeStatus,nativeBody,finishes=0;const sent=[];
  await page.route('**/api/auth/passkeys/register/finish',async route=>{
   finishes++;const body=route.request().postDataJSON();sent.push(body.ceremony_id);const modified=structuredClone(body);
   canaries.push(body.credential.response.attestationObject,body.credential.response.clientDataJSON,body.credential.id,body.credential.rawId);
   if(fault==='expired'){const state=savedAuth(env);state.webauthn_ceremonies.find(c=>c.id===body.ceremony_id).expires_at='2000-01-01T00:00:00Z';writeFileSync(env.authPath,JSON.stringify(state));}
   if(fault==='consumed'){
    const invalid=structuredClone(body);invalid.credential.response.attestationObject='AAAA';
    const consumed=await route.fetch({postData:JSON.stringify(invalid)});expect(consumed.status()).toBe(400);expect(savedAuth(env).passkeys).toEqual(original.passkeys);
   }
   if(fault==='other-session'){modified.ceremony_id=foreign.start.body.ceremony_id;modified.credential=foreign.credential;}
   if(fault==='origin'){
    const client=JSON.parse(Buffer.from(modified.credential.response.clientDataJSON,'base64url'));client.origin='https://evil.example';modified.credential.response.clientDataJSON=Buffer.from(JSON.stringify(client)).toString('base64url');
   }
   if(fault==='rp'){
    // Replace only the RP hash in a real, otherwise unchanged CBOR attestation.
    // Same length preserves CBOR structure; no malformed-blob shortcut.
    const rp=original.passkeys[0].rp_id;expect(rp).toBeTruthy();
    const attestation=Buffer.from(modified.credential.response.attestationObject,'base64url'),hash=createHash('sha256').update(rp).digest();
    const offset=attestation.indexOf(hash);expect(offset).toBeGreaterThan(0);expect(attestation.indexOf(hash,offset+1)).toBe(-1);
    createHash('sha256').update('evil.example').digest().copy(attestation,offset);modified.credential.response.attestationObject=attestation.toString('base64url');
   }
   if(fault==='proof')modified.credential.response.attestationObject='AAAA';
   const response=await route.fetch({postData:JSON.stringify(modified)});nativeStatus=response.status();nativeBody=await response.json();await route.fulfill({response});
  });
  const startResponse=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/register/start')&&r.request().method()==='POST');
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Rejected key');await page.getByRole('button',{name:'Add passkey',exact:true}).click();
  const start=await startResponse;expect(start.status()).toBe(200);const options=await start.json();canaries.push(options.ceremony_id,options.options.publicKey.challenge);
  if(fault==='revoked'){
   await expect.poll(()=>page.evaluate(()=>window.__registrationCreated)).toBe(true);const state=savedAuth(env);state.sessions=[];writeFileSync(env.authPath,JSON.stringify(state));
   await page.evaluate(()=>window.__releaseRegistration());
  }
  await expect(page.getByRole('alert')).toContainText('Not registered on the server');expect(nativeStatus).toBe(fault==='revoked'?401:400);expect(finishes).toBe(1);
  const error=fault==='revoked'?'browser owner sign-in required':['expired','consumed','other-session'].includes(fault)?'passkey ceremony expired, consumed or invalid':'passkey verification failed';
  expect(nativeBody).toEqual({error});await expect(page.getByRole('alert')).toContainText('local credential may remain');await expect(page.getByText('Passkey registered.',{exact:true})).toHaveCount(0);
  await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Retained laptop']);expect(savedAuth(env).passkeys).toEqual(original.passkeys);
  if(fault==='other-session'){
   // Neither the foreign binding nor the unused local ceremony is consumed.
   const ids=savedAuth(env).webauthn_ceremonies.map(c=>c.id);expect(ids).toContain(foreign.start.body.ceremony_id);expect(ids).toContain(options.ceremony_id);
  }
  else if(fault!=='revoked')expect((savedAuth(env).webauthn_ceremonies||[]).some(c=>c.id===options.ceremony_id)).toBe(false);
  const local=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials;expect(local).toHaveLength(1);canaries.push(local[0].privateKey);
  const surface=await page.evaluate(()=>JSON.stringify({html:document.documentElement.outerHTML,url:location.href,local:{...localStorage},session:{...sessionStorage}}));
  for(const value of canaries){expect(typeof value).toBe('string');expect(value.length).toBeGreaterThan(10);expect(surface).not.toContain(value);expect(JSON.stringify(nativeBody)).not.toContain(value);}
  await page.waitForTimeout(150);expect(finishes).toBe(1);expect(sent).toEqual([options.ceremony_id]);await page.unrouteAll({behavior:'wait'});
  // Every rejection retains a genuinely usable earlier key, not only a row.
  await context.clearCookies();await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:laptop});
  await page.reload();await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();await expect(page.locator('.compose-box textarea')).toBeVisible();await openAuthentication(page);await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Retained laptop']);
  expect(savedAuth(env).passkeys[0].credential.id).toBe(original.passkeys[0].credential.id);expect(savedAuth(env).passkeys[0].credential.publicKey).toBe(original.passkeys[0].credential.publicKey);
  if(fault==='other-session'){
   // Identical foreign proof is valid in its own owner session. A malformed
   // or mismatched challenge cannot make the cross-session rejection pass.
   expect((await foreign.finish()).status).toBe(200);expect(savedAuth(env).passkeys.map(k=>k.name)).toEqual(['Retained laptop','Other browser']);
  }
 }finally{await otherAuth?.cdp.detach();await otherContext?.close();await auth.cdp.detach();await env.close();}
});

test('Settings first passkey waits for native confirmation and retains fresh TOTP sign-in without chat enrolment',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);let release;const gate=new Promise(r=>release=r);
 try{
  await loginTOTP(page,env);const before=savedAuth(env);expect(before.passkeys||[]).toHaveLength(0);
  const messages=()=>page.evaluate(async id=>{const response=await fetch(`/api/sessions/${id}/messages`);if(!response.ok)throw new Error('Cannot read fixture messages');return response.json();},env.main.id);const chatBefore=await messages();
  const paths=[],urls=[];page.on('request',r=>{if(r.method()==='POST')paths.push(new URL(r.url()).pathname);if(r.isNavigationRequest())urls.push(r.url());});
  await page.locator('.compose-box textarea').fill('First passkey draft');await openAuthentication(page);let held=false;
  await page.route('**/api/auth/passkeys/register/finish',async route=>{held=true;await gate;await route.continue();});
  const startResponse=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/register/start'));await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Laptop');await page.getByRole('button',{name:'Add passkey',exact:true}).click();
  const started=await startResponse;expect(started.status()).toBe(200);const start=await started.json();await expect.poll(()=>held).toBe(true);
  await expect(page.locator('.gi-passkey-row')).toHaveCount(0);await expect(page.getByText('Passkey registered.',{exact:true})).toHaveCount(0);expect(savedAuth(env).passkeys||[]).toHaveLength(0);
  release();await expect(page.getByText('Passkey registered.',{exact:true})).toBeVisible();await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop']);
  const after=savedAuth(env);expect(after.totp_enabled).toBe(true);expect(after.totp_secret).toBe(before.totp_secret);expect(after.sessions).toEqual(before.sessions);expect(after.passkeys).toHaveLength(1);
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('First passkey draft');expect(await messages()).toEqual(chatBefore);
  expect(paths).toEqual(['/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish']);expect(new URL(page.url()).search).toBe('');expect(new URL(page.url()).hash).toBe('');
  for(const value of [env.secret,start.ceremony_id,start.options.publicKey.challenge]){expect(JSON.stringify(urls)).not.toContain(value);expect(JSON.stringify(await messages())).not.toContain(value);}
  // TOTP availability is tested with a fresh cookie-free sign-in, not only
  // the policy flag or an already authenticated session.
  await context.clearCookies();await page.reload();await expect(page.getByRole('button',{name:'Sign in with passkey',exact:true})).toBeVisible();
  await page.getByRole('textbox',{name:'Authentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('First passkey draft');
  await openAuthentication(page);await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop']);expect(savedAuth(env).passkeys).toEqual(after.passkeys);
 }finally{release();await auth.cdp.detach();await env.close();}
});

test('Settings adds a second distinct key with passkey-only proof and no configured TOTP',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');await changePolicy(page,'passkey-only');
  // No TOTP-removal UI is claimed: seed this precondition only in the private
  // disposable owner store, then obtain real fresh passkey proof by signing in.
  const state=savedAuth(env);state.totp_enabled=false;state.totp_secret='';writeFileSync(env.authPath,JSON.stringify(state));
  await context.clearCookies();await page.reload();await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toHaveCount(0);await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();await expect(page.locator('.compose-box textarea')).toBeVisible();
  await openAuthentication(page);await expect(page.getByRole('button',{name:'Verify code',exact:true})).toHaveCount(0);await expect(page.getByRole('textbox',{name:'Reauthentication code',exact:true})).toHaveCount(0);
  const before=savedAuth(env);expect(before.passkeys).toHaveLength(1);expect(before.totp_enabled).toBe(false);expect(before.totp_secret||'').toBe('');const key=(await list(page)).body.passkeys[0];const writes=[];
  page.on('request',r=>{if(r.method()==='POST')writes.push(new URL(r.url()).pathname);});
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Backup key');
  await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop','Backup key']);await expect(credentialRow(page,key.id).locator('strong')).toHaveText('Laptop');
  const after=savedAuth(env);expect(after.passkeys).toHaveLength(2);expect(after.passkeys[0]).toEqual(before.passkeys[0]);expect(after.passkeys[1].credential.id).not.toBe(before.passkeys[0].credential.id);expect(after.passkeys[1].credential.publicKey).not.toBe(before.passkeys[0].credential.publicKey);expect(after.sessions).toEqual(before.sessions);expect(after.login_policy).toBe('passkey-only');expect(after.totp_enabled).toBe(false);expect(after.totp_secret||'').toBe('');
  expect(writes).toEqual(['/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish']);await expect(page.locator('.gi-authentication-pane a')).toHaveCount(0);expect(new URL(page.url()).search).toBe('');
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await context.clearCookies();await page.reload();await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toHaveCount(0);await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();await expect(page.locator('.compose-box textarea')).toBeVisible();
  await openAuthentication(page);await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop','Backup key']);
 }finally{await auth.cdp.detach();await env.close();}
});

test('Settings confirmed removal rejects a fresh removed-key assertion and accepts only the surviving key',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);let release;const gate=new Promise(r=>release=r);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');const laptop=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Backup key');const backup=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  const before=savedAuth(env),keys=(await list(page)).body.passkeys;let removes=0,held=false;
  await page.route('**/api/auth/passkeys/remove',async route=>{removes++;expect(route.request().postDataJSON()).toEqual({id:keys[0].id});held=true;await gate;await route.continue();});
  await page.getByRole('button',{name:'Remove Laptop',exact:true}).click();await expect(page.getByRole('group',{name:'Confirm passkey removal'})).toContainText(`Laptop (${keys[0].id})`);expect(removes).toBe(0);
  await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect.poll(()=>held).toBe(true);await expect(page.locator('.gi-passkey-row')).toHaveCount(2);await expect(page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toHaveCount(0);expect(savedAuth(env)).toEqual(before);
  release();await expect(page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toBeVisible();await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Backup key']);expect(savedAuth(env).passkeys).toEqual([before.passkeys[1]]);expect(savedAuth(env).sessions).toEqual(before.sessions);expect(removes).toBe(1);await page.unrouteAll({behavior:'wait'});
  await context.clearCookies();await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:laptop});await page.reload();
  // Normal allowCredentials excludes the removed key. Alter only the options
  // sent to the authenticator so it signs a genuine assertion from that key;
  // the native server still uses its own surviving credential inventory.
  await page.route('**/api/auth/passkeys/login/start',async route=>{const response=await route.fetch();expect(response.status()).toBe(200);const body=await response.json();expect(body.options.publicKey.allowCredentials.map(c=>c.id)).toEqual([keys[1].id]);body.options.publicKey.allowCredentials=[{type:'public-key',id:keys[0].id}];await route.fulfill({response,json:body});});
  const rejected=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/login/finish'));await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();expect((await rejected).status()).toBe(400);await expect(page.getByRole('alert')).toBeVisible();await expect(page.locator('.compose-box textarea')).toHaveCount(0);expect((await context.cookies()).find(c=>c.name==='gi_session')).toBeUndefined();expect(await page.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(401);expect(savedAuth(env).passkeys).toEqual([before.passkeys[1]]);
  await page.unrouteAll({behavior:'wait'});await context.clearCookies();await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:backup});await page.reload();
  const accepted=page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/login/finish'));await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();expect((await accepted).status()).toBe(200);await expect(page.locator('.compose-box textarea')).toBeVisible();await openAuthentication(page);await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Backup key']);
  const surviving=savedAuth(env).passkeys;expect(surviving).toHaveLength(1);expect(surviving[0].name).toBe(before.passkeys[1].name);expect(surviving[0].rp_id).toBe(before.passkeys[1].rp_id);expect(surviving[0].credential.id).toBe(before.passkeys[1].credential.id);expect(surviving[0].credential.publicKey).toBe(before.passkeys[1].credential.publicKey);
 }finally{release();await auth.cdp.detach();await env.close();}
});

for(const returnVia of ['pane switch','Settings reopen'])test(`Settings ${returnVia} reconciles a lost registration response without replay`,async({page},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 let release;const readGate=new Promise(resolve=>release=resolve);
 try{
  await loginTOTP(page,env);const composer=page.locator('.compose-box textarea');await composer.fill('Uncertain enrolment keeps this draft Ω');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'reconcile.txt',mimeType:'text/plain',buffer:Buffer.from('Retained uncertain-enrolment bytes')});await expect(page.locator('.compose-file-pill[title="reconcile.txt"]')).toBeVisible();
  await openAuthentication(page);await addFromSettings(page,'Laptop');const original=savedAuth(env).passkeys[0];
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});
  const writes=[];let committed;
  page.on('request',request=>{if(request.method()==='POST'&&new URL(request.url()).pathname.startsWith('/api/auth/'))writes.push(new URL(request.url()).pathname);});
  await page.route('**/api/auth/passkeys/register/finish',async route=>{
   const response=await route.fetch();expect(response.status()).toBe(200);expect(await response.json()).toEqual({ok:true});committed=savedAuth(env);await route.abort('failed');
  });
  await page.getByRole('textbox',{name:'New passkey name',exact:true}).fill('Uncertain key');await page.getByRole('button',{name:'Add passkey',exact:true}).click();
  await expect(page.getByRole('alert')).toContainText('could not be confirmed');await expect(page.getByText('Passkey registered.',{exact:true})).toHaveCount(0);await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop']);
  expect(committed.passkeys.map(k=>k.name)).toEqual(['Laptop','Uncertain key']);expect(committed.passkeys[0]).toEqual(original);
  const newID=Buffer.from(committed.passkeys[1].credential.id,'base64').toString('base64url');
  expect((await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials).toHaveLength(1);
  // Leave before installing the held inventory read. Never click Refresh: a
  // fresh pane must ask the native server itself, not reuse the prior snapshot.
  if(returnVia==='pane switch'){await page.getByRole('button',{name:'General',exact:true}).click();await expect(page.getByRole('heading',{name:'General',exact:true})).toBeVisible();}
  else{await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(composer).toHaveValue('Uncertain enrolment keeps this draft Ω');}
  await expect(page.locator('.gi-authentication-pane')).toHaveCount(0);let reads=0;
  await page.route('**/api/auth/passkeys',async route=>{expect(route.request().method()).toBe('GET');reads++;await readGate;await route.continue();});
  if(returnVia==='Settings reopen'){await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Settings',exact:true}).click();}
  await page.getByRole('button',{name:'Authentication',exact:true}).click();await expect.poll(()=>reads).toBe(1);
  await expect(page.getByRole('status').filter({hasText:'Loading passkeys…'})).toBeVisible();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeDisabled();
  await expect(page.locator('.gi-passkey-row')).toHaveCount(0);await expect(page.getByText('No passkeys registered.',{exact:true})).toHaveCount(0);await expect(page.getByText('Passkey registered.',{exact:true})).toHaveCount(0);
  expect(writes).toEqual(['/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish']);expect(savedAuth(env)).toEqual(committed);
  release();await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop','Uncertain key']);await expect(credentialRow(page,newID)).toHaveCount(1);await expect(page.getByRole('alert')).toHaveCount(0);
  expect(reads).toBe(1);expect(savedAuth(env)).toEqual(committed);expect(writes).toEqual(['/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish']);
  // A second return and full-page reload are further reconciliation boundaries,
  // not timing-only observations. Neither may replay the consumed ceremony.
  await page.getByRole('button',{name:'General',exact:true}).click();await expect(page.locator('.gi-authentication-pane')).toHaveCount(0);await page.getByRole('button',{name:'Authentication',exact:true}).click();
  await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();await expect(credentialRow(page,newID)).toHaveCount(1);expect(reads).toBe(2);
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await page.reload();await expect(composer).toHaveValue('Uncertain enrolment keeps this draft Ω');await expect(page.locator('.compose-file-pill[title="reconcile.txt"]')).toBeVisible();
  await openAuthentication(page);await expect(page.locator('.gi-passkey-row strong')).toHaveText(['Laptop','Uncertain key']);await expect(credentialRow(page,newID)).toHaveCount(1);expect(reads).toBe(3);
  expect(savedAuth(env)).toEqual(committed);expect(writes).toEqual(['/api/auth/passkeys/register/start','/api/auth/passkeys/register/finish']);
 }finally{release();await page.unrouteAll({behavior:'wait'});await auth.cdp.detach();await env.close();}
});

async function changePolicy(page,value){
 await page.getByRole('combobox',{name:'Accepted sign-in methods',exact:true}).selectOption(value);
 await page.getByRole('button',{name:'Change sign-in policy',exact:true}).click();
 await page.getByRole('button',{name:'Confirm policy change',exact:true}).click();
 await expect(page.getByText(`Current policy: ${value}. Changing accepted factors does not remove credentials or sign out existing sessions.`,{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
}

test('Settings changes accepted factors without lockout and reauthenticates under the current policy',async({page,context},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);
 try{
  await loginTOTP(page,env);await page.locator('.compose-box textarea').fill('Policy changes retain this draft');await openAuthentication(page);
  await page.getByRole('combobox',{name:'Accepted sign-in methods',exact:true}).selectOption('passkey-only');await expect(page.getByRole('button',{name:'Change sign-in policy',exact:true})).toBeDisabled();
  await page.getByRole('combobox',{name:'Accepted sign-in methods',exact:true}).selectOption('either');await addFromSettings(page,'Laptop');
  const before=JSON.parse(readFileSync(env.authPath,'utf8'));const cookies=await context.cookies();
  await page.getByRole('combobox',{name:'Accepted sign-in methods',exact:true}).selectOption('passkey-only');await page.getByRole('button',{name:'Change sign-in policy',exact:true}).click();await page.getByRole('button',{name:'Cancel policy change',exact:true}).click();
  expect(JSON.parse(readFileSync(env.authPath,'utf8')).login_policy||'either').toBe('either');await expect(page.getByRole('button',{name:'Change sign-in policy',exact:true})).toBeFocused();
  await changePolicy(page,'passkey-only');expect(await context.cookies()).toEqual(cookies);await expect(page.getByRole('button',{name:'Verify code',exact:true})).toHaveCount(0);
  expect(JSON.parse(readFileSync(env.authPath,'utf8')).sessions).toEqual(before.sessions);
  // The former TOTP proof cannot manage a passkey-only account.
  await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeDisabled();await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();
  await page.getByRole('button',{name:'Remove Laptop',exact:true}).click();await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect(page.getByRole('alert')).toContainText('another accepted sign-in method');await expect(page.locator('.gi-passkey-row')).toHaveCount(1);await page.getByRole('button',{name:'Cancel removal',exact:true}).click();await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();
  await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();await changePolicy(page,'totp-only');
  // The pane remains usable with passkey actions disabled by policy.
  await expect(page.getByRole('button',{name:'Verify code',exact:true})).toBeVisible();await expect(page.getByRole('button',{name:'Add passkey',exact:true})).toBeDisabled();await expect(page.locator('.gi-passkey-row')).toHaveCount(1);await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeDisabled();
  await page.getByRole('textbox',{name:'Reauthentication code',exact:true}).fill(totp(env.secret));await page.getByRole('button',{name:'Verify code',exact:true}).click();await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();
  await changePolicy(page,'either');expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys[0].credential.publicKey).toEqual(before.passkeys[0].credential.publicKey);
  await page.getByRole('button',{name:'Close settings',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('Policy changes retain this draft');
  await env.restart();await context.clearCookies();await page.reload();await expect(page.getByRole('button',{name:'Sign in with passkey',exact:true})).toBeVisible();await expect(page.getByRole('textbox',{name:'Authentication code',exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();await expect(page.locator('.compose-box textarea')).toHaveValue('Policy changes retain this draft');
 }finally{await auth.cdp.detach();await env.close();}
});

test('Two Settings views concurrently remove different keys without deleting the last sign-in factor',async({page,browser},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);const otherContext=await browser.newContext();const other=await otherContext.newPage();const otherAuth=await authenticator(other);
 let release;const gate=new Promise(resolve=>release=resolve);
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  const laptop=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await addFromSettings(page,'Backup key');
  const backup=(await auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:auth.id})).credentials[0];
  await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);await loginTOTP(other,env);
  await changePolicy(page,'passkey-only');
  await auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:auth.id});await auth.cdp.send('WebAuthn.addCredential',{authenticatorId:auth.id,credential:laptop});
  await otherAuth.cdp.send('WebAuthn.addCredential',{authenticatorId:otherAuth.id,credential:backup});await openAuthentication(other);
  for(const p of [page,other]){
   await p.getByRole('button',{name:'Verify with passkey',exact:true}).click();await expect(p.getByText('Authentication verified.',{exact:true})).toBeVisible();await expect(p.locator('.gi-passkey-row')).toHaveCount(2);
  }
  const before=savedAuth(env);expect(before.login_policy).toBe('passkey-only');const keys=(await list(page)).body.passkeys;
  const cookies=await page.context().cookies(),otherCookies=await otherContext.cookies();
  const views=[{page,auth,index:0},{page:other,auth:otherAuth,index:1}];const requests=[];
  for(const view of views){
   await view.page.route('**/api/auth/passkeys/remove',async route=>{requests.push({view:view.index,body:route.request().postDataJSON()});await gate;await route.continue();});
   // Each browser uses the credential it attempts to remove: the loser's
   // proof remains valid for a later explicit retry, independent of ordering.
   await credentialRow(view.page,keys[view.index].id).getByRole('button',{name:`Remove ${keys[view.index].name}`,exact:true}).click();
  }
  const responses=Promise.all(views.map(view=>view.page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/remove')&&r.request().method()==='POST')));
  await Promise.all(views.map(view=>view.page.getByRole('button',{name:'Confirm removal',exact:true}).click()));
  await expect.poll(()=>requests.length).toBe(2);expect(requests.sort((a,b)=>a.view-b.view)).toEqual(keys.map((key,index)=>({view:index,body:{id:key.id}})));
  expect(savedAuth(env)).toEqual(before);for(const view of views){await expect(view.page.locator('.gi-passkey-row')).toHaveCount(2);await expect(view.page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toHaveCount(0);}
  release();const finished=await responses;expect(finished.map(r=>r.status()).sort()).toEqual([200,409]);
  const winner=views[finished.findIndex(r=>r.status()===200)],loser=views[finished.findIndex(r=>r.status()===409)];
  // The non-blocking state-file lock may reject simultaneous writers before
  // the factor check. A serialised arrival instead reaches last-factor safety.
  // Keep those outcomes distinct; require last-factor refusal on explicit retry.
  const failure=await finished[loser.index].json();expect([{error:'Authentication state changed; retry'},{error:'add another accepted sign-in method before removing this passkey'}]).toContainEqual(failure);
  await expect(winner.page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toBeVisible();await expect(winner.page.locator('.gi-passkey-row')).toHaveCount(1);
  await expect(loser.page.getByRole('alert')).toHaveText(failure.error);await expect(loser.page.getByText('Passkey removed. Existing login sessions are not signed out.',{exact:true})).toHaveCount(0);
  await expect(loser.page.getByText('The displayed list is the last confirmed snapshot. Refresh before making changes.',{exact:true})).toBeVisible();await expect(loser.page.locator('.gi-passkey-row')).toHaveCount(2);
  const survivor=keys[loser.index];expect(savedAuth(env).passkeys).toEqual([before.passkeys[loser.index]]);expect(savedAuth(env).sessions).toEqual(before.sessions);
  expect(await page.context().cookies()).toEqual(cookies);expect(await otherContext.cookies()).toEqual(otherCookies);
  for(const view of views){
   await view.page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(view.page.getByRole('button',{name:'Refresh passkeys',exact:true})).toBeEnabled();
   await expect(view.page.locator('.gi-passkey-row')).toHaveCount(1);await expect(credentialRow(view.page,survivor.id).locator('strong')).toHaveText(survivor.name);await expect(view.page.getByRole('alert')).toHaveCount(0);
   expect((await list(view.page)).body.passkeys).toEqual([survivor]);expect(await view.page.evaluate(async()=>(await fetch('/api/sessions')).status)).toBe(200);
  }
  expect(requests).toHaveLength(2); // refresh never retries a removal
  await loser.page.getByRole('button',{name:`Remove ${survivor.name}`,exact:true}).click();
  const retry=loser.page.waitForResponse(r=>r.url().endsWith('/api/auth/passkeys/remove')&&r.request().method()==='POST');
  await loser.page.getByRole('button',{name:'Confirm removal',exact:true}).click();const refused=await retry;
  expect(refused.status()).toBe(409);expect(await refused.json()).toEqual({error:'add another accepted sign-in method before removing this passkey'});
  await expect(loser.page.getByRole('alert')).toContainText('another accepted sign-in method');expect(requests).toHaveLength(3);
  expect(savedAuth(env).passkeys).toEqual([before.passkeys[loser.index]]);expect(savedAuth(env).sessions).toEqual(before.sessions);
  // Copy the current virtual counter, not the pre-reauth snapshot, into the
  // removed-key browser and prove a genuinely fresh login with the survivor.
  const credential=(await loser.auth.cdp.send('WebAuthn.getCredentials',{authenticatorId:loser.auth.id})).credentials[0];
  await winner.auth.cdp.send('WebAuthn.clearCredentials',{authenticatorId:winner.auth.id});await winner.auth.cdp.send('WebAuthn.addCredential',{authenticatorId:winner.auth.id,credential});
  await winner.page.context().clearCookies();await winner.page.reload();await expect(winner.page.getByRole('textbox',{name:'Authentication code',exact:true})).toHaveCount(0);
  await winner.page.getByRole('button',{name:'Sign in with passkey',exact:true}).click();await expect(winner.page.locator('.compose-box textarea')).toBeVisible();await openAuthentication(winner.page);
  await expect(winner.page.locator('.gi-passkey-row')).toHaveCount(1);await expect(credentialRow(winner.page,survivor.id).locator('strong')).toHaveText(survivor.name);
 }finally{release();await otherAuth.cdp.detach();await otherContext.close();await auth.cdp.detach();await env.close();}
});

test('Policy changes reject stale browser revisions and protect a pending last-key removal',async({page,browser},info)=>{
 const env=await authEnvironment(page,info,{passkeys:true});const auth=await authenticator(page);const otherContext=await browser.newContext();const other=await otherContext.newPage();
 try{
  await loginTOTP(page,env);await openAuthentication(page);await addFromSettings(page,'Laptop');
  await other.addInitScript(id=>localStorage.setItem('gi_session_id',id),env.main.id);await loginTOTP(other,env);await openAuthentication(other);
  await other.getByRole('combobox',{name:'Accepted sign-in methods',exact:true}).selectOption('totp-only');await other.getByRole('button',{name:'Change sign-in policy',exact:true}).click();
  await changePolicy(page,'totp-only');await changePolicy(page,'either');
  await other.getByRole('button',{name:'Confirm policy change',exact:true}).click();await expect(other.getByRole('alert')).toContainText('state changed');expect(JSON.parse(readFileSync(env.authPath,'utf8')).login_policy).toBe('either');
  await other.getByRole('button',{name:'Cancel policy change',exact:true}).click();await other.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(other.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();
  await page.getByRole('button',{name:'Verify with passkey',exact:true}).click();await expect(page.getByText('Authentication verified.',{exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Remove Laptop',exact:true}).click();
  await changePolicy(other,'passkey-only');
  await page.getByRole('button',{name:'Confirm removal',exact:true}).click();await expect(page.getByRole('alert')).toContainText('another accepted sign-in method');await expect(page.locator('.gi-passkey-row')).toHaveCount(1);
  expect(JSON.parse(readFileSync(env.authPath,'utf8')).passkeys).toHaveLength(1);
  await page.getByRole('button',{name:'Cancel removal',exact:true}).click();await page.getByRole('button',{name:'Refresh passkeys',exact:true}).click();await expect(page.getByRole('button',{name:'Remove Laptop',exact:true})).toBeEnabled();
  // A failed policy write never displays optimistic acceptance.
  await page.route('**/api/auth/policy',route=>route.request().method()==='POST'?route.fulfill({status:503,contentType:'application/json',body:'{"error":"Policy unavailable"}'}):route.continue());
  await page.getByRole('combobox',{name:'Accepted sign-in methods',exact:true}).selectOption('either');await page.getByRole('button',{name:'Change sign-in policy',exact:true}).click();await page.getByRole('button',{name:'Confirm policy change',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Policy unavailable');expect(JSON.parse(readFileSync(env.authPath,'utf8')).login_policy).toBe('passkey-only');
  await page.unroute('**/api/auth/policy');
 }finally{await otherContext.close();await auth.cdp.detach();await env.close();}
});
