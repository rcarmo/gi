import {test,expect} from '@playwright/test';
import {waitForAppShell} from './helpers';

// Protocol checks do not claim dock/tab/popout UI acceptance. That host remains
// owned by the shared frontend; these tests use a real browser WebSocket/PTY.
test('terminal API executes in workspace, resizes and transfers without losing its shell',async({page,request})=>{
 const client=`functional-terminal-${Date.now()}`;
 const infoResponse=await request.get('/terminal/session',{headers:{'x-piclaw-terminal-client':client}});expect(infoResponse.status()).toBe(200);
 const info=await infoResponse.json();expect(info).toMatchObject({enabled:true,transport:'websocket',active:false});
 await page.goto('/');await waitForAppShell(page);
 await page.evaluate(async client=>{
  (window as any).__terminalFrames=[];
  const open=(handoff='')=>new Promise<void>((resolve,reject)=>{
   const url=new URL('/terminal/ws',location.href);url.protocol=location.protocol==='https:'?'wss:':'ws:';url.searchParams.set('client',client);if(handoff)url.searchParams.set('handoff',handoff);
   const ws=new WebSocket(url);(window as any).__terminalSocket=ws;
   ws.onmessage=e=>{const p=JSON.parse(e.data);(window as any).__terminalFrames.push(p);if(p.type==='session')resolve();};ws.onerror=()=>reject(new Error('terminal WS failed'));
  });
  (window as any).__openTerminal=open;await open();
  (window as any).__terminalSocket.send(JSON.stringify({type:'resize',cols:91,rows:27}));
  (window as any).__terminalSocket.send(JSON.stringify({type:'input',data:"stty -echo; export TERMINAL_TEST_STATE=retained; printf '\\nWORKSPACE=%s\\n' \"$PWD\"; stty size; printf 'PTY_%s\\n' READY\r"}));
 },client);
 await expect.poll(()=>page.evaluate(()=>(window as any).__terminalFrames.filter((p:any)=>p.type==='output').map((p:any)=>p.data).join(''))).toContain('PTY_READY');
 const first=await page.evaluate(()=>(window as any).__terminalFrames.find((p:any)=>p.type==='session'));
 const output=await page.evaluate(()=>(window as any).__terminalFrames.filter((p:any)=>p.type==='output').map((p:any)=>p.data).join(''));
 expect(output).toContain(`WORKSPACE=${info.cwd}`);expect(output).toContain('27 91');
 const response=await request.post('/terminal/handoff',{headers:{'x-piclaw-terminal-client':client}});expect(response.status()).toBe(200);const {handoff}=await response.json();
 await page.evaluate(async token=>{await (window as any).__openTerminal(token);(window as any).__terminalSocket.send(JSON.stringify({type:'input',data:"printf 'TRANSFER_%s\\n' \"$TERMINAL_TEST_STATE\"\r"}));},handoff.token);
 await expect.poll(()=>page.evaluate(()=>(window as any).__terminalFrames.filter((p:any)=>p.type==='output').map((p:any)=>p.data).join(''))).toContain('TRANSFER_retained');
 expect(await page.evaluate(()=>(window as any).__terminalFrames.filter((p:any)=>p.type==='session').at(-1).session_id)).toBe(first.session_id);
 await page.evaluate(()=>(window as any).__terminalSocket.close());
 await expect.poll(async()=>(await (await request.get('/terminal/session',{headers:{'x-piclaw-terminal-client':client}})).json()).active,{timeout:10000}).toBe(false);
});

test('terminal API rejects cross-origin shell access and missing client identity',async({request})=>{
 expect((await request.get('/terminal/session')).status()).toBe(400);
 expect((await request.get('/terminal/session',{headers:{Origin:'https://evil.example','x-piclaw-terminal-client':'test-client-123'}})).status()).toBe(403);
 expect((await request.post('/terminal/handoff',{headers:{'x-piclaw-terminal-client':'not-active-123'}})).status()).toBe(409);
});
