import {test,expect} from '@playwright/test';
import {spawn} from 'node:child_process';
import {mkdtempSync,mkdirSync,rmSync,writeFileSync,createWriteStream} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
import {createServer,request as httpRequest} from 'node:http';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function environment(page,info,options={}){
 const dir=mkdtempSync(join(tmpdir(),'gi-reconnect-'));
 const reserve=createServer();await new Promise(r=>reserve.listen(0,'127.0.0.1',r));const port=reserve.address().port;await new Promise(r=>reserve.close(r));
 const origin=`http://127.0.0.1:${port}`;let child;
 mkdirSync(resolve('test-results/ux-parity'),{recursive:true});
 const log=createWriteStream(resolve('test-results/ux-parity',`reconnect-${info.project.name}-${Date.now()}.log`));
 const start=async()=>{
  child=spawn(resolve('bin/gi-ux-steer'),[],{env:{...process.env,GI_UX_STATE_DIR:dir,GI_UX_LISTEN:`127.0.0.1:${port}`,GI_UX_QUEUE_GATES:dir},stdio:['ignore','pipe','pipe']});child.stdout.pipe(log,{end:false});child.stderr.pipe(log,{end:false});
  await expect.poll(async()=>{try{return (await fetch(origin+'/api/runtime/config')).status}catch{return 0}},{timeout:10000}).toBe(200);
 };
 const stop=async()=>{if(!child||child.exitCode!==null)return;const done=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await done;};
 let blocked=false;const connections=new Set();
 let releaseInitial;const initialReady=new Promise(r=>releaseInitial=r);
 if(!options.holdInitial)releaseInitial();
 const proxy=createServer(async(req,res)=>{
  await initialReady;if(res.destroyed)return;
  res.setHeader('Access-Control-Allow-Origin','*');if(blocked){res.writeHead(503);res.end();return;}
  connections.add(res);res.on('close',()=>connections.delete(res));
  const upstream=httpRequest(new URL(req.url,origin),source=>{res.writeHead(source.statusCode,{'Content-Type':'text/event-stream','Cache-Control':'no-cache','Access-Control-Allow-Origin':'*'});source.on('aborted',()=>res.destroy());source.on('error',()=>res.destroy());source.pipe(res);});
  upstream.on('error',()=>res.destroy());res.on('close',()=>upstream.destroy());upstream.end();
 });
 await new Promise(r=>proxy.listen(0,'127.0.0.1',r));
 await page.addInitScript(origin=>{const Native=window.EventSource;window.EventSource=class extends Native{constructor(url,options){const u=new URL(url,location.href);super(u.pathname==='/sse/stream'?origin+u.pathname+u.search:url,options);}};},`http://127.0.0.1:${proxy.address().port}`);
 await start();
 const api=async(path,method='GET',body)=>{const res=await fetch(origin+path,{method,headers:{'Content-Type':'application/json'},...(body?{body:JSON.stringify(body)}:{})});expect(res.ok).toBe(true);return res.json();};
 const main=await api('/api/sessions','POST',{agent_id:`reconnect-${Date.now()}`,title:'main'});
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto(origin);const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 await expect(page.locator('.compose-connection-status')).toHaveCount(0);
 return {origin,main,input,api,ready:()=>releaseInitial(),release:token=>writeFileSync(join(dir,token),'go'),drop(){blocked=true;for(const res of connections)res.destroy();},resume(){blocked=false;},stop,start,
  async close(){releaseInitial();for(const res of connections)res.destroy();await new Promise(r=>proxy.close(r));await stop();log.end();rmSync(dir,{recursive:true,force:true});}};
}
async function source(info,id){const scenario=loadCorpus().find(x=>x.id===id);await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});}

