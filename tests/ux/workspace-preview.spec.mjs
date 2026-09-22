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
