import { test, expect } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

const inputName = 'Message (Enter to send, Shift+Enter for newline)...';

test('Gi filtered picker keeps native text editing, Tab and keyboard selection', async ({ page, request }, info) => {
  const rootName = `keyboard-root-${info.project.name}`, leafName = `keyboard-leaf-${info.project.name}`;
  const response = await request.post('/api/sessions', { data: { title: `@${rootName}`, agent_id: rootName } });
  const main = await response.json();
  const fork = await (await request.post(`/api/sessions/${main.id}/fork`, { data: { title: leafName, agent_id: leafName } })).json();
  const child = fork.branch.chat_jid.slice(3);
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  const trigger = page.getByRole('button', { name: /Manage sessions for/ }).last();
  const search = page.getByRole('searchbox', { name: 'Search sessions', exact: true });
  const popup = page.getByRole('listbox', { name: 'Sessions and agents', exact: true });
  await trigger.click();
  await search.pressSequentially(leafName);
  await expect(search).toHaveValue(leafName);
  // Home/End and printable keys edit the search rather than jumping rows.
  await page.keyboard.press('Home');
  await expect.poll(() => search.evaluate(el => el.selectionStart)).toBe(0);
  await page.keyboard.press('End');
  await expect.poll(() => search.evaluate(el => el.selectionStart)).toBe(leafName.length);
  await expect(popup.locator('.active')).toContainText('keyboard-leaf');
  await page.keyboard.press('ArrowUp');
  await expect(popup.locator('.active')).toContainText('keyboard-root');
  await page.keyboard.press('ArrowDown');
  await expect(popup.locator('.active')).toContainText('keyboard-leaf');
  await page.keyboard.press('Enter');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  await expect(search).toHaveCount(0);

  await trigger.click();
  await search.fill(rootName);
  const root = popup.getByRole('option').filter({ hasText: `gi:${main.id}` });
  await expect(popup.getByRole('option')).toHaveCount(1);
  await expect(root).toBeVisible();
  await expect(search).toBeFocused();
  await page.keyboard.press('Tab');
  // The reference row leads with Pin; preserve native sequential focus.
  await expect(popup.getByRole('button',{name:`Pin @${rootName}`,exact:true})).toBeFocused();
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  await page.keyboard.press('Tab');
  await expect(root).toBeFocused();
  // Tab must only move focus, never switch a chat.
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  await page.keyboard.press('Enter');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
});

test('Gi mounted timeline admits whitespace-only selection but blocks selected text on touch swipe',async({page,request})=>{
 const token=`selection-swipe-${Date.now()}`;
 const create=async name=>{const r=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`${token}-${name}`}});expect(r.status()).toBe(201);return(await r.json()).id;};
 const current=await create('current'),other=await create('other');
 const sessions=(await(await request.get('/api/sessions')).json()).sessions;
 const ordered=sessions.filter(s=>!s.state?.archived_at).sort((a,b)=>{
  const active=s=>s.state?.status==='running'||s.state?.status==='queued'||Number(s.state?.queue_count||0)>0;
  return Number(active(b))-Number(active(a))||`gi:${a.id}`.localeCompare(`gi:${b.id}`);
 }).map(s=>s.id);
 expect(ordered).toContain(current);expect(ordered).toContain(other);
 const adjacent=ordered[(ordered.indexOf(current)+1)%ordered.length];
 await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'});},current);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('selection draft');
 const timeline=page.locator('.timeline').first();await expect(timeline).toBeVisible();
 const gesture=async text=>timeline.evaluate((el,value)=>{
  const span=document.createElement('span');span.style.whiteSpace='pre';span.textContent=value==='   '?'A   B':value;el.appendChild(span);
  const range=document.createRange();if(value==='   '){range.setStart(span.firstChild,1);range.setEnd(span.firstChild,4);}else range.selectNodeContents(span);
  const selection=window.getSelection();selection.removeAllRanges();selection.addRange(range);
  const observed=selection.toString();
  for(const [name,x] of [['touchstart',190],['touchmove',85],['touchend',85]]){
   const touch={identifier:1,target:el,clientX:x,clientY:150},event=new Event(name,{bubbles:true,cancelable:true});
   Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[touch]});Object.defineProperty(event,'changedTouches',{value:[touch]});el.dispatchEvent(event);
  }
  return observed;
 },text);
 const selected=()=>page.evaluate(()=>localStorage.getItem('gi_session_id'));
 expect(await gesture('Selected words')).toBe('Selected words');await page.waitForTimeout(250);expect(await selected()).toBe(current);await expect(input).toHaveValue('selection draft');
 expect(await gesture('   ')).toBe('   ');await expect.poll(selected).toBe(adjacent);await expect(input).toHaveValue('');
});

test('Gi delayed mutation failure stays with its originating picker', async ({ page, request }, info) => {
  const agent = `mutation-race-${info.project.name}`;
  const main = await (await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } })).json();
  const fork = await (await request.post(`/api/sessions/${main.id}/fork`, { data: { title: `${agent}-child`, agent_id: `${agent}-child` } })).json();
  const child = fork.branch.chat_jid.slice(3);
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  const trigger = page.getByRole('button', { name: /Manage sessions for/ }).last();
  await trigger.click();
  await page.locator(`[data-session-jid="gi:${child}"]`).getByRole('button', { name: /^Rename / }).click();
  await page.getByRole('textbox', { name: 'Session name', exact: true }).fill('   ');
  let release, delivered;
  const gate = new Promise(resolve => { release = resolve; });
  const delivery = new Promise(resolve => { delivered = resolve; });
  let held = false;
  await page.route(`**/api/sessions/${child}`, async route => {
    if (route.request().method() !== 'PATCH') return route.continue();
    const response = await route.fetch();
    expect(response.status()).toBe(400); // Real validation error, not a fake payload.
    held = true;
    await gate;
    await route.fulfill({ response });
    delivered();
  });
  try {
    await page.getByRole('button', { name: 'Save name', exact: true }).click();
    await expect.poll(() => held).toBe(true);
    await page.keyboard.press('Escape'); // Cancel edit, not the network operation.
    await page.keyboard.press('Escape'); // Dismiss the originating picker.
    const compose = page.getByRole('textbox', { name: inputName, exact: true });
    await compose.fill('Keep my focus and draft');
    release();
    await delivery;
    await page.unroute(`**/api/sessions/${child}`);
    await trigger.click();
    await expect(page.getByRole('alert')).toHaveCount(0);
    await expect(page.getByRole('status').filter({ hasText: 'Renamed session.' })).toHaveCount(0);
    await page.keyboard.press('Escape');
    await expect(compose).toHaveValue('Keep my focus and draft');
    await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
    const after = await (await request.get(`/api/sessions/${child}`)).json();
    expect(after.title).toBe(`${agent}-child`);
  } finally {
    release();
  }
});

