import { test, expect } from '@playwright/test';
import { loadCorpus } from './support/catalogue.mjs';

const cases = loadCorpus().filter(scenario => ['@ux-original-001', '@ux-original-002'].includes(scenario.id));

for (const scenario of cases) {
  test(`${scenario.id} ${scenario.name}`, async ({ page, request }, testInfo) => {
    await testInfo.attach('gherkin', { body: `${scenario.uri}:${scenario.line}\n${scenario.steps.join('\n')}`, contentType: 'text/plain' });
    // Seed only through Gi's native APIs in the dedicated parity process.
    const create = async (agent) => {
      const response = await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } });
      expect(response.status()).toBe(201);
      return response.json();
    };
    const main = await create('web');
    const research = await create('research');
    expect(main.id).not.toBe(research.id);
    await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
    await page.goto('/');
    const input = page.getByRole('textbox', { name: 'Message (Enter to send, Shift+Enter for newline)...', exact: true });
    await expect(input).toBeVisible();
    const submissions = [];
    page.on('request', req => {
      if (req.method() === 'POST' && /\/prompt$/.test(new URL(req.url()).pathname)) submissions.push(req.postData());
    });
    const trigger = page.getByTestId('hamburger');
    const menu = page.locator('.timeline-menu-dropdown[role="menu"]');
    if (scenario.id === '@ux-original-001') {
      await expect(menu).toHaveCount(0);
      await trigger.click();
      await expect(menu).toBeVisible();
      await expect(menu).toHaveCount(1);
      await page.keyboard.press('Escape');
      await expect(menu).toHaveCount(0);
      await trigger.focus();
      await page.keyboard.press('Enter');
      await expect(menu).toBeVisible();
      // Click visible inert timeline content, never force a hidden control.
      const timeline = page.locator('.timeline');
      const bounds = await timeline.boundingBox();
      expect(bounds).toBeTruthy();
      await timeline.click({ position: { x: bounds.width - 12, y: bounds.height - 12 } });
      await expect(menu).toHaveCount(0);
    } else {
      const draft = 'Canonical unsent draft: preserve on workspace toggle.';
      await input.fill(draft);
      const sidebar = page.locator('.workspace-sidebar');
      const wasVisible = await sidebar.isVisible();
      await trigger.click();
      await menu.getByRole('menuitem', { name: wasVisible ? 'Hide workspace' : 'Show workspace', exact: true }).click();
      await expect(sidebar).toBeVisible({ visible: !wasVisible });
      await expect(menu).toHaveCount(0);
      await expect(input).toHaveValue(draft);
      await trigger.click();
      await menu.getByRole('menuitem', { name: wasVisible ? 'Show workspace' : 'Hide workspace', exact: true }).click();
      await expect(sidebar).toBeVisible({ visible: wasVisible });
      await expect(menu).toHaveCount(0);
      await expect(input).toHaveValue(draft);
    }
    expect(submissions).toEqual([]);
    const stored = await request.get(`/api/sessions/${main.id}/messages`);
    expect(stored.ok()).toBeTruthy();
    // Gi serializes an empty Go slice as null; both forms mean no stored rows.
    expect((await stored.json()).messages ?? []).toEqual([]);
    expect(await page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
    await testInfo.attach('checkpoint', { body: await page.screenshot(), contentType: 'image/png' });
  });
}

