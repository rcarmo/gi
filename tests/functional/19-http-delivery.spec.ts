import { test, expect, chromium, webkit } from '@playwright/test';
import http from 'node:http';
import { networkInterfaces } from 'node:os';
import { BASE_URL } from './helpers';

let proxy:http.Server,origin:string;
test.beforeAll(async()=>{
 const address=Object.values(networkInterfaces()).flat().find(x=>x?.family==='IPv4'&&!x.internal)?.address;if(!address)throw Error('Non-loopback IPv4 required');
 proxy=http.createServer((req,res)=>{const up=http.request(new URL(req.url!,BASE_URL),{method:req.method,headers:req.headers},r=>{res.writeHead(r.statusCode!,r.headers);r.pipe(res)});up.on('error',()=>{res.writeHead(502);res.end()});req.pipe(up);res.on('close',()=>up.destroy())});
 await new Promise<void>(resolve=>proxy.listen(0,'0.0.0.0',resolve));origin=`http://${address}:${(proxy.address() as any).port}`;
});
test.afterAll(async()=>{proxy?.closeAllConnections();await new Promise<void>(resolve=>proxy.close(()=>resolve()))});

for(const[name,type]of[['chromium',chromium],['webkit',webkit]] as const){
 test(`${name} HTTP attachment bytes survive upload, send, response and reload`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();const page=await context.newPage();const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
  try{
   await page.goto(origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();expect(await page.evaluate(()=>isSecureContext)).toBe(false);
   const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));const text=`attachment-${name}-${Date.now()}`,bytes=Buffer.from('Real HTTP attachment 中文🙂\nexact bytes\n');
   await page.locator('.compose-box input[type=file]').setInputFiles({name:'human-note.txt',mimeType:'text/plain',buffer:bytes});await expect(page.locator('.compose-file-pill')).toHaveCount(1);await input.fill(text);
   const accepted=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${id}/prompt`));await page.locator('.compose-box .send-btn').click();const response=await accepted;expect(response.status()).toBe(202);const result=await response.json();
   await expect.poll(async()=>((await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns||[]).find((t:any)=>t.id===result.turn_id)?.status).toBe('completed');
   const turns=(await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns;const turn=turns.find((t:any)=>t.id===result.turn_id);expect(turn.prompt).toContain(text);expect(turn.metadata.media).toHaveLength(1);
   const raw=await context.request.get(`${origin}/api/media/${turn.metadata.media[0].media_id}/raw`);expect(Buffer.from(await raw.body())).toEqual(bytes);
   await expect(page.locator('.post.agent-post .post-content').filter({hasText:`Gi received: ${text}`})).toHaveCount(1);await expect(input).toHaveValue('');await expect(page.locator('.compose-box .compose-file-pill')).toHaveCount(0);
   await page.reload();await expect(page.locator('.post.agent-post .post-content').filter({hasText:text})).toHaveCount(1);expect((await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns.filter((t:any)=>t.id===result.turn_id)).toHaveLength(1);expect(errors).toEqual([]);
  }finally{await context.close();await browser.close()}
 });
 test(`${name} HTTP lost accepted acknowledgement never restores accepted text over newer draft`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();const page=await context.newPage();const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));let release!:()=>void;const gate=new Promise<void>(r=>release=r);
  try{
   await page.goto(origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
   const text=`lost-ack-${name}-${Date.now()}`,newer='newer unsent 中文🙂';let admitted:any=null,posts=0;let payload:any;
   await page.route(`**/api/sessions/${id}/prompt`,async route=>{posts++;payload=route.request().postDataJSON();const response=await route.fetch();expect(response.status()).toBe(202);admitted=await response.json();await gate;await route.abort('failed')});
   await input.fill(text);await input.press('Enter');await expect.poll(()=>admitted?.turn_id).toBeTruthy();await expect(page.getByRole('status').filter({hasText:'Sending message'})).toBeVisible();await input.fill(newer);
   await expect(page.locator('.post.agent-post .post-content').filter({hasText:`Gi received: ${text}`})).toHaveCount(1);release();
   await expect(page.getByRole('status').filter({hasText:'Sending message'})).toHaveCount(0);await expect(input).toHaveValue(newer);expect(payload.client_request_id).toBeTruthy();expect(posts).toBe(1);
   await page.reload();await expect(input).toHaveValue(newer);await expect(page.locator('.post.agent-post .post-content').filter({hasText:text})).toHaveCount(1);
   const turns=(await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns;expect(turns.filter((t:any)=>t.prompt===text)).toHaveLength(1);expect(errors).toEqual([]);
  }finally{release();await context.close();await browser.close()}
 });
}

for(const[name,type]of[['chromium',chromium],['webkit',webkit]] as const){
 test(`${name} HTTP failed receipt read remains unknown without resending accepted work`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();const page=await context.newPage();
  try{
   await page.goto(origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));const text=`unknown-ack-${name}-${Date.now()}`;let posts=0,receiptReads=0;
   await page.route(`**/api/sessions/${id}/turns`,r=>{receiptReads++;return r.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error:'Receipt lookup unavailable'})})});
   await page.route(`**/api/sessions/${id}/prompt`,async r=>{posts++;const accepted=await r.fetch();expect(accepted.status()).toBe(202);await r.abort('failed')});
   await input.fill(text);await input.press('Enter');await expect(page.getByText(/Delivery is unknown/).first()).toBeVisible();await expect(input).toHaveValue(text);expect(receiptReads).toBeGreaterThan(0);expect(posts).toBe(1);
   await expect(page.locator('.post.agent-post .post-content').filter({hasText:text})).toHaveCount(1);const turns=(await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns;expect(turns.filter((t:any)=>t.prompt===text)).toHaveLength(1);
   // Do not press Return here: unknown delivery is not permission for a retry.
  }finally{await context.close();await browser.close()}
 });
 test(`${name} HTTP New session and return preserve parent draft and child history`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();const page=await context.newPage();
  try{
   await page.goto(origin);const input=page.locator('.compose-box textarea');await expect(input).toBeVisible();const parent=await page.evaluate(()=>localStorage.getItem('gi_session_id'));const draft=`parent unsent ${name} 中文`;
   await input.fill(draft);const created=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${parent}/fork`));
   await page.locator('.compose-session-trigger-top button').click();await page.locator('[data-session-entry-key="action:new"]').click();const response=await created;expect(response.status()).toBe(201);const child=(await response.json()).branch.chat_jid.replace(/^gi:/,'');await expect(input).toBeFocused();await expect(input).toHaveValue('');
   const text=`child-${name}-${Date.now()}`;await input.fill(text);await input.press('Enter');await expect(page.locator('.post.agent-post .post-content').filter({hasText:`Gi received: ${text}`})).toHaveCount(1);
   await page.locator('.compose-session-trigger-top button').click();await page.locator(`[data-session-jid="gi:${parent}"]`).click();await expect(input).toHaveValue(draft);await page.reload();await expect(input).toHaveValue(draft);
   await page.locator('.compose-session-trigger-top button').click();await page.locator(`[data-session-jid="gi:${child}"]`).click();await expect(page.locator('.post.agent-post .post-content').filter({hasText:`Gi received: ${text}`})).toHaveCount(1);
   const turns=(await(await context.request.get(`${origin}/api/sessions/${child}/turns`)).json()).turns;expect(turns.filter((t:any)=>t.prompt===text)).toHaveLength(1);
  }finally{await context.close();await browser.close()}
 });
}
