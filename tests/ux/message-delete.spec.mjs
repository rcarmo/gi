import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function fixture(page,request,info){
 const key=`delete-${info.project.name}-${Date.now()}`;
 const session=await(await request.post('/api/sessions',{data:{agent_id:key,title:key}})).json();
 const accepted=await(await request.post(`/api/sessions/${session.id}/prompt`,{data:{prompt:'remove native orchid',model:'test-model'}})).json();
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns??[]).find(t=>t.id===accepted.turn_id)?.status).toBe('completed');
 await expect.poll(async()=>{const s=await(await request.get(`/api/sessions/${session.id}/activity`)).json();return s.status;}).toBe('idle');
 const messages=async()=>((await(await request.get(`/api/sessions/${session.id}/messages`)).json()).messages??[]);
 const stored=await messages(),target=stored.find(m=>m.role==='assistant');
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id)},session.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('draft survives deletion');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'retain.txt',mimeType:'text/plain',buffer:Buffer.from('unsent')});
 const post=page.locator(`[id="post-${target.id}"]`);await expect(post).toBeVisible();
 return {session,stored,target,post,input,messages};
}

test('@ux-timeline-017 Delete an idle single message only after native success and retain deletion on reload',async({page,request},info)=>{
 await info.attach('gherkin',{body:loadCorpus().find(c=>c.id==='@ux-timeline-017').steps.join('\n'),contentType:'text/plain'});
 const {session,stored,target,post,input,messages}=await fixture(page,request,info);
 let release,held=false;const gate=new Promise(r=>release=r);
 await page.route(`**/api/sessions/${session.id}/messages/${target.id}?cascade=false`,async route=>{const res=await route.fetch();expect(res.status()).toBe(200);held=true;await gate;await route.fulfill({response:res});});
 await page.evaluate(id=>{window.__removedStates=[];new MutationObserver(()=>{const el=document.getElementById('post-'+id);window.__removedStates.push(el?el.className:'absent')}).observe(document.querySelector('.timeline'),{attributes:true,childList:true,subtree:true});},target.id);
 try{
  await post.getByRole('button',{name:'Delete message',exact:true}).click();await expect.poll(()=>held).toBe(true);await expect(post).toBeVisible();await expect(post).not.toHaveClass(/removing/);
  await expect.poll(async()=> (await messages()).some(m=>m.id===target.id)).toBe(false);release();
  await expect.poll(()=>page.evaluate(()=>window.__removedStates.some(s=>s.includes('removing')))).toBe(true);await expect(post).toHaveCount(0);
  expect((await messages()).map(m=>m.id)).toEqual(stored.filter(m=>m.id!==target.id).map(m=>m.id));await expect(input).toHaveValue('draft survives deletion');await expect(page.locator('.compose-file-pill').filter({hasText:'retain.txt'})).toBeVisible();
  await page.screenshot({path:info.outputPath('message-deleted.png')});await page.reload();await expect(post).toHaveCount(0);await expect(input).toHaveValue('draft survives deletion');
  const hits=await(await request.get(`/api/sessions/${session.id}/search?q=orchid&scope=current`)).json();expect(hits).toHaveProperty('messages');expect((hits.messages??[]).some(m=>m.id===target.id)).toBe(false);
 }finally{release()}
});

test('Gi held pre-delete timeline response cannot resurrect a removed row',async({page,request},info)=>{
 const {session,target,post,input,messages}=await fixture(page,request,info);
 let release,held=false,delivered;const gate=new Promise(r=>release=r),done=new Promise(r=>delivered=r);
 const pattern=`**/api/sessions/${session.id}/messages?*`;
 await page.route(pattern,async route=>{const response=await route.fetch();if(!held){held=true;await gate;await route.fulfill({response});delivered();}else await route.fulfill({response});});
 try{
  // The live refresh runs every 10 seconds. Hold the real pre-delete page,
  // then acknowledge deletion while that response is still in flight.
  await expect.poll(()=>held,{timeout:16000}).toBe(true);
  await post.getByRole('button',{name:'Delete message',exact:true}).click();await expect(post).toHaveCount(0);
  expect((await messages()).some(m=>m.id===target.id)).toBe(false);
  release();await done;await page.unroute(pattern);await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  await expect(post).toHaveCount(0);await expect(input).toHaveValue('draft survives deletion');
  await page.reload();await expect(post).toHaveCount(0);
 }finally{release()}
});