test('@ux-reconnect-002 Refresh authoritative chat state after reconnect',async({page},info)=>{
 await source(info,'@ux-reconnect-002');const env=await environment(page,info);const{main,input,api}=env;
 let unblock,held=false,delivered;const gate=new Promise(r=>unblock=r),delivery=new Promise(r=>delivered=r);
 try{
  const oldToken=`old-${Date.now()}`;const active=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:`UX steer gate:${oldToken}`,model:'ux-local/gate'});
  await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeVisible();
  const oldQ=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:'old queued row',intent:'queue',model:'ux-local/gate'});await expect(page.locator(`[data-queue-id="${oldQ.turn_id}"]`)).toBeVisible();
  await input.fill('draft survives outage');await page.locator('.compose-box input[type=file]').setInputFiles({name:'offline.txt',mimeType:'text/plain',buffer:Buffer.from('bytes')});
  await page.route(`**/api/sessions/${main.id}/messages?*`,async route=>{const response=await route.fetch();const isHeld=!held;if(isHeld){held=true;await gate;}await route.fulfill({response});if(isHeld)delivered();});
  // Native mutation schedules a timeline refresh; hold its real old snapshot.
  const extra=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:'trigger old refresh',intent:'queue',model:'ux-local/gate'});await expect.poll(()=>held).toBe(true);
  env.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});await expect(page.getByRole('button',{name:'Stop response',exact:true})).toHaveCount(0);
  await api(`/api/sessions/${main.id}/queue/${oldQ.turn_id}`,'DELETE');await api(`/api/sessions/${main.id}/queue/${extra.turn_id}`,'DELETE');env.release(oldToken);
  await expect.poll(async()=> (await api(`/api/sessions/${main.id}/turns`)).turns.find(t=>t.id===active.turn_id).status).toBe('completed');
  const newToken=`new-${Date.now()}`;const next=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:`UX steer gate:${newToken}`,model:'ux-local/gate'});
  const queue=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:'new authoritative queued row',intent:'queue',model:'ux-local/gate'});
  env.resume();await expect(page.locator('.compose-connection-status')).toHaveCount(0,{timeout:15000});
  await expect(page.locator(`[data-queue-id="${queue.turn_id}"]`)).toBeVisible();await expect(page.locator(`[data-queue-id="${oldQ.turn_id}"]`)).toHaveCount(0);
  const stop=page.getByRole('button',{name:'Stop response',exact:true});await expect(stop).toBeEnabled();
  await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 32K tokens (0%)');
  await expect(page.getByText(`UX steer gate:${newToken}`,{exact:true})).toBeVisible();
  const lateResponse=page.waitForResponse(r=>r.url().includes(`/api/sessions/${main.id}/messages?`));unblock();await delivery;await(await lateResponse).finished();await page.unrouteAll({behavior:'wait'});
  await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  await expect(page.getByText(`UX steer gate:${newToken}`,{exact:true})).toBeVisible();await expect(input).toHaveValue('draft survives outage');await expect(page.locator('.compose-file-pill[title="offline.txt"]')).toBeVisible();
  const cancel=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${main.id}/activity`)&&r.request().method()==='POST');await stop.click();expect((await cancel).request().postDataJSON().turn_id).toBe(next.turn_id);
  env.release(newToken);
 }finally{unblock();await env.close();}
});

test('@ux-reconnect-004 Show version drift without automatically reloading',async({page},info)=>{
 await source(info,'@ux-reconnect-004');const env=await environment(page,info);const{input}=env;
 try{
  const old=await page.locator('script[src*="/dist/app.bundle.js"]').getAttribute('src');
  await input.fill('unsaved editor draft');let navigations=0;page.on('framenavigated',frame=>{if(frame===page.mainFrame())navigations++;});
  await env.stop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});await env.start();
  const warning=page.getByRole('status').filter({hasText:'New UI available'});await expect(warning).toHaveCount(1,{timeout:15000});await expect(warning).toContainText('Reload manually');expect(navigations).toBe(0);await expect(input).toHaveValue('unsaved editor draft');
  expect(await page.locator('script[src*="/dist/app.bundle.js"]').getAttribute('src')).toBe(old);
  await input.fill('');env.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});env.resume();await expect(page.locator('.compose-connection-status')).toHaveCount(0,{timeout:15000});await expect(warning).toHaveCount(1);expect(navigations).toBe(0);
  await page.reload();await expect(warning).toHaveCount(0);expect(await page.locator('script[src*="/dist/app.bundle.js"]').getAttribute('src')).not.toBe(old);
 }finally{await env.close();}
});

test('Gi pre-disconnect activity failure cannot overwrite healthy reconnect state',async({page},info)=>{
 const env=await environment(page,info);const{main,input,api}=env;
 let unblock,held=false,done;const gate=new Promise(r=>unblock=r),delivery=new Promise(r=>done=r);
 try{
  const token=`late-${Date.now()}`;await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:`UX steer gate:${token}`,model:'ux-local/gate'});await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeEnabled();
  await input.fill('keep through late error');
  await page.route(`**/api/sessions/${main.id}/activity`,async route=>{if(route.request().method()==='GET'&&!held){held=true;await gate;await route.abort('failed');done();return;}await route.continue();});
  await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:'queue update',intent:'queue',model:'ux-local/gate'});await expect.poll(()=>held).toBe(true);
  env.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});env.resume();await expect(page.locator('.compose-connection-status')).toHaveCount(0,{timeout:15000});await expect(page.getByRole('button',{name:'Stop response',exact:true})).toBeEnabled();
  const failed=page.waitForEvent('requestfailed',r=>r.url().endsWith(`/api/sessions/${main.id}/activity`));unblock();await delivery;await failed;await page.unrouteAll({behavior:'wait'});await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  await expect(page.getByRole('alert')).toHaveCount(0);await expect(input).toHaveValue('keep through late error');env.release(token);
 }finally{unblock();await env.close();}
});

test('@ux-reconnect-003 Search survives reconnect without a main-timeline refresh',async({page},info)=>{
 await source(info,'@ux-reconnect-003');const env=await environment(page,info);const{main,input,api}=env;
 let releaseOld;const oldGate=new Promise(r=>releaseOld=r);let held=false,delivered;const oldDone=new Promise(r=>delivered=r);
 try{
  const initial=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:'needle original',model:'ux-local/gate'});await expect.poll(async()=> (await api(`/api/sessions/${main.id}/turns`)).turns.find(t=>t.id===initial.turn_id).status).toBe('completed');
  await input.fill('preserved search draft');await page.locator('.compose-box input[type=file]').setInputFiles({name:'search.txt',mimeType:'text/plain',buffer:Buffer.from('keep')});
  await page.route(`**/api/sessions/${main.id}/messages?*`,async route=>{const response=await route.fetch();if(!held){held=true;await oldGate;await route.fulfill({response});delivered();}else await route.fulfill({response});});
  const token=`search-${Date.now()}`;const active=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:`UX steer gate:${token}`,model:'ux-local/gate'});await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:'Search',exact:true}).click();const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});await search.fill('needle');await search.press('Enter');
  await expect(page.getByText('needle original',{exact:true})).toBeVisible();
  const oldResponse=page.waitForResponse(r=>r.url().includes(`/api/sessions/${main.id}/messages?`));releaseOld();await oldDone;await(await oldResponse).finished();await page.unrouteAll({behavior:'wait'});
  await expect(page.getByText(`UX steer gate:${token}`,{exact:true})).toHaveCount(0);
  let timelines=0;page.on('request',r=>{if(r.url().includes(`/api/sessions/${main.id}/messages?`))timelines++;});
  env.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});
  const queued=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:'needle offline queue',intent:'queue',model:'ux-local/gate'});env.release(token);
  await expect.poll(async()=> (await api(`/api/sessions/${main.id}/turns`)).turns.find(t=>t.id===queued.turn_id).status).toBe('completed');
  const nextToken=`search-next-${Date.now()}`;await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:`UX steer gate:${nextToken}`,model:'ux-local/gate'});
  const nextQueue=await api(`/api/sessions/${main.id}/prompt`,'POST',{prompt:'queue after search outage',intent:'queue',model:'ux-local/gate'});
  const fresh=new Set();page.on('request',r=>{for(const suffix of ['/activity','/queue','/model'])if(r.url().endsWith(`/api/sessions/${main.id}${suffix}`))fresh.add(suffix);});
  env.resume();await expect(page.locator('.compose-connection-status')).toHaveCount(0,{timeout:15000});await expect(search).toHaveValue('needle');await expect(page.getByText('needle offline queue',{exact:true})).toBeVisible();expect(timelines).toBe(0);
  await expect.poll(()=>fresh.size).toBe(3);await expect(page.locator(`[data-queue-id="${nextQueue.turn_id}"]`)).toBeVisible();
  await expect(page.getByText(`UX steer gate:${nextToken}`,{exact:true})).toHaveCount(0);
  await search.press('Escape');await expect(input).toHaveValue('preserved search draft');await expect(page.locator('.compose-file-pill[title="search.txt"]')).toBeVisible();await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 32K tokens (0%)');await expect(page.getByText(`UX steer gate:${nextToken}`,{exact:true})).toBeVisible();
  env.release(nextToken);
 }finally{releaseOld();await env.close();}
});

test('Gi search scopes, literal query, stale responses and no prompt submission',async({page},info)=>{
 const env=await environment(page,info);const{main,input,api}=env;let unblock,held=false,delivered;const gate=new Promise(r=>unblock=r),done=new Promise(r=>delivered=r);
 try{
  const child=(await api(`/api/sessions/${main.id}/fork`,'POST',{agent_id:'search-child',title:'child'})).branch.chat_jid.slice(3);
  const other=await api('/api/sessions','POST',{agent_id:'search-outside',title:'outside'});
  for(const [id,prompt]of [[main.id,'needle main 100%_'],[child,'needle child'],[other.id,'needle outside']]){const t=await api(`/api/sessions/${id}/prompt`,'POST',{prompt,model:'ux-local/gate'});await expect.poll(async()=> (await api(`/api/sessions/${id}/turns`)).turns.find(x=>x.id===t.turn_id).status).toBe('completed');}
  await input.fill('scope draft');await page.getByRole('button',{name:'Search',exact:true}).click();const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});await search.fill('needle');await search.press('Enter');
  await expect(page.getByText('needle main 100%_',{exact:true})).toBeVisible();await expect(page.getByText('needle child',{exact:true})).toHaveCount(0);
  await page.locator('.compose-search-scope-select').selectOption('root');await expect(page.getByText('needle child',{exact:true})).toBeVisible();await expect(page.getByText('needle outside',{exact:true})).toHaveCount(0);
  await page.locator('.compose-search-scope-select').selectOption('all');await expect(page.getByText('needle outside',{exact:true})).toBeVisible();
  await search.fill('100%_');await search.press('Enter');await expect(page.getByText('needle main 100%_',{exact:true})).toBeVisible();await expect(page.getByText('needle child',{exact:true})).toHaveCount(0);
  await page.route(`**/api/sessions/${main.id}/search?*`,async route=>{const response=await route.fetch();if(!held&&new URL(route.request().url()).searchParams.get('q')==='100%_'){held=true;await gate;await route.fulfill({response});delivered();}else await route.fulfill({response});});
  await search.fill('100%_');await search.press('Enter');await expect.poll(()=>held).toBe(true);await search.fill('no such message');await search.press('Enter');await expect(page.getByText('No matching messages.',{exact:true})).toBeVisible();unblock();await done;await page.unrouteAll({behavior:'wait'});await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));await expect(page.getByText('needle main 100%_',{exact:true})).toHaveCount(0);
  await page.getByRole('button',{name:'Close search',exact:true}).click();await expect(input).toHaveValue('scope draft');expect((await api(`/api/sessions/${main.id}/turns`)).turns).toHaveLength(1);
 }finally{unblock();await env.close();}
});

test('Gi failed search remains scoped and reconnect does not erase its query',async({page},info)=>{
 const env=await environment(page,info);const{main,input}=env;
 try{
  await input.fill('draft before search error');await page.getByRole('button',{name:'Search',exact:true}).click();const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});
  const pattern=`**/api/sessions/${main.id}/search?*`;await page.route(pattern,route=>route.abort('failed'));
  await search.fill('unavailable query');await search.press('Enter');await expect(page.getByRole('alert').filter({hasText:'Search failed:'})).toBeVisible();await expect(search).toHaveValue('unavailable query');
  await page.unroute(pattern);env.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});env.resume();await expect(page.getByText('No matching messages.',{exact:true})).toBeVisible();await expect(search).toHaveValue('unavailable query');await expect(page.getByRole('alert')).toHaveCount(0);
  await search.press('Escape');await expect(input).toHaveValue('draft before search error');
 }finally{await env.close();}
});

function observeRefreshes(page){
 const paths=['messages','activity','queue','model','compaction'];const counts=new Map();
 page.on('request',r=>{const m=new URL(r.url()).pathname.match(/^\/api\/sessions\/([^/]+)\/([^/]+)$/);if(r.method()==='GET'&&m&&paths.includes(m[2])){const key=m[1]+'/'+m[2];counts.set(key,(counts.get(key)||0)+1);}});
 return {count:(id,name)=>counts.get(id+'/'+name)||0,paths};
}
test('@ux-reconnect-005 Initial activation and SSE readiness do not duplicate refresh',async({page},info)=>{
 await source(info,'@ux-reconnect-005');const requests=observeRefreshes(page);const env=await environment(page,info,{holdInitial:true});const{main,input,api}=env;
 try{
  // Shell is already rendered but the native subscription is deliberately held.
  await input.fill('draft before first subscription');
  for(const name of requests.paths)expect(requests.count(main.id,name)).toBe(0);
  env.ready();await expect.poll(()=>requests.count(main.id,'compaction')).toBe(1);
  await expect(page.locator('.compose-context-pie')).toBeVisible();
  for(const name of requests.paths)expect(requests.count(main.id,name)).toBe(1);
  await expect(input).toHaveValue('draft before first subscription');
  // Each selection (including A -> B -> A) is a new activation, not a cache hit.
  const other=(await api(`/api/sessions/${main.id}/fork`,'POST',{agent_id:'initial-child',title:'child'})).branch.chat_jid.slice(3);
  // Refresh the picker through the normal native fork invalidation/periodic path.
  await page.reload();await expect(input).toHaveValue('draft before first subscription');
  const baseline=requests.count(main.id,'activity');
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${other}"]`).getByRole('menuitem').click();
  await expect.poll(()=>requests.count(other,'activity')).toBe(1);await expect(page.locator('.compose-context-pie')).toBeVisible();
  for(const name of requests.paths)expect(requests.count(other,name)).toBe(1);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();
  await expect.poll(()=>requests.count(main.id,'activity')).toBe(baseline+1);await expect(input).toHaveValue('draft before first subscription');
  const before=Object.fromEntries(requests.paths.map(name=>[name,requests.count(main.id,name)]));env.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});env.resume();await expect(page.locator('.compose-connection-status')).toHaveCount(0,{timeout:15000});
  for(const name of requests.paths)await expect.poll(()=>requests.count(main.id,name)).toBe(before[name]+1);
 }finally{await env.close();}
});

test('Gi failed initial state refresh remains retryable on real reconnect',async({page},info)=>{
 const env=await environment(page,info,{holdInitial:true});const{main,input}=env;
 try{
  const pattern=`**/api/sessions/${main.id}/activity`;await page.route(pattern,route=>route.abort('failed'));
  await input.fill('draft survives first failure');env.ready();await expect(page.getByRole('alert')).toBeVisible();
  await page.unroute(pattern);env.drop();await expect(page.locator('.compose-connection-status')).toBeVisible({timeout:15000});env.resume();await expect(page.getByRole('alert')).toHaveCount(0,{timeout:15000});await expect(page.locator('.compose-context-pie')).toBeVisible();await expect(input).toHaveValue('draft survives first failure');
 }finally{await env.close();}
});
