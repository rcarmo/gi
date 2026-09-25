import { test, expect } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { loadCorpus } from './support/catalogue.mjs';

const inputName = 'Message (Enter to send, Shift+Enter for newline)...';
const scenario = loadCorpus().find(row => row.id === '@ux-original-014');

test('@ux-original-013 Open the session picker from pointer or keyboard and close it with Escape', async ({ page, request }, info) => {
  const scenario = loadCorpus().find(row => row.id === '@ux-original-013');
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
  const create = async agent => {
    const response = await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } });
    expect(response.status()).toBe(201);
    return response.json();
  };
  const suffix = info.project.name;
  const main = await create(`picker-main-${suffix}`), other = await create(`picker-other-${suffix}`);
  const fork = async (parent, name) => {
    const response = await request.post(`/api/sessions/${parent}/fork`, { data: { title: name, agent_id: name } });
    expect(response.status()).toBe(201);
    return (await response.json()).branch.chat_jid.slice(3);
  };
  const child = await fork(main.id, `picker-child-${suffix}`);
  await fork(child, `picker-leaf-${suffix}`);
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), child);
  await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true });
  const triggers = page.getByRole('button', { name: /Manage sessions for/ });
  const popup = page.getByRole('menu', { name: 'Sessions and agents', exact: true });
  const search = page.getByRole('searchbox', { name: 'Search sessions', exact: true });
  await expect(triggers.last()).toBeVisible();
  await input.fill('Do not submit this draft');

  // Both visible pointer triggers must restore the specific opening button.
  for (const trigger of [triggers.first(), triggers.last()]) {
    await trigger.click();
    await expect(search).toBeFocused();
    await expect(popup.getByRole('group', { name: 'Current', exact: true }).getByRole('menuitem')).toContainText('picker-child');
    const tree = popup.getByRole('group', { name: 'This session tree', exact: true });
    await expect(tree.getByRole('menuitem').filter({ hasText: 'picker-main' })).toBeVisible();
    await expect(tree.getByRole('menuitem').filter({ hasText: 'picker-leaf' })).toBeVisible();
    await expect(popup.getByRole('group', { name: 'Other sessions', exact: true }).getByRole('menuitem').filter({ hasText: `gi:${other.id}` })).toBeVisible();
    await search.fill(`GI:${other.id}`); // Case-insensitive JID search.
    await expect(popup.getByRole('menuitem')).toHaveCount(1);
    await expect(popup.getByRole('menuitem')).toContainText('picker-other');
    await search.fill(`picker-leaf-${suffix}`); // Matching descendants keep their ancestors.
    await expect(tree.getByRole('menuitem').filter({ hasText: 'picker-leaf' })).toBeVisible();
    await expect(popup.getByRole('menuitem')).toHaveCount(3);
    await search.fill('no-session-matches-this-query');
    await expect(popup.getByRole('status')).toHaveText('No sessions match your search.');
    await page.keyboard.press('Enter');
    await expect(search).toBeFocused();
    await page.keyboard.press('Escape');
    await expect(search).toHaveCount(0);
    await expect(trigger).toBeFocused();
    await expect(input).toHaveValue('Do not submit this draft');
    await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  }

  // Native keyboard activation, not a programmatic click.
  await page.keyboard.press('Enter');
  await expect(search).toBeFocused();
  await expect(search).toHaveValue('');
  await page.keyboard.press('Escape');
  await expect(triggers.last()).toBeFocused();
  await input.fill('');
  await page.keyboard.press('@');
  await expect(search).toBeFocused();
  await expect(input).toHaveValue('');
  await page.keyboard.press('Escape');
  await expect(triggers.first()).toBeFocused();
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  const messages = await (await request.get(`/api/sessions/${child}/messages`)).json();
  expect(messages.messages || []).toEqual([]);
  await info.attach('checkpoint', { body: await page.screenshot(), contentType: 'image/png' });
});

test('@ux-session-006 Dismiss the session picker without choosing an entry', async ({ page, request }, info) => {
  const frozen = loadCorpus().find(row => row.id === '@ux-session-006');
  expect(frozen).toBeTruthy();
  await info.attach('gherkin', { body: frozen.steps.join('\n'), contentType: 'text/plain' });
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
  const popup = page.getByRole('menu', { name: 'Sessions and agents', exact: true });
  await expect(input).toBeVisible(); await input.fill('unsent dismissal draft');
  for (const trigger of [triggers.first(), triggers.last()]) {
    await trigger.click(); await expect(search).toBeFocused();
    await search.pressSequentially(other.id);
    await expect(popup.getByRole('menuitem')).toHaveCount(1);
    await expect(popup.getByRole('menuitem')).toContainText(other.id);
    await page.keyboard.press('Escape');
    await expect(search).toHaveCount(0); await expect(trigger).toBeFocused();
    await expect(input).toHaveValue('unsent dismissal draft');
    expect(await page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
    await trigger.click(); await expect(search).toBeFocused();
    await expect(search).toHaveValue('');
    await expect(popup.getByRole('menuitem').filter({ hasText: `gi:${other.id}` })).toBeVisible();
    // The dismissed search and typeahead must not preselect the old result.
    await expect(popup.getByRole('group', { name: 'Current', exact: true }).getByRole('menuitem')).toContainText(main.id);
    await page.keyboard.press('Escape'); await expect(trigger).toBeFocused();
  }
  expect((await (await request.get(`/api/sessions/${main.id}/messages`)).json()).messages).toEqual(before);
  await expect(page.locator('.post-content').filter({ hasText: `retained history ${token}` }).first()).toBeVisible();
  await expect(input).toHaveValue('unsent dismissal draft');
});

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
  const popup = page.getByRole('menu', { name: 'Sessions and agents', exact: true });
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
  const root = popup.getByRole('menuitem').filter({ hasText: `gi:${main.id}` });
  await expect(popup.getByRole('menuitem')).toHaveCount(1);
  await expect(root).toBeVisible();
  await expect(search).toBeFocused();
  await page.keyboard.press('Tab');
  await expect(root).toBeFocused();
  // Tab must only move focus, never switch a chat.
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  await page.keyboard.press('Enter');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
});

