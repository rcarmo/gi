import { test, expect, chromium, webkit } from '@playwright/test';
import http from 'node:http';
import { networkInterfaces } from 'node:os';
import { BASE_URL } from './helpers';

let proxy:http.Server,origin:string;
// Hold actual upstream acknowledgements, not a route.fetch clone of a paused
// browser POST (page close can release that original request and send twice).
const heldReplies=new Map<string,{gate:Promise<void>,admitted:any,count:number}>();
test.beforeAll(async()=>{
 const address=Object.values(networkInterfaces()).flat().find(x=>x?.family==='IPv4'&&!x.internal)?.address;if(!address)throw Error('Non-loopback IPv4 required');
 proxy=http.createServer((req,res)=>{const held=req.method==='POST'?heldReplies.get(req.url!):undefined;if(held)held.count++;
  const up=http.request(new URL(req.url!,BASE_URL),{method:req.method,headers:req.headers},r=>{
   if(!held){res.writeHead(r.statusCode!,r.headers);r.pipe(res);return}
   const chunks:Buffer[]=[];r.on('data',chunk=>chunks.push(chunk));r.on('end',()=>{const body=Buffer.concat(chunks);held.admitted=JSON.parse(body.toString());void held.gate.then(()=>{if(!res.destroyed){res.writeHead(r.statusCode!,r.headers);res.end(body)}})});
  });up.on('error',()=>{if(!res.destroyed){res.writeHead(502);res.end()}});req.pipe(up);res.on('close',()=>up.destroy())});
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
   await page.route(`**/api/sessions/${id}/send-receipt?*`,r=>{receiptReads++;return r.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error:'Receipt lookup unavailable'})})});
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

for(const[name,type]of[['chromium',chromium],['webkit',webkit]] as const){
 test(`${name} HTTP close before accepted reply reconciles persisted capture without restoring accepted attachments`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();let page=await context.newPage();let release!:()=>void;const gate=new Promise<void>(r=>release=r);let held:{gate:Promise<void>,admitted:any,count:number}|undefined;let heldPath='';
  try{
   await page.goto(origin);let input=page.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
   const text=`closed-ack-${name}-${Date.now()}`,newer='new draft survives close 中文🙂';
   await page.locator('.compose-box input[type=file]').setInputFiles({name:'already-sent.txt',mimeType:'text/plain',buffer:Buffer.from('accepted attachment bytes')});await expect(page.locator('.compose-file-pill')).toHaveCount(1);
   heldPath=`/api/sessions/${id}/prompt`;held={gate,admitted:null,count:0};heldReplies.set(heldPath,held);
   await input.fill(text);await input.press('Enter');await expect.poll(()=>held!.admitted?.turn_id).toBeTruthy();await input.fill(newer);
   const acceptedTurn=(await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns.find((t:any)=>t.id===held!.admitted.turn_id);
   const saved=()=>page.evaluate(id=>new Promise<any>((resolve,reject)=>{const open=indexedDB.open('gi-session-drafts');open.onerror=()=>reject(open.error);open.onsuccess=()=>{const db=open.result;const read=db.transaction('drafts','readonly').objectStore('drafts').get(id);read.onsuccess=()=>{resolve(read.result);db.close()};read.onerror=()=>reject(read.error)};}),id);
   await expect.poll(async()=>{const row=await saved();return [row?.draft?.text,row?.pending?.[0]?.id]}).toEqual([newer,acceptedTurn.metadata.client_request_id]);
   await expect(page.locator('.post.agent-post .post-content').filter({hasText:text})).toHaveCount(1);await page.close();release();
   page=await context.newPage();const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));await page.goto(origin);input=page.locator('.compose-box textarea');await expect(input).toHaveValue(newer);await expect(page.locator('.compose-box .compose-file-pill')).toHaveCount(0);await expect(page.locator('.post.agent-post .post-content').filter({hasText:text})).toHaveCount(1);
   const row=await saved();expect(row.pending).toEqual([]);expect(row.draft.media).toEqual([]);expect(held.count).toBe(1);
   await page.reload();await expect(input).toHaveValue(newer);const turns=(await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns;expect(turns.filter((t:any)=>t.prompt.startsWith(text))).toHaveLength(1);expect(errors).toEqual([]);
  }finally{release();heldReplies.delete(heldPath);await context.close();await browser.close()}
 });
 test(`${name} HTTP reopen with failed receipt lookup keeps unknown text once and never auto-submits`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();let page=await context.newPage();let release!:()=>void;const gate=new Promise<void>(r=>release=r);let held:{gate:Promise<void>,admitted:any,count:number}|undefined;let heldPath='';
  try{
   await page.goto(origin);let input=page.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));const text=`closed-unknown-${name}-${Date.now()}`;
   heldPath=`/api/sessions/${id}/prompt`;held={gate,admitted:null,count:0};heldReplies.set(heldPath,held);await input.fill(text);await input.press('Enter');await expect.poll(()=>held!.admitted?.turn_id).toBeTruthy();await expect(page.locator('.post.agent-post .post-content').filter({hasText:text})).toHaveCount(1);await page.close();release();
   page=await context.newPage();await page.route(`**/api/sessions/${id}/send-receipt?*`,r=>r.fulfill({status:503,contentType:'application/json',body:'{"error":"Receipt unavailable"}'}));await page.goto(origin);input=page.locator('.compose-box textarea');await expect(input).toHaveValue(text);await expect(page.getByText(/Recovered an unacknowledged send. Delivery is unknown/)).toBeVisible();expect(held.count).toBe(1);
   await page.reload();await expect(input).toHaveValue(text);expect((await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns.filter((t:any)=>t.prompt===text)).toHaveLength(1);expect(held.count).toBe(1);
  }finally{release();heldReplies.delete(heldPath);await context.close();await browser.close()}
 });
}

