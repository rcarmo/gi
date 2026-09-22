import { test, expect } from '@playwright/test';
import { loadCorpus } from './support/catalogue.mjs';

const inputName = 'Message (Enter to send, Shift+Enter for newline)...';
const attachment = name => ({ name, mimeType: 'text/plain', buffer: Buffer.from(`contents of ${name}`) });
async function fixture(page, request, info) {
  const agent = `draft-${info.project.name}-${info.title.replace(/[^a-z0-9]/gi, '').slice(0,24)}`;
  const main = await (await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } })).json();
  const fork = await (await request.post(`/api/sessions/${main.id}/fork`, { data: { title: `${agent}-child`, agent_id: `${agent}-child` } })).json();
  const child = fork.branch.chat_jid.slice(3);
  await page.addInitScript(id => { if (!localStorage.getItem('gi_session_id')) localStorage.setItem('gi_session_id', id); }, main.id);
  await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true });
  await expect(input).toBeVisible();
  const switchTo = async id => {
    await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
    await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();
    await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(id);
  };
  return { main, child, input, switchTo };
}
async function storedDraft(page, id) {
  return page.evaluate(async id => {
    const db = await new Promise((resolve,reject) => {
      const request = indexedDB.open('gi-session-drafts',1);
      request.onsuccess = () => resolve(request.result); request.onerror = () => reject(request.error);
    });
    return new Promise((resolve,reject) => {
      const tx = db.transaction('drafts','readonly'); const req = tx.objectStore('drafts').get(id);
      tx.oncomplete = () => { db.close(); resolve(req.result ? { text:req.result.draft.text, media:req.result.draft.media.map(f=>f.name), refs:req.result.draft.messageRefs, pending:req.result.pending.length } : null); };
      tx.onerror = () => reject(tx.error);
    });
  }, id);
}
async function evidence(info, id) {
  const scenario = loadCorpus().find(row => row.id === id);
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType:'text/plain' });
}
async function attachWorkspaceFile(page, request) {
  const response = await request.post('/api/tools/execute', {data:{tool:'write', input:{path:'draft-reference.txt',content:'Draft reference fixture'}}});
  expect(response.ok()).toBe(true);
  await page.getByTestId('hamburger').click();
  await page.getByRole('menuitem', {name:'Show workspace',exact:true}).click();
  await page.locator('.workspace-row[data-path="draft-reference.txt"]').click();
  await page.getByTestId('hamburger').click();
  await page.getByRole('menuitem', {name:'Hide workspace',exact:true}).click();
  await expect(page.locator('.compose-file-pill[title="draft-reference.txt"]')).toBeVisible();
}
async function history(request, session) {
  const response = await request.post(`/api/sessions/${session}/prompt`, { data: { prompt:`Reference ${session}`, model:'test-model' } });
  expect(response.status()).toBe(202);
  await expect.poll(async () => {
    const { messages } = await (await request.get(`/api/sessions/${session}/messages`)).json();
    return messages?.length || 0;
  }).toBe(2);
}

test('Gi drafts retain text, file bytes and message references across reload and session switches', async ({page,request},info) => {
  const { main,child,input,switchTo } = await fixture(page,request,info);
  await history(request,main.id); await page.reload();
  await page.locator('.post-time').first().click();
  await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);
  await attachWorkspaceFile(page,request);
  await input.fill('A durable draft\n第二行');
  await page.locator('.compose-box input[type=file]').setInputFiles(attachment('draft-a.txt'));
  await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({ text:'A durable draft\n第二行', media:['draft-a.txt'] });
  await switchTo(child); await expect(input).toHaveValue('');
  await input.fill('B durable draft');
  await expect.poll(()=>storedDraft(page,child)).toMatchObject({ text:'B durable draft' });
  await page.reload(); await expect(input).toHaveValue('B durable draft');
  await switchTo(main.id);
  await expect(input).toHaveValue('A durable draft\n第二行');
  await expect(page.locator('.compose-file-pill[title="draft-reference.txt"]')).toBeVisible();
  await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(1);
  await expect(page.locator('.compose-file-pill').filter({hasText:'draft-a.txt'})).toBeVisible();
  // Real upload of the restored File proves bytes survive structured cloning.
  await input.press('Enter');
  await expect.poll(async () => {
    const {media} = await (await request.get(`/api/sessions/${main.id}/media`)).json(); return media?.length || 0;
  }).toBe(1);
  const {media} = await (await request.get(`/api/sessions/${main.id}/media`)).json();
  const bytes = await request.get(`/api/sessions/${main.id}/media/${media[0].id}`);
  expect(await bytes.text()).toBe('contents of draft-a.txt');
  await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'',media:[],pending:0});
  await page.reload(); await expect(input).toHaveValue('');
  await expect(page.locator('.compose-file-pill')).toHaveCount(0);
});

