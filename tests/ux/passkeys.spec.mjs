import {test,expect} from '@playwright/test';
import {readFileSync,writeFileSync} from 'node:fs';
import {createHash,createPrivateKey,sign} from 'node:crypto';
import {authEnvironment,totp} from './support/auth-environment.mjs';

// Real browser credential API and signatures, native Go HTTP/crypto/storage.
// Direct API integration only: no Settings/native physical prompt parity credit.
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