test('@ux-session-002 Group picker entries using current native session metadata', async ({ page, request }, info) => {
  const frozen=loadCorpus().find(row=>row.id==='@ux-session-002');expect(frozen).toBeTruthy();
  await info.attach('gherkin',{body:frozen.steps.join('\n'),contentType:'text/plain'});
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
  const gate=resolve('test-results/ux-parity/queue-gates',`${token}-busy`);mkdirSync(resolve(gate,'..'),{recursive:true});
  const run=await request.post(`/api/sessions/${active}/prompt`,{data:{prompt:`UX queue gate:${token}-busy`,model:'test-model'}});
  expect(run.status()).toBe(202);const {turn_id}=await run.json();
  const stored=async id=>(await(await request.get(`/api/sessions/${id}`)).json());
  try {
    await expect.poll(async()=> (await stored(active)).state.status).toBe('running');
    await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main);await page.goto('/');
    const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('grouping draft');
    await page.getByRole('button',{name:/Manage sessions for/}).last().click();
    const popup=page.getByRole('menu',{name:'Sessions and agents',exact:true});
    const row=id=>popup.locator(`[data-session-jid="gi:${id}"]`);
    const groups=[['Current',main],['Pinned',pinned],['Active',active],['This session tree',tree],['Other sessions',other],['Archived',archived]];
    for(const [label,id] of groups){
      const group=popup.getByRole('group',{name:label,exact:true});await expect(group).toBeVisible();await expect(group.locator(`[data-session-jid="gi:${id}"]`)).toBeVisible();
      await expect(row(id).getByRole('menuitem')).toContainText(id);
    }
    const order=await popup.getByRole('group').evaluateAll(elements=>elements.map(el=>el.getAttribute('aria-label')));
    expect(order).toEqual(groups.map(([label])=>label));
    await expect(row(main).getByRole('menuitem')).toHaveAttribute('aria-current','true');
    await expect(popup.locator('[role="menuitem"][aria-current="true"]')).toHaveCount(1);
    await expect(row(active).getByRole('button',{name:/^Archive /})).toHaveCount(0);
    await expect(input).toHaveValue('grouping draft');expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main);
  } finally {
    writeFileSync(gate,'release');
    await expect.poll(async()=>((await(await request.get(`/api/sessions/${active}/turns`)).json()).turns||[]).find(turn=>turn.id===turn_id)?.status,{timeout:15000}).toBe('completed');
  }
});

test('@ux-session-005 Touch swipe keeps native carousel order and target/selection exclusions', async ({ page, request }, info) => {
  const frozen=loadCorpus().find(row=>row.id==='@ux-session-005');expect(frozen).toBeTruthy();
  await info.attach('gherkin',{body:frozen.steps.join('\n'),contentType:'text/plain'});
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
  const gate=resolve('test-results/ux-parity/queue-gates',`${token}-busy`);mkdirSync(resolve(gate,'..'),{recursive:true});
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
    await page.locator(`[data-session-jid="gi:${current}"]`).getByRole('menuitem').click();
    await expect.poll(selected).toBe(current);
    await expect(input).toHaveValue('swipe draft');
  } finally {
    writeFileSync(gate,'release');
    await expect.poll(async()=>((await(await request.get(`/api/sessions/${next}/turns`)).json()).turns||[]).find(turn=>turn.id===turn_id)?.status,{timeout:15000}).toBe('completed');
  }
});

test('@ux-mobile-001 Eligible timeline swipe selects adjacent session and wraps at the catalogue end',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-mobile-001');expect(scenario).toBeTruthy();
 await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
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
 await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${last}"]`).getByRole('menuitem').click();
 await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(last);await expect(input).toHaveValue('wrap draft');
});