test('@ux-compose-001 Clear captured content while allowing a new draft', async ({page,request},info) => {
  await evidence(info,'@ux-compose-001');
  const {main,input} = await fixture(page,request,info);
  await history(request,main.id); await page.reload();
  await page.locator('.post-time').first().click();
  await attachWorkspaceFile(page,request);
  await input.fill('captured message');
  await page.locator('.compose-box input[type=file]').setInputFiles(attachment('captured.txt'));
  let release, held=false; const gate=new Promise(resolve=>{release=resolve;});
  await page.route(`**/api/sessions/${main.id}/media`,async route=>{
    const response=await route.fetch(); held=true; await gate; await route.fulfill({response});
  });
  try {
    await input.press('Enter'); await expect(input).toHaveValue('');
    await expect(page.locator('.compose-file-pill')).toHaveCount(0);
    await expect.poll(()=>held).toBe(true);
    await input.fill('newer message');
    await page.locator('.compose-box input[type=file]').setInputFiles(attachment('newer.txt'));
    release();
    await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'newer message',media:['newer.txt'],pending:0});
    await expect.poll(async()=>{
      const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();
      return turns.find(t=>t.prompt.startsWith('captured message'))?.status;
    }).toBe('completed');
    const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();
    const sent=turns.find(t=>t.prompt.startsWith('captured message'));
    expect(sent.prompt).toContain('Referenced messages:'); expect(sent.prompt).toContain('captured.txt'); expect(sent.prompt).toContain('Files:\n- draft-reference.txt');
    expect(sent.prompt).not.toContain('newer'); expect(sent.metadata.media).toHaveLength(1);
    await expect(input).toHaveValue('newer message');
  } finally {release();}
});

test('@ux-compose-002 Restore a failed submission alongside newer text', async ({page,request},info) => {
  await evidence(info,'@ux-compose-002');
  const {main,child,input,switchTo}=await fixture(page,request,info);
  await history(request,main.id); await page.reload();
  await page.locator('.post-time').first().click();
  await attachWorkspaceFile(page,request);
  await input.fill('failed original');
  await page.locator('.compose-box input[type=file]').setInputFiles(attachment('original.txt'));
  let release,held=false; const gate=new Promise(resolve=>{release=resolve;});
  await page.route(`**/api/sessions/${main.id}/prompt`,async route=>{held=true;await gate;await route.abort('failed');});
  try {
    await input.press('Enter'); await expect.poll(()=>held).toBe(true);
    await input.fill('newer draft');
    await page.locator('.post-time').last().click();
    await page.locator('.compose-box input[type=file]').setInputFiles(attachment('newer.txt'));
    await switchTo(child); await input.fill('B must stay untouched');
    release();
    await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'failed original\n\nnewer draft',media:['original.txt','newer.txt'],pending:0});
    await expect(input).toHaveValue('B must stay untouched');
    await expect(page.getByRole('alert')).toHaveCount(0);
    await switchTo(main.id);
    await expect(input).toHaveValue('failed original\n\nnewer draft');
    await expect(page.locator('.compose-file-pill[title="draft-reference.txt"]')).toBeVisible();
    await expect(page.locator('.compose-file-pill[title^="Message reference:"]')).toHaveCount(2);
    await expect(page.getByRole('alert')).toBeVisible();
    await page.reload();
    await expect(input).toHaveValue('failed original\n\nnewer draft');
    const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();
    expect(turns.filter(t=>t.prompt.startsWith('failed original'))).toHaveLength(0);
  } finally {release();}
});

