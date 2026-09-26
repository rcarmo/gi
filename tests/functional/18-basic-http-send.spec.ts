import { test, expect, chromium, webkit } from '@playwright/test';
import http from 'node:http';
import { networkInterfaces } from 'node:os';
import { BASE_URL } from './helpers';

// Real insecure HTTP origin, no isSecureContext/crypto overrides. The proxy
// forwards to the disposable test instance, never the live service.
let proxy: http.Server, remoteOrigin: string;
test.beforeAll(async()=>{
 const address=Object.values(networkInterfaces()).flat().find(x=>x?.family==='IPv4'&&!x.internal)?.address;
 if(!address)throw new Error('A non-loopback IPv4 address is required for real HTTP-origin send acceptance');
 proxy=http.createServer((req,res)=>{
  const upstream=http.request(new URL(req.url!,BASE_URL),{method:req.method,headers:req.headers},response=>{
   res.writeHead(response.statusCode!,response.headers);response.pipe(res);
  });upstream.on('error',e=>{res.writeHead(502);res.end(e.message)});req.pipe(upstream);res.on('close',()=>upstream.destroy());
 });
 await new Promise<void>(resolve=>proxy.listen(0,'0.0.0.0',resolve));
 remoteOrigin=`http://${address}:${(proxy.address() as any).port}`;
});
test.afterAll(async()=>{proxy?.closeAllConnections();await new Promise<void>(resolve=>proxy?.close(()=>resolve()));});

for(const [engine,type] of [['chromium',chromium],['webkit',webkit]] as const){
 for(const insecure of [false,true]){
  test(`${engine} ${insecure?'HTTP host':'localhost'} Return and Send admit exact messages, respond and survive reload`,async()=>{
   const browser=await type.launch();const context=await browser.newContext();const page=await context.newPage();const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
   const origin=insecure?remoteOrigin:BASE_URL;
   try{
    await page.goto(origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();await expect(input).toBeEnabled();
    const caps=await page.evaluate(()=>({secure:isSecureContext,uuid:typeof crypto.randomUUID,random:typeof crypto.getRandomValues}));
    expect(caps.secure).toBe(!insecure);expect(caps.random).toBe('function');if(insecure)expect(caps.uuid).toBe('undefined');
    const sent:string[]=[];
    for(const method of ['Return','Send']){
     const text=`${engine}-${insecure}-${method}-${Date.now()} unique 中文🙂`;sent.push(text);
     const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));expect(id).toBeTruthy();
     const before=await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json();
     const responsePromise=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${id}/prompt`));
     await input.fill(text);
     if(method==='Return')await input.press('Enter');else await page.locator('.compose-box .send-btn').click();
     const response=await responsePromise;expect(response.status()).toBe(202);const admitted=await response.json();expect(admitted.turn_id).toBeTruthy();expect(admitted.session_id).toBe(id);
     await expect.poll(async()=>{const data=await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json();return data.turns.find((t:any)=>t.id===admitted.turn_id)?.status;}).toBe('completed');
     const after=await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json();
     expect(after.turns).toHaveLength((before.turns||[]).length+1);expect(after.turns.filter((t:any)=>t.prompt===text)).toHaveLength(1);
     await expect(page.locator('.post:not(.agent-post) .post-content').filter({hasText:text})).toHaveCount(1);
     await expect(page.locator('.post.agent-post .post-content').filter({hasText:`Gi received: ${text}`})).toHaveCount(1);
     await expect(input).toHaveValue('');await page.reload();await expect(input).toBeVisible();
     for(const prior of sent){await expect(page.locator('.post:not(.agent-post) .post-content').filter({hasText:prior})).toHaveCount(1);await expect(page.locator('.post.agent-post .post-content').filter({hasText:`Gi received: ${prior}`})).toHaveCount(1)}
    }
    expect(errors).toEqual([]);
   }finally{await context.close();await browser.close()}
  });
 }
 test(`${engine} HTTP rejection and network loss retain text with honest feedback and explicit retry`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();const page=await context.newPage();const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
  try{
   await page.goto(remoteOrigin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
   const count=async()=>((await(await context.request.get(`${remoteOrigin}/api/sessions/${id}/turns`)).json()).turns||[]).length;const before=await count();
   const endpoint=`**/api/sessions/${id}/prompt`,text=`recover-${engine}-${Date.now()}`;
   await page.route(endpoint,r=>r.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error:'Admission rejected for test'})}));
   await input.fill(text);await input.press('Enter');await expect(page.getByText(/Admission rejected for test/).first()).toBeVisible();await expect(input).toHaveValue(text);expect(await count()).toBe(before);
   await page.unroute(endpoint);await page.route(endpoint,r=>r.abort('failed'));
   await input.press('Enter');await expect(page.getByText(/Delivery is unknown/).first()).toBeVisible();await expect(input).toHaveValue(text);expect(await count()).toBe(before);
   await page.unroute(endpoint);
   const responsePromise=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${id}/prompt`));await input.press('Enter');expect((await responsePromise).status()).toBe(202);
   await expect(page.locator('.post.agent-post .post-content').filter({hasText:`Gi received: ${text}`})).toHaveCount(1);expect(await count()).toBe(before+1);await expect(input).toHaveValue('');expect(errors).toEqual([]);
  }finally{await context.close();await browser.close()}
 });
}