test('@ux-mobile-005 Primarily vertical movement cancels the current timeline swipe',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-mobile-005');expect(scenario).toBeTruthy();
 await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
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
 const scenario=loadCorpus().find(row=>row.id==='@ux-mobile-006');expect(scenario).toBeTruthy();
 await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
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
  const frozen = loadCorpus().find(row => row.id === '@ux-session-003'); expect(frozen).toBeTruthy();
  await info.attach('gherkin', { body: frozen.steps.join('\n'), contentType: 'text/plain' });
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
  const popup=page.getByRole('menu',{name:'Sessions and agents',exact:true});
  await expect(search).toBeFocused();
  await search.fill(`GI:${sibling.id}`.toUpperCase());await expect(popup.getByRole('menuitem')).toHaveCount(1);
  await expect(popup.getByRole('menuitem')).toContainText(sibling.id);
  // Session metadata is refreshed by the native 10-second safety poll. Wait
  // for the accepted external model change; do not fabricate picker records.
  await search.fill(`bootstrap ${sibling.id}`);
  await expect(popup.getByRole('menuitem')).toHaveCount(1,{timeout:16000});
  await expect(popup.getByRole('menuitem')).toContainText(sibling.id);
  await search.fill(`filter ${token}`);await expect(popup.getByRole('menuitem')).toHaveCount(2);
  const active=popup.locator('[data-session-entry-key].active');
  await expect(active).toContainText(main.id);
  await search.press('ArrowDown');await expect(active).toContainText(sibling.id);
  await search.press('ArrowDown');await expect(active).toContainText(main.id);
  await search.press('ArrowUp');await expect(active).toContainText(sibling.id);
  expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
  await expect(input).toHaveValue('metadata search draft');
  await search.fill(outsider.id);await expect(popup.getByRole('menuitem')).toHaveCount(1);
  await expect(popup.getByRole('menuitem')).toContainText(outsider.id);
  await expect(active).toContainText(outsider.id);
  await search.press('Enter');await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(outsider.id);
  await expect(input).toHaveValue('');
  await trigger.click();await search.fill(main.id);
  await expect(popup.getByRole('menuitem')).toHaveCount(1);
  await expect(active).toContainText(main.id);
  await search.press('Enter');
  await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
  await expect(input).toHaveValue('metadata search draft');
});

test('@ux-original-014 Select another session through the picker', async ({ page, request }, info) => {
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
  const create = async agent => {
    const response = await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } });
    expect(response.status()).toBe(201);
    return response.json();
  };
  const main = await create('web'), research = await create('research');
  const mainText = `Main history ${main.id}`, researchText = `Research history ${research.id}`;
  for (const [session, text] of [[main, mainText], [research, researchText]]) {
    const response = await request.post(`/api/sessions/${session.id}/prompt`, { data: { prompt: text, model: 'test-model' } });
    expect(response.status()).toBe(202);
    await expect.poll(async () => {
      const stored = await (await request.get(`/api/sessions/${session.id}/messages`)).json();
      return (stored.messages || []).some(message => message.role === 'assistant' && message.content === `Gi received: ${text}`);
    }).toBe(true);
  }
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true });
  const trigger = page.getByRole('button', { name: /Manage sessions for/ }).last();
  const popup = page.getByRole('menu', { name: 'Sessions and agents', exact: true });
  await expect(input).toBeVisible();
  await expect(trigger).toBeVisible();
  await input.fill('Main unsent draft');

  // Hold a real main-session response, not a fabricated payload. The action to
  // switch sessions still goes through the visible native picker.
  let release;
  const gate = new Promise(resolve => { release = resolve; });
  let held = false;
  let delivered;
  const deliveredResponse = new Promise(resolve => { delivered = resolve; });
  await page.route(`**/api/sessions/${main.id}/messages*`, async route => {
    const response = await route.fetch();
    held = true;
    await gate;
    await route.fulfill({ response });
    delivered();
  });
  // Gi's native refresh interval will start the captured request.
  await expect.poll(() => held, { timeout: 15000 }).toBe(true);
  const paths = [];
  page.on('request', req => paths.push(new URL(req.url()).pathname));
  await trigger.click();
  const target = popup.locator(`[data-session-jid="gi:${research.id}"]`).getByRole('menuitem');
  await expect(target).toBeVisible();
  await target.focus();
  await page.keyboard.press('Enter');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(research.id);
  await expect(input).toHaveValue('');
  await input.fill('Research unsent draft');
  await expect.poll(() => paths.includes(`/api/sessions/${research.id}/messages`)).toBe(true);
  await expect.poll(() => paths.includes(`/api/sessions/${research.id}/queue`)).toBe(true);
  await expect(page.locator('.post-content').filter({ hasText: researchText }).first()).toBeVisible();
  release();
  await deliveredResponse;
  await page.unroute(`**/api/sessions/${main.id}/messages*`);
  await expect(input).toHaveValue('Research unsent draft');
  await expect(trigger).toHaveAttribute('aria-label', 'Manage sessions for @research');
  await expect(page.locator('.post-content').filter({ hasText: mainText })).toHaveCount(0);
  await expect(page.locator('.post-content').filter({ hasText: researchText }).first()).toBeVisible();
  await trigger.click();
  await popup.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();
  await expect(input).toHaveValue('Main unsent draft');
  await trigger.click();
  await popup.locator(`[data-session-jid="gi:${research.id}"]`).getByRole('menuitem').click();
  await expect(input).toHaveValue('Research unsent draft');
  const messages = await request.get(`/api/sessions/${research.id}/messages`);
  expect((await messages.json()).messages.map(message => message.content)).toEqual([researchText, `Gi received: ${researchText}`]);
  await info.attach('checkpoint', { body: await page.screenshot(), contentType: 'image/png' });
});