test('Gi failed send merges into a newer draft on an A-B-A revisit', async ({page,request},info) => {
  const {main,child,input,switchTo}=await fixture(page,request,info);
  let release,held=false; const gate=new Promise(resolve=>{release=resolve;});
  await page.route(`**/api/sessions/${main.id}/prompt`,async route=>{held=true;await gate;await route.abort('failed');});
  try {
    await input.fill('old A send');await input.press('Enter');await expect.poll(()=>held).toBe(true);
    await switchTo(child);await input.fill('B draft');await switchTo(main.id);
    await input.fill('new visit draft');release();
    await expect(input).toHaveValue('old A send\n\nnew visit draft');
    await input.fill('edited after recovery');
    await switchTo(child);await expect(input).toHaveValue('B draft');
    await switchTo(main.id);await expect(input).toHaveValue('edited after recovery');
    await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'edited after recovery',pending:0});
  } finally {release();}
});

test('Gi acknowledgement cleanup failure warns without restoring a delivered draft', async ({page,request},info) => {
  await page.addInitScript(()=>{
    const put=IDBObjectStore.prototype.put;
    let captured=false;
    IDBObjectStore.prototype.put=function(value,...args){
      if(this.name==='drafts') {
        if(value.pending?.length)captured=true;
        else if(captured)throw new DOMException('Cleanup quota failure','QuotaExceededError');
      }
      return put.call(this,value,...args);
    };
  });
  const {main,input}=await fixture(page,request,info);
  await input.fill('acknowledged once');await input.press('Enter');
  await expect(page.getByRole('alert').filter({hasText:'Send acknowledged, but draft cleanup failed'})).toBeVisible();
  await expect(input).toHaveValue('');
  const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();
  expect(turns.filter(t=>t.prompt==='acknowledged once')).toHaveLength(1);
});

test('@ux-compose-003 Reject an entirely empty submission', async ({page,request},info) => {
  await evidence(info,'@ux-compose-003');
  const {main,input}=await fixture(page,request,info);
  let submits=0; page.on('request',req=>{if(req.method()==='POST' && req.url().endsWith('/prompt'))submits++;});
  await input.fill('   '); await input.press('Enter');
  await expect(input).toHaveValue('   ');
  await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'   ',pending:0});
  expect(submits).toBe(0);
  const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();expect(turns||[]).toHaveLength(0);
});

test('@ux-compose-006 Submit captures the destination chat', async ({page,request},info) => {
  await evidence(info,'@ux-compose-006');
  const {main,child,input,switchTo}=await fixture(page,request,info);
  await input.fill('origin only'); await page.locator('.compose-box input[type=file]').setInputFiles(attachment('origin.txt'));
  let release,held=false;const gate=new Promise(resolve=>{release=resolve;});
  await page.route(`**/api/sessions/${main.id}/media`,async route=>{const response=await route.fetch();held=true;await gate;await route.fulfill({response});});
  try {
    await input.press('Enter');await expect.poll(()=>held).toBe(true);
    await switchTo(child);await input.fill('target draft');release();
    await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'',pending:0});
    await expect.poll(async()=>{
      const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();return turns?.[0]?.status;
    }).toBe('completed');
    await expect(input).toHaveValue('target draft');
    const {turns}=await (await request.get(`/api/sessions/${child}/turns`)).json();expect(turns||[]).toHaveLength(0);
    const {media}=await (await request.get(`/api/sessions/${child}/media`)).json();expect(media||[]).toHaveLength(0);
  } finally {release();}
});

test('Gi reload recovers an unacknowledged send without resubmitting it', async ({page,request},info) => {
  const {main,input}=await fixture(page,request,info);
  let release,held=false;const gate=new Promise(resolve=>{release=resolve;});
  await page.route(`**/api/sessions/${main.id}/prompt`,async route=>{held=true;await gate;await route.abort('failed').catch(()=>{});});
  try {
    await input.fill('uncertain send');await input.press('Enter');await expect.poll(()=>held).toBe(true);
    await input.fill('new draft');await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'new draft',pending:1});
    await page.reload();
    await expect(input).toHaveValue('uncertain send\n\nnew draft');
    await expect(page.getByRole('alert')).toContainText('Delivery is unknown');
    release();await page.unroute(`**/api/sessions/${main.id}/prompt`);
    await page.reload();await expect(input).toHaveValue('uncertain send\n\nnew draft');
    const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();expect(turns||[]).toHaveLength(0);
  } finally {release();}
});

