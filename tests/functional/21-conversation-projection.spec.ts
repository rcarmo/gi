import {test,expect,chromium,webkit} from '@playwright/test';
import {spawnSync} from 'node:child_process';
import {resolve, join} from 'node:path';
import {BASE_URL} from './helpers';

for(const name of ['chromium','webkit'] as const)test(`${name} conversation excludes tool internals, attributes notices and clears idle activity`,async()=>{
 const browser=await ({chromium,webkit}[name]).launch({headless:true});const page=await browser.newPage();const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 const api=page.request;const created=await api.post(BASE_URL+'/api/sessions',{data:{agent_id:`projection-${name}`,title:`projection-${name}`}});expect(created.ok()).toBe(true);const {id}=await created.json();
 const prefix=`projection-${name}-${Date.now()}`;const message=(suffix:string)=>`${prefix}-${suffix}`;
 const rows=[['user','user','question',{}],['assistant','assistant','**Visible reply**\n\n- ordered item',{}],['intermediate','assistant','Let me inspect.\n[tool_call: shell]',{kind:'tool_calls'}],['tool','tool_result','PRIVATE TOOL OUTPUT\n---\nfile listing',{}],['error','system','Inference error: fixture failure',{}]] as const;
 const quote=(s:string)=>`'${s.replaceAll("'","''")}'`;
 const values=rows.map(([key,role,content,payload],i)=>`(${quote(message(key))},${quote(id)},${quote(role)},${quote(content)},${quote(JSON.stringify(payload))},'2026-01-01T00:00:0${i}Z')`).join(',');
 // This target is run only by make test-ux, against its fresh isolated DB.
 const db=resolve(process.env.GI_TEST_DB||join(process.env.GI_TEST_RUN_ROOT!, 'instance', 'gi.db'));const seed=spawnSync('sqlite3',['-cmd','.timeout 5000',db,`insert into messages(id,session_id,role,content,payload_json,created_at) values ${values};`],{encoding:'utf8'});expect(seed.status,seed.stderr).toBe(0);
 await page.addInitScript(s=>{localStorage.setItem('gi_session_id',s);const Native=window.EventSource;(window as any).__conversationStreams=[];window.EventSource=class extends Native{constructor(url:string|URL,options?:EventSourceInit){super(url,options);(window as any).__conversationStreams.push(this)}};},id);
 let active=false;
 await page.route(`**/api/sessions/${id}/activity`,r=>r.fulfill({json:{status:active?'running':'idle',phase:active?'running':'failed',turn_id:'fixture-turn',tool:{tool_call_id:'old',name:'shell',state:'completed',duration_ms:0}}}));
 const assertConversation=async()=>{
  await expect(page.locator('.post')).toHaveCount(4);
  await expect(page.locator(`#post-${message('user')} .post-author`)).toHaveText('Test User');
  await expect(page.locator(`#post-${message('assistant')}`)).toHaveClass(/agent-post/);
  await expect(page.locator(`#post-${message('assistant')} strong`)).toHaveText('Visible reply');
  await expect(page.locator(`#post-${message('assistant')} li`)).toHaveText('ordered item');
  await expect(page.locator(`#post-${message('intermediate')} .post-content`)).toHaveText('Let me inspect.');
  await expect(page.locator(`#post-${message('error')} .post-author`)).toHaveText('System');
  await expect(page.locator('.timeline')).not.toContainText('[tool_call:');await expect(page.locator('.timeline')).not.toContainText('PRIVATE TOOL OUTPUT');
  await expect(page.locator('.gi-tool-activity')).toHaveCount(0);
 };
 try{
  await page.goto(BASE_URL);await assertConversation();await expect(page.locator('.agent-status-panel')).toHaveCount(0);
  await page.locator('.compose-box textarea').fill('unsent draft retained');await page.reload();await assertConversation();await expect(page.locator('.compose-box textarea')).toHaveValue('unsent draft retained');
  const raw=await(await api.get(BASE_URL+`/api/sessions/${id}/messages`)).json();expect(raw.messages).toHaveLength(5);expect(raw.messages.find((m:any)=>m.role==='tool_result').content).toContain('PRIVATE TOOL OUTPUT');
  const search=await(await api.get(BASE_URL+`/api/sessions/${id}/search?q=PRIVATE&view=conversation`)).json();expect(search.messages).toEqual([]);
  const originalMessages=await(await api.get(BASE_URL+`/api/sessions/${id}/messages?view=conversation&limit=50`)).json();
  const liveId=message('live-error');const persisted={id:liveId,session_id:id,role:'system',content:'Live system notice',created_at:'2026-01-02T00:00:00Z',payload:{}};
  await page.route(`**/api/sessions/${id}/messages?*`,r=>r.fulfill({json:{...originalMessages,messages:[...originalMessages.messages,persisted]}}));
  // Deliver through the real mounted EventSource listener, then reconcile to
  // an equivalent read response. No backend mutation is forged by the event.
  await page.evaluate(({id,liveId})=>{for(const s of (window as any).__conversationStreams)s.dispatchEvent(new MessageEvent('new_post',{data:JSON.stringify({id:liveId,chat_jid:`gi:${id}`,sender:'system',timestamp:'2026-01-02T00:00:00Z',data:{type:'system_message',content:'Live system notice'}})}));},{id,liveId});
  await expect(page.locator(`#post-${liveId} .post-author`)).toHaveText('System');await expect(page.locator(`#post-${liveId}`)).toHaveClass(/agent-post/);
  await page.unroute(`**/api/sessions/${id}/messages?*`);
  active=true;await page.reload();await expect(page.getByText('Waiting for model...',{exact:false})).toBeVisible();await expect(page.locator('.gi-tool-activity')).toHaveCount(0);
  await page.evaluate(id=>{for(const s of (window as any).__conversationStreams){s.dispatchEvent(new MessageEvent('agent_status',{data:JSON.stringify({chat_jid:`gi:${id}`,turn_id:'fixture-turn',status:'running',title:'Writing response'})}));s.dispatchEvent(new MessageEvent('agent_draft_delta',{data:JSON.stringify({chat_jid:`gi:${id}`,turn_id:'fixture-turn',delta:'PARTIAL DRAFT PREVIEW'})}));s.dispatchEvent(new MessageEvent('agent_thought_delta',{data:JSON.stringify({chat_jid:`gi:${id}`,turn_id:'fixture-turn',delta:'PARTIAL THOUGHT PREVIEW'})}));}},id);
  await expect(page.locator('.agent-status-panel')).toContainText('PARTIAL DRAFT PREVIEW');
  active=false;
  await page.evaluate(id=>{for(const s of (window as any).__conversationStreams)s.dispatchEvent(new MessageEvent('agent_status',{data:JSON.stringify({chat_jid:`gi:${id}`,turn_id:'fixture-turn',status:'idle',title:'',phase:'failed'})}));},id);
  await expect(page.locator('.agent-status-panel')).toHaveCount(0);await expect(page.locator('.compose-box textarea')).toHaveValue('unsent draft retained');
  await page.reload();await expect(page.locator('.agent-status-panel')).toHaveCount(0);expect(errors).toEqual([]);
 }finally{await browser.close();}
});