test('@ux-session-001 Show the selected chat timeline and ignore a superseded read', async ({ page, request }, info) => {
  const frozen = loadCorpus().find(row => row.id === '@ux-session-001'); expect(frozen).toBeTruthy();
  await info.attach('gherkin', { body: frozen.steps.join('\n'), contentType: 'text/plain' });
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
    await page.locator(`[data-session-jid="gi:${research.id}"]`).getByRole('menuitem').click();
    await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(research.id);
    await expect.poll(()=>seen.includes(`/api/sessions/${research.id}/messages`)).toBe(true);
    await expect(researchPost).toBeVisible();await expect(mainPost).toHaveCount(0);
    await expect(input).toHaveValue('');await input.fill('research draft');
    release();await done;await page.unroute(pattern);
    await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
    await expect(researchPost).toBeVisible();await expect(mainPost).toHaveCount(0);await expect(input).toHaveValue('research draft');
    await picker.click();await page.locator(`[data-session-jid="gi:${main.id}"]`).getByRole('menuitem').click();
    await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);
    await expect(mainPost).toBeVisible();await expect(researchPost).toHaveCount(0);await expect(input).toHaveValue('main draft');
    const response = await request.get(`/api/sessions/${research.id}/messages`);
    expect((await response.json()).messages.some(message=>message.content===researchText)).toBe(true);
  } finally { release(); }
});

test('@ux-original-015 Use the session actions actually supplied by the client', async ({ page, request }, info) => {
  const scenario = loadCorpus().find(row => row.id === '@ux-original-015');
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
  const agent = `mutations-${info.project.name}`;
  const main = await (await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } })).json();
  const fork = await (await request.post(`/api/sessions/${main.id}/fork`, { data: { title: `${agent}-child`, agent_id: `${agent}-child` } })).json();
  const child = fork.branch.chat_jid.slice(3);
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  const compose = page.getByRole('textbox', { name: inputName, exact: true });
  await compose.fill('Mutation must not submit or lose this draft');
  const trigger = page.getByRole('button', { name: /Manage sessions for/ }).last();
  await trigger.click();
  const row = id => page.locator(`[data-session-jid="gi:${id}"]`);
  const childRow = row(child), rootRow = row(main.id);
  await expect(childRow).toBeVisible();
  await expect(rootRow.getByRole('button', { name: /^Archive / })).toHaveCount(0);
  await expect(childRow.getByRole('button', { name: /^Restore / })).toHaveCount(0);
  await expect(page.getByRole('button', { name: /Delete current/ })).toHaveCount(0);
  const stored = async () => (await request.get(`/api/sessions/${child}`)).json();

  await childRow.getByRole('button', { name: /^Pin / }).click();
  await expect(page.getByRole('group', { name: 'Pinned', exact: true }).locator(`[data-session-jid="gi:${child}"]`)).toBeVisible();
  await expect.poll(async () => (await stored()).state.pinned).toBe(true);
  await childRow.getByRole('button', { name: /^Rename / }).click();
  const name = page.getByRole('textbox', { name: 'Session name', exact: true });
  await expect(name).toBeFocused();
  await name.fill('   ');
  await page.getByRole('button', { name: 'Save name', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('title must be');
  await expect.poll(async () => (await stored()).title).toBe(`${agent}-child`);
  await expect(page.getByRole('status').filter({ hasText: 'Renamed session.' })).toHaveCount(0);
  const renamed = `${agent}-renamed`;
  await name.fill(renamed);
  await page.getByRole('button', { name: 'Save name', exact: true }).click();
  await expect(childRow.getByRole('menuitem')).toContainText(renamed);
  await expect.poll(async () => (await stored()).title).toBe(renamed);
  await childRow.getByRole('button', { name: /^Archive / }).click();
  await page.getByRole('button', { name: 'Cancel', exact: true }).click();
  await expect.poll(async () => (await stored()).state.archived_at || null).toBe(null);
  await childRow.getByRole('button', { name: /^Archive / }).click();
  await page.getByRole('button', { name: 'Confirm archive', exact: true }).click();
  await expect(page.getByRole('group', { name: 'Archived', exact: true }).locator(`[data-session-jid="gi:${child}"]`)).toBeVisible();
  await expect(childRow.getByRole('button', { name: /^Rename / })).toHaveCount(0);
  await expect(childRow.getByRole('button', { name: /^(Pin|Unpin) / })).toHaveCount(0);
  await expect.poll(async () => Boolean((await stored()).state.archived_at)).toBe(true);
  await page.keyboard.press('Escape');
  await expect(compose).toHaveValue('Mutation must not submit or lose this draft');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);

  await page.reload(); // Persistent pin/archive metadata, not page-local state.
  await trigger.click();
  await expect(page.getByRole('group', { name: 'Archived', exact: true }).locator(`[data-session-jid="gi:${child}"]`)).toBeVisible();
  await childRow.getByRole('button', { name: /^Restore / }).click();
  await expect(page.getByRole('group', { name: 'Pinned', exact: true }).locator(`[data-session-jid="gi:${child}"]`)).toBeVisible();
  await expect.poll(async () => (await stored()).state.archived_at || null).toBe(null);
  await childRow.getByRole('button', { name: /^Unpin / }).click();
  await expect.poll(async () => (await stored()).state.pinned).toBe(false);
  await expect(page.getByRole('group', { name: 'This session tree', exact: true }).locator(`[data-session-jid="gi:${child}"]`)).toBeVisible();
  await page.keyboard.press('Escape');
  // A display-name change must not rewrite the native @agent routing handle.
  await compose.fill(`@${agent}-chi`);
  const mention = page.locator('.slash-item').filter({ hasText: `gi:${child}` });
  await expect(mention).toBeVisible();
  await expect(mention.locator('.slash-name')).toHaveText(`@${agent}-child`);
  await mention.click();
  await expect(compose).toHaveValue(`@${agent}-child `);
  await trigger.click();
  await childRow.getByRole('menuitem').click();
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  const after = await stored();
  expect(after.scope.agent_id).toBe(`${agent}-child`);
  expect(after.parent_session_id).toBe(main.id);
  expect(after.title).toBe(renamed);
  const messages = await (await request.get(`/api/sessions/${main.id}/messages`)).json();
  expect(messages.messages || []).toEqual([]);
});

