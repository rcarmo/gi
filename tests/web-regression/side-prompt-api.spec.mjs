import {test,expect} from '@playwright/test';
import {writeFileSync} from 'node:fs';
import {journeyEnvironment} from '../ux/support/journey-environment.mjs';

// Backend acceptance only. Composer/panel/dock integration is upstream-owned.
test('BTW API snapshots session context, retries, aborts and persists only normal injection',async({page},info)=>{
 const previous=process.env.GI_UX_SIDE_PROFILE_DIR;
 process.env.GI_UX_SIDE_PROFILE_DIR=info.outputPath('server-profile');
 const cdp=await page.context().newCDPSession(page);
 await cdp.send('Profiler.enable');await cdp.send('Profiler.start');
 await cdp.send('HeapProfiler.enable');await cdp.send('HeapProfiler.startSampling',{samplingInterval:4096});
 let h;
 try{
  h=await journeyEnvironment(info);
  await page.goto(h.origin);await expect(page.locator('.compose-box textarea')).toBeVisible();
  const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));expect(id).toBeTruthy();
  const read=async suffix=>(await(await page.request.get(h.origin+`/api/sessions/${id}/${suffix}`)).json());
  const admitted=await(await page.request.post(h.origin+`/api/sessions/${id}/prompt`,{data:{prompt:'history-alpha'}})).json();
  await expect.poll(async()=>(await read('turns')).turns.find(t=>t.id===admitted.turn_id)?.status).toBe('completed');
  const other=await(await page.request.post(h.origin+'/api/sessions',{data:{title:'other'}})).json();
  const foreign=await(await page.request.post(h.origin+`/api/sessions/${other.id}/prompt`,{data:{prompt:'other-session-secret'}})).json();
  await expect.poll(async()=>(await(await page.request.get(h.origin+`/api/sessions/${other.id}/turns`)).json()).turns.find(t=>t.id===foreign.turn_id)?.status).toBe('completed');
  const beforeMessages=await read('messages'),beforeTurns=await read('turns');
  for(let i=0;i<2;i++){
   const stream=await page.evaluate(async id=>{const r=await fetch('/agent/side-prompt/stream',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({prompt:'side question',chat_jid:`gi:${id}`})});return{status:r.status,text:await r.text()};},id);
   expect(stream.status).toBe(200);expect(stream.text).toContain('side_prompt_start');expect(stream.text).toContain('side_prompt_thinking_delta');expect(stream.text).toContain('side_prompt_text_delta');expect(stream.text).toContain('side_prompt_done');expect(stream.text).toContain('history-alpha');expect(stream.text).not.toContain('other-session-secret');
  }
  const error=await page.request.post(h.origin+'/agent/side-prompt/stream',{data:{prompt:'fail-side',chat_jid:`gi:${id}`}});
  expect(await error.text()).toContain('side_prompt_error');
  const aborted=await page.evaluate(async id=>{
   const controller=new AbortController();const r=await fetch('/agent/side-prompt/stream',{method:'POST',signal:controller.signal,headers:{'Content-Type':'application/json'},body:JSON.stringify({prompt:'wait-side',chat_jid:`gi:${id}`})});
   const reader=r.body.getReader();let text='';
   try{while(!text.includes('side_prompt_text_delta')){const {value,done}=await reader.read();if(done)throw new Error('stream ended early');text+=new TextDecoder().decode(value);}controller.abort();try{await reader.cancel();}catch(error){if(error.name!=='AbortError')throw error;}return text;}finally{controller.abort();}
  },id);expect(aborted).toContain('side_prompt_text_delta');
  const retry=await page.request.post(h.origin+'/agent/side-prompt',{data:{prompt:'retry question',chat_jid:`gi:${id}`}});expect(retry.status()).toBe(200);expect(await retry.json()).toMatchObject({status:'success',thinking:'side thought'});
  expect(await read('messages')).toEqual(beforeMessages);expect(await read('turns')).toEqual(beforeTurns);
  const injection=await(await page.request.post(h.origin+`/api/sessions/${id}/prompt`,{data:{prompt:'BTW question: side question\n\nBTW answer: injected-side-answer'}})).json();
  await expect.poll(async()=>(await read('turns')).turns.find(t=>t.id===injection.turn_id)?.status).toBe('completed');
  expect(JSON.stringify(await read('messages'))).toContain('injected-side-answer');expect((await read('turns')).turns).toHaveLength(beforeTurns.turns.length+1);
 }finally{
  await h?.close();
  if(previous===undefined)delete process.env.GI_UX_SIDE_PROFILE_DIR;else process.env.GI_UX_SIDE_PROFILE_DIR=previous;
  const {profile:cpu}=await cdp.send('Profiler.stop');writeFileSync(info.outputPath('browser-cpu.json'),JSON.stringify(cpu));
  const {profile:heap}=await cdp.send('HeapProfiler.stopSampling');writeFileSync(info.outputPath('browser-heap.json'),JSON.stringify(heap));await cdp.detach();
 }
});