test('Gi storage failure retains draft and prevents an unprotected send', async ({page,request},info) => {
  // Storage fault injection only; UI state and timeline remain native.
  await page.addInitScript(()=>{const original=IDBObjectStore.prototype.put;IDBObjectStore.prototype.put=function(...args){if(this.name==='drafts')throw new DOMException('Test quota exhausted','QuotaExceededError');return original.apply(this,args);};});
  const {main,input}=await fixture(page,request,info);
  await input.fill('keep on quota failure');await input.press('Enter');
  await expect(input).toHaveValue('keep on quota failure');
  await expect(page.getByRole('alert').filter({hasText:'Draft not saved'})).toBeVisible();
  const {turns}=await (await request.get(`/api/sessions/${main.id}/turns`)).json();expect(turns||[]).toHaveLength(0);
});

// Hold only a real accepted HTTP response; native persistence, SSE and provider
// execution continue. No invented post payload or replacement EventSource.
async function holdAcknowledgement(page, session) {
  let release, held = false, responseStatus;
  const gate = new Promise(resolve => { release = resolve; });
  const pattern = `**/api/sessions/${session}/prompt`;
  await page.route(pattern, async route => {
    const response = await route.fetch();
    responseStatus = response.status(); held = true;
    await gate; await route.fulfill({ response });
  });
  return { release, held: () => held, status: () => responseStatus,
    async close() { release(); await page.unrouteAll({ behavior: 'wait' }); } };
}
async function completedPrompt(request, session, prompt) {
  await expect.poll(async () => {
    const { turns } = await (await request.get(`/api/sessions/${session}/turns`)).json();
    return turns?.find(turn => turn.prompt === prompt)?.status;
  }).toBe('completed');
  // The completed row precedes release of the engine's active claim. A new
  // prompt sent before idle is steering, not a fresh history-building turn.
  await expect.poll(async () => (await (await request.get(`/api/sessions/${session}/activity`)).json()).status).toBe('idle');
}
async function messages(request, session) {
  return (await (await request.get(`/api/sessions/${session}/messages`)).json()).messages || [];
}

test('@ux-compose-007 Accepted response notifies refresh and displays the stored message', async ({page,request},info) => {
  await evidence(info,'@ux-compose-007');
  const {main,input} = await fixture(page,request,info);
  const ack = await holdAcknowledgement(page,main.id);
  let reads = 0;
  page.on('request',req => { if(new URL(req.url()).pathname === `/api/sessions/${main.id}/messages`) reads++; });
  try {
    await input.fill('accepted native text'); await input.press('Enter');
    await expect.poll(ack.held).toBe(true); expect(ack.status()).toBe(202);
    await completedPrompt(request,main.id,'accepted native text');
    const stored = (await messages(request,main.id)).find(m => m.role === 'user');
    expect(stored.content).toBe('accepted native text');
    await expect(page.locator(`#post-${stored.id}`)).toContainText(stored.content);
    // SSE has already delivered the completed turn; HTTP acknowledgement still
    // owns a refresh independently (and never adds a synthetic duplicate).
    const before = reads; ack.release();
    await expect.poll(() => reads).toBeGreaterThan(before);
    await expect.poll(() => storedDraft(page,main.id)).toMatchObject({pending:0});
    await expect(page.locator(`#post-${stored.id}`)).toHaveCount(1);
    await page.reload(); await expect(page.locator(`#post-${stored.id}`)).toContainText(stored.content);
  } finally { await ack.close(); }
});

