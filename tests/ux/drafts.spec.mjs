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
