import { test, expect } from 'bun:test';
import { recoverSubmittedPrompt, recoverPendingSends } from '../../../web/src/gi-send-recovery';
import { pendingSendKey } from '../../../web/src/gi-drafts';
const result={turn_id:'turn-1',session_id:'A',status:'running',queued:false};
const receipt={confirmed:true,source_session_id:'A',client_request_id:'token',result};
test('recover source-owned native local, routed and steered successful reply without history scans',async()=>{
 for(const value of [result,{...result,session_id:'B',source_session_id:'A',routed:true,target_agent_id:'peer',created_session:true}]){
  const paths:string[]=[];expect(await recoverSubmittedPrompt('A','token',async path=>{paths.push(path);return{...receipt,result:value}})).toEqual(value);
  expect(paths).toEqual(['/api/sessions/A/send-receipt?client_request_id=token']);
 }
});
test('unknown, foreign, mismatched or malformed receipts remain unknown',async()=>{
 for(const value of [null,{}, {...receipt,confirmed:false},{...receipt,source_session_id:'B'},{...receipt,client_request_id:'other'},{...receipt,result:{...result,turn_id:''}},{...receipt,result:{...result,session_id:'B'}},{...receipt,result:{...result,session_id:'B',source_session_id:'A',routed:false}}]){
  expect(await recoverSubmittedPrompt('A','token',async()=>value)).toBeNull();
 }
 expect(await recoverSubmittedPrompt('A','token',async()=>{throw Error('offline')})).toBeNull();
});
test('startup recovery uses at most six bounded reads and two workers',async()=>{
 const pending=Array.from({length:12},(_,i)=>({sessionId:'A',token:'token'+i}));let reads=0,active=0,max=0;
 const got=await recoverPendingSends(pending,async path=>{reads++;active++;max=Math.max(max,active);await new Promise(r=>setTimeout(r,1));active--;const token=new URL(path,'http://test').searchParams.get('client_request_id');return{...receipt,client_request_id:token}});
 expect(reads).toBe(6);expect(max).toBeLessThanOrEqual(2);expect(got.size).toBe(6);expect(got.has(pendingSendKey('A','token0'))).toBe(true);expect(got.has(pendingSendKey('A','token6'))).toBe(false);
});