test('Gi new-session action allocates a distinct child chat', async ({ page, request }) => {
  const response = await request.post('/api/sessions', { data: { title: '@web', agent_id: 'web' } });
  const main = await response.json();
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
  const created = page.waitForResponse(res => res.url().endsWith(`/api/sessions/${main.id}/fork`) && res.request().method() === 'POST');
  await page.getByRole('button', { name: 'New', exact: true }).click();
  const fork = await (await created).json();
  const child = fork.branch.chat_jid.slice(3);
  expect(child).not.toBe(main.id);
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  const stored = await (await request.get(`/api/sessions/${child}`)).json();
  expect(stored.parent_session_id).toBe(main.id);
});

test('@gi-swipe-001 @gi-swipe-002 Gi rapid reverse swipe uses the committed session while catalogue responses are held', async ({ page, request }, info) => {
  const token = `rapid-${info.project.name}-${Date.now()}`;
  const create = async name => { const r = await request.post('/api/sessions', { data: { agent_id: `${token}-${name}`, title: `${token}-${name}` } }); expect(r.status()).toBe(201); return (await r.json()).id; };
  const a = await create('first'), b = await create('last');
  for (const [id, marker] of [[a, 'A'], [b, 'B']]) {
    expect((await request.post(`/api/sessions/${id}/prompt`, { data: { prompt: `${token} native ${marker}`, model: 'test-model' } })).status()).toBe(202);
    await expect.poll(async () => ((await (await request.get(`/api/sessions/${id}/turns`)).json()).turns || [])[0]?.status).toBe('completed');
  }
  const sessions = (await (await request.get('/api/sessions')).json()).sessions;
  const active = s => s.state?.status === 'running' || s.state?.status === 'queued' || Number(s.state?.queue_count || 0) > 0;
  const order = sessions.filter(s => !s.state?.archived_at).sort((x,y) => Number(active(y))-Number(active(x)) || `gi:${x.id}`.localeCompare(`gi:${y.id}`)).map(s => s.id);
  expect(order[(order.indexOf(a)+1)%order.length]).toBe(b);
  await page.addInitScript(id => { localStorage.setItem('gi_session_id',id); Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'}); },a);
  await page.goto('/'); const input = page.getByRole('textbox',{name:inputName,exact:true}); await expect(input).toBeVisible(); await input.fill('rapid A draft');
  await expect(page.locator('.timeline .post-content').filter({ hasText: `${token} native A` }).first()).toBeVisible();
  await page.locator('.compose-box input[type=file]').setInputFiles({ name: 'rapid.txt', mimeType: 'text/plain', buffer: Buffer.from('rapid native attachment') });
  await expect(page.locator('.compose-file-pill[title="rapid.txt"]')).toBeVisible();
  // Establish initial catalogue using visible native picker, then never open it
  // during the rapid gestures or wait for target HTTP/catalogue completion.
  await page.getByRole('button',{name:/Manage sessions for/}).last().click();
  await expect(page.locator('.compose-session-popup [data-session-jid]')).toHaveCount(sessions.length);
  await page.keyboard.press('Escape');
  let release; const gate = new Promise(r => { release = r; }); let held = 0, heldB = 0, delivered = 0;
  const hold = async route => { const response = await route.fetch(); held++; if (new URL(route.request().url()).pathname === `/api/sessions/${b}/messages`) heldB++; await gate; await route.fulfill({response}); delivered++; };
  await page.route('**/api/sessions',hold);
  await page.route(`**/api/sessions/${b}/messages**`,hold);
  const gestureSequence = async deltas => page.evaluate(async deltas => {
    const swipe = delta => {
      const el = document.querySelector('.timeline');
      for(const [name,x] of [['touchstart',190],['touchmove',190+delta],['touchend',190+delta]]) {
        const touch={identifier:1,target:el,clientX:x,clientY:150},event=new Event(name,{bubbles:true,cancelable:true});
        Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[touch]});Object.defineProperty(event,'changedTouches',{value:[touch]});el.dispatchEvent(event);
      }
    };
    const frames = [];
    for(const delta of deltas) {
      swipe(delta); frames.push(localStorage.getItem('gi_session_id'));
      // One rendered frame: current selection has committed, passive effects
      // may not yet have run. This is a fresh contact, not two swipes in one task.
      await new Promise(requestAnimationFrame);
    }
    return frames;
  },deltas);
  try {
    const transitions = await gestureSequence(Array.from({length:4},()=>[-105,105]).flat());
    expect(transitions).toEqual(Array.from({length:4},()=>[b,a]).flat());
    await expect(input).toHaveValue('rapid A draft');
    expect(await gestureSequence([-105])).toEqual([b]);
    await expect.poll(() => heldB).toBeGreaterThan(0);
    await input.fill('rapid B draft');
    expect(await gestureSequence([105])).toEqual([a]);
    await expect(input).toHaveValue('rapid A draft');
    const count = held; release(); await expect.poll(() => delivered).toBeGreaterThanOrEqual(count);
    await page.unroute('**/api/sessions',hold); await page.unroute(`**/api/sessions/${b}/messages**`,hold);
    await expect(page.locator('.timeline .post-content').filter({ hasText: `${token} native A` }).first()).toBeVisible();
    await expect(page.locator('.timeline .post-content').filter({ hasText: `${token} native B` })).toHaveCount(0);
    expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(a);
    await expect(input).toHaveValue('rapid A draft');
    await expect(page.locator('.compose-file-pill[title="rapid.txt"]')).toBeVisible();
    for (const id of [a,b]) expect((await (await request.get(`/api/sessions/${id}/turns`)).json()).turns || []).toHaveLength(1);
  } finally { release(); }
});

