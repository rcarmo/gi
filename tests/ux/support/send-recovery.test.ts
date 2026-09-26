import { test, expect } from 'bun:test';
import { recoverSubmittedPrompt } from '../../../web/src/gi-send-recovery';
const turn={id:'turn-1',session_id:'A',status:'completed',metadata:{client_request_id:'token'}};
const event={type:'turn.submitted',turn_id:'turn-1',session_id:'A'};

test('recover lost reply using exact native identity and durable post-rollback event',async()=>{
 const paths:string[]=[];
 const result=await recoverSubmittedPrompt('A','token',async path=>{paths.push(path);return path.endsWith('/turns')?{turns:[turn]}:{events:[event]}});
 expect(result).toEqual({turn_id:'turn-1',session_id:'A',status:'completed',queued:false});
 expect(paths).toEqual(['/api/sessions/A/turns','/api/turns/turn-1/events']);
});
test('absent, duplicate, foreign, malformed and unaudited admission remains unknown',async()=>{
 for(const turns of [null,[],[turn,turn],[{...turn,session_id:'B'}],[{...turn,metadata:{client_request_id:'other'}}],[{...turn,id:null}]]){
  expect(await recoverSubmittedPrompt('A','token',async()=>({turns,events:[event]}))).toBeNull();
 }
 for(const events of [null,[],[{...event,turn_id:'other'}],[{...event,session_id:'B'}],[{...event,type:'turn.started'}]]){
  expect(await recoverSubmittedPrompt('A','token',async path=>path.endsWith('/turns')?{turns:[turn]}:{events})).toBeNull();
 }
});
test('receipt read failure never invents success',async()=>{
 expect(await recoverSubmittedPrompt('A','token',async()=>{throw new TypeError('offline')})).toBeNull();
 expect(await recoverSubmittedPrompt('A','token',async path=>{if(path.endsWith('/turns'))return {turns:[turn]};throw new DOMException('timed out','TimeoutError')})).toBeNull();
});

test('startup recovery is capped at six captures, caches session snapshot, preserves session identity',async()=>{
 const {recoverPendingSends}=await import('../../../web/src/gi-send-recovery');
 const {pendingSendKey}=await import('../../../web/src/gi-drafts');
 const pending=Array.from({length:12},(_,i)=>({sessionId:'A',token:'token'+i}));let snapshots=0,audits=0,active=0,maxActive=0;
 const result=await recoverPendingSends(pending,async path=>{
  active++;maxActive=Math.max(maxActive,active);await new Promise(r=>setTimeout(r,1));active--;
  if(path.endsWith('/turns')){snapshots++;return{turns:pending.map((p,i)=>({id:'turn-'+i,session_id:p.sessionId,status:'completed',metadata:{client_request_id:p.token}}))}}
  audits++;const id=path.split('/')[3];return{events:[{type:'turn.submitted',turn_id:id,session_id:'A'}]};
 });
 expect(snapshots).toBe(1);expect(audits).toBe(6);expect(maxActive).toBeLessThanOrEqual(2);expect(result.size).toBe(6);expect(result.has(pendingSendKey('A','token0'))).toBe(true);expect(result.has(pendingSendKey('A','token6'))).toBe(false);
});