for(const[name,type]of[['chromium',chromium],['webkit',webkit]] as const){
 for(const reopen of [false,true])test(`${name} HTTP routed send ${reopen?'after close':'lost acknowledgement'} recovers source receipt without cross-chat scans`,async()=>{
  const browser=await type.launch();const context=await browser.newContext();let page=await context.newPage();let release!:()=>void;const gate=new Promise<void>(r=>release=r);let path='';let held:{gate:Promise<void>,admitted:any,count:number};
  try{
   await page.goto(origin);let input=page.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
   const before=(await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns||[];
   const text=`@peer routed-${name}-${Date.now()}`,newer='source newer draft';path=`/api/sessions/${id}/prompt`;held={gate,admitted:null,count:0};
   if(reopen)heldReplies.set(path,held);else await page.route(`**${path}`,async r=>{held.count++;const response=await r.fetch();expect(response.status()).toBe(202);held.admitted=await response.json();await gate;await r.abort('failed')});
   await input.fill(text);await input.press('Enter');await expect.poll(()=>held.admitted?.turn_id).toBeTruthy();expect(held.admitted.routed).toBe(true);expect(held.admitted.source_session_id).toBe(id);expect(held.admitted.session_id).not.toBe(id);await input.fill(newer);
   await expect.poll(async()=>{const data=await(await context.request.get(`${origin}/api/sessions/${held.admitted.session_id}/turns`)).json();return data.turns.find((t:any)=>t.id===held.admitted.turn_id)?.status}).toBe('completed');
   if(reopen){await page.close();release();page=await context.newPage();await page.goto(origin);input=page.locator('.compose-box textarea')}
   else release();
   await expect(input).toHaveValue(newer);await expect(page.getByRole('status').filter({hasText:'Sending message'})).toHaveCount(0);expect(held.count).toBe(1);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);
   await page.reload();await expect(input).toHaveValue(newer);
   const turns=(await(await context.request.get(`${origin}/api/sessions/${held.admitted.session_id}/turns`)).json()).turns;expect(turns.filter((t:any)=>t.id===held.admitted.turn_id)).toHaveLength(1);
   expect((await(await context.request.get(`${origin}/api/sessions/${id}/turns`)).json()).turns||[]).toHaveLength(before.length);
  }finally{release();heldReplies.delete(path);await context.close();await browser.close()}
 });
}