for(const [id,opening,dismissal] of [['@shared-1','pointer','outside pointer'],['@shared-2','keyboard','Escape']])test(`${id} Workspace menu ${opening} opening and ${dismissal} dismissal never activate underneath`,async({page,request},info)=>{
 const source=loadCorpus('shared').find(s=>s.id===id);expect(source.name).toBe('Open and dismiss the workspace menu');expect(source.steps.join('\n')).toContain(opening);await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const token=`shell-${opening}-${info.project.name}-${Date.now()}`;const created=await request.post('/api/sessions',{data:{agent_id:token,title:token}});expect(created.status()).toBe(201);const main=await created.json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true}),trigger=page.getByTestId('hamburger'),menu=page.locator('.timeline-menu-dropdown[role=menu]');
 await expect(input).toBeVisible();await input.fill('menu dismissal must not send');await page.locator('.compose-box input[type=file]').setInputFiles({name:'menu-draft.txt',mimeType:'text/plain',buffer:Buffer.from('retained menu bytes')});
 let submissions=0;page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${main.id}/prompt`)submissions++;});
 await page.locator('.compose-box').evaluate(el=>{window.__shellClicks=0;el.addEventListener('click',()=>window.__shellClicks++);});
 const open=async()=>{await expect(menu).toHaveCount(0);if(opening==='pointer')await trigger.click();else{await trigger.focus();await trigger.press('Enter');}await expect(menu).toHaveCount(1);await expect(menu).toBeVisible();await expect(trigger).toHaveAttribute('aria-expanded','true');};
 await open();
 // Real Tab navigation, not direct DOM focus, must reach every enabled item.
 const enabled=await menu.getByRole('menuitem').evaluateAll(nodes=>nodes.filter(n=>!n.disabled).map(n=>n.textContent.trim()));expect(enabled.length).toBeGreaterThan(2);
 await trigger.focus();const seen=new Set();
 for(let n=0;n<enabled.length+8;n++){
  await page.keyboard.press('Tab');
  const label=await page.evaluate(()=>{const e=document.activeElement;return e?.closest('.timeline-menu-dropdown')&&e.getAttribute('role')==='menuitem'?e.textContent.trim():null;});if(label)seen.add(label);
  if(enabled.every(label=>seen.has(label)))break;
 }
 expect([...seen].sort()).toEqual([...enabled].sort());
 const send=page.getByRole('button',{name:'Send message',exact:true});await expect(send).toBeEnabled();
 if(dismissal==='outside pointer'){
  const box=await send.boundingBox();const point={x:box.x+box.width/2,y:box.y+box.height/2};expect(await send.evaluate((el,p)=>el.contains(document.elementFromPoint(p.x,p.y)),point)).toBe(true);
  await page.mouse.click(point.x,point.y);
 }else await page.keyboard.press('Escape');
 await expect(menu).toHaveCount(0);await expect(trigger).toHaveAttribute('aria-expanded','false');
 expect(submissions).toBe(0);expect(await page.evaluate(()=>window.__shellClicks)).toBe(0);await expect(input).toHaveValue('menu dismissal must not send');await expect(page.locator('.compose-file-pill[title="menu-draft.txt"]')).toHaveCount(1);await expect(trigger).toBeFocused();
 // Listener cleanup: a new activation works and a later independent pointer
 // on the composer is not swallowed by the dismissal guard.
 await open();await menu.getByRole('menuitem',{name:'Settings',exact:true}).focus();await page.keyboard.press('Escape');await expect(menu).toHaveCount(0);await expect(trigger).toBeFocused();await input.click();await expect(input).toBeFocused();expect(await page.evaluate(()=>window.__shellClicks)).toBe(1);
 expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).toEqual([]);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);expect(submissions).toBe(0);
});

test.describe('Gi trusted touch menu dismissal',()=>{
 test.use({hasTouch:true});
 test('Gi touch outside the menu dismisses once without leaking a send or consuming the next gesture',async({page,request},info)=>{
  const token=`menu-touch-${info.project.name}-${Date.now()}`,created=await request.post('/api/sessions',{data:{agent_id:token,title:token}});expect(created.status()).toBe(201);const main=await created.json();
  await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
  const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true}),trigger=page.getByTestId('hamburger'),menu=page.locator('.timeline-menu-dropdown'),send=page.getByRole('button',{name:'Send message',exact:true});
  await input.fill('trusted touch stays unsent');let submissions=0;page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.endsWith('/prompt'))submissions++;});
  await page.locator('.compose-box').evaluate(el=>{window.__shellTouches=0;el.addEventListener('click',()=>window.__shellTouches++);});
  const tap=async target=>{const b=await target.boundingBox();await page.touchscreen.tap(b.x+b.width/2,b.y+b.height/2);};
  await tap(trigger);await expect(menu).toHaveCount(1);await tap(send);await expect(menu).toHaveCount(0);await expect(trigger).toBeFocused();expect(submissions).toBe(0);expect(await page.evaluate(()=>window.__shellTouches)).toBe(0);await expect(input).toHaveValue('trusted touch stays unsent');
  await tap(input);await expect(input).toBeFocused();expect(await page.evaluate(()=>window.__shellTouches)).toBe(1);
  // A cancelled pointer sequence must not close/arm a later unrelated action.
  // This is a synthetic cancel-path check, distinct from the trusted tap above.
  await tap(trigger);await expect(menu).toHaveCount(1);await send.dispatchEvent('pointerdown',{pointerId:9,pointerType:'touch',isPrimary:true});await send.dispatchEvent('pointercancel',{pointerId:9,pointerType:'touch',isPrimary:true});await expect(menu).toHaveCount(1);
  await tap(send);await expect(menu).toHaveCount(0);expect(submissions).toBe(0);await tap(input);expect(await page.evaluate(()=>window.__shellTouches)).toBe(2);
  expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).toEqual([]);
 });
});