test('@ux-session-004 Archive and restore through the supplied session actions', async ({ page, request }, info) => {
  const frozen=loadCorpus().find(row=>row.id==='@ux-session-004');expect(frozen).toBeTruthy();
  await info.attach('gherkin',{body:frozen.steps.join('\n'),contentType:'text/plain'});
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

test('@ux-mobile-004 Native pinned and active overlap yields one stable active-first carousel', async ({ page, request }, info) => {
  test.setTimeout(60000);
  const scenario = loadCorpus().find(row => row.id === '@ux-mobile-004'); expect(scenario).toBeTruthy();
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
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
      const key = `${token}-gate-${index}`, gate = resolve('test-results/ux-parity/queue-gates', key); mkdirSync(resolve(gate, '..'), { recursive: true });
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
      await expect(popup.locator(`[data-session-jid="gi:${id}"] [role="menuitem"]`)).toHaveAttribute('aria-current', 'true');
      await page.keyboard.press('Escape'); await expect(popup).toHaveCount(0);
    };
    const pick = async id => {
      await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
      const popup = page.locator('.compose-session-popup'); await popup.getByRole('searchbox', { name: 'Search sessions', exact: true }).fill(id);
      const row = popup.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem'); await expect(row).toBeVisible(); await row.click(); await expect.poll(selected).toBe(id); await ready(id);
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

for(const [id,method] of [['@shared-23','pointer'],['@shared-24','keyboard']]) test(`${id} Session picker ${method} opening focuses before first visible frame and keeps its composer anchor`,async({page,request},info)=>{
  const source=loadCorpus('shared').find(row=>row.id===id);expect(source).toBeTruthy();expect(source.steps.join('\n')).toContain(method);await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
  const token=`shared-picker-${info.project.name}-${method}-${Date.now()}`;
  const create=async title=>{const response=await request.post('/api/sessions',{data:{agent_id:`${token}-${title}`,title}});expect(response.status()).toBe(201);return(await response.json()).id;};
  const main=await create('main'),research=await create('research');
  await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main);await page.goto('/');
  const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('shared picker draft');
  const triggerButtons=page.locator('.compose-session-trigger-group button');await expect(triggerButtons).toHaveCount(2);
  for(let index=0;index<2;index++){
    const trigger=triggerButtons.nth(index);await expect(trigger).toBeEnabled();
    // Observation only: no focus, style or event handlers in application code
    // are replaced. Capture insertion-time and first-frame focus separately.
    await page.evaluate(()=>{
      window.__pickerFrame=null;window.__pickerInsertion=null;
      const observer=new MutationObserver(()=>{
        const popup=document.querySelector('.compose-session-popup');if(!popup)return;
        observer.disconnect();const search=popup.querySelector('input[type="search"]');
        window.__pickerInsertion={focused:document.activeElement===search};
        requestAnimationFrame(()=>{const box=popup.getBoundingClientRect(),style=getComputedStyle(popup);window.__pickerFrame={focused:document.activeElement===search,visible:box.width>0&&box.height>0&&style.display!=='none'&&style.visibility!=='hidden',count:document.querySelectorAll('.compose-session-popup').length};});
      });observer.observe(document.body,{childList:true,subtree:true});window.__pickerObserver=observer;
    });
    if(method==='pointer')await trigger.click();else{await trigger.focus();await trigger.press('Enter');}
    await expect.poll(()=>page.evaluate(()=>window.__pickerFrame)).not.toBeNull();
    expect(await page.evaluate(()=>window.__pickerInsertion)).toEqual({focused:true});expect(await page.evaluate(()=>window.__pickerFrame)).toEqual({focused:true,visible:true,count:1});
    const popup=page.locator('.compose-session-popup'),search=popup.getByRole('searchbox',{name:'Search sessions',exact:true});
    const anchor=async()=>{
      const geometry=await popup.evaluate(el=>{const p=el.getBoundingClientRect(),host=el.closest('.compose-input-main'),a=host?.getBoundingClientRect();return {position:getComputedStyle(el).position,anchor:!!host,left:p.left,bottom:p.bottom,right:p.right,anchorLeft:a?.left,anchorTop:a?.top,viewport:innerWidth};});
      expect(geometry.anchor).toBe(true);
      // Current Classic reference uses fixed mobile bounds; desktop retains
      // the composer anchor and six-pixel gap. Focus/query checks stay intact.
      if(geometry.viewport<=639){expect(geometry.position).toBe('fixed');expect(Math.abs(geometry.left-8)).toBeLessThanOrEqual(1);expect(Math.abs(geometry.right-(geometry.viewport-8))).toBeLessThanOrEqual(1);}
      else{expect(geometry.position).toBe('absolute');expect(Math.abs(geometry.left-geometry.anchorLeft)).toBeLessThanOrEqual(1);expect(Math.abs(geometry.bottom-(geometry.anchorTop-6))).toBeLessThanOrEqual(1);}
      expect(geometry.left).toBeGreaterThanOrEqual(0);expect(geometry.right).toBeLessThanOrEqual(geometry.viewport+1);
    };
    await anchor();for(const sid of [main,research])await expect(popup.locator(`[data-session-jid="gi:${sid}"]`)).toBeVisible();
    await search.fill(`gi:${research}`);await expect(popup.locator(`[data-session-jid="gi:${research}"]`)).toBeVisible();await expect(popup.locator(`[data-session-jid="gi:${main}"]`)).toHaveCount(0);
    const width=info.project.use.viewport.width;await page.setViewportSize({width:width<700?820:390,height:900});await anchor();await expect(search).toHaveValue(`gi:${research}`);await expect(search).toBeFocused();
    await page.keyboard.press('Escape');await expect(popup).toHaveCount(0);await expect(trigger).toBeFocused();await expect(input).toHaveValue('shared picker draft');
    expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main);await page.setViewportSize(info.project.use.viewport);
    await page.evaluate(()=>window.__pickerObserver?.disconnect());
  }
  for(const sid of [main,research])expect((await(await request.get(`/api/sessions/${sid}/turns`)).json()).turns||[]).toHaveLength(0);
});

test('@shared-26 Expose only native session mutations and recover a rejected edit',async({page,request},info)=>{
 const source=loadCorpus('shared').find(s=>s.id==='@shared-26');expect(source.name).toBe('Expose only supported session mutations');await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const token=`capabilities-${info.project.name}-${Date.now()}`;
 const created=await request.post('/api/sessions',{data:{agent_id:token,title:token}});expect(created.status()).toBe(201);const main=(await created.json()).id;
 const fork=async name=>{const r=await request.post(`/api/sessions/${main}/fork`,{data:{agent_id:`${token}-${name}`,title:`${token}-${name}`}});expect(r.status()).toBe(201);return(await r.json()).branch.chat_jid.slice(3);};
 const research=await fork('research'),busy=await fork('busy');
 const get=async id=>{const r=await request.get(`/api/sessions/${id}`);expect(r.status()).toBe(200);return r.json();};
 const gate=resolve('test-results/ux-parity/queue-gates',token);mkdirSync(resolve(gate,'..'),{recursive:true});
 const reject=async(id,data)=>{const before=await get(id),r=await request.patch(`/api/sessions/${id}`,{data});expect(r.status()).toBe(409);expect((await r.json()).error).toContain('session mutation conflicts');expect(await get(id)).toEqual(before);};
 const idempotent=async(id,action)=>{const {updated_at:beforeTime,...before}=await get(id),r=await request.patch(`/api/sessions/${id}`,{data:{action}});expect(r.status()).toBe(200);const {updated_at:afterTime,...after}=await r.json();expect(after).toEqual(before);};
 try {
  const running=await request.post(`/api/sessions/${busy}/prompt`,{data:{prompt:`UX queue gate:${token}`,model:'test-model'}});expect(running.status()).toBe(202);const turn=(await running.json()).turn_id;
  await expect.poll(async()=>(await get(busy)).state.status).toBe('running');
  await reject(main,{action:'archive'});await reject(busy,{action:'archive'});
  // Restore-on-idle is supported and idempotent, but needs no visible action.
  await idempotent(research,'restore');const initial=await get(research);
  // Native session records genuinely have no message counts. Do not remove
  // metadata in interception just to manufacture the unknown-count condition.
  const catalogue=(await(await request.get('/api/sessions')).json()).sessions;
  for(const id of [main,research,busy]){
   const entry=catalogue.find(s=>s.id===id);expect(entry).toBeTruthy();
   for(const key of ['message_count','messageCount']){expect(entry).not.toHaveProperty(key);expect(entry.state).not.toHaveProperty(key);}
   const before=await get(id),denied=await request.delete(`/api/sessions/${id}`);expect(denied.status()).toBe(405);expect(denied.headers().allow).toBe('GET, PATCH');expect(await get(id)).toEqual(before);
  }
  await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main);await page.goto('/');
  const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('main capability draft');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'capability-draft.txt',mimeType:'text/plain',buffer:Buffer.from('retained capability draft bytes')});
  const trigger=page.getByRole('button',{name:/Manage sessions for/}).last(),picker=page.locator('.compose-session-popup');
  const row=id=>picker.locator(`[data-session-jid="gi:${id}"]`);
  const action=(id,name)=>row(id).getByRole('button',{name:new RegExp(`^${name} @`)});
  const selection=async id=>{await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);};
  const noDelete=async()=>{await expect(picker.getByRole('button',{name:/^Delete /})).toHaveCount(0);await expect(picker.locator('.compose-model-popup-item-delete')).toHaveCount(0);};
  const mutate=async(id,body,activate,status=200)=>{
   const result=page.waitForResponse(r=>r.request().method()==='PATCH'&&new URL(r.url()).pathname===`/api/sessions/${id}`);
   await activate();const r=await result;expect(r.status()).toBe(status);expect(r.request().postDataJSON()).toEqual(body);return r.json();
  };
  const draft=async()=>{await expect(input).toHaveValue('main capability draft');await expect(page.locator('.compose-input-main .compose-file-pill[title="capability-draft.txt"]')).toHaveCount(1);};
  let deletes=0;page.on('request',r=>{if(r.method()==='DELETE'&&/\/api\/sessions\/[^/]+$/.test(new URL(r.url()).pathname))deletes++;});
  await trigger.click();await noDelete();
  await expect(action(main,'Archive')).toHaveCount(0);await expect(action(research,'Archive')).toBeEnabled();await expect(action(research,'Restore')).toHaveCount(0);
  await expect(action(busy,'Archive')).toHaveCount(0);await expect(action(busy,'Pin')).toBeEnabled();await expect(action(busy,'Rename')).toBeEnabled();
  // Check deletion absence when the running branch and unknown-count idle
  // branch are each current, covering the header action as well as row actions.
  for(const id of [busy,research,main]){await row(id).getByRole('menuitem').click();await selection(id);await trigger.click();await noDelete();}
  await draft();
  await mutate(research,{action:'pin',pinned:true},()=>action(research,'Pin').click());
  await expect(picker.getByRole('group',{name:'Pinned',exact:true}).locator(`[data-session-jid="gi:${research}"]`)).toBeVisible();expect((await get(research)).state.pinned).toBe(true);
  await action(research,'Rename').click();const name=picker.getByRole('textbox',{name:'Session name',exact:true});await expect(name).toBeFocused();await name.fill('   ');
  const rejected=await mutate(research,{action:'rename',title:'   '},()=>picker.getByRole('button',{name:'Save name',exact:true}).click(),400);
  expect(rejected.error).toContain('title must be');await expect(picker.getByRole('alert')).toContainText('title must be');await expect(name).toHaveValue('   ');await expect(picker).toBeVisible();await selection(main);await draft();expect((await get(research)).title).toBe(initial.title);
  const title=`${token}-renamed`;await name.fill(title);await mutate(research,{action:'rename',title},()=>picker.getByRole('button',{name:'Save name',exact:true}).click());
  await expect(row(research).getByRole('menuitem')).toContainText(title);await expect(picker.getByRole('alert')).toHaveCount(0);expect((await get(research)).title).toBe(title);
  await action(research,'Archive').click();await mutate(research,{action:'archive'},()=>picker.getByRole('button',{name:'Confirm archive',exact:true}).click());
  await expect(picker.getByRole('group',{name:'Archived',exact:true}).locator(`[data-session-jid="gi:${research}"]`)).toBeVisible();
  for(const name of ['Pin','Unpin','Rename','Archive'])await expect(action(research,name)).toHaveCount(0);
  await expect(action(research,'Restore')).toBeEnabled();await noDelete();expect((await get(research)).state.archived_at).toBeTruthy();await selection(main);await draft();
  await reject(research,{action:'rename',title:'must remain archived'});await reject(research,{action:'pin',pinned:true});await reject(research,{action:'pin',pinned:false});
  // Repeated archive retains the stored archive timestamp and identity.
  await idempotent(research,'archive');
  await page.keyboard.press('Escape');await page.reload();await draft();await trigger.click();await expect(action(research,'Restore')).toBeEnabled();await noDelete();
  await mutate(research,{action:'restore'},()=>action(research,'Restore').click());await expect(action(research,'Unpin')).toBeEnabled();expect((await get(research)).state.archived_at||null).toBe(null);
  await mutate(research,{action:'pin',pinned:false},()=>action(research,'Unpin').click());await expect(action(research,'Pin')).toBeEnabled();expect((await get(research)).state.pinned).toBe(false);
  const after=await get(research);expect(after.scope).toEqual(initial.scope);expect(after.parent_session_id).toBe(main);expect(after.title).toBe(title);
  // The enabled New action must allocate a real child and preserve the origin
  // draft. No callback or branch identity is fabricated in the browser.
  const newSession=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===`/api/sessions/${main}/fork`);
  await expect(picker.getByRole('button',{name:'New',exact:true})).toBeEnabled();await picker.getByRole('button',{name:'New',exact:true}).click();
  const response=await newSession;expect(response.status()).toBe(201);const child=(await response.json()).branch.chat_jid.slice(3);expect([main,research,busy]).not.toContain(child);await selection(child);expect((await get(child)).parent_session_id).toBe(main);await expect(input).toHaveValue('');await expect(page.locator('.compose-input-main .compose-file-pill')).toHaveCount(0);
  await trigger.click();await noDelete();await row(main).getByRole('menuitem').click();await selection(main);await draft();await page.reload();await draft();
  for(const id of [main,research,child])expect((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns||[]).toEqual([]);
  const turns=(await(await request.get(`/api/sessions/${busy}/turns`)).json()).turns;expect(turns).toHaveLength(1);expect(turns[0]).toMatchObject({id:turn,status:'running'});expect(deletes).toBe(0);
 }finally{writeFileSync(gate,'release');}
});