for (const name of ['chromium', 'webkit'] as const) test(`${name} failed provider prefix survives reload without exposing raw partials`, async () => {
 const browser = await ({chromium, webkit}[name]).launch({headless: true});
 const page = await browser.newPage();
 const api = page.request;
 try {
  const created = await api.post(BASE_URL + '/api/sessions', {data: {agent_id: `partial-${name}`, title: `partial-${name}`}});
  expect(created.ok()).toBe(true);
  const {id} = await created.json();
  const rowId = `partial-${name}-${Date.now()}`;
  const content = 'Retained streamed prefix after error';
  const payload = {kind: 'chat', source: 'inference', adopted: false, partial: true,
   provider_message: {role: 'assistant', content: [], stopReason: 'error', errorMessage: 'provider failed'},
   provider_partial: {role: 'assistant', content: [{type: 'text', text: content}, {type: 'thinking', thinkingSignature: 'PRIVATE_PARTIAL_SIGNATURE'}, {type: 'toolCall', name: 'shell', arguments: {secret: 'PRIVATE_TOOL_PREFIX'}}]}};
  const quote = (s: string) => `'${s.replaceAll("'", "''")}'`;
  const db = resolve(process.env.GI_TEST_DB || join(process.env.GI_TEST_RUN_ROOT!, 'instance', 'gi.db'));
  const seed = spawnSync('sqlite3', ['-cmd', '.timeout 5000', db, `insert into messages(id,session_id,role,content,payload_json,created_at) values(${quote(rowId)},${quote(id)},'assistant',${quote(content)},${quote(JSON.stringify(payload))},'2026-01-02T00:00:00Z');`], {encoding: 'utf8'});
  expect(seed.status, seed.stderr).toBe(0);
  await page.addInitScript(s => localStorage.setItem('gi_session_id', s), id);
  await page.goto(BASE_URL);
  for (let reload = 0; reload < 2; reload++) {
   if (reload) await page.reload();
   await expect(page.locator(`#post-${rowId} .post-content`)).toContainText(content);
   await expect(page.locator('.timeline')).not.toContainText('PRIVATE_TOOL_PREFIX');
   for (const suffix of ['messages?view=conversation', 'search?view=conversation&q=Retained']) {
    const response = await api.get(BASE_URL + `/api/sessions/${id}/` + suffix);
    expect(response.ok()).toBe(true);
    const body = await response.text();
    expect(body).toContain(content);
    expect(body).not.toContain('provider_partial');
    expect(body).not.toContain('PRIVATE_PARTIAL_SIGNATURE');
    expect(body).not.toContain('PRIVATE_TOOL_PREFIX');
   }
  }
  const raw = await (await api.get(BASE_URL + `/api/sessions/${id}/messages`)).text();
  expect(raw).toContain('PRIVATE_PARTIAL_SIGNATURE');
 } finally { await browser.close(); }
});
