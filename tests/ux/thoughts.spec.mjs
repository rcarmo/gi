import { test, expect } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { loadCorpus } from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function setup(page,request,info) {
  const token=`preview-${info.project.name}-${Date.now()}`,gate=resolve('test-results/ux-parity/queue-gates',token);mkdirSync(resolve(gate,'..'),{recursive:true});
  const create=async name=>{const r=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`${token}-${name}`}});expect(r.status()).toBe(201);return(await r.json()).id;};
  const a=await create('a'),b=await create('b');await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),a);await page.goto('/');
  const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
  const posted=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${a}/prompt`);
  await input.fill(`UX preview expand UX steer gate:${token}`);await input.press('Enter');const r=await posted;expect(r.status()).toBe(202);const turn=(await r.json()).turn_id;
  await expect(input).toHaveValue('');await input.fill('preview unsent draft');
  const panel=kind=>page.locator('.agent-thinking').filter({has:page.locator('.agent-thinking-title').filter({hasText:kind})});
  for(const kind of ['Thoughts','Draft'])await expect(panel(kind).getByRole('button',{name:'▸ 4 more lines',exact:true})).toBeVisible();
  const more=()=>writeFileSync(gate+'.more','release');
  const finish=async()=>{more();writeFileSync(gate,'release');await expect.poll(async()=>((await(await request.get(`/api/sessions/${a}/turns`)).json()).turns||[]).find(t=>t.id===turn)?.status,{timeout:15000}).toBe('completed');};
  return {a,b,turn,token,gate,input,panel,more,finish};
}
for(const id of ['@gi-preview-002','@ux-thoughts-002','@ux-thoughts-003','@ux-thoughts-004','@ux-thoughts-005']) test(`${id} Native streamed thought and draft disclosures retain text, scrolling and keyboard scope`,async({page,request},info)=>{
  if (id.startsWith('@ux-')) { const source=loadCorpus().find(row=>row.id===id);expect(source).toBeTruthy();await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'}); }
  const f=await setup(page,request,info);
  try {
    const thought=f.panel('Thoughts'),draft=f.panel('Draft');
    for(const panel of [thought,draft]) {
      await expect(panel).toHaveAttribute('data-expanded','false');
      const body=panel.locator('.agent-thinking-body');await expect(body).toHaveCSS('overflow-y','hidden');
      expect(await body.evaluate(el=>parseFloat(getComputedStyle(el).maxHeight))).toBeGreaterThan(0);
      expect(await body.evaluate(el=>el.scrollHeight>el.clientHeight)).toBe(true);
    }
    f.more();
    for(const panel of [thought,draft]) {await expect(panel.getByRole('button',{name:'▸ 8 more lines',exact:true})).toBeVisible();await expect(panel).toHaveAttribute('data-expanded','false');}
    let text=await thought.locator('.agent-thinking-body').textContent();expect(text).toContain('Thought line 16');
    await thought.getByRole('button',{name:'▸ 8 more lines',exact:true}).click();await expect(thought).toHaveAttribute('data-expanded','true');
    await expect(thought.locator('.agent-thinking-body')).toHaveCSS('max-height','none');
    await expect(draft).toHaveAttribute('data-expanded','false');
    if (id === '@ux-thoughts-002') {
      writeFileSync(f.gate+'.expanded','release');await expect(thought.locator('.agent-thinking-body')).toContainText('Thought line 20');
      await expect(thought).toHaveAttribute('data-expanded','true');await expect(draft).toHaveAttribute('data-expanded','false');
      text=await thought.locator('.agent-thinking-body').textContent();
    }
    const omitted=id==='@ux-thoughts-002'?12:8;
    const title=thought.locator('.agent-thinking-title');await title.click();await page.keyboard.press('Escape');await expect(thought).toHaveAttribute('data-expanded','false');
    await thought.getByRole('button',{name:`▸ ${omitted} more lines`,exact:true}).click();
    // Escape in an editable target and modified Escape do not consume disclosure.
    await f.input.focus();await page.keyboard.press('Escape');await expect(thought).toHaveAttribute('data-expanded','true');
    await title.click();await page.keyboard.press('Shift+Escape');await expect(thought).toHaveAttribute('data-expanded','true');
    await draft.getByRole('button',{name:`▸ ${omitted} more lines`,exact:true}).click();await expect(draft).toHaveAttribute('data-expanded','true');
    await draft.locator('.agent-thinking-title').click();await page.keyboard.press('Escape');await expect(draft).toHaveAttribute('data-expanded','false');await expect(thought).toHaveAttribute('data-expanded','true');
    await thought.getByRole('button',{name:'Close Thoughts panel',exact:true}).click();await expect(thought).toHaveAttribute('data-expanded','false');
    await thought.getByRole('button',{name:`▸ ${omitted} more lines`,exact:true}).click();
    await expect(thought.locator('.agent-thinking-body')).toHaveText(text);
    const width=info.project.use.viewport.width;await page.setViewportSize({width:width>720?600:1000,height:700});await expect(thought).toHaveAttribute('data-expanded','true');await expect(thought.locator('.agent-thinking-body')).toHaveText(text);
    await page.locator('.agent-status-panel').evaluate(el=>{el.scrollTop=el.scrollHeight;});expect(await page.locator('.agent-status-panel').evaluate(el=>el.scrollTop)).toBeGreaterThan(0);
    if(process.env.GI_THOUGHTS_CAPTURE&&id==='@ux-thoughts-005'&&info.project.name.startsWith('chromium-')){mkdirSync('test-results/thought-captures',{recursive:true});await page.screenshot({path:`test-results/thought-captures/${info.project.name}.png`});}
    await expect(f.input).toHaveValue('preview unsent draft');
    const composer=await f.input.boundingBox();expect(composer.y+composer.height).toBeLessThanOrEqual(701);
    expect((await(await request.get(`/api/sessions/${f.a}/turns`)).json()).turns).toHaveLength(1);
  } finally {await f.finish();}
});

test('@gi-preview-001 Expansion and buffers reset across session switches and native turns',async({page,request},info)=>{
  const f=await setup(page,request,info);let secondGate;
  const pick=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();const popup=page.locator('.compose-session-popup');await popup.getByRole('searchbox',{name:'Search sessions',exact:true}).fill(id);const row=popup.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem');await expect(row).toBeVisible();const activity=page.waitForResponse(r=>r.request().method()==='GET'&&new URL(r.url()).pathname===`/api/sessions/${id}/activity`);await row.click();await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);await activity;await expect(page.locator('.compose-connection-status')).toHaveCount(0);};
  try {
    await f.panel('Thoughts').getByRole('button',{name:'▸ 4 more lines',exact:true}).click();await expect(f.panel('Thoughts')).toHaveAttribute('data-expanded','true');
    await pick(f.b);await f.input.fill('B preview draft');f.more();
    await page.waitForTimeout(200);await expect(page.locator('.agent-thinking').filter({hasText:'Thought line'})).toHaveCount(0);await expect(f.input).toHaveValue('B preview draft');
    await pick(f.a);await expect(f.input).toHaveValue('preview unsent draft');
    writeFileSync(f.gate+'.expanded','release');await expect(f.panel('Thoughts').locator('.agent-thinking-body')).toContainText('Thought line 20');
    await expect(f.panel('Thoughts')).toHaveAttribute('data-expanded','false');
    await expect(f.panel('Thoughts').locator('.agent-thinking-body')).not.toContainText('Thought line 01');
    // Only newly received bytes return after a session revisit; no cached old
    // previews or expansion. Final native message remains the durable history.
    await f.finish();await expect(page.locator('.agent-thinking').filter({hasText:'Thought line'})).toHaveCount(0);
    const key=f.token+'-next';secondGate=resolve('test-results/ux-parity/queue-gates',key);
    const posted=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${f.a}/prompt`);
    await f.input.fill(`UX preview expand UX steer gate:${key}`);await f.input.press('Enter');expect((await posted).status()).toBe(202);await f.input.fill('next turn unsent');
    await expect(f.panel('Thoughts').getByRole('button',{name:'▸ 4 more lines',exact:true})).toBeVisible();await expect(f.panel('Thoughts')).toHaveAttribute('data-expanded','false');
    await expect(f.panel('Thoughts').locator('.agent-thinking-body')).not.toContainText('Thought line 20');
    await f.panel('Thoughts').getByRole('button',{name:'▸ 4 more lines',exact:true}).click();
    const restored=page.waitForResponse(r=>r.request().method()==='GET'&&new URL(r.url()).pathname===`/api/sessions/${f.a}/activity`);
    await page.reload();await restored;await expect(page.locator('.compose-connection-status')).toHaveCount(0);await expect(f.input).toHaveValue('next turn unsent');await expect(page.locator('.agent-thinking').filter({hasText:'Thought line'})).toHaveCount(0);
    writeFileSync(secondGate+'.more','release');await expect(f.panel('Thoughts').locator('.agent-thinking-body')).toContainText('Thought line 16');await expect(f.panel('Thoughts')).toHaveAttribute('data-expanded','false');
  } finally {
    if(secondGate){writeFileSync(secondGate+'.more','release');writeFileSync(secondGate,'release');await expect.poll(async()=>((await(await request.get(`/api/sessions/${f.a}/turns`)).json()).turns||[]).every(t=>t.status==='completed'),{timeout:15000}).toBe(true);}
    await f.finish();
  }
});
