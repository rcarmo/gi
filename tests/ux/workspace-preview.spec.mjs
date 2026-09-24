import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function fixture(page,request,info){
 const id=`preview-${info.project.name}`;
 for(const [path,content] of Object.entries({[`${id}.md`]:'# Native preview\n\n**stored Markdown**', [`${id}.txt`]:'<img src=x onerror="window.__previewUnsafe=true">\nplain <tag> & text', [`${id}.svg`]:'<svg xmlns="http://www.w3.org/2000/svg"><script>window.__previewUnsafe=true</script></svg>'})) {
  const r=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content}}});expect(r.ok()).toBe(true);
 }
 // Binary and image fixtures use the native shell tool; no preview responses are mocked.
 const root=(await (await request.get('/api/runtime/config')).json()).workspace_root;
 expect(typeof root).toBe('string');const cwd="'"+root.replaceAll("'","'\\''")+"'";
 const png='iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jKuoAAAAASUVORK5CYII=';
 const r=await request.post('/api/tools/execute',{data:{tool:'shell',input:{command:`cd ${cwd} && printf '%s' '${png}' | base64 -d > ${id}.png; printf '\\000\\377BINARY' > ${id}.bin; printf 'large text %.0s' $(seq 1 3000) > ${id}-large.txt`}}});expect(r.ok()).toBe(true);expect(await r.json()).not.toHaveProperty('error');
 const s=await (await request.post('/api/sessions',{data:{title:id,agent_id:id}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),s.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await input.fill('preview draft retained');
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 const pick=async ext=>{const path=`${id}${ext}`;await page.locator(`.workspace-row[data-path="${path}"] .workspace-label-text`).click();await expect(page.locator('.workspace-preview-meta')).toContainText(path);return path;};
 return {id,input,pick};
}
test('@ux-workspace-008 Native previews render Markdown, escaped text, images, binary and metadata',async({page,request},info)=>{
 await info.attach('gherkin',{body:loadCorpus().find(s=>s.id==='@ux-workspace-008').steps.join('\n'),contentType:'text/plain'});
 const f=await fixture(page,request,info),host=page.locator('.workspace-preview-body');
 await f.pick('.md');await expect(host.locator('h1')).toHaveText('Native preview');await expect(host.locator('strong').filter({hasText:'stored Markdown'})).toBeVisible();
 for(const token of ['kind: markdown','extension: md','type: text/markdown','size:','modified:','path:']) await expect(host).toContainText(token);
 await f.pick('.txt');await expect(host.locator('pre code')).toHaveText('<img src=x onerror="window.__previewUnsafe=true">\nplain <tag> & text');await expect(host.locator('img')).toHaveCount(0);
 await f.pick('.png');await expect(host.locator('img')).toBeVisible();await expect.poll(()=>host.locator('img').evaluate(el=>el.complete&&el.naturalWidth===1)).toBe(true);await expect(host).toContainText('kind: image');await expect(host).toContainText('type: image/png');
 await page.screenshot({path:info.outputPath('workspace-image-preview.png')});
 await f.pick('.bin');await expect(host).toContainText('Binary file — download to view.');await expect(host).toContainText('kind: binary');await expect(host).toContainText('extension: bin');
 await f.pick('.svg');await expect(host.locator('pre code')).toContainText('<script>');expect(await page.evaluate(()=>window.__previewUnsafe)).toBeUndefined();
 await f.pick('-large.txt');await expect(host).toContainText('truncated');expect((await host.locator('pre code').textContent()).length).toBe(20000);
 // Opening workspace files deliberately adds references; it must not submit or clear text.
 await expect(f.input).toHaveValue('preview draft retained');
 await page.screenshot({path:info.outputPath('workspace-text-preview.png')});
 await page.reload();await expect(f.input).toHaveValue('preview draft retained');
});

test('@ux-workspace-004 Hidden toggle persists and reloads root plus every expanded subtree',async({page,request},info)=>{
 await info.attach('gherkin',{body:loadCorpus().find(s=>s.id==='@ux-workspace-004').steps.join('\n'),contentType:'text/plain'});
 const folder=`tree-${info.project.name}`,deep=`${folder}/nested/deep`;
 const paths=[`${folder}/visible.txt`,`${folder}/.hidden.txt`,`${folder}/nested/visible.txt`,`${folder}/nested/.nested-hidden.txt`,`${deep}/leaf.txt`,`${deep}/.deep-hidden.txt`,`.root-${info.project.name}.txt`];
 for(const path of paths){const r=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:`stored ${path}`}}});expect((await r.json()).error).toBeFalsy();}
 const session=await (await request.post('/api/sessions',{data:{title:folder,agent_id:folder}})).json();
 await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);},session.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await input.fill('tree draft preserved');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'tree-draft.txt',mimeType:'text/plain',buffer:Buffer.from('unsent')});
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 const row=path=>page.locator(`.workspace-row[data-path="${path}"]`);
 for(const path of [folder,`${folder}/nested`,deep]){await row(path).locator('.workspace-caret').click();}
 await expect(row(`${deep}/leaf.txt`)).toBeVisible();
 for(const path of paths.filter(p=>p.split('/').at(-1).startsWith('.')))await expect(row(path)).toHaveCount(0);
 const calls=[];page.on('request',r=>{const u=new URL(r.url());if(u.pathname==='/api/workspace/tree')calls.push({path:u.searchParams.get('path'),hidden:u.searchParams.get('show_hidden'),depth:u.searchParams.get('depth')});});
 const toggle=async(show)=>{
  calls.length=0;await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:show?'Show hidden files':'Hide hidden files',exact:true}).click();
  await expect.poll(()=>page.evaluate(()=>localStorage.getItem('workspaceShowHidden'))).toBe(String(show));
  for(const path of ['.',folder,`${folder}/nested`,deep])await expect.poll(()=>calls.some(r=>r.path===path&&r.hidden===String(show)&&r.depth==='1')).toBe(true);
 };
 await toggle(true);
 for(const path of paths)await expect(row(path)).toBeVisible();
 await page.screenshot({path:info.outputPath('workspace-hidden-visible.png')});
 await toggle(false);for(const path of paths.filter(p=>p.split('/').at(-1).startsWith('.')))await expect(row(path)).toHaveCount(0);await expect(row(`${deep}/leaf.txt`)).toBeVisible();
 await toggle(true);await expect(row(`${deep}/.deep-hidden.txt`)).toBeVisible();
 await page.reload();await expect(input).toHaveValue('tree draft preserved');await expect(page.locator('.compose-file-pill').filter({hasText:'tree-draft.txt'})).toBeVisible();
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 await expect(row(`.root-${info.project.name}.txt`)).toBeVisible();expect(await page.evaluate(()=>localStorage.getItem('workspaceShowHidden'))).toBe('true');
 expect((await (await request.get(`/api/sessions/${session.id}/messages`)).json()).messages ?? []).toEqual([]);
});

