import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';

const composer=page=>page.locator('.compose-box textarea');
const selected=page=>page.evaluate(()=>localStorage.getItem('gi_session_id'));
const read=(page,path)=>page.evaluate(async path=>{const response=await fetch(path);if(!response.ok)throw new Error(`Read ${path}: ${response.status}`);return response.json();},path);
const turns=async(page,id)=>(await read(page,`/api/sessions/${id}/turns`)).turns || [];
async function boot(page,env){
 const created=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/sessions'&&r.request().method()==='POST');
 await page.goto(env.origin);const response=await created;expect(response.status()).toBe(201);const session=await response.json();
 await expect(composer(page)).toBeVisible();await expect.poll(()=>selected(page)).toBe(session.id);
 expect((await read(page,'/api/sessions')).sessions.map(s=>s.id)).toEqual([session.id]);return session.id;
}
async function assertTurn(page,id,turnId,prompt){
 await expect.poll(async()=>{const rows=await turns(page,id);return rows.map(t=>[t.id,t.status]);}).toEqual([[turnId,'completed']]);
 const messages=(await read(page,`/api/sessions/${id}/messages`)).messages;
 expect(messages.filter(m=>m.role==='user').map(m=>m.content)).toEqual([prompt]);
 const [turn]=await turns(page,id);expect(turn.id).toBe(turnId);expect(turn.session_id).toBe(id);
 expect(messages.filter(m=>m.role==='assistant')).toHaveLength(1);
 await expect(page.locator('.timeline .post:not(.agent-post) .post-content')).toHaveText([prompt]);
 await expect(page.locator('.timeline .post.agent-post')).toHaveCount(1);
}

test('Clean browser boot focuses the composer and Return admits exactly one first turn',async({page},info)=>{
 const env=await journeyEnvironment(info);const admissions=[];
 page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))admissions.push(r);});
 try{
  const id=await boot(page,env);await expect(composer(page)).toBeFocused();
  const text=`first Return ${info.project.name} Ω`;await page.keyboard.type(text);
  const reply=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${id}/prompt`));await page.keyboard.press('Enter');const response=await reply;expect(response.status()).toBe(202);const {turn_id}=await response.json();expect(turn_id).toBeTruthy();
  await assertTurn(page,id,turn_id,text);expect(admissions).toHaveLength(1);expect(admissions[0].postDataJSON().prompt).toBe(text);await expect(composer(page)).toHaveValue('');
  await page.reload();await expect(composer(page)).toBeVisible();expect(await selected(page)).toBe(id);expect(admissions).toHaveLength(1);expect((await read(page,'/api/sessions')).sessions).toHaveLength(1);
 }finally{await page.close();await env.close();}
});

test('First Return keeps its draft through failed admission and admits only one turn while pending',async({page},info)=>{
 const env=await journeyEnvironment(info);let release=()=>{};let admissions=0;
 try{
  const id=await boot(page,env);await expect(composer(page)).toBeFocused();const prompt=`explicit retry ${info.project.name}`;await page.keyboard.type(prompt);
  await page.route(`**/api/sessions/${id}/prompt`,route=>{admissions++;return route.fulfill({status:503,contentType:'application/json',body:'{"error":"Fixture admission unavailable"}'});});
  await page.keyboard.press('Enter');await expect(page.getByRole('alert')).toContainText('Fixture admission unavailable');await expect(composer(page)).toHaveValue(prompt);expect(await turns(page,id)).toEqual([]);expect(admissions).toBe(1);
  await page.unroute(`**/api/sessions/${id}/prompt`);const gate=new Promise(r=>release=r);let held=false;
  await page.route(`**/api/sessions/${id}/prompt`,async route=>{admissions++;held=true;await gate;await route.continue();});
  const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${id}/prompt`));await page.keyboard.press('Enter');await expect.poll(()=>held).toBe(true);expect(await turns(page,id)).toEqual([]);await expect(composer(page)).toHaveValue('');await page.keyboard.press('Enter');await page.keyboard.press('Enter');expect(admissions).toBe(2);
  release();const accepted=await response;expect(accepted.status()).toBe(202);const {turn_id}=await accepted.json();await assertTurn(page,id,turn_id,prompt);await expect(composer(page)).toHaveValue('');expect(admissions).toBe(2);
 }finally{release();await page.unrouteAll({behavior:'wait'});await page.close();await env.close();}
});

test('Clean boot holds readiness and explicitly retries a failed initial session request',async({page},info)=>{
 const env=await journeyEnvironment(info);let release=()=>{};let creates=0;
 try{
  const gate=new Promise(r=>release=r);let held=false;
  await page.route('**/api/sessions',async route=>{if(route.request().method()!=='POST')return route.continue();creates++;held=true;await gate;await route.fulfill({status:503,contentType:'application/json',body:'{"error":"Fixture bootstrap unavailable"}'});});
  await page.goto(env.origin);await expect.poll(()=>held).toBe(true);await expect(composer(page)).toHaveCount(0);release();
  await expect(page.getByRole('alert')).toContainText('Unable to open a chat');await expect(page.getByRole('button',{name:'Retry opening chat',exact:true})).toBeVisible();expect(creates).toBe(1);expect((await read(page,'/api/sessions')).sessions).toEqual([]);
  await page.unrouteAll({behavior:'wait'});const response=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/sessions'&&r.request().method()==='POST');await page.getByRole('button',{name:'Retry opening chat',exact:true}).click();const accepted=await response;expect(accepted.status()).toBe(201);const {id}=await accepted.json();await expect(composer(page)).toBeFocused();expect(await selected(page)).toBe(id);expect((await read(page,'/api/sessions')).sessions.map(s=>s.id)).toEqual([id]);
 }finally{release();await page.unrouteAll({behavior:'wait'});await page.close();await env.close();}
});