test('@ux-compose-009 Multiple native uploads retain filename, identifier and byte association', async ({page,request},info) => {
  await evidence(info,'@ux-compose-009');
  const {main,input} = await fixture(page,request,info);
  const files = [attachment('first α.txt'), attachment('second β.txt'), attachment('third.txt')];
  const uploads = [];
  page.on('response', response => {
    if(response.request().method()==='POST' && response.url().endsWith(`/api/sessions/${main.id}/media`))
      uploads.push(response.json().then(data=>data.media));
  });
  const sent = page.waitForRequest(req => req.method()==='POST' && req.url().endsWith(`/api/sessions/${main.id}/prompt`));
  await input.fill('  attachment batch  ');
  await page.locator('.compose-box input[type=file]').setInputFiles(files);
  await input.press('Enter'); const body = (await sent).postDataJSON();
  expect(uploads).toHaveLength(files.length);
  const uploaded = await Promise.all(uploads);
  expect(new Set(uploaded.map(u=>u.id)).size).toBe(files.length);
  expect(body.media).toEqual(uploaded.map(u=>({media_id:u.id,session_id:main.id})));
  expect(body.prompt).toBe('attachment batch\n\nAttachments:\n'+uploaded.map((u,i)=>`- attachment:${u.id} (${files[i].name})`).join('\n'));
  await completedPrompt(request,main.id,body.prompt);
  const {media} = await (await request.get(`/api/sessions/${main.id}/media`)).json();
  for(let i=0;i<uploaded.length;i++) {
    const native = media.find(m=>m.id===uploaded[i].id); expect(native.filename).toBe(files[i].name);
    const bytes = await request.get(`/api/sessions/${main.id}/media/${uploaded[i].id}`);
    expect(await bytes.body()).toEqual(files[i].buffer);
  }
  const stored = (await messages(request,main.id)).find(m=>m.role==='user');
  expect(stored.content).toBe(body.prompt);
  await expect(page.locator(`#post-${stored.id}`)).toContainText('second β.txt');
  await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'',media:[],pending:0});
});

test('@ux-compose-010 Earlier acknowledgement never resets newer typing or attachments', async ({page,request},info) => {
  await evidence(info,'@ux-compose-010');
  const {main,input} = await fixture(page,request,info);
  const ack = await holdAcknowledgement(page,main.id);
  try {
    await input.fill('earlier captured'); await input.press('Enter');
    await expect(input).toHaveValue(''); await expect.poll(ack.held).toBe(true);
    await completedPrompt(request,main.id,'earlier captured');
    await input.fill('new draft\n第二行');
    await page.locator('.compose-box input[type=file]').setInputFiles(attachment('new-unsent.txt'));
    await input.press('Home'); await input.press('ArrowRight');
    const cursor = await input.evaluate(el => [el.selectionStart,el.selectionEnd]);
    ack.release();
    await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'new draft\n第二行',media:['new-unsent.txt'],pending:0});
    await expect(input).toHaveValue('new draft\n第二行');
    expect(await input.evaluate(el => [el.selectionStart,el.selectionEnd])).toEqual(cursor);
    const stored = await messages(request,main.id);
    expect(stored.filter(m=>m.role==='user').map(m=>m.content)).toEqual(['earlier captured']);
    await page.reload(); await expect(input).toHaveValue('new draft\n第二行');
    await expect(page.locator('.compose-file-pill').filter({hasText:'new-unsent.txt'})).toBeVisible();
  } finally { await ack.close(); }
});

