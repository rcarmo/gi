/**
 * 06-workspace.spec.ts — Verify workspace browser and file APIs.
 *
 * Tests that the workspace tree API works, files can be read, and the
 * workspace explorer component renders in the UI.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, apiGet } from './helpers';

test('Native workspace editor preserves the conversation draft while switching retained tabs',async({page,request})=>{
 const token=Date.now(),paths=[`functional-transition-${token}.md`,`functional-second-${token}.md`];
 for(const path of paths){const r=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'# Functional preview'}}});expect((await r.json()).error).toBeFalsy();}
 await page.setViewportSize({width:390,height:844});await page.goto(BASE_URL);await waitForAppShell(page);const input=page.locator('.compose-box textarea');await input.fill('Functional retained conversation');
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 for(const path of paths){await page.locator(`.workspace-row[data-path="${path}"] .workspace-label-text`).click();await page.locator('.workspace-edit').click();await expect(page.getByRole('tab').filter({hasText:path})).toBeVisible();}
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Hide workspace',exact:true}).click();
 const tabs=page.getByRole('tablist');await expect(tabs.getByRole('tab')).toHaveCount(2);
 await tabs.getByRole('tab').filter({hasText:paths[0]}).locator('.tab-label').click();await expect(page.locator('.cm-editor')).toContainText('Functional preview');await expect(input).toHaveValue('Functional retained conversation');
 for(const path of paths){const close=tabs.getByRole('button',{name:`Close ${path}`,exact:true});await close.focus();await close.press('Enter');}
 await expect(tabs).toHaveCount(0);await expect(input).toBeFocused();await expect(input).toHaveValue('Functional retained conversation');
});

test.describe('Workspace', () => {

  test('workspace tree API returns valid structure', async ({ request }) => {
    const tree = await apiGet(request, '/api/workspace/tree');
    expect(tree.name).toBeTruthy();
    expect(tree.type).toBe('dir');
    expect(tree.path).toBe('.');
  });

  test('workspace tree includes seeded config files', async ({ request }) => {
    const tree = await apiGet(request, '/api/workspace/tree');
    const childNames = (tree.children || []).map((c: any) => c.name);
    // The test instance seeds .piclaw/ and .pi/
    expect(childNames).toContain('.piclaw');
    expect(childNames).toContain('.pi');
  });

  test('workspace file API reads seeded config', async ({ request }) => {
    const file = await apiGet(request, '/api/workspace/file?path=.piclaw/config.json');
    expect(file.path).toBe('.piclaw/config.json');
    expect(file.content).toContain('Gi Test');
  });

  test('workspace file API rejects path traversal', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/workspace/file?path=../../etc/passwd`);
    expect(res.status()).toBe(400);
  });

  test('workspace file API returns error for missing file', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/workspace/file?path=nonexistent.txt`);
    expect(res.status()).toBe(500); // os.ReadFile error
  });

  test('workspace toggle button exists in UI', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const toggle = page.locator('.workspace-toggle-tab');
    await expect(toggle).toBeVisible({ timeout: 5000 });
  });
});

test('workspace preview exposes bounded metadata and safe raw downloads',async({page,request})=>{
 const path='functional-preview.md';
 const written=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content:'# Native preview\n\n**verified**'}}});expect((await written.json()).error).toBeFalsy();
 const response=await request.get(`${BASE_URL}/api/workspace/file?path=${path}&max_bytes=8`);expect(response.ok()).toBe(true);
 const preview=await response.json();expect(preview).toMatchObject({path,kind:'text',content_type:'text/markdown',text:'# Native',truncated:true});expect(preview.mtime).toBeTruthy();expect(preview.size).toBeGreaterThan(8);
 const raw=await request.get(`${BASE_URL}/api/workspace/raw?path=${path}`);expect(await raw.text()).toBe('# Native preview\n\n**verified**');expect(raw.headers()['content-disposition']).toContain('attachment;');expect(raw.headers()['content-security-policy']).toContain('sandbox');
 await page.goto(BASE_URL);await waitForAppShell(page);await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 await page.locator(`.workspace-row[data-path="${path}"] .workspace-label-text`).click();await expect(page.locator('.workspace-preview-body h1')).toHaveText('Native preview');
});

test('workspace subtree queries honour depth and hidden files through the visible menu',async({page,request})=>{
 const folder='functional-hidden-tree';
 for(const path of [`${folder}/nested/deep/leaf.txt`,`${folder}/.hidden.txt`]){
  const r=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content:path}}});expect((await r.json()).error).toBeFalsy();
 }
 const query=async(hidden:boolean)=>(await request.get(`${BASE_URL}/api/workspace/tree?path=${folder}&depth=1&show_hidden=${hidden}`)).json();
 expect((await query(false)).children.map((n:any)=>n.name)).toEqual(['nested']);
 expect((await query(true)).children.map((n:any)=>n.name)).toEqual(['nested','.hidden.txt']);
 const deep=await (await request.get(`${BASE_URL}/api/workspace/tree?path=${folder}/nested/deep&depth=1&show_hidden=false`)).json();expect(deep.children[0].name).toBe('leaf.txt');
 await page.goto(BASE_URL);await waitForAppShell(page);await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 await page.locator(`.workspace-row[data-path="${folder}"] .workspace-caret`).click();
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show hidden files',exact:true}).click();
 await expect(page.locator(`.workspace-row[data-path="${folder}/.hidden.txt"]`)).toBeVisible();
});

test('explicit native workspace reindex supplies scoped lexical results without chat submission',async({page,request})=>{
 for(const [path,content] of [['notes/functional-index.md','functionalorchid source'],['.pi/skills/functional/SKILL.md','functionalviolet skill']]){
  const r=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content}}});expect((await r.json()).error).toBeFalsy();
 }
 await page.goto(BASE_URL);await waitForAppShell(page);await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 const response=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/workspace/index');
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Reindex workspace',exact:true}).click();
 const indexed=await response;expect(indexed.status()).toBe(200);const ready=await indexed.json();expect(ready.state).toBe('ready');expect(ready.indexed_file_count).toBeGreaterThanOrEqual(2);
 const result=await (await request.get(`${BASE_URL}/api/workspace/search?scope=all&q=functionalorchid`)).json();expect(result.mode).toBe('fts');expect(result.hits.map((h:any)=>h.path)).toEqual(['notes/functional-index.md']);expect(result.hits[0].start_line).toBe(1);
 const saved=await (await request.get(`${BASE_URL}/api/workspace/index`)).json();expect(saved.generation).toBe(ready.generation);await expect(page.locator('.workspace-index-status-row')).toHaveCount(0);
});

test('workspace index runtime defaults add no implicit extra roots; absent built-in roots are optional',async({request})=>{
 const runtime=await (await request.get(`${BASE_URL}/api/runtime/config`)).json();
 expect(runtime.workspace_index).toEqual({extraRoots:null,extraExtensions:null,optionalRoots:null});
 const state=await (await request.get(`${BASE_URL}/api/workspace/index`)).json();
 expect(state.roots).toEqual(['.pi/skills','notes']);
 // Built-in roots missing from the workspace are optional instead of failing
 // every scan (465d581d); present ones stay required.
 const optional=state.optional_roots??[];
 for(const root of optional)expect(['.pi/skills','notes']).toContain(root);
 expect(state.required_roots).toBe(optional.length===0);
});

test('workspace index reads keep edited bytes stale until the next explicit application refresh',async({request})=>{
 const path='notes/lifecycle-snapshot.md';
 const write=async(content:string)=>{const r=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content}}});expect((await r.json()).error).toBeFalsy();};
 const refresh=async()=>{const r=await request.post(`${BASE_URL}/api/workspace/index?scope=notes`);expect(r.status()).toBe(200);return r.json();};
 const search=async(q:string)=>(await request.get(`${BASE_URL}/api/workspace/search?scope=notes&q=${q}`)).json();
 await write('lifecycleoldorchid');const before=await refresh();expect(before.state).toBe('ready');
 await write('lifecyclenewviolet');
 expect((await search('lifecycleoldorchid')).hits.map((h:any)=>h.path)).toEqual([path]);
 expect((await search('lifecyclenewviolet')).hits).toEqual([]);
 const unchanged=await (await request.get(`${BASE_URL}/api/workspace/index?scope=notes`)).json();
 expect(unchanged.state).toBe('stale');expect(unchanged.generation).toBe(before.generation);expect(unchanged.last_indexed_at).toBe(before.last_indexed_at);
 const after=await refresh();expect(after.generation).toBe(before.generation+1);expect(after.state).toBe('ready');
 expect((await search('lifecyclenewviolet')).hits.map((h:any)=>h.path)).toEqual([path]);expect((await search('lifecycleoldorchid')).hits).toEqual([]);
});

test('Native Markdown editor preview renders source and refreshes clean external writes',async({page,request})=>{
 const path=`functional-tab-${Date.now()}.md`;const written=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'# Native tab proof\n\nUse `inline proof` as code.'}}});expect((await written.json()).error).toBeFalsy();
 await page.goto(BASE_URL);const input=page.locator('.compose-box textarea');await input.fill('retained tab draft');
 // Exercise the native editor directly; explorer actions have shared coverage.
 await page.evaluate(path=>window.dispatchEvent(new CustomEvent('piclaw-extension-ui:request',{detail:{payload:{request_id:'markdown-preview-open',kind:'custom',options:{action:'open_workspace_file',path,target:'tab'}}}})),path);
 const tab=page.getByRole('tab').filter({hasText:path});await expect(page.locator('.cm-editor')).toContainText('Native tab proof');await tab.click({button:'right'});await page.locator('.tab-context-menu').getByRole('button',{name:'Preview',exact:true}).click();
 const preview=page.locator('.md-preview-panel');await expect(preview.getByRole('heading',{name:'Native tab proof',exact:true})).toBeVisible();await expect(preview.locator('[contenteditable=true],textarea')).toHaveCount(0);
 const code=preview.locator('p code');await expect(code).toHaveText('inline proof');expect(await code.evaluate(el=>getComputedStyle(el).fontFamily)).toContain('monospace');expect(await code.evaluate(el=>getComputedStyle(el).fontFamily===getComputedStyle(el.parentElement!).fontFamily)).toBe(false);
 const updated=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'# Refreshed tab bytes\n\nγ'}}});expect((await updated.json()).error).toBeFalsy();
 await expect(preview.getByRole('heading',{name:'Refreshed tab bytes'})).toBeVisible();await expect(preview.getByRole('heading',{name:'Native tab proof'})).toHaveCount(0);
 await preview.getByRole('button',{name:'Close preview',exact:true}).click();await expect(preview).toHaveCount(0);await tab.getByRole('button',{name:`Close ${path}`,exact:true}).click();await expect(input).toHaveValue('retained tab draft');await expect(input).toBeFocused();
});

test('Native pinned workspace tab survives Close All and preserves draft', async ({page,request}) => {
 const token=`pin-${Date.now()}`,paths=[`${token}-a.md`,`${token}-b.csv`];
 for(const path of paths){const r=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content:'# Pinned native file'}}});expect((await r.json()).error).toBeFalsy();}
 await page.goto(BASE_URL);const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('pin keeps draft');
 const tabs=page.getByRole('tablist'),menu=page.locator('.tab-context-menu');
 for(const path of paths){
  await page.getByTestId('hamburger').click();const show=page.getByRole('menuitem',{name:'Show workspace',exact:true});if(await show.count())await show.click();else await page.keyboard.press('Escape');
  await page.locator('.workspace-tree').getByText(path,{exact:true}).click();await page.locator('.workspace-edit').click();
 }
 if(!await page.locator('.app-shell.workspace-collapsed').count()){await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Hide workspace',exact:true}).click();}
 const a=page.locator('.tab-item').filter({has:page.locator('.tab-label',{hasText:paths[0]})});await a.click({button:'right'});await menu.getByRole('button',{name:'Pin',exact:true}).click();await expect(a).toHaveClass(/pinned/);
 await a.click({button:'right'});await menu.getByRole('button',{name:'Close All',exact:true}).click();await expect(page.locator('.tab-item')).toHaveCount(1);await expect(a).toHaveClass(/active/);
 await a.getByRole('button',{name:`Close ${paths[0]}`,exact:true}).click();await expect(tabs).toHaveCount(0);await expect(input).toHaveValue('pin keeps draft');await expect(input).toBeFocused();
});

test('editor compatibility API reads complete text and saves without changing unchanged files', async ({request}) => {
 const path=`functional-editor-api-${Date.now()}.md`,content='αβ editor bytes\n'.repeat(3000);
 expect((await request.post('/api/workspace/file',{data:{path:'.',name:path,content}})).status()).toBe(200);
 try {
  const original=await (await request.get(`/workspace/file?path=${path}&mode=edit&max=1000000`)).json();
  expect(original).toMatchObject({path,text:content,content,truncated:false});
  const unchanged=await request.put('/workspace/file',{data:{path,content,expected_revision:original.revision}});expect(unchanged.status()).toBe(200);expect((await unchanged.json()).mtime).toBe(original.mtime);
  const changed=await request.put('/workspace/file',{data:{path,content:'saved by native editor API',expected_revision:original.revision}});expect(changed.status()).toBe(200);
  expect(await (await request.get(`/workspace/raw?path=${path}`)).text()).toBe('saved by native editor API');
  expect((await (await request.get(`/workspace/stat?path=${path}`)).json()).size).toBe(26);
  expect((await request.get('/workspace/file?path=../../etc/passwd&mode=edit')).status()).toBe(400);
  const oversize=await request.put('/workspace/file',{data:{path,content:'x'.repeat(262145)}});expect(oversize.status()).toBe(400);
  expect(await (await request.get(`/workspace/raw?path=${path}`)).text()).toBe('saved by native editor API');
 } finally { expect((await request.delete(`/api/workspace/file?path=${path}`)).status()).toBe(200); }
});

test('editor workspace SSE observes external tool writes without publishing file contents', async ({page,request}) => {
 const created=await request.post('/api/sessions',{data:{agent_id:`editor-sse-${Date.now()}`}});expect(created.status()).toBe(201);const session=(await created.json()).id;
 const path=`editor-external-${Date.now()}.md`;
 await page.goto(BASE_URL);await waitForAppShell(page);
 await page.evaluate(async id=>{
  const source=new EventSource(`/sse/stream?chat_jid=gi:${id}`);
  (window as any).__editorChanges=[];(window as any).__editorEvents=source;
  source.addEventListener('workspace_update',e=>(window as any).__editorChanges.push(JSON.parse((e as MessageEvent).data)));
  await new Promise<void>((resolve,reject)=>{source.addEventListener('connected',()=>resolve(),{once:true});source.onerror=()=>reject(new Error('SSE failed'));});
 },session);
 try {
  const written=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'private editor contents'}}});expect((await written.json()).error).toBeFalsy();
  await expect.poll(()=>page.evaluate(p=>(window as any).__editorChanges.some((x:any)=>x.updates.some((u:any)=>u.changed_paths?.includes(p)||u.changed_paths?.includes('.'))),path)).toBe(true);
  const updates=await page.evaluate(()=>(window as any).__editorChanges);expect(JSON.stringify(updates)).not.toContain('private editor contents');
 } finally { await page.evaluate(()=>(window as any).__editorEvents.close());await request.delete(`/api/workspace/file?path=${path}`); }
});

import { checkWorkspaceMotion } from '../ux/support/workspace-motion.mjs';
test('workspace close stays left-anchored throughout the animation', async ({ page }, info) => {
  await page.setViewportSize({width:1440,height:900});
  await page.goto('/');
  const input=page.locator('.compose-box textarea');
  await expect(input).toBeVisible();
  const text='workspace motion must not submit this draft';
  await input.fill(text);
  let posts=0;
  page.on('request',r=>{if(r.method()==='POST'&&/\/api\/sessions\/[^/]+\/(prompt|queue|steer)$/.test(new URL(r.url()).pathname))posts++;});
  await checkWorkspaceMotion(page,info,async()=>{await expect(input).toHaveValue(text);expect(posts).toBe(0);});
});