test('Gi scoped index reindex/query and missing-root failure preserve session drafts',async({page,request},info)=>{
 const id=info.project.name.replaceAll('-',''),filename=`notes/${id}.md`;
 const write=async(path,content)=>{const r=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content}}});expect((await r.json()).error).toBeFalsy();};
 await write(filename,`orchid${id} stored source`);await write(`.pi/skills/${id}/SKILL.md`,`violet${id} skill`);
 const root=(await (await request.get('/api/runtime/config')).json()).workspace_root;
 const quote=s=>"'"+s.replaceAll("'","'\\''")+"'";
 const shell=async command=>{const r=await request.post('/api/tools/execute',{data:{tool:'shell',input:{command:`cd ${quote(root)} && ${command}`}}});expect((await r.json()).error).toBeFalsy();};
 const session=await (await request.post('/api/sessions',{data:{title:id,agent_id:id}})).json();
 const child=(await (await request.post(`/api/sessions/${session.id}/fork`,{data:{title:'other-index',agent_id:`other-${id}`}})).json()).branch.chat_jid.slice(3);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await input.fill('index draft retained');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'index-unsent.txt',mimeType:'text/plain',buffer:Buffer.from('retain')});
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 const state=async()=>(await (await request.get('/api/workspace/index?scope=all')).json());
 const query=async(q,scope='all')=>(await (await request.get(`/api/workspace/search?scope=${scope}&q=${encodeURIComponent(q)}`)).json());
 const run=async()=>{await page.getByTestId('hamburger').click();const result=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/workspace/index');await page.getByRole('menuitem',{name:'Reindex workspace',exact:true}).click();return result;};
 const result=await run();expect(result.status()).toBe(200);const ready=await result.json();expect(ready.state).toBe('ready');expect(ready.indexed_file_count).toBeGreaterThanOrEqual(2);expect(ready.last_indexed_at).toBeTruthy();
 expect((await query(`orchid${id}`)).hits.map(h=>h.path)).toEqual([filename]);expect((await query(`violet${id}`)).hits.map(h=>h.path)).toEqual([`.pi/skills/${id}/SKILL.md`]);expect((await state()).generation).toBe(ready.generation);
 await expect(page.locator('.workspace-index-status-row')).toHaveCount(0);
 // Removing a required scope root induces the real native scanner failure.
 await shell(`mv notes notes-index-held`);
 try {
  const failed=await run();expect(failed.status()).toBe(500);await expect(page.locator('.workspace-index-status-row')).toContainText('Workspace index failed');
  const saved=await state();expect(saved.state).toBe('failed');expect(saved.generation).toBe(ready.generation);expect(saved.last_indexed_at).toBe(ready.last_indexed_at);
  expect((await query(`orchid${id}`)).hits.map(h=>h.path)).toEqual([filename]);expect((await state()).state).toBe('failed');
  await expect(input).toHaveValue('index draft retained');await expect(page.locator('.compose-file-pill').filter({hasText:'index-unsent.txt'})).toBeVisible();
  await page.screenshot({path:info.outputPath('workspace-index-failure.png')});
  await page.reload();await expect(input).toHaveValue('index draft retained');expect((await state()).state).toBe('failed');
  await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();await expect(page.locator('.workspace-index-status-row')).toContainText('Workspace index failed');
 } finally {await shell('mv notes-index-held notes')}
 await write(filename,`changed${id} new source`);const retried=await run();expect(retried.status()).toBe(200);expect((await retried.json()).generation).toBe(ready.generation+1);expect((await query(`orchid${id}`)).hits).toEqual([]);expect((await query(`changed${id}`)).hits.map(h=>h.path)).toEqual([filename]);
 await expect(input).toHaveValue('index draft retained');
 await page.locator('.workspace-toggle-tab.open').click();
 const select=async target=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${target}"]`).getByRole('menuitem').click();};
 await select(child);await input.fill('other draft');await select(session.id);await expect(input).toHaveValue('index draft retained');await expect(page.locator('.compose-file-pill').filter({hasText:'index-unsent.txt'})).toBeVisible();
 expect((await (await request.get(`/api/sessions/${session.id}/messages`)).json()).messages??[]).toEqual([]);
 // The visible Refresh action reaches the explorer's existing refresh handler.
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 await write(`fresh-${id}.txt`,'tree refresh');await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Refresh tree',exact:true}).click();await expect(page.locator(`.workspace-row[data-path="fresh-${id}.txt"]`)).toBeVisible();
});