async function storedDraft(page:any,id:string){
 return page.evaluate((id:string)=>new Promise((resolve,reject)=>{const open=indexedDB.open('gi-session-drafts');open.onerror=()=>reject(open.error);open.onsuccess=()=>{const db=open.result,tx=db.transaction('drafts','readonly'),read=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();resolve(read.result)};tx.onabort=()=>{db.close();reject(tx.error)}}}),id);
}
for(const[name,type]of[['chromium',chromium],['webkit',webkit]] as const){
 test(`${name} HTTP stale tab cannot erase pending capture or dispatch after draft conflict`,async()=>{
  const browser=await type.launch(),context=await browser.newContext();let a=await context.newPage(),b=await context.newPage();let release!:()=>void;const gate=new Promise<void>(r=>release=r);let path='';
  try{
   await a.goto(origin);const input=a.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await a.evaluate(()=>localStorage.getItem('gi_session_id'));await b.goto(origin);const other=b.locator('.compose-box textarea');await expect(other).toBeVisible();
   const text=`cross-tab-${name}-${Date.now()}`,bytes=Buffer.from('cross tab attachment bytes Ω');
   await a.locator('.compose-box input[type=file]').setInputFiles({name:'cross-tab.txt',mimeType:'text/plain',buffer:bytes});await input.fill(text);
   path=`/api/sessions/${id}/prompt`;const held={gate,admitted:null as any,count:0};heldReplies.set(path,held);
   await input.press('Enter');await expect.poll(()=>held.admitted?.turn_id).toBeTruthy();
   await expect.poll(async()=>(await storedDraft(a,id))?.pending?.length).toBe(1);
   await other.fill('other tab unsaved');await expect(b.getByText(/Draft not saved:.*another tab/)).toBeVisible();
   let otherPosts=0;b.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))otherPosts++});
   await other.press('Enter');await expect(other).toHaveValue('other tab unsaved');await expect(b.getByRole('status').filter({hasText:'Sending message'})).toHaveCount(0);expect(otherPosts).toBe(0);
   const row:any=await storedDraft(b,id);expect(row.pending).toHaveLength(1);expect(row.pending[0].draft.text).toBe(text);expect(row.pending[0].draft.media).toHaveLength(1);
   await a.close();await b.close();release();
   a=await context.newPage();await a.route(`**/api/sessions/${id}/send-receipt?*`,r=>r.fulfill({status:503,contentType:'application/json',body:'{"error":"unavailable"}'}));await a.goto(origin);
   await expect(a.locator('.compose-box textarea')).toHaveValue(text);await expect(a.locator('.compose-file-pill[title="cross-tab.txt"]')).toBeVisible();await expect(a.getByText(/Recovered an unacknowledged send/)).toBeVisible();
   const restored=await a.evaluate(async(id:string)=>{const db=await new Promise<IDBDatabase>((resolve,reject)=>{const q=indexedDB.open('gi-session-drafts');q.onsuccess=()=>resolve(q.result);q.onerror=()=>reject(q.error)});const row:any=await new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>resolve(r.result);tx.onabort=()=>reject(tx.error)});db.close();return Array.from(new Uint8Array(row.draft.media[0].bytes))},id);expect(restored).toEqual([...bytes]);expect(held.count).toBe(1);
  }finally{release();heldReplies.delete(path);await context.close();await browser.close()}
 });
 test(`${name} HTTP stale unknown recovery cannot restore after another tab confirms`,async()=>{
  const browser=await type.launch(),context=await browser.newContext(),a=await context.newPage(),b=await context.newPage();let releaseAck!:()=>void,releaseRead!:()=>void;
  const ack=new Promise<void>(r=>releaseAck=r),read=new Promise<void>(r=>releaseRead=r);let path='',reading=false;
  try{
   await a.goto(origin);const input=a.locator('.compose-box textarea');await expect(input).toBeVisible();const id=await a.evaluate(()=>localStorage.getItem('gi_session_id'));
   path=`/api/sessions/${id}/prompt`;const held={gate:ack,admitted:null as any,count:0};heldReplies.set(path,held);await input.fill(`confirmed-${name}-${Date.now()}`);await input.press('Enter');await expect.poll(()=>held.admitted?.turn_id).toBeTruthy();
   await b.route(`**/api/sessions/${id}/send-receipt?*`,async r=>{reading=true;await read;await r.fulfill({status:200,contentType:'application/json',body:'{"confirmed":false}'})});
   await b.goto(origin);await expect.poll(()=>reading).toBe(true);releaseAck();
   await expect.poll(async()=>(await storedDraft(a,id))?.pending?.length).toBe(0);const committed:any=await storedDraft(a,id);releaseRead();
   await expect(b.getByText(/Draft recovery unavailable:.*another tab/)).toBeVisible();await expect(b.locator('.compose-box textarea')).toHaveValue('');
   expect((await storedDraft(a,id))?.revision).toBe(committed.revision);expect((await storedDraft(a,id))?.draft.text).toBe('');expect(held.count).toBe(1);
   await b.unrouteAll({behavior:'wait'});await b.reload();await expect(b.locator('.compose-box textarea')).toHaveValue('');await expect(b.getByText(/Recovered an unacknowledged send/)).toHaveCount(0);
  }finally{releaseAck();releaseRead();heldReplies.delete(path);await context.close();await browser.close()}
 });
 test(`${name} draft storage upgrade fences old writers and preserves version-one media`,async()=>{
  const browser=await type.launch(),context=await browser.newContext(),page=await context.newPage();
  try{
   // Seed a real v1 database before loading the application.
   await page.goto(origin+'/favicon.ico');await page.evaluate(async()=>{const db=await new Promise<IDBDatabase>((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onupgradeneeded=()=>r.result.createObjectStore('drafts',{keyPath:'sessionId'});r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error)});await new Promise<void>((resolve,reject)=>{const tx=db.transaction('drafts','readwrite');tx.objectStore('drafts').put({sessionId:'legacy',draft:{text:'legacy Ω',media:[{name:'old.txt',type:'text/plain',lastModified:1,bytes:new Uint8Array([1,2,3]).buffer}],fileRefs:['one'],messageRefs:[]},pending:[]});tx.oncomplete=()=>resolve();tx.onabort=()=>reject(tx.error)});db.close()});
   await page.goto(origin);await expect(page.locator('.compose-box textarea')).toBeVisible();
   const state=await page.evaluate(async()=>{const db=await new Promise<IDBDatabase>((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts');r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error)});const version=db.version;const row:any=await new Promise(resolve=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get('legacy');tx.oncomplete=()=>resolve(r.result)});db.close();const old=await new Promise(resolve=>{const r=indexedDB.open('gi-session-drafts',1);r.onerror=()=>resolve(r.error?.name);r.onsuccess=()=>{r.result.close();resolve('unexpected success')}});return{version,text:row.draft.text,bytes:Array.from(new Uint8Array(row.draft.media[0].bytes)),old}});
   expect(state).toEqual({version:2,text:'legacy Ω',bytes:[1,2,3],old:'VersionError'});
  }finally{await context.close();await browser.close()}
 });
}

for(const[name,type]of[['chromium',chromium],['webkit',webkit]] as const){
 for(const closes of [true,false])test(`${name} old draft writer ${closes?'closes on upgrade':'blocks upgrade without losing data'}`,async()=>{
  const browser=await type.launch(),context=await browser.newContext(),old=await context.newPage(),page=await context.newPage();
  try{
   await old.goto(origin+'/favicon.ico');await old.evaluate(async closes=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onupgradeneeded=()=>r.result.createObjectStore('drafts',{keyPath:'sessionId'});r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error)});
    (window as any).__oldDraftDB=db;(window as any).__versionChanged=false;
    db.onversionchange=()=>{(window as any).__versionChanged=true;if(closes)db.close()};
    await new Promise<void>((resolve,reject)=>{const tx=db.transaction('drafts','readwrite');tx.objectStore('drafts').put({sessionId:'legacy',draft:{text:'old durable',media:[],fileRefs:[],messageRefs:[]},pending:[]});tx.oncomplete=()=>resolve();tx.onabort=()=>reject(tx.error)});
   },closes);
   await page.goto(origin);await expect.poll(()=>old.evaluate(()=>(window as any).__versionChanged)).toBe(true);
   if(closes){
    await expect(page.locator('.compose-box textarea')).toBeVisible();
    expect(await old.evaluate(()=>{try{(window as any).__oldDraftDB.transaction('drafts','readwrite');return 'unsafe'}catch(e){return e.name}})).toBe('InvalidStateError');
   }else{
    await expect(page.getByText(/Draft recovery unavailable:.*upgrade blocked/)).toBeVisible();
    const input=page.locator('.compose-box textarea');await input.fill('unsaved during blocked upgrade');let posts=0;page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))posts++});
    await input.press('Enter');await expect(input).toHaveValue('unsaved during blocked upgrade');await expect(page.getByRole('status').filter({hasText:'Sending message'})).toHaveCount(0);expect(posts).toBe(0);
    await old.evaluate(()=>(window as any).__oldDraftDB.close());await page.reload();await expect(page.getByText(/Draft recovery unavailable/)).toHaveCount(0);await expect(input).toBeVisible();
   }
   expect((await storedDraft(page,'legacy'))?.draft.text).toBe('old durable');
  }finally{await context.close();await browser.close()}
 });
}