test('Gi search deletion targets the result origin and ignores a held pre-delete search response',async({page,request},info)=>{
 const {session,target,input,messages}=await fixture(page,request,info);
 const child=(await(await request.post(`/api/sessions/${session.id}/fork`,{data:{agent_id:'delete-search-child-'+Date.now()}})).json()).branch.chat_jid.slice(3);
 const accepted=await(await request.post(`/api/sessions/${child}/prompt`,{data:{prompt:'search-only violet',model:'test-model'}})).json();
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${child}/turns`)).json()).turns??[]).find(t=>t.id===accepted.turn_id)?.status).toBe('completed');
 await expect.poll(async()=> (await(await request.get(`/api/sessions/${child}/activity`)).json()).status).toBe('idle');
 const childPosts=async()=>((await(await request.get(`/api/sessions/${child}/messages`)).json()).messages??[]);
 const childTarget=(await childPosts()).find(m=>m.role==='user'&&m.content==='search-only violet');expect(childTarget).toBeTruthy();
 await page.getByRole('button',{name:'Search',exact:true}).click();
 const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});
 await page.locator('.compose-search-scope-select').selectOption('root');await search.fill('violet');await search.press('Enter');
 const childPost=page.locator(`[id="post-${childTarget.id}"]`);await expect(childPost).toBeVisible();
 let release,held=false,delivered;const gate=new Promise(r=>release=r),done=new Promise(r=>delivered=r);
 const pattern=`**/api/sessions/${session.id}/search?*`;
 await page.route(pattern,async route=>{const response=await route.fetch();if(!held){held=true;await gate;await route.fulfill({response});delivered();}else await route.fulfill({response});});
 try{
  await search.press('Enter');await expect.poll(()=>held).toBe(true);
  await childPost.getByRole('button',{name:'Delete message',exact:true}).click();await expect(childPost).toHaveCount(0);
  expect((await childPosts()).some(m=>m.id===childTarget.id)).toBe(false);
  expect((await messages()).some(m=>m.id===target.id)).toBe(true);
  release();await done;await page.unroute(pattern);await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  await expect(childPost).toHaveCount(0);await page.getByRole('button',{name:'Close search',exact:true}).click();await expect(input).toHaveValue('draft survives deletion');
  await page.reload();await page.getByRole('button',{name:'Search',exact:true}).click();await page.locator('.compose-search-scope-select').selectOption('root');
  const reopened=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});await reopened.fill('violet');await reopened.press('Enter');await expect(childPost).toHaveCount(0);
 }finally{release()}
});

test('Gi deletion failure preserves the post and remains with the originating session',async({page,request},info)=>{
 const {session,target,post,input,messages}=await fixture(page,request,info);
 const child=(await(await request.post(`/api/sessions/${session.id}/fork`,{data:{agent_id:'delete-child-'+Date.now()}})).json()).branch.chat_jid.slice(3);
 await page.reload();await expect(post).toBeVisible();
 let release,held=false;const gate=new Promise(r=>release=r);
 await page.route(`**/api/sessions/${session.id}/messages/${target.id}?cascade=false`,async route=>{held=true;await gate;await route.abort('failed');});
 try{
  await post.getByRole('button',{name:'Delete message',exact:true}).click();await expect.poll(()=>held).toBe(true);await expect(post).not.toHaveClass(/removing/);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${child}"]`).getByRole('menuitem').click();await input.fill('child draft');release();
  await expect(input).toHaveValue('child draft');await expect(page.getByRole('alert')).toHaveCount(0);
  expect((await messages()).some(m=>m.id===target.id)).toBe(true);
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${session.id}"]`).getByRole('menuitem').click();await expect(post).toBeVisible();await expect(input).toHaveValue('draft survives deletion');
 }finally{release()}
});
