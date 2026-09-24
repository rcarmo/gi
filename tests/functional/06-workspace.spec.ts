/**
 * 06-workspace.spec.ts — Verify workspace browser and file APIs.
 *
 * Tests that the workspace tree API works, files can be read, and the
 * workspace explorer component renders in the UI.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, apiGet } from './helpers';

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

test('workspace index runtime defaults are strict with no implicit extra roots',async({request})=>{
 const runtime=await (await request.get(`${BASE_URL}/api/runtime/config`)).json();
 expect(runtime.workspace_index).toEqual({extraRoots:null,extraExtensions:null,optionalRoots:null});
 const state=await (await request.get(`${BASE_URL}/api/workspace/index`)).json();
 expect(state.required_roots).toBe(true);expect(state.optional_roots).toBeNull();expect(state.roots).toEqual(['.pi/skills','notes']);
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

test('Read-only workspace tab mounts native content and closes back to the draft',async({page,request})=>{
 const path=`functional-tab-${Date.now()}.md`;const written=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content:'# Native tab proof\n\nUse `inline proof` as code.'}}});expect(written.ok()).toBe(true);expect((await written.json()).error).toBeFalsy();
 await page.goto(BASE_URL);const input=page.locator('textarea').last();await expect(input).toBeVisible();await input.fill('retained tab draft');await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();await page.locator(`.workspace-row[data-path="${path}"] .workspace-label-text`).click();await page.getByRole('button',{name:'Open read-only tab',exact:true}).click();
 const preview=page.getByRole('region',{name:`Read-only preview: ${path}`,exact:true});await expect(preview.getByRole('heading',{name:'Native tab proof',exact:true})).toBeVisible();await expect(preview.locator('[contenteditable=true],textarea')).toHaveCount(0);
 const code=preview.locator('.workspace-preview-text p code');await expect(code).toHaveText('inline proof');expect(await code.evaluate(el=>getComputedStyle(el).fontFamily)).toContain('monospace');expect(await code.evaluate(el=>getComputedStyle(el).fontFamily===getComputedStyle(el.parentElement!).fontFamily)).toBe(false);
 const updated=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content:'# Refreshed tab bytes\n\nγ'}}});expect((await updated.json()).error).toBeFalsy();
 await preview.getByRole('button',{name:'Refresh preview',exact:true}).click();await expect(preview.getByRole('heading',{name:'Refreshed tab bytes'})).toBeVisible();await expect(preview.getByRole('heading',{name:'Native tab proof'})).toHaveCount(0);
 await preview.getByRole('button',{name:'Close preview',exact:true}).click();await expect(preview).toHaveCount(0);await expect(input).toHaveValue('retained tab draft');
});

test('Read-only pinned workspace tab survives Close All and preserves draft', async ({page,request}) => {
 const token=`pin-${Date.now()}`,paths=[`${token}-a.md`,`${token}-b.csv`];
 for(const path of paths){const r=await request.post(`${BASE_URL}/api/tools/execute`,{data:{tool:'write',input:{path,content:'# Pinned native file'}}});expect((await r.json()).error).toBeFalsy();}
 await page.goto(BASE_URL);const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('pin keeps draft');
 const tabs=page.getByRole('tablist',{name:'Editor tabs',exact:true}),menu=page.locator('.tab-context-menu');
 for(const path of paths){
  await page.getByTestId('hamburger').click();const show=page.getByRole('menuitem',{name:'Show workspace',exact:true});if(await show.count())await show.click();else await page.keyboard.press('Escape');
  await page.locator('.workspace-tree').getByText(path,{exact:true}).click();await page.getByRole('button',{name:'Open read-only tab',exact:true}).click();
 }
 if(!await page.locator('.app-shell.workspace-collapsed').count()){await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Hide workspace',exact:true}).click();}
 const a=page.locator('.tab-item').filter({has:page.locator('.tab-label',{hasText:paths[0]})});await a.click({button:'right'});await menu.getByRole('button',{name:'Pin',exact:true}).click();await expect(a).toHaveClass(/pinned/);
 await a.click({button:'right'});await menu.getByRole('button',{name:'Close All',exact:true}).click();await expect(page.locator('.tab-item')).toHaveCount(1);await expect(a).toHaveClass(/active/);
 await a.getByRole('button',{name:`Close ${paths[0]}`,exact:true}).click();await expect(tabs).toHaveCount(0);await expect(input).toHaveValue('pin keeps draft');await expect(input).toBeFocused();
});
