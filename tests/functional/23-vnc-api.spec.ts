import {test,expect} from '@playwright/test';
import {waitForAppShell} from './helpers';

test('VNC API rejects unconfigured targets and cross-origin connections',async({request})=>{
 const response=await request.get('/vnc/session');expect(response.status()).toBe(200);const info=await response.json();expect(info.host_policy).toBe('allowlist');expect(info.direct_connect_enabled).toBe(false);
 expect((await request.get('/vnc/ws?target=127.0.0.1:22')).status()).toBe(404);
 expect((await request.get('/vnc/session',{headers:{Origin:'https://evil.example'}})).status()).toBe(403);
});

test('VNC API bridges binary data and transfers the same upstream connection',async({page,request})=>{
 test.skip(process.env.GI_VNC_PROTOCOL_FIXTURE!=='1','requires disposable local TCP fixture');
 const response=await request.get('/vnc/session?target=fixture');expect(response.status()).toBe(200);expect((await response.json()).target.id).toBe('fixture');
 await page.goto('/');await waitForAppShell(page);
 await page.evaluate(async()=>{
  (window as any).__vncFrames=[];
  (window as any).__openVNC=(token='')=>new Promise<void>((resolve,reject)=>{
   const url=new URL('/vnc/ws',location.href);url.protocol=location.protocol==='https:'?'wss:':'ws:';url.searchParams.set('target','fixture');if(token)url.searchParams.set('handoff',token);
   const ws=new WebSocket(url);ws.binaryType='arraybuffer';(window as any).__vncSocket=ws;
   ws.onmessage=e=>{if(typeof e.data==='string'){const p=JSON.parse(e.data);(window as any).__vncFrames.push(p);if(p.type==='vnc.connected')resolve();}else{(window as any).__vncFrames.push({bytes:[...new Uint8Array(e.data)]});}};ws.onerror=()=>reject(new Error('VNC socket failed'));
  });
  await (window as any).__openVNC();
 });
 await expect.poll(()=>page.evaluate(()=>(window as any).__vncFrames.filter((x:any)=>x.bytes).map((x:any)=>String.fromCharCode(...x.bytes)).join(''))).toContain('RFB 003.008');
 const handoff=await request.post('/vnc/handoff?target=fixture');expect(handoff.status()).toBe(200);const token=(await handoff.json()).handoff.token;
 await page.evaluate(async token=>{await (window as any).__openVNC(token);(window as any).__vncSocket.send(new Uint8Array([0,1,2,255]));(window as any).__vncSocket.send(JSON.stringify({type:'ping'}));},token);
 await expect.poll(()=>page.evaluate(()=>(window as any).__vncFrames.some((x:any)=>JSON.stringify(x.bytes)==='[0,1,2,255]'))).toBe(true);
 await expect.poll(()=>page.evaluate(()=>(window as any).__vncFrames.some((x:any)=>x.type==='pong'))).toBe(true);
 await page.evaluate(()=>(window as any).__vncSocket.close());
});
