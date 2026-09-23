import { test, expect } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { loadCorpus } from './support/catalogue.mjs';

const inputName = 'Message (Enter to send, Shift+Enter for newline)...';
async function swipe(target, points = [['touchstart',190,150],['touchmove',85,150],['touchend',85,150]]) {
  await target.evaluate((el,points) => {
    for (const [name,x,y] of points) {
      const touch = {identifier:1,target:el,clientX:x,clientY:y}, event = new Event(name,{bubbles:true,cancelable:true});
      Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[touch]});Object.defineProperty(event,'changedTouches',{value:[touch]});el.dispatchEvent(event);
    }
  },points);
}

for (const kind of ['draft','thought']) test(`${kind === 'draft' ? '@ux-mobile-003' : '@gi-swipe-003'} Native streaming ${kind} panel links allow swipes but retain selection and direction guards`, async ({page,request},info) => {
  const scenario = loadCorpus().find(row=>row.id==='@ux-mobile-003'); expect(scenario).toBeTruthy(); await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
  const token = `panels-${kind}-${info.project.name}-${Date.now()}`, gate = resolve('test-results/ux-parity/queue-gates',token); mkdirSync(resolve(gate,'..'),{recursive:true});
  const create = async name => {const r=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`${token}-${name}`}});expect(r.status()).toBe(201);return(await r.json()).id;};
  const a=await create('a');await create('b');
  await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'});},a);
  await page.goto('/'); const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
  // Submit through the real composer so native catalogue refresh follows its
  // normal accepted-message lifecycle (not an out-of-band API-only mutation).
  const posted=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${a}/prompt`);
  await input.fill(`UX status panels UX steer gate:${token}`);await input.press('Enter');
  const run=await posted;expect(run.status()).toBe(202);const turn=(await run.json()).turn_id;await expect(input).toHaveValue('');await input.fill('status swipe draft');
  const selected=()=>page.evaluate(()=>localStorage.getItem('gi_session_id'));
  try {
    const panel=page.locator('.agent-status-panel'); await expect(panel.getByText(`Native ${kind}`,{exact:false}).first()).toBeVisible();
    const thinking=panel.locator('.agent-thinking').filter({hasText:`Native ${kind}`});
    const link=thinking.getByRole('link',{name:'details',exact:true});await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href',`https://example.invalid/${kind}`);
    const sessions=(await(await request.get('/api/sessions')).json()).sessions;
    const active=s=>s.state?.status==='running'||s.state?.status==='queued'||Number(s.state?.queue_count||0)>0;
    const ordered=sessions.filter(s=>!s.state?.archived_at).sort((x,y)=>Number(active(y))-Number(active(x))||`gi:${x.id}`.localeCompare(`gi:${y.id}`)).map(s=>s.id);
    expect(ordered[0]).toBe(a);const next=ordered[1];expect(next).toBeTruthy();
    // Genuine Markdown link, rendered from streamed provider content.
    await swipe(link,[['touchstart',190,150],['touchmove',180,195],['touchmove',65,195],['touchend',65,195]]);
    await page.waitForTimeout(200);expect(await selected()).toBe(a);
    await thinking.locator('.agent-thinking-body').evaluate(el=>{const range=document.createRange();range.selectNodeContents(el);const selection=window.getSelection();selection.removeAllRanges();selection.addRange(range);});
    expect(await page.evaluate(()=>window.getSelection()?.toString())).toContain(`Native ${kind}`);await swipe(link);await page.waitForTimeout(200);expect(await selected()).toBe(a);
    await page.evaluate(()=>window.getSelection()?.removeAllRanges());
    // The wider listener boundary must not intercept target handlers/defaults
    // for excluded controls. These listeners observe real delivery only.
    await input.evaluate(el=>{window.__inputTouches=0;window.__inputWheels=0;el.addEventListener('touchstart',()=>window.__inputTouches++);el.addEventListener('wheel',()=>window.__inputWheels++);});
    await swipe(input);await page.waitForTimeout(200);expect(await selected()).toBe(a);
    const wheelDefault=await input.evaluate(el=>el.dispatchEvent(new WheelEvent('wheel',{bubbles:true,cancelable:true,deltaX:110})));
    expect(wheelDefault).toBe(true);expect(await page.evaluate(()=>[window.__inputTouches,window.__inputWheels])).toEqual([1,1]);
    await page.keyboard.press('Control+,');const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});await expect(dialog).toBeVisible();
    await swipe(dialog.getByRole('button',{name:'Close settings',exact:true}));await page.waitForTimeout(150);expect(await selected()).toBe(a);await page.keyboard.press('Escape');
    expect((await(await request.get(`/api/sessions/${a}/turns`)).json()).turns).toHaveLength(1);
    if (process.env.GI_STATUS_CAPTURE && kind === 'draft' && info.project.name.startsWith('chromium-')) {
      mkdirSync('test-results/status-swipe-captures',{recursive:true});
      await page.screenshot({path:`test-results/status-swipe-captures/${info.project.name}.png`});
    }
    // Excluded pen contacts must not poison the helper's finger-only state.
    await input.evaluate(el=>el.dispatchEvent(new PointerEvent('pointerdown',{bubbles:true,pointerType:'pen'})));
    await swipe(link);await expect.poll(selected).toBe(next);await expect(input).toHaveValue('');
    await swipe(page.locator('.timeline').first(),[['touchstart',85,150],['touchmove',190,150],['touchend',190,150]]);await expect.poll(selected).toBe(a);await expect(input).toHaveValue('status swipe draft');
  } finally {
    writeFileSync(gate,'release');await expect.poll(async()=>((await(await request.get(`/api/sessions/${a}/turns`)).json()).turns||[]).find(row=>row.id===turn)?.status,{timeout:15000}).toBe('completed');
  }
});