test('@shared-3 Workspace show/hide and narrow backdrop preserve the native composer',async({page,request},info)=>{
 const source=loadCorpus('shared').find(s=>s.id==='@shared-3');expect(source.name).toBe('Show and hide the native workspace');await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const token=`workspace-shared-${info.project.name}-${Date.now()}`,path=`${token}.txt`;
 const write=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content:'native tree file'}}});expect(write.ok()).toBe(true);expect((await write.json()).error).toBeFalsy();
 const create=async agent=>{const r=await request.post('/api/sessions',{data:{agent_id:agent,title:agent}});expect(r.status()).toBe(201);return r.json();};
 const main=await create(token),research=await create(`${token}-research`);
 const sent=await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'native workspace reference history',model:'test-model'}});expect(sent.status()).toBe(202);const turn=(await sent.json()).turn_id;
 const get=async(id,part)=>{const r=await request.get(`/api/sessions/${id}/${part}`);expect(r.status()).toBe(200);return r.json();};
 await expect.poll(async()=>(await get(main.id,'turns')).turns.find(t=>t.id===turn)?.status).toBe('completed');
 // Terminal status is persisted before final timestamps/claim cleanup. Capture
 // the stable native fixture, not an intermediate completion snapshot.
 await expect.poll(async()=>Boolean((await get(main.id,'turns')).turns.find(t=>t.id===turn)?.finished_at)).toBe(true);
 await expect.poll(async()=>(await get(main.id,'activity')).status).toBe('idle');
 const before={messages:await get(main.id,'messages'),turns:await get(main.id,'turns'),other:await get(research.id,'messages')};
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const toggle=async show=>{await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:show?'Show workspace':'Hide workspace',exact:true}).click();await expect(page.locator('.timeline-menu-dropdown')).toHaveCount(0);};
 await toggle(true);await page.locator(`.workspace-row[data-path="${path}"]`).click();await toggle(false);
 const link=page.locator('.post .post-time').first(),messageId=(await link.getAttribute('href')).replace(/^#msg-/,'');await link.click();
 const text='workspace draft\nretain text and references',bytes='native unsent attachment bytes';
 await input.fill(text);await page.locator('.compose-box input[type=file]').setInputFiles({name:'workspace-unsent.txt',mimeType:'text/plain',buffer:Buffer.from(bytes)});
 const titles=()=>page.locator('.compose-input-main .compose-file-pill').evaluateAll(nodes=>nodes.map(n=>n.title));
 const expected=[`Message reference: ${messageId}`,path,'workspace-unsent.txt'];await expect.poll(titles).toEqual(expected);
 const stored=()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;if(!s)return resolve(null);resolve({...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:Array.from(new Uint8Array(f.bytes))}))}});};tx.onerror=()=>reject(tx.error);});
 },main.id);
 await expect.poll(stored).toMatchObject({draft:{text,fileRefs:[path],messageRefs:[messageId],media:[{name:'workspace-unsent.txt',bytes:Array.from(Buffer.from(bytes))}]},pending:[]});const draftBefore=await stored();
 const unchanged=async()=>{await expect(input).toHaveValue(text);await expect.poll(titles).toEqual(expected);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);expect(await stored()).toEqual(draftBefore);};
 let posts=0;page.on('request',r=>{if(r.method()==='POST'&&/\/(prompt|queue|steer)$/.test(new URL(r.url()).pathname))posts++;});
 await toggle(true);await expect(page.locator('.workspace-sidebar')).toBeVisible();await expect(page.locator(`.workspace-row[data-path="${path}"]`)).toBeVisible();await unchanged();
 // Gi has no Plan opener. Its absence is explicit; this negative backdrop
 // criterion does not grant any Plan editing or submission feature credit.
 await expect(page.getByRole('button',{name:/^(Open |Show )?Plan$/i})).toHaveCount(0);await expect(page.getByRole('menuitem',{name:/^(Open |Show )?Plan$/i})).toHaveCount(0);
 await toggle(false);await expect(page.locator('.app-shell')).toHaveClass(/workspace-collapsed/);await unchanged();
 const drawer=await page.evaluate(()=>matchMedia('(max-width: 1023px), (orientation: portrait)').matches);
 if(drawer){
  // Observers count real controls' click handlers, without changing DOM,
  // dispatching synthetic clicks or making unsupported controls actionable.
  await page.locator('.compose-box').evaluate(el=>{window.__workspaceComposerClicks=0;el.addEventListener('click',()=>window.__workspaceComposerClicks++);});
  // Use controls in the exposed backdrop strip; left-hand picker controls
  // lie under the sidebar itself on a phone and are not backdrop targets.
  for(const target of [input,page.getByRole('button',{name:'Send message',exact:true})]){
   const box=await target.boundingBox();expect(box).toBeTruthy();const point={x:box.x+box.width-8,y:box.y+box.height/2};
   expect(await target.evaluate((el,p)=>el.contains(document.elementFromPoint(p.x,p.y)),point)).toBe(true);
   await toggle(true);await expect(page.locator('.workspace-drawer-backdrop')).toBeVisible();
   await expect.poll(()=>page.evaluate(p=>document.elementFromPoint(p.x,p.y)?.classList.contains('workspace-drawer-backdrop'),point)).toBe(true);
   await page.mouse.click(point.x,point.y);await expect(page.locator('.workspace-drawer-backdrop')).toHaveCount(0);await expect(page.locator('.app-shell')).toHaveClass(/workspace-collapsed/);
   expect(await page.evaluate(()=>window.__workspaceComposerClicks)).toBe(0);
   await expect(page.locator('.compose-session-popup,.compose-model-popup')).toHaveCount(0);await expect(input).not.toBeFocused();await unchanged();expect(posts).toBe(0);
  }
 }else{
  await toggle(true);await expect(page.locator('.workspace-drawer-backdrop')).toBeHidden();await expect(page.locator('.workspace-sidebar')).toBeVisible();await toggle(false);await unchanged();
 }
 await page.reload();await unchanged();expect(posts).toBe(0);
 expect(await get(main.id,'messages')).toEqual(before.messages);expect(await get(main.id,'turns')).toEqual(before.turns);expect(await get(research.id,'messages')).toEqual(before.other);expect((await get(research.id,'turns')).turns||[]).toEqual([]);
});