test('@ux-session-006 Dismiss the session picker without choosing an entry', async ({ page, request }, info) => {

  const token = `${info.project.name}-${Date.now()}`;
  const create = async name => {
    const res = await request.post('/api/sessions', { data: { agent_id: name, title: `@${name}` } });
    expect(res.status()).toBe(201);
    return res.json();
  };
  const main = await create(`dismiss-main-${token}`), other = await create(`dismiss-other-${token}`);
  const prompt = await request.post(`/api/sessions/${main.id}/prompt`, { data: { prompt: `retained history ${token}`, model: 'test-model' } });
  expect(prompt.status()).toBe(202);
  const { turn_id } = await prompt.json();
  await expect.poll(async () => ((await (await request.get(`/api/sessions/${main.id}/turns`)).json()).turns || []).find(turn => turn.id === turn_id)?.status).toBe('completed');
  const before = (await (await request.get(`/api/sessions/${main.id}/messages`)).json()).messages;
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true });
  const triggers = page.getByRole('button', { name: /Manage sessions for/ });
  const search = page.getByRole('searchbox', { name: 'Search sessions', exact: true });
  const popup = page.getByRole('listbox', { name: 'Sessions and agents', exact: true });
  await expect(input).toBeVisible(); await input.fill('unsent dismissal draft');
  for (const trigger of [triggers.first(), triggers.last()]) {
    await trigger.click(); await expect(search).toBeFocused();
    await search.pressSequentially(other.id);
    await expect(popup.getByRole('option')).toHaveCount(1);
    await expect(popup.getByRole('option')).toContainText(other.id);
    await page.keyboard.press('Escape');
    await expect(search).toHaveCount(0); await expect(trigger).toBeFocused();
    await expect(input).toHaveValue('unsent dismissal draft');
    expect(await page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
    await trigger.click(); await expect(search).toBeFocused();
    await expect(search).toHaveValue('');
    await expect(popup.getByRole('option').filter({ hasText: `gi:${other.id}` })).toBeVisible();
    // The dismissed search and typeahead must not preselect the old result.
    await expect(popup.getByRole('group', { name: 'Current', exact: true }).getByRole('option')).toContainText(main.id);
    await page.keyboard.press('Escape'); await expect(trigger).toBeFocused();
  }
  expect((await (await request.get(`/api/sessions/${main.id}/messages`)).json()).messages).toEqual(before);
  await expect(page.locator('.post-content').filter({ hasText: `retained history ${token}` }).first()).toBeVisible();
  await expect(input).toHaveValue('unsent dismissal draft');
});

test('@ux-session-002 Group picker entries using current native session metadata', async ({ page, request }, info) => {

  const token=`groups-${info.project.name}-${Date.now()}`;
  const create=async(agent,parent=null)=>{
    const response=parent
      ?await request.post(`/api/sessions/${parent}/fork`,{data:{agent_id:agent,title:`@${agent}`}})
      :await request.post('/api/sessions',{data:{agent_id:agent,title:`@${agent}`}});
    expect(response.status()).toBe(201);
    const body=await response.json();return parent?body.branch.chat_jid.slice(3):body.id;
  };
  const main=await create(`${token}-main`),pinned=await create(`${token}-pinned`),active=await create(`${token}-active`);
  const tree=await create(`${token}-tree`,main),archived=await create(`${token}-archived`,main);
  const other=await create(`${token}-other`);
  const mutation=async(id,body)=>{const response=await request.patch(`/api/sessions/${id}`,{data:body});expect(response.status()).toBe(200);};
  await mutation(pinned,{action:'pin',pinned:true});await mutation(archived,{action:'archive'});
  const gate=resolve(process.env.GI_UX_QUEUE_GATES || `${process.env.GI_TEST_RUN_ROOT}/results/ux-parity/queue-gates`,`${token}-busy`);mkdirSync(resolve(gate,'..'),{recursive:true});
  const run=await request.post(`/api/sessions/${active}/prompt`,{data:{prompt:`UX queue gate:${token}-busy`,model:'test-model'}});
  expect(run.status()).toBe(202);const {turn_id}=await run.json();
  const stored=async id=>(await(await request.get(`/api/sessions/${id}`)).json());
  try {
    await expect.poll(async()=> (await stored(active)).state.status).toBe('running');
    await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main);await page.goto('/');
    const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('grouping draft');
    await page.getByRole('button',{name:/Manage sessions for/}).last().click();
    const popup=page.getByRole('listbox', { name: 'Sessions and agents',exact:true});
    const row=id=>popup.locator(`[data-session-jid="gi:${id}"]`);
    const groups=[['Current',main],['Pinned',pinned],['Active',active],['This session tree',tree],['Other sessions',other],['Archived',archived]];
    for(const [label,id] of groups){
      const group=popup.getByRole('group',{name:label,exact:true});await expect(group).toBeVisible();await expect(group.locator(`[data-session-jid="gi:${id}"]`)).toBeVisible();
      await expect(row(id).getByRole('option')).toContainText(id);
    }
    const order=await popup.getByRole('group').evaluateAll(elements=>elements.map(el=>el.getAttribute('aria-label')));
    expect(order).toEqual(groups.map(([label])=>label));
    await expect(row(main).getByRole('option')).toHaveAttribute('aria-current','true');
    await expect(popup.locator('[role="option"][aria-current="true"]')).toHaveCount(1);
    await expect(row(active).getByRole('button',{name:/^Archive /})).toHaveCount(0);
    await expect(input).toHaveValue('grouping draft');expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main);
  } finally {
    writeFileSync(gate,'release');
    await expect.poll(async()=>((await(await request.get(`/api/sessions/${active}/turns`)).json()).turns||[]).find(turn=>turn.id===turn_id)?.status,{timeout:15000}).toBe('completed');
  }
});