test('@ux-compose-011 Native posts reconcile once and respect current history-reading position', async ({page,request},info) => {
  test.setTimeout(90000); await evidence(info,'@ux-compose-011');
  const {main,input} = await fixture(page,request,info);
  for(let i=0;i<16;i++) {
    const prompt = `Reader history ${i}\n` + Array.from({length:8},(_,n)=>`Historical line ${n}`).join('\n');
    expect((await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt,model:'test-model'}})).status()).toBe(202);
    await completedPrompt(request,main.id,prompt);
  }
  await page.reload(); const timeline = page.locator('.timeline');
  await expect(page.locator('.timeline .post')).toHaveCount(32);
  const ack = await holdAcknowledgement(page,main.id);
  try {
    await input.fill('send before reading history'); await input.press('Enter');
    await expect.poll(ack.held).toBe(true); await completedPrompt(request,main.id,'send before reading history');
    await expect(page.locator('.timeline .post')).toHaveCount(34);
    await input.fill('next unsent reader draft');
    await timeline.hover(); await page.mouse.wheel(0,-700);
    await expect.poll(()=>timeline.evaluate(el=>el.scrollTop)).toBeLessThan(-100);
    // WebKit wheel/entry animations must finish before measuring a reading anchor.
    await page.waitForTimeout(350);
    const anchor = await timeline.evaluate(root=>{
      const box=root.getBoundingClientRect();
      const node=[...root.querySelectorAll('.post[id]')].find(el=>{const r=el.getBoundingClientRect();return r.top>=box.top&&r.top<box.bottom;});
      return {id:node.id,top:node.getBoundingClientRect().top};
    });
    ack.release(); await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({pending:0});
    const position = () => page.locator(`[id="${anchor.id}"]`).evaluate(el=>el.getBoundingClientRect().top);
    await expect.poll(async()=>Math.abs(await position()-anchor.top)).toBeLessThanOrEqual(1);
    // Real later arrivals follow the same policy, not a separate response-only path.
    expect((await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'later native arrival',model:'test-model'}})).status()).toBe(202);
    await completedPrompt(request,main.id,'later native arrival');
    await expect(page.locator('.timeline .post')).toHaveCount(36);
    await expect.poll(async()=>Math.abs(await position()-anchor.top)).toBeLessThanOrEqual(1);
    await expect(input).toHaveValue('next unsent reader draft');
    await timeline.hover(); await page.mouse.wheel(0,100000);
    await expect.poll(()=>timeline.evaluate(el=>Math.abs(el.scrollTop))).toBeLessThanOrEqual(1);
    await input.fill('near-bottom submission'); await input.press('Enter');
    await completedPrompt(request,main.id,'near-bottom submission');
    await expect(page.locator('.timeline .post')).toHaveCount(38);
    await expect.poll(()=>timeline.evaluate(el=>Math.abs(el.scrollTop))).toBeLessThanOrEqual(1);
    const native = await messages(request,main.id);
    const ids = await page.locator('.timeline .post').evaluateAll(nodes=>nodes.map(n=>n.id.slice(5)));
    expect(ids).toEqual(native.map(m=>m.id)); expect(new Set(ids).size).toBe(ids.length);
  } finally { await ack.close(); }
});

test('Gi delayed acknowledgement refreshes the current search rather than the hidden timeline', async ({page,request},info) => {
  const {main,input} = await fixture(page,request,info);
  const ack = await holdAcknowledgement(page,main.id);
  try {
    await input.fill('searchable accepted text'); await input.press('Enter');
    await expect.poll(ack.held).toBe(true); await completedPrompt(request,main.id,'searchable accepted text');
    await input.fill('draft beneath search');
    await page.getByRole('button',{name:'Search',exact:true}).click();
    const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});
    await search.fill('no-such-message'); await search.press('Enter');
    await expect(page.getByText('No matching messages.',{exact:true})).toBeVisible();
    let searches=0,pages=0;
    page.on('request',req=>{const path=new URL(req.url()).pathname;if(path.endsWith(`/${main.id}/search`))searches++;if(path.endsWith(`/${main.id}/messages`))pages++;});
    ack.release(); await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({pending:0});
    await expect.poll(()=>searches).toBeGreaterThan(0); expect(pages).toBe(0);
    await expect(search).toHaveValue('no-such-message'); await expect(page.locator('.timeline .post')).toHaveCount(0);
    await search.press('Escape'); await expect(input).toHaveValue('draft beneath search');
    const stored=(await messages(request,main.id)).find(m=>m.role==='user');
    await expect(page.locator(`#post-${stored.id}`)).toContainText(stored.content);
  } finally { await ack.close(); }
});

test('Gi accepted origin response cannot refresh a newly selected chat', async ({page,request},info) => {
  const {main,child,input,switchTo} = await fixture(page,request,info);
  const ack = await holdAcknowledgement(page,main.id);
  try {
    await input.fill('accepted in origin'); await input.press('Enter');
    await expect.poll(ack.held).toBe(true); await completedPrompt(request,main.id,'accepted in origin');
    await switchTo(child); await input.fill('child draft remains');
    await expect.poll(()=>storedDraft(page,child)).toMatchObject({text:'child draft remains'});
    ack.release(); await expect.poll(()=>storedDraft(page,main.id)).toMatchObject({text:'',pending:0});
    await expect(input).toHaveValue('child draft remains'); await expect(page.locator('.timeline .post')).toHaveCount(0);
    expect(await messages(request,child)).toEqual([]);
    await switchTo(main.id); await expect(input).toHaveValue('');
    const stored=(await messages(request,main.id)).find(m=>m.role==='user');
    await expect(page.locator(`#post-${stored.id}`)).toContainText('accepted in origin');
    await expect(page.locator(`#post-${stored.id}`)).toHaveCount(1);
  } finally { await ack.close(); }
});