test('@shared-33 Native session picker searches and activates enabled entries with incremental typeahead',async({page,request},info)=>{
 const source=loadCorpus('shared').find(s=>s.id==='@shared-33');expect(source.steps.join('\n')).toContain('session picker');await info.attach('gherkin',{body:source.steps.join('\n'),contentType:'text/plain'});
 const token=`typeahead-${info.project.name}-${Date.now()}`;
 const create=async(title,agent)=>{const r=await request.post('/api/sessions',{data:{agent_id:agent,title}});expect(r.status()).toBe(201);return r.json();};
 const main=await create(`${token}-main`,`${token}-main`),substring=await create(`z-alpha-${token}`,`${token}-substring`),alpha=await create(`alpha-${token}`,`${token}-alpha`),alpine=await create(`alpine-${token}`,`${token}-alpine`);
 const others=[];for(let i=0;i<8;i++)others.push(await create(`other-${i}-${token}`,`${token}-${i}`));
 expect((await request.patch(`/api/sessions/${substring.id}`,{data:{action:'pin',pinned:true}})).status()).toBe(200);
 expect((await request.patch(`/api/sessions/${alpha.id}/model`,{data:{model:'test/bootstrap'}})).status()).toBe(200);
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('typeahead unsent draft');await page.locator('.compose-box input[type=file]').setInputFiles({name:'typeahead.txt',mimeType:'text/plain',buffer:Buffer.from('retained bytes')});
 const trigger=page.getByRole('button',{name:/Manage sessions for/}).last(),popup=page.locator('.compose-session-popup'),search=page.getByRole('searchbox',{name:'Search sessions',exact:true});
 const row=id=>popup.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem');
 const visible=()=>popup.locator('[data-session-jid]').evaluateAll(nodes=>nodes.map(n=>n.dataset.sessionJid));
 const highlighted=()=>popup.locator('[data-session-entry-key].active').getAttribute('data-session-entry-key');
 let writes=0;page.on('request',r=>{if(!['GET','HEAD'].includes(r.method())&&new URL(r.url()).pathname.startsWith('/api/sessions'))writes++;});
 await trigger.click();await expect(search).toBeFocused();await search.fill(token);
 const native=(await(await request.get('/api/sessions')).json()).sessions.filter(s=>s.title.includes(token));
 const expected=[main.id,...native.filter(s=>s.id!==main.id&&s.state?.pinned).map(s=>s.id),...native.filter(s=>s.id!==main.id&&!s.state?.pinned).map(s=>s.id)].map(id=>`gi:${id}`);
 await expect.poll(visible).toEqual(expected);
 for(const [query,result] of [[alpha.id,[`gi:${alpha.id}`]],[`alpha-${token}`,[`gi:${substring.id}`,`gi:${alpha.id}`]],[`${token} bootstrap`,[`gi:${alpha.id}`]]]){
  await search.fill(query);await expect.poll(visible).toEqual(result);await expect(search).toBeFocused();
 }
 await search.fill(token);await expect.poll(visible).toEqual(expected.filter(id=>[main,substring,alpha,alpine,...others].some(s=>`gi:${s.id}`===id)));
 const filtered=await visible();expect(filtered.length).toBe(12);await expect(popup.getByRole('button',{name:/^Delete /})).toHaveCount(0);
 // Begin on a pinned substring match. Prefix matching must beat it, then the
 // second/third character must resolve among two similar prefix names.
 await row(substring.id).focus();await row(substring.id).press('a');await expect.poll(highlighted).toBe(`session:gi:${alpha.id}`);await expect(row(alpha.id)).toBeFocused();
 await page.keyboard.type('lp');await expect(row(alpha.id)).toBeFocused();await page.keyboard.press('i');await expect.poll(highlighted).toBe(`session:gi:${alpine.id}`);await expect(row(alpine.id)).toBeFocused();await expect(search).toHaveValue(token);expect(await visible()).toEqual(filtered);
 // Navigation stays inside enabled filtered results, including page steps.
 await row(alpha.id).focus();await page.keyboard.press('Home');await expect.poll(highlighted).toBe(`session:${filtered[0]}`);await expect(search).toBeFocused();
 await search.press('ArrowUp');await expect.poll(highlighted).toBe(`session:${filtered.at(-1)}`);await search.press('ArrowDown');await expect.poll(highlighted).toBe(`session:${filtered[0]}`);
 await search.press('PageDown');await expect.poll(highlighted).toBe(`session:${filtered[8]}`);await search.press('PageUp');await expect.poll(highlighted).toBe(`session:${filtered[0]}`);
 await row(alpha.id).focus();await page.keyboard.press('End');await expect.poll(highlighted).toBe(`session:${filtered.at(-1)}`);
 await page.keyboard.press('Escape');await expect(popup).toHaveCount(0);await expect(trigger).toBeFocused();expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main.id);await expect(input).toHaveValue('typeahead unsent draft');
 await trigger.click();await search.fill(token);await row(substring.id).focus();await page.keyboard.type('alpi');await expect(row(alpine.id)).toBeFocused();await expect.poll(highlighted).toBe(`session:gi:${alpine.id}`);
 await page.evaluate(()=>{window.__typeaheadSelections=[];const original=Storage.prototype.setItem;Storage.prototype.setItem=function(k,v){if(k==='gi_session_id')window.__typeaheadSelections.push(v);return original.call(this,k,v);};});
 await page.keyboard.press('Enter');await expect(popup).toHaveCount(0);await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(alpine.id);await expect(input).toHaveValue('');expect(await page.evaluate(()=>window.__typeaheadSelections)).toEqual([alpine.id]);
 await trigger.click();await row(main.id).click();await expect(input).toHaveValue('typeahead unsent draft');await expect(page.locator('.compose-input-main .compose-file-pill[title="typeahead.txt"]')).toHaveCount(1);await page.reload();await expect(input).toHaveValue('typeahead unsent draft');await expect(page.locator('.compose-input-main .compose-file-pill[title="typeahead.txt"]')).toHaveCount(1);
 expect(writes).toBe(0);for(const id of [main.id,alpha.id,alpine.id])expect((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns||[]).toEqual([]);
});
