import {test,expect} from '@playwright/test';
import {readFileSync} from 'node:fs';

test('Gi read-only Piclaw host context/callbacks use bounded native reads and fence old instances',async({page,request},info)=>{
 const path=`pane-${info.project.name}-${Date.now()}.txt`,body='αβ Unicode preview\n'.repeat(2000);
 const written=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:body}}});expect((await written.json()).error).toBeFalsy();
 const session=await(await request.post('/api/sessions',{data:{agent_id:path,title:path}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('host draft β');
 const native=await(await request.get(`/api/workspace/file?path=${path}&max_bytes=20000`)).json();expect(native.truncated).toBe(true);
 const writes=[];page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/')&&!['GET','HEAD'].includes(r.method())&&!r.url().includes('/frontend/log'))writes.push(r.url());});
 // Run the real WorkspaceTab + lifecycle against a test-only Piclaw extension.
 // Intercept only the harness asset; file responses remain native.
 await page.route('**/__pane-conformance.js',r=>r.fulfill({contentType:'text/javascript',body:readFileSync('test-results/pane-host-fixture.js','utf8')}));
 await page.addScriptTag({type:'module',url:'/__pane-conformance.js'});
 await expect.poll(()=>page.evaluate(()=>Boolean(window.__paneHarness))).toBe(true);
 let release,held=false;const gate=new Promise(r=>release=r);
 await page.route('**/api/workspace/file?*',async route=>{const r=await route.fetch();held=true;await gate;await route.fulfill({response:r});});
 try{
  await page.evaluate(p=>window.__paneHarness.mount(p),path);await expect.poll(()=>held).toBe(true);await input.focus();release();
  const pane=page.locator('[data-conformance]');await expect(pane.locator('[data-content]')).toHaveText(native.text);await expect(input).toBeFocused();await page.unroute('**/api/workspace/file?*');
  const events=()=>page.evaluate(()=>window.__paneHarness.events);
  await expect.poll(events).toContainEqual(expect.objectContaining({type:'mount',context:{path,mode:'view',content:native.text,mtime:native.mtime,size:native.size,preview:native}}));
  await expect.poll(async()=>(await events()).filter(e=>e.type==='resize').length).toBeGreaterThan(0);
  const initial=(await events()).filter(e=>e.type==='resize').at(-1).width;
  await page.evaluate(()=>window.__paneHarness.resize(320));await expect.poll(async()=>(await events()).filter(e=>e.type==='resize').at(-1).width).toBeGreaterThan(initial);
  const update=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'refreshed native bytes γ'}}});expect((await update.json()).error).toBeFalsy();
  await pane.getByRole('button',{name:'Refresh preview'}).click();await expect(pane.locator('[data-content]')).toHaveText('refreshed native bytes γ');
  expect((await events()).filter(e=>e.type==='dispose')).toHaveLength(1);
  await page.evaluate(()=>window.__paneHarness.closeOld(0));await expect(pane.locator('[data-content]')).toHaveText('refreshed native bytes γ');
  await pane.getByRole('button',{name:'Pane requests close'}).click();await expect(pane.locator('section')).toHaveCount(0);
  expect((await events()).filter(e=>e.type==='closed')).toHaveLength(1);expect((await events()).filter(e=>e.type==='dispose')).toHaveLength(2);
  const before=await events();await page.evaluate(()=>{window.__paneHarness.closeOld(1);window.__paneHarness.resize(220);window.dispatchEvent(new Event('resize'));});await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));expect(await events()).toEqual(before);
  await page.evaluate(p=>window.__paneHarness.mount(p,['preview','edit']),path);await expect(pane.getByRole('alert')).toContainText('outside Gi’s read-only preview host');expect((await events()).filter(e=>e.type==='mount')).toHaveLength(2);
  await expect(input).toHaveValue('host draft β');expect(writes).toEqual([]);expect((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns||[]).toEqual([]);expect((await events()).filter(e=>e.type==='focus')).toEqual([]);
  await info.attach('pane-host-events',{body:JSON.stringify(await events(),null,2),contentType:'application/json'});
 }finally{release();}
});
