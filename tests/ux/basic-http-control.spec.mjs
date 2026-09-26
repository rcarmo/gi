import {test,expect} from '@playwright/test';
import http from 'node:http';import{networkInterfaces}from'node:os';
import{journeyEnvironment}from'./support/journey-environment.mjs';

async function setup(page,info){
 const fixture=await journeyEnvironment(info),address=Object.values(networkInterfaces()).flat().find(x=>x?.family==='IPv4'&&!x.internal)?.address;if(!address)throw Error('Nonloopback interface required');
 let blockSSE=false;const streams=new Set();
 const server=http.createServer((req,res)=>{
  const sse=req.url.startsWith('/sse/stream');if(sse&&blockSSE){res.writeHead(503);res.end();return}
  if(sse){streams.add(res);res.on('close',()=>streams.delete(res))}
  const up=http.request(new URL(req.url,fixture.origin),{method:req.method,headers:req.headers},r=>{res.writeHead(r.statusCode,r.headers);r.pipe(res)});up.on('error',()=>res.destroy());req.pipe(up);res.on('close',()=>up.destroy());
 });await new Promise(r=>server.listen(0,'0.0.0.0',r));const origin=`http://${address}:${server.address().port}`;
 const read=async path=>{const r=await page.request.get(origin+path);expect(r.status()).toBe(200);return r.json()};
 await page.goto(origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();expect(await page.evaluate(()=>isSecureContext)).toBe(false);const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
 const token=`basic-${info.project.name}-${Date.now()}`;const accepted=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${id}/prompt`));await input.fill(`UX steer gate:${token}`);await input.press('Enter');const response=await accepted;expect(response.status()).toBe(202);const turn=(await response.json()).turn_id;
 await expect.poll(async()=>(await read(`/api/sessions/${id}/turns`)).turns.find(x=>x.id===turn)?.status).toBe('running');await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeEnabled();
 return{origin,input,id,turn,read,release:()=>fixture.release(token),drop(){blockSSE=true;for(const r of streams)r.destroy()},resume(){blockSSE=false},async close(){fixture.release(token);server.closeAllConnections();await new Promise(r=>server.close(r));await fixture.close()}};
}

test('HTTP Stop cancels the addressed running turn without submitting or clearing the next draft',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('next unsent draft 中文');const posts=[];page.on('request',r=>{if(r.method()==='POST')posts.push({path:new URL(r.url()).pathname,body:r.postDataJSON()})});
  const stopped=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/activity`));await page.getByRole('button',{name:'Stop response',exact:true}).click();expect((await stopped).status()).toBe(200);
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===h.turn)?.status).toBe('cancelled');
  await expect(h.input).toHaveValue('next unsent draft 中文');expect(posts).toEqual([{path:`/api/sessions/${h.id}/activity`,body:{turn_id:h.turn}}]);expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(1);
  await page.reload();await expect(h.input).toHaveValue('next unsent draft 中文');await expect(page.getByRole('button',{name:'Send message',exact:true})).toBeEnabled();
  const next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${h.id}/prompt`));await h.input.press('Enter');const response=await next;expect(response.status()).toBe(202);const nextTurn=(await response.json()).turn_id;
  await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===nextTurn)?.status).toBe('completed');expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(2);await expect(h.input).toHaveValue('');
 }finally{await h.close()}
});

test('HTTP SSE disconnect and reconnect preserves draft and recovers native completion without resending',async({page},info)=>{
 const h=await setup(page,info);try{
  await h.input.fill('offline draft 中文');h.drop();await expect(page.getByText('Reconnecting',{exact:true})).toBeVisible();await expect(h.input).toHaveValue('offline draft 中文');
  h.release();await expect.poll(async()=>(await h.read(`/api/sessions/${h.id}/turns`)).turns.find(t=>t.id===h.turn)?.status).toBe('completed');h.resume();await expect(page.getByText('Reconnecting',{exact:true})).toHaveCount(0);
  await expect(h.input).toHaveValue('offline draft 中文');await expect(page.locator('.post.agent-post')).toHaveCount(1);expect((await h.read(`/api/sessions/${h.id}/turns`)).turns).toHaveLength(1);
  await page.reload();await expect(h.input).toHaveValue('offline draft 中文');await expect(page.locator('.post.agent-post')).toHaveCount(1);
 }finally{await h.close()}
});
