import {expect} from '@playwright/test';
import {spawn} from 'node:child_process';
import {mkdtempSync,rmSync,createWriteStream,writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
import {createServer} from 'node:net';

// Empty database and browser storage: no API-created session, owner, localStorage
// selection, or replaced runtime/bootstrap responses. Only inference is local.
export async function journeyEnvironment(info){
 const dir=mkdtempSync(join(tmpdir(),'gi-journey-'));
 const reserve=createServer();await new Promise(r=>reserve.listen(0,'127.0.0.1',r));const port=reserve.address().port;await new Promise(r=>reserve.close(r));
 const origin=`http://127.0.0.1:${port}`,log=createWriteStream(info.outputPath('server.log'));
 const child=spawn(resolve(process.env.GI_UX_SERVER_BIN||'bin/gi-ux-steer'),[],{env:{...process.env,GI_UX_STATE_DIR:dir,GI_UX_LISTEN:`127.0.0.1:${port}`,GI_UX_QUEUE_GATES:dir},stdio:['ignore','pipe','pipe']});
 child.stdout.pipe(log,{end:false});child.stderr.pipe(log,{end:false});
 const close=async()=>{if(child.exitCode===null){const done=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await done;}log.end();rmSync(dir,{recursive:true,force:true});};
 try{
  await expect.poll(async()=>{if(child.exitCode!==null)throw new Error(`Journey fixture exited ${child.exitCode}`);try{return(await fetch(origin+'/api/runtime/config')).status;}catch{return 0;}},{timeout:10000}).toBe(200);
  expect((await(await fetch(origin+'/api/sessions')).json()).sessions).toEqual([]);
  return {origin,release:token=>writeFileSync(join(dir,token),'ready'),close};
 }catch(error){await close();throw error;}
}
