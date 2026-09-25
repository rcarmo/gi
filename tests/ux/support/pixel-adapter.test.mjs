import {test,expect} from 'bun:test';
import {EventEmitter} from 'node:events';
import {installPixelHost} from './pixel-adapter.mjs';
import state from '../fixtures/compose-pixel-state.json';
import reference from '../fixtures/compose-pixel-reference.json';
class Page extends EventEmitter{
 async addInitScript(){} async route(_,handler){this.handler=handler;} async unroute(){this.handler=null;}
 async request(origin,path,method='GET'){
  let response,continued=false;
  await this.handler({request:()=>({url:()=>origin+path,method:()=>method}),fulfill:async value=>{response=value;},continue:async()=>{continued=true;}});
  return {response,continued};
 }
}
async function fixture(host,run){const page=new Page(),adapter=await installPixelHost({page,host,root:process.cwd(),state,reference});try{await run(page,adapter);}finally{await adapter.dispose();expect(page.listenerCount('pageerror')).toBe(0);expect(page.listenerCount('requestfailed')).toBe(0);expect(page.handler).toBeNull();}}
test('native snapshot has array errors and equal model/session identities',()=>fixture('piclaw',async(page,a)=>{
 const {response}=await page.request(a.origin,'/agent/status?chat_jid=gi%3Amain&ui=1');
 expect(response.json.errors).toEqual([]);expect(response.json.model).toEqual(state.model);expect(response.json.agent_name).toBe(state.agentName);
 const sessions=await page.request(a.origin,'/agent/active-chats');expect(sessions.response.json.chats[0].agent_name).toBe(state.sessionLabel);expect(sessions.response.json.chats[0].model).toBe(state.model.current);
 expect(a.failures).toEqual([]);
}));
test('undeclared writes, origins and wrong session scopes fail closed',()=>fixture('gi',async(page,a)=>{
 await page.request(a.origin,'/api/sessions','POST');await page.request('https://example.invalid','/api/sessions');await page.request(a.origin,'/api/sessions?chat_jid=wrong');
 expect(a.failures).toHaveLength(3);expect(()=>a.assert()).toThrow();
}));
test('page errors and failed network requests are capture failures',()=>fixture('gi',async(page,a)=>{
 page.emit('pageerror',Error('boom'));page.emit('requestfailed',{url:()=>a.origin+'/dist/app.js',failure:()=>({errorText:'net::ERR_ABORTED'})});
 expect(a.failures).toHaveLength(2);expect(()=>a.assert()).toThrow();
}));
test('expected stream abort is recorded but cannot replace a live stream',()=>fixture('gi',async(page,a)=>{
 page.emit('requestfailed',{url:()=>a.origin+'/sse/stream',failure:()=>({errorText:'net::ERR_ABORTED'})});
 expect(a.streamAborts).toHaveLength(1);expect(a.failures).toEqual([]);expect(()=>a.assert()).toThrow('No connected fixture stream');
 page.emit('requestfailed',{url:()=>a.origin+'/sse/stream',failure:()=>({errorText:'net::ERR_ABORTED'})});expect(a.failures).toHaveLength(1);
}));
test('Piclaw stream and topic aborts are never waived',async()=>{
 for(const host of ['piclaw','gi'])await fixture(host,async(page,a)=>{
  page.emit('requestfailed',{url:()=>a.origin+'/sse/topics',failure:()=>({errorText:'net::ERR_ABORTED'})});
  expect(a.failures).toHaveLength(1);
  if(host==='piclaw'){page.emit('requestfailed',{url:()=>a.origin+'/sse/stream',failure:()=>({errorText:'net::ERR_ABORTED'})});expect(a.failures).toHaveLength(2);}
 });
});