test('@ux-session-005 Touch swipe keeps native carousel order and target/selection exclusions', async ({ page, request }, info) => {

  const token=`swipe-${info.project.name}-${Date.now()}`;
  const create=async(name)=>{
    const response=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`@${token}-${name}`}});
    expect(response.status()).toBe(201);return (await response.json()).id;
  };
  const other=await create('other'),current=await create('current'),next=await create('active');
  const archive=(await request.post(`/api/sessions/${current}/fork`,{data:{agent_id:`${token}-child`,title:'swipe archived child'}}));
  expect(archive.status()).toBe(201);const child=(await archive.json()).branch.chat_jid.slice(3);
  const archivedResult=await request.patch(`/api/sessions/${child}`,{data:{action:'archive'}});expect(archivedResult.status()).toBe(200);
  const read=await request.post(`/api/sessions/${current}/prompt`,{data:{prompt:`Swipe selectable text ${token}`,model:'test-model'}});
  expect(read.status()).toBe(202);
  await expect.poll(async()=>((await(await request.get(`/api/sessions/${current}/messages`)).json()).messages||[]).some(message=>message.role==='assistant')).toBe(true);
  const gate=resolve(process.env.GI_UX_QUEUE_GATES || `${process.env.GI_TEST_RUN_ROOT}/results/ux-parity/queue-gates`,`${token}-busy`);mkdirSync(resolve(gate,'..'),{recursive:true});
  const run=await request.post(`/api/sessions/${next}/prompt`,{data:{prompt:`UX queue gate:${token}-busy`,model:'test-model'}});
  expect(run.status()).toBe(202);const {turn_id}=await run.json();
  const selected=()=>page.evaluate(()=>localStorage.getItem('gi_session_id'));
  const gesture=async(target,delta=-105,type='touch')=>{
    await target.evaluate((el,{delta,type})=>{
      if(type==='wheel'){
        el.dispatchEvent(new WheelEvent('wheel',{bubbles:true,cancelable:true,deltaX:-delta,deltaY:0}));return;
      }
      const point=(x)=>({identifier:1,target:el,clientX:x,clientY:150,pageX:x,pageY:150,screenX:x,screenY:150});
      const dispatch=(name,x)=>{
        const touch=point(x), event=new Event(name,{bubbles:true,cancelable:true});
        Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[touch]});
        Object.defineProperty(event,'changedTouches',{value:[touch]});
        el.dispatchEvent(event);
      };
      dispatch('touchstart',190);dispatch('touchmove',190+delta);dispatch('touchend',190+delta);
    },{delta,type});
  };
  try {
    await expect.poll(async()=> (await(await request.get(`/api/sessions/${next}`)).json()).state.status).toBe('running');
    await page.addInitScript(id=>{
      localStorage.setItem('gi_session_id',id);
      Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'});
    },current);
    await page.goto('/');
    const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('swipe draft');
    const timeline=page.locator('.timeline').first();await expect(timeline).toBeVisible();
    await expect(timeline.locator('.post-content').filter({hasText:`Swipe selectable text ${token}`}).first()).toBeVisible();
    const copy=timeline.getByRole('button',{name:'Copy message',exact:true}).first();await expect(copy).toBeVisible();
    // Use the full persisted catalogue: Playwright projects share one test server.
    const sessions=(await(await request.get('/api/sessions')).json()).sessions;
    const archivedRow=sessions.find(session=>session.id===child);
    expect(archivedRow.state.archived_at).toBeTruthy();
    const candidates=sessions.filter(session=>!session.state?.archived_at).sort((a,b)=>{
      const active=s=>s.state?.status==='running'||s.state?.status==='queued'||Number(s.state?.queue_count||0)>0;
      return Number(active(b))-Number(active(a))||`gi:${a.id}`.localeCompare(`gi:${b.id}`);
    }).map(session=>session.id);
    expect(candidates).not.toContain(child);
    expect(candidates[0]).toBe(next);
    const neighbour=id=>candidates[(candidates.indexOf(id)+1)%candidates.length];
    // Reader selection, interactive targets and archived rows cannot enter the carousel.
    await gesture(copy);await expect.poll(selected).toBe(current);
    await gesture(input);await expect.poll(selected).toBe(current);
    await page.evaluate(()=>{
      const text=document.querySelector('.timeline .post-content');
      if(text){const range=document.createRange();range.selectNodeContents(text);const selection=window.getSelection();selection?.removeAllRanges();selection?.addRange(range);}
    });
    await expect.poll(()=>page.evaluate(()=>window.getSelection()?.toString())).toContain('Swipe selectable text');
    await gesture(timeline);await expect.poll(selected).toBe(current);
    await page.evaluate(()=>window.getSelection()?.removeAllRanges());
    expect(neighbour(current)).toBe(next);
    await gesture(timeline);
    await expect.poll(selected).toBe(next);
    await expect(input).toHaveValue('');
    const following=neighbour(next);
    await gesture(page.locator('.timeline').first());
    await expect.poll(selected).toBe(following);
    expect(following).not.toBe(child);
    await page.getByRole('button',{name:/Manage sessions for/}).last().click();
    await page.locator(`[data-session-jid="gi:${current}"]`).getByRole('option').click();
    await expect.poll(selected).toBe(current);
    await expect(input).toHaveValue('swipe draft');
  } finally {
    writeFileSync(gate,'release');
    await expect.poll(async()=>((await(await request.get(`/api/sessions/${next}/turns`)).json()).turns||[]).find(turn=>turn.id===turn_id)?.status,{timeout:15000}).toBe('completed');
  }
});

