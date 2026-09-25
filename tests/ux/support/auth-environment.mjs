import {expect} from '@playwright/test';
import {spawn} from 'node:child_process';
import {mkdtempSync,mkdirSync,rmSync,createWriteStream,writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
import {createServer} from 'node:http';
import {createHmac} from 'node:crypto';

export function totp(secret) {
 const alphabet='ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';let bits='';
 for(const char of secret.toUpperCase().replace(/=+$/,''))bits+=alphabet.indexOf(char).toString(2).padStart(5,'0');
 const key=Buffer.from((bits.match(/.{8}/g)||[]).map(byte=>parseInt(byte,2)));
 const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)));
 const digest=createHmac('sha1',key).update(counter).digest(),offset=digest.at(-1)&15;
 return String((digest.readUInt32BE(offset)&0x7fffffff)%1000000).padStart(6,'0');
}

// One native process/workspace per case; no production auth files or database.
export async function authEnvironment(page,info,{passkeys=false}={}) {
 const dir=mkdtempSync(join(tmpdir(),'gi-auth-'));
 const reserve=createServer();await new Promise(r=>reserve.listen(0,'127.0.0.1',r));const port=reserve.address().port;await new Promise(r=>reserve.close(r));
 const origin=`http://${passkeys?'localhost':'127.0.0.1'}:${port}`;
 if(passkeys){mkdirSync(join(dir,'.pi'),{recursive:true});writeFileSync(join(dir,'.pi','settings.json'),JSON.stringify({passkeys:{rp_id:'localhost',origins:[origin]}}));}
 mkdirSync(resolve('test-results/ux-parity'),{recursive:true});
 const log=createWriteStream(resolve('test-results/ux-parity',`auth-${info.project.name}-${Date.now()}.log`));
 const start=()=>{const p=spawn(resolve(process.env.GI_UX_SERVER_BIN||'bin/gi-ux-steer'),[],{env:{...process.env,GI_UX_STATE_DIR:dir,GI_UX_LISTEN:`127.0.0.1:${port}`,GI_UX_QUEUE_GATES:dir},stdio:['ignore','pipe','pipe']});p.stdout.pipe(log,{end:false});p.stderr.pipe(log,{end:false});return p;};
 let child=start();
 const stop=async()=>{if(child.exitCode===null){const done=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await done;}};
 const ready=()=>expect.poll(async()=>{try{return(await fetch(origin+'/api/auth/status')).status}catch{return 0}},{timeout:10000}).toBe(200);
 const restart=async()=>{await stop();child=start();await ready();};
 const close=async()=>{await page.close();await stop();log.end();rmSync(dir,{recursive:true,force:true});};
 try {
  await expect.poll(async()=>{try{return(await fetch(origin+'/api/auth/status')).status}catch{return 0}},{timeout:10000}).toBe(200);
  const api=async(path,body)=>{const response=await fetch(origin+path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});expect(response.ok).toBe(true);return response.json();};
  const main=await api('/api/sessions',{agent_id:'auth-browser',title:'Auth fixture'});
  const pending=await api('/api/auth/enroll/start',{username:'admin'});
  await api('/api/auth/enroll/verify',{username:'admin',code:totp(pending.secret)});
  await page.addInitScript(id=>{
   localStorage.setItem('gi_session_id',id);
   const Native=window.EventSource;window.__authConnected=0;
   window.EventSource=class extends Native{constructor(url,options){super(url,options);this.addEventListener('connected',()=>window.__authConnected++);}};
  },main.id);
  return {origin,main,secret:pending.secret,authPath:join(dir,'.gi','auth.json'),restart,close};
 } catch(error){await close();throw error;}
}
