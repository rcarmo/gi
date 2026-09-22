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