test('@ux-mobile-001 Eligible timeline swipe selects adjacent session and wraps at the catalogue end',async({page,request},info)=>{

 const token=`mobile-${info.project.name}-${Date.now()}`;
 const create=async name=>{const response=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`@${token}-${name}`}});expect(response.status()).toBe(201);return (await response.json()).id;};
 const first=await create('first'),last=await create('last');
 const read=await request.post(`/api/sessions/${last}/prompt`,{data:{prompt:`Wrap source ${token}`,model:'test-model'}});expect(read.status()).toBe(202);
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${last}/messages`)).json()).messages||[]).some(message=>message.role==='assistant')).toBe(true);
 // The final assistant message precedes native claim/session cleanup. Wait
 // for authoritative idle before asserting this new ID ends the idle group.
 await expect.poll(async()=>(await(await request.get(`/api/sessions/${last}`)).json()).state.status).toBe('idle');
 const sessions=(await(await request.get('/api/sessions')).json()).sessions;
 const ordered=sessions.filter(s=>!s.state?.archived_at).sort((a,b)=>{
  const active=s=>s.state?.status==='running'||s.state?.status==='queued'||Number(s.state?.queue_count||0)>0;
  return Number(active(b))-Number(active(a))||`gi:${a.id}`.localeCompare(`gi:${b.id}`);
 }).map(s=>s.id);
 expect(ordered).toContain(first);expect(ordered.at(-1)).toBe(last);
 await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'});},last);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('wrap draft');
 const timeline=page.locator('.timeline').first();await expect(timeline.locator('.post-content').filter({hasText:`Wrap source ${token}`}).first()).toBeVisible();
 const swipe=async()=>timeline.evaluate(el=>{
  const point=x=>({identifier:1,target:el,clientX:x,clientY:150});
  for(const [name,x] of [['touchstart',190],['touchmove',85],['touchend',85]]){
   const touch=point(x),event=new Event(name,{bubbles:true,cancelable:true});
   Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[touch]});
   Object.defineProperty(event,'changedTouches',{value:[touch]});el.dispatchEvent(event);
  }
 });
 expect(await page.evaluate(()=>window.getSelection()?.toString()||'')).toBe('');
 await swipe();await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(ordered[0]);
 await expect(input).toHaveValue('');
 await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${last}"]`).getByRole('option').click();
 await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(last);await expect(input).toHaveValue('wrap draft');
});

test('@ux-mobile-005 Primarily vertical movement cancels the current timeline swipe',async({page,request},info)=>{

 const token=`vertical-${info.project.name}-${Date.now()}`;
 const create=async name=>{const response=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`@${token}-${name}`}});expect(response.status()).toBe(201);return (await response.json()).id;};
 await create('other');const current=await create('current');
 const sessions=(await(await request.get('/api/sessions')).json()).sessions;
 const ordered=sessions.filter(s=>!s.state?.archived_at).sort((a,b)=>{
  const active=s=>s.state?.status==='running'||s.state?.status==='queued'||Number(s.state?.queue_count||0)>0;
  return Number(active(b))-Number(active(a))||`gi:${a.id}`.localeCompare(`gi:${b.id}`);
 }).map(s=>s.id);
 expect(ordered).toContain(current);expect(ordered.length).toBeGreaterThan(1);
 const adjacent=ordered[(ordered.indexOf(current)+1)%ordered.length];expect(adjacent).not.toBe(current);
 await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'});},current);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('vertical draft');
 const timeline=page.locator('.timeline').first();await expect(timeline).toBeVisible();
 const gesture=async points=>timeline.evaluate((el,points)=>{
  for(const [name,x,y] of points){
   const touch={identifier:1,target:el,clientX:x,clientY:y},event=new Event(name,{bubbles:true,cancelable:true});
   Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[touch]});
   Object.defineProperty(event,'changedTouches',{value:[touch]});el.dispatchEvent(event);
  }
 },points);
 const selected=()=>page.evaluate(()=>localStorage.getItem('gi_session_id'));
 expect(await page.evaluate(()=>window.getSelection()?.toString()||'')).toBe('');
 // A vertical first move cancels this contact, even if a later move is far enough to be horizontal.
 await gesture([['touchstart',190,150],['touchmove',180,190],['touchmove',60,190],['touchend',60,190]]);
 await page.waitForTimeout(500); // Allow any erroneous async session switch to settle before the negative assertion.
 await expect.poll(selected).toBe(current);await expect(input).toHaveValue('vertical draft');
 // Prove the same native timeline listener still navigates for a fresh eligible contact.
 await gesture([['touchstart',190,150],['touchmove',85,150],['touchend',85,150]]);
 await expect.poll(selected).toBe(adjacent);await expect(input).toHaveValue('');
});