test('A lost initial session response is recovered by lookup without duplicating the chat',async({page},info)=>{
 const env=await journeyEnvironment(info);let createdId,creates=0;
 try{
  page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname==='/api/sessions')creates++;});
  await page.route('**/api/sessions',async route=>{if(route.request().method()!=='POST')return route.continue();const response=await route.fetch();expect(response.status()).toBe(201);createdId=(await response.json()).id;await route.abort('failed');});
  await page.goto(env.origin);await expect(page.getByRole('button',{name:'Retry opening chat',exact:true})).toBeVisible();expect((await read(page,'/api/sessions')).sessions.map(s=>s.id)).toEqual([createdId]);
  await page.unrouteAll({behavior:'wait'});await page.getByRole('button',{name:'Retry opening chat',exact:true}).click();await expect(composer(page)).toBeFocused();expect(await selected(page)).toBe(createdId);expect(creates).toBe(1);
 }finally{await page.close();await env.close();}
});

test('Malformed runtime bootstrap after session creation retries without a second chat',async({page},info)=>{
 const env=await journeyEnvironment(info);let release=()=>{},creates=0;
 try{
  page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname==='/api/sessions')creates++;});
  const gate=new Promise(r=>release=r);await page.route('**/api/runtime/config',async route=>{await gate;await route.fulfill({status:200,contentType:'application/json',body:'{invalid'});});
  const response=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/sessions'&&r.request().method()==='POST');await page.goto(env.origin);const initial=await response;expect(initial.status()).toBe(201);const {id}=await initial.json();await expect(composer(page)).toHaveCount(0);release();
  await expect(page.getByRole('button',{name:'Retry opening chat',exact:true})).toBeVisible();expect((await read(page,'/api/sessions')).sessions.map(s=>s.id)).toEqual([id]);await page.unrouteAll({behavior:'wait'});await page.getByRole('button',{name:'Retry opening chat',exact:true}).click();await expect(composer(page)).toBeFocused();expect(await selected(page)).toBe(id);expect(creates).toBe(1);
 }finally{release();await page.unrouteAll({behavior:'wait'});await page.close();await env.close();}
});

test('New session failure retains the parent and a delayed success cannot steal Settings focus',async({page},info)=>{
 const env=await journeyEnvironment(info);let release=()=>{};
 try{
  const parent=await boot(page,env);await composer(page).fill('preserved parent');let forks=0;
  page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith(`/api/sessions/${parent}/fork`))forks++;});
  await page.route(`**/api/sessions/${parent}/fork`,route=>route.fulfill({status:503,contentType:'application/json',body:'{"error":"Fixture fork unavailable"}'}));
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.getByRole('button',{name:'New',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Fixture fork unavailable');expect(await selected(page)).toBe(parent);await expect(composer(page)).toHaveValue('preserved parent');expect((await read(page,'/api/sessions')).sessions.map(s=>s.id)).toEqual([parent]);expect(forks).toBe(1);
  await page.unrouteAll({behavior:'wait'});let held=false;const gate=new Promise(r=>release=r);
  await page.route(`**/api/sessions/${parent}/fork`,async route=>{held=true;await gate;await route.continue();});
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.getByRole('button',{name:'New',exact:true}).click();await expect.poll(()=>held).toBe(true);
  await page.keyboard.press('Control+,');const close=page.getByRole('button',{name:'Close settings',exact:true});await expect(close).toBeFocused();release();await expect.poll(()=>selected(page)).not.toBe(parent);await expect(close).toBeFocused();await expect(page.locator('.settings-dialog')).toBeVisible();expect(forks).toBe(2);
 }finally{release();await page.unrouteAll({behavior:'wait'});await page.close();await env.close();}
});

test('New session restores composer focus and Shift+Enter cannot submit the parent draft',async({page},info)=>{
 const env=await journeyEnvironment(info);const admissions=[];
 page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))admissions.push(r);});
 try{
  const parent=await boot(page,env);await composer(page).fill('parent unsent draft');
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await expect(page.locator('.compose-session-popup')).toBeVisible();
  const fork=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${parent}/fork`)&&r.request().method()==='POST');await page.getByRole('button',{name:'New',exact:true}).click();const result=await fork;expect(result.status()).toBe(201);const child=(await result.json()).branch.chat_jid.slice(3);
  expect(child).not.toBe(parent);await expect.poll(()=>selected(page)).toBe(child);await expect(composer(page)).toHaveValue('');await expect(composer(page)).toBeFocused();
  await page.keyboard.type('child first line');await page.keyboard.press('Shift+Enter');await page.keyboard.type('second line');await expect(composer(page)).toHaveValue('child first line\nsecond line');expect(admissions).toHaveLength(0);expect(await turns(page,parent)).toEqual([]);expect(await turns(page,child)).toEqual([]);
  const reply=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${child}/prompt`));await page.keyboard.press('Enter');const response=await reply;expect(response.status()).toBe(202);const {turn_id}=await response.json();await assertTurn(page,child,turn_id,'child first line\nsecond line');expect(admissions).toHaveLength(1);expect(await turns(page,parent)).toEqual([]);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();const row=page.locator(`[data-session-entry-key="session:gi:${parent}"]`);await expect(row).toBeVisible();await row.click();await expect.poll(()=>selected(page)).toBe(parent);await expect(composer(page)).toHaveValue('parent unsent draft');expect(admissions).toHaveLength(1);
 }finally{await page.close();await env.close();}
});