test('@ux-mobile-006 Horizontal wheel only navigates on desktop Safari, never non-Safari or iOS',async({page,request},info)=>{

 const token=`wheel-${info.project.name}-${Date.now()}`;
 const create=async name=>{const response=await request.post('/api/sessions',{data:{agent_id:`${token}-${name}`,title:`@${token}-${name}`}});expect(response.status()).toBe(201);return (await response.json()).id;};
 await create('other');const current=await create('current');
 const response=await request.post(`/api/sessions/${current}/prompt`,{data:{prompt:`Wheel timeline ${token}`,model:'test-model'}});expect(response.status()).toBe(202);
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${current}/messages`)).json()).messages||[]).some(message=>message.role==='assistant')).toBe(true);
 const sessions=(await(await request.get('/api/sessions')).json()).sessions;
 const ordered=sessions.filter(s=>!s.state?.archived_at).sort((a,b)=>{
  const active=s=>s.state?.status==='running'||s.state?.status==='queued'||Number(s.state?.queue_count||0)>0;
  return Number(active(b))-Number(active(a))||`gi:${a.id}`.localeCompare(`gi:${b.id}`);
 }).map(s=>s.id);
 expect(ordered).toContain(current);expect(ordered.length).toBeGreaterThan(1);
 const adjacent=ordered[(ordered.indexOf(current)+1)%ordered.length];expect(adjacent).not.toBe(current);
 await page.addInitScript(id=>{
  localStorage.setItem('gi_session_id',id);
  const mode=localStorage.getItem('wheel_browser_mode')||'chrome';
  const userAgent=mode==='ios'?'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0) AppleWebKit/605.1.15 Safari/604.1':mode==='safari'?'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Version/17.0 Safari/605.1.15':'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0 Safari/537.36';
  Object.defineProperty(navigator,'userAgent',{configurable:true,value:userAgent});
  Object.defineProperty(navigator,'platform',{configurable:true,value:'Win32'});
  Object.defineProperty(navigator,'maxTouchPoints',{configurable:true,value:0});
 },current);
 const selected=()=>page.evaluate(()=>localStorage.getItem('gi_session_id'));
 const wheel=async()=>page.locator('.timeline').first().evaluate(el=>el.dispatchEvent(new WheelEvent('wheel',{bubbles:true,cancelable:true,deltaX:110,deltaY:0})));
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('wheel draft');
 await expect(page.locator('.timeline .post-content').filter({hasText:`Wheel timeline ${token}`}).first()).toBeVisible();
 for(const mode of ['chrome','ios']){
  await page.evaluate(value=>localStorage.setItem('wheel_browser_mode',value),mode);
  await page.reload();await expect(input).toHaveValue('wheel draft');
  await wheel();await page.waitForTimeout(550);await expect.poll(selected).toBe(current);
  await expect(input).toHaveValue('wheel draft');
 }
 // Positive control: the same native listener can navigate with the same wheel delta on desktop Safari.
 await page.evaluate(()=>localStorage.setItem('wheel_browser_mode','safari'));
 await page.reload();await expect(input).toHaveValue('wheel draft');
 await wheel();await expect.poll(selected).toBe(adjacent);await expect(input).toHaveValue('');
});

test('@ux-session-003 Filter session entries using their search metadata', async ({ page, request }, info) => {

  const token = `${info.project.name}-${Date.now()}`;
  const create = async (agent,title) => {
    const result=await request.post('/api/sessions',{data:{agent_id:agent,title}});expect(result.status()).toBe(201);return result.json();
  };
  const main=await create(`filter-main-${token}`,`@filter-main-${token}`);
  const sibling=await create(`filter-sibling-${token}`,`@filter-sibling-${token}`);
  const outsider=await create(`outside-${token}`,`@outside-${token}`);
  const selected=await request.patch(`/api/sessions/${sibling.id}/model`,{data:{model:'test/bootstrap'}});
  expect(selected.status()).toBe(200);
  const storedSibling=await(await request.get(`/api/sessions/${sibling.id}`)).json();
  expect(storedSibling.state.selected_model).toBe('bootstrap');
  await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
  const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('metadata search draft');
  const trigger=page.getByRole('button',{name:/Manage sessions for/}).last();await trigger.click();
  const search=page.getByRole('searchbox',{name:'Search sessions',exact:true});
  const popup=page.getByRole('listbox', { name: 'Sessions and agents',exact:true});
  await expect(search).toBeFocused();
  await search.fill(`GI:${sibling.id}`.toUpperCase());await expect(popup.getByRole('option')).toHaveCount(1);
  await expect(popup.getByRole('option')).toContainText(sibling.id);
  // Session metadata is refreshed by the native 10-second safety poll. Wait
  // for the accepted external model change; do not fabricate picker records.
  await search.fill(`bootstrap ${sibling.id}`);
  await expect(popup.getByRole('option')).toHaveCount(1,{timeout:16000});
  await expect(popup.getByRole('option')).toContainText(sibling.id);
  await search.fill(`filter ${token}`);await expect(popup.getByRole('option')).toHaveCount(2);
  const active=popup.locator('[data-session-entry-key].active');
  await expect(active).toContainText(main.id);
  await search.press('ArrowDown');await expect(active).toContainText(sibling.id);
  await search.press('ArrowDown');await expect(active).toContainText(main.id);
  await search.press('ArrowUp');await expect(active).toContainText(sibling.id);
  expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
  await expect(input).toHaveValue('metadata search draft');
  await search.fill(outsider.id);await expect(popup.getByRole('option')).toHaveCount(1);
  await expect(popup.getByRole('option')).toContainText(outsider.id);
  await expect(active).toContainText(outsider.id);
  await search.press('Enter');await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(outsider.id);
  await expect(input).toHaveValue('');
  await trigger.click();await search.fill(main.id);
  await expect(popup.getByRole('option')).toHaveCount(1);
  await expect(active).toContainText(main.id);
  await search.press('Enter');
  await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
  await expect(input).toHaveValue('metadata search draft');
});

test('@ux-session-001 Show the selected chat timeline and ignore a superseded read', async ({ page, request }, info) => {

  const token = `${info.project.name}-${Date.now()}`;
  const create = async agent => {
    const response = await request.post('/api/sessions', { data: { agent_id: agent, title: `@${agent}` } });
    expect(response.status()).toBe(201); return response.json();
  };
  const main = await create(`timeline-main-${token}`), research = await create(`timeline-research-${token}`);
  const posts = [[main,`Main timeline ${token}`],[research,`Research timeline ${token}`]];
  for (const [session,prompt] of posts) {
    const response = await request.post(`/api/sessions/${session.id}/prompt`, { data: { prompt, model: 'test-model' } });
    expect(response.status()).toBe(202); const { turn_id } = await response.json();
    await expect.poll(async () => ((await (await request.get(`/api/sessions/${session.id}/turns`)).json()).turns || []).find(turn => turn.id === turn_id)?.status).toBe('completed');
  }
  const [mainText,researchText] = posts.map(([,text])=>text);
  await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
  const input=page.getByRole('textbox',{name:inputName,exact:true}); await expect(input).toBeVisible();
  const mainPost=page.locator('.post-content').filter({hasText:mainText}).first();
  const researchPost=page.locator('.post-content').filter({hasText:researchText}).first();
  await expect(mainPost).toBeVisible();await expect(researchPost).toHaveCount(0);await input.fill('main draft');
  let release,held=false,delivered;const gate=new Promise(resolve=>release=resolve),done=new Promise(resolve=>delivered=resolve);
  const pattern=`**/api/sessions/${main.id}/messages?*`;
  await page.route(pattern,async route=>{const response=await route.fetch();if(!held){held=true;await gate;await route.fulfill({response});delivered();}else await route.fulfill({response});});
  const seen=[];page.on('request',req=>seen.push(new URL(req.url()).pathname));
  try {
    await expect.poll(()=>held,{timeout:15000}).toBe(true);
    const picker=page.getByRole('button',{name:/Manage sessions for/}).last();await picker.click();
    await page.locator(`[data-session-jid="gi:${research.id}"]`).getByRole('option').click();
    await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(research.id);
    await expect.poll(()=>seen.includes(`/api/sessions/${research.id}/messages`)).toBe(true);
    await expect(researchPost).toBeVisible();await expect(mainPost).toHaveCount(0);
    await expect(input).toHaveValue('');await input.fill('research draft');
    release();await done;await page.unroute(pattern);
    await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
    await expect(researchPost).toBeVisible();await expect(mainPost).toHaveCount(0);await expect(input).toHaveValue('research draft');
    await picker.click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('option').click();
    await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
    await expect(mainPost).toBeVisible();await expect(researchPost).toHaveCount(0);await expect(input).toHaveValue('main draft');
    const response = await request.get(`/api/sessions/${research.id}/messages`);
    expect((await response.json()).messages.some(message=>message.content===researchText)).toBe(true);
  } finally { release(); }
});

test('@ux-session-004 Archive and restore through the supplied session actions', async ({ page, request }, info) => {

  const token=`archive-${info.project.name}-${Date.now()}`;
  const mainResponse=await request.post('/api/sessions',{data:{agent_id:token,title:`@${token}`}});expect(mainResponse.status()).toBe(201);
  const main=await mainResponse.json();
  const branchResponse=await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:`${token}-child`,title:`@${token}-child`}});expect(branchResponse.status()).toBe(201);
  const child=(await branchResponse.json()).branch.chat_jid.slice(3);
  const stored=async()=>(await(await request.get(`/api/sessions/${child}`)).json());
  await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
  const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('origin archive draft');
  const trigger=page.getByRole('button',{name:/Manage sessions for/}).last();await trigger.click();
  const row=page.locator(`[data-session-jid="gi:${child}"]`);await expect(row).toBeVisible();
  const path=`**/api/sessions/${child}`;
  let release,held=false,delivered;const gate=new Promise(resolve=>release=resolve),done=new Promise(resolve=>delivered=resolve);
  await page.route(path,async route=>{
    if(route.request().method()!=='PATCH')return route.continue();
    const response=await route.fetch();held=true;await gate;await route.fulfill({response});delivered();
  });
  try {
    await row.getByRole('button',{name:/^Archive /}).click();
    const accepted=page.waitForResponse(response=>response.url().endsWith(`/api/sessions/${child}`)&&response.request().method()==='PATCH');
    await page.getByRole('button',{name:'Confirm archive',exact:true}).click();
    await expect.poll(()=>held).toBe(true);
    // Real accepted native mutation is held at the browser delivery boundary;
    // the picker must not present an optimistic Archived group or success.
    await expect(page.getByRole('group',{name:'Archived',exact:true}).locator(`[data-session-jid="gi:${child}"]`)).toHaveCount(0);
    await expect(page.getByRole('status').filter({hasText:'Archived session.'})).toHaveCount(0);
    release();await done;const archivedResponse=await accepted;expect(archivedResponse.status()).toBe(200);
    expect(archivedResponse.request().postDataJSON()).toMatchObject({action:'archive'});await page.unroute(path);
    await expect(page.getByRole('group',{name:'Archived',exact:true}).locator(`[data-session-jid="gi:${child}"]`)).toBeVisible();
    await expect.poll(async()=>Boolean((await stored()).state.archived_at)).toBe(true);
    await expect(input).toHaveValue('origin archive draft');
    await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
    const restored=page.waitForResponse(response=>response.url().endsWith(`/api/sessions/${child}`)&&response.request().method()==='PATCH');
    await row.getByRole('button',{name:/^Restore /}).click();const restoredResponse=await restored;
    expect(restoredResponse.status()).toBe(200);expect(restoredResponse.request().postDataJSON()).toMatchObject({action:'restore'});
    await expect(page.getByRole('group',{name:'This session tree',exact:true}).locator(`[data-session-jid="gi:${child}"]`)).toBeVisible();
    await expect.poll(async()=> (await stored()).state.archived_at||null).toBe(null);
    await expect(input).toHaveValue('origin archive draft');
    // A transport failure rejects the captured native callback without
    // inventing a successful response or changing the authoritative row.
    await row.getByRole('button',{name:/^Archive /}).click();
    let rejected=false;
    await page.route(path,async route=>{if(route.request().method()!=='PATCH')return route.continue();
      expect(route.request().postDataJSON()).toMatchObject({action:'archive'});
      rejected=true;await route.abort('failed');});
    await page.getByRole('button',{name:'Confirm archive',exact:true}).click();
    await expect.poll(()=>rejected).toBe(true);
    await expect(page.locator('.compose-session-mutation-error[role="alert"]')).toBeVisible();
    await expect(page.locator('.compose-session-mutation-error')).toContainText(/Load failed|Failed to fetch/);
    await expect.poll(async()=> (await stored()).state.archived_at||null).toBe(null);
    await expect(row).toBeVisible();await expect(page.getByRole('group',{name:'Archived',exact:true}).locator(`[data-session-jid="gi:${child}"]`)).toHaveCount(0);
    await expect(input).toHaveValue('origin archive draft');
    expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
    await page.unroute(path);
  } finally {release();}
});

test('@ux-mobile-004 Native pinned and active overlap yields one stable active-first carousel', async ({ page, request }, info) => {
  test.setTimeout(60000);

  const token = `order-${info.project.name}-${Date.now()}`;
  const create = async name => {
    const response = await request.post('/api/sessions', { data: { agent_id: `${token}-${name}`, title: `@${token}-${name}` } });
    expect(response.status()).toBe(201); return (await response.json()).id;
  };
  const ordinary = await create('ordinary'), pinnedIdle = await create('pinned'), activeA = await create('active-a'), activeB = await create('active-b');
  for (const id of [pinnedIdle, activeB]) expect((await request.patch(`/api/sessions/${id}`, { data: { action: 'pin', pinned: true } })).status()).toBe(200);
  const fork = await request.post(`/api/sessions/${ordinary}/fork`, { data: { agent_id: `${token}-archived`, title: `@${token}-archived` } });
  expect(fork.status()).toBe(201); const archived = (await fork.json()).branch.chat_jid.slice(3);
  expect((await request.patch(`/api/sessions/${archived}`, { data: { action: 'archive' } })).status()).toBe(200);
  const runs = [];
  const selected = () => page.evaluate(() => localStorage.getItem('gi_session_id'));
  const swipe = async delta => page.locator('.timeline').first().evaluate((el, delta) => {
    for (const [name, x] of [['touchstart', 190], ['touchmove', 190 + delta], ['touchend', 190 + delta]]) {
      const touch = { identifier: 1, target: el, clientX: x, clientY: 150 }, event = new Event(name, { bubbles: true, cancelable: true });
      Object.defineProperty(event, 'touches', { value: name === 'touchend' ? [] : [touch] });
      Object.defineProperty(event, 'changedTouches', { value: [touch] }); el.dispatchEvent(event);
    }
  }, delta);
  try {
    for (const [index, id] of [activeA, activeB].entries()) {
      const key = `${token}-gate-${index}`, gate = resolve(process.env.GI_UX_QUEUE_GATES || `${process.env.GI_TEST_RUN_ROOT}/results/ux-parity/queue-gates`, key); mkdirSync(resolve(gate, '..'), { recursive: true });
      const response = await request.post(`/api/sessions/${id}/prompt`, { data: { prompt: `UX queue gate:${key}`, model: 'test-model' } });
      expect(response.status()).toBe(202); runs.push({ id, gate, turn: (await response.json()).turn_id });
      await expect.poll(async () => (await (await request.get(`/api/sessions/${id}`)).json()).state.status).toBe('running');
    }
    const sessions = (await (await request.get('/api/sessions')).json()).sessions;
    expect(sessions.find(s => s.id === activeB).state.pinned).toBe(true);
    expect(sessions.find(s => s.id === pinnedIdle).state.pinned).toBe(true);
    expect(sessions.find(s => s.id === archived).state.archived_at).toBeTruthy();
    const active = s => s.state?.status === 'running' || s.state?.status === 'queued' || Number(s.state?.queue_count || 0) > 0;
    const ordered = sessions.filter(s => !s.state?.archived_at).sort((a,b) => Number(active(b)) - Number(active(a)) || `gi:${a.id}`.localeCompare(`gi:${b.id}`)).map(s => s.id);
    expect(ordered.slice(0,2)).toEqual([activeA, activeB].sort((a,b) => `gi:${a}`.localeCompare(`gi:${b}`)));
    expect(new Set(ordered).size).toBe(ordered.length); expect(ordered).not.toContain(archived);
    await page.addInitScript(id => { localStorage.setItem('gi_session_id', id); Object.defineProperty(navigator, 'userAgent', { configurable: true, value: 'iPhone Safari' }); }, activeB);
    await page.goto('/'); const input = page.getByRole('textbox', { name: inputName, exact: true }); await expect(input).toBeVisible(); await input.fill('pinned active draft');
    const ready = async id => {
      // localStorage changes before session activation/catalogue refresh. Use
      // the actual populated picker as readiness, not an arbitrary sleep.
      await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
      const popup = page.locator('.compose-session-popup');
      await popup.getByRole('searchbox', { name: 'Search sessions', exact: true }).fill('');
      await expect(popup.locator('[data-session-jid]')).toHaveCount(sessions.length);
      await expect(popup.locator(`[data-session-jid="gi:${id}"] [role="option"]`)).toHaveAttribute('aria-current', 'true');
      await page.keyboard.press('Escape'); await expect(popup).toHaveCount(0);
    };
    const pick = async id => {
      await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
      const popup = page.locator('.compose-session-popup'); await popup.getByRole('searchbox', { name: 'Search sessions', exact: true }).fill(id);
      const row = popup.locator(`[data-session-jid="gi:${id}"]`).getByRole('option'); await expect(row).toBeVisible(); await row.click(); await expect.poll(selected).toBe(id); await ready(id);
    };
    // Overlap is real persisted state: activeB belongs to pinned, active and
    // ordinary catalogue membership. Both directions must visit it once, never
    // stop at a duplicate; current selection must not reorder the carousel.
    await ready(activeB);
    for (const id of [activeA, activeB, ordinary, pinnedIdle, ordered.at(-1)]) {
      if (await selected() !== id) await pick(id);
      const next = ordered[(ordered.indexOf(id) + 1) % ordered.length];
      await swipe(-105); await expect.poll(selected).toBe(next); await ready(next);
      await swipe(105); await expect.poll(selected).toBe(id); await ready(id);
    }
    await pick(activeB); await expect(input).toHaveValue('pinned active draft');
    // Pin changes affect picker grouping, not active/JID carousel precedence.
    expect((await request.patch(`/api/sessions/${activeB}`, { data: { action: 'pin', pinned: false } })).status()).toBe(200);
    await page.reload(); await expect(input).toHaveValue('pinned active draft'); await ready(activeB);
    await swipe(-105); await expect.poll(selected).toBe(ordered[(ordered.indexOf(activeB) + 1) % ordered.length]);
    expect(await selected()).not.toBe(archived);
  } finally {
    for (const run of runs) writeFileSync(run.gate, 'release');
    for (const run of runs) await expect.poll(async () => ((await (await request.get(`/api/sessions/${run.id}/turns`)).json()).turns || []).find(t => t.id === run.turn)?.status, { timeout: 15000 }).toBe('completed');
  }
});
