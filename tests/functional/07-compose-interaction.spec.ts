/**
 * 07-compose-interaction.spec.ts — Verify compose box interaction details.
 *
 * Tests keyboard behavior, history navigation, and input state management
 * beyond basic send/receive.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, getComposeInput, sendMessage } from './helpers';

test.describe('Compose interaction', () => {

  test('native slash suggestions own Tab and Escape without invoking Quick Actions or sending',async({page})=>{
    await page.goto(BASE_URL);await waitForAppShell(page);const input=getComposeInput(page);const posts:string[]=[];page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))posts.push(r.url());});
    await input.fill('/');await expect(page.locator('.slash-name')).toContainText(['/model','/compact']);await expect(page.locator('.slash-name').filter({hasText:'/shutdown'})).toHaveCount(0);await expect(page.locator('.timeline-quick-actions')).toHaveCount(0);
    await input.fill('/m preserved argument');await expect(page.locator('.slash-autocomplete')).toBeVisible();await input.press('Tab');await expect(input).toHaveValue('/model preserved argument');await expect(input).toBeFocused();expect(posts).toEqual([]);
    await input.fill('/m');await input.press('Escape');await expect(page.locator('.slash-autocomplete')).toHaveCount(0);await expect(input).toHaveValue('/m');expect(posts).toEqual([]);
  });

  test('mobile picker geometry exposes touch-sized dismissal without sending', async ({page})=>{
    await page.setViewportSize({width:390,height:700});await page.goto(BASE_URL);await waitForAppShell(page);const input=getComposeInput(page);await input.fill('Mobile picker functional draft');
    const writes:string[]=[];page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))writes.push(r.url());});
    for(const kind of ['session','model']){
      const trigger=kind==='session'?page.getByRole('button',{name:/Manage sessions for/}).last():page.locator('.compose-model-hint-btn');await trigger.click();const popup=page.locator('.compose-model-popup');await expect(popup).toBeVisible();
      const g=await popup.evaluate(e=>({x:e.getBoundingClientRect().x,width:e.getBoundingClientRect().width,position:getComputedStyle(e).position}));expect(g).toEqual({x:8,width:374,position:'fixed'});
      await page.getByRole('button',{name:`Close ${kind} picker`,exact:true}).click();await expect(popup).toHaveCount(0);await expect(trigger).toBeFocused();await expect(input).toHaveValue('Mobile picker functional draft');
    }
    expect(writes).toEqual([]);
  });

  test('Shift+Enter does not submit', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    const admissions:string[]=[];page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/prompt'))admissions.push(r.url());});
    await input.fill('line one');
    await input.press('Shift+Enter');
    await input.type('line two');
    await expect(input).toHaveValue('line one\nline two');
    expect(admissions).toEqual([]);
  });

  test('empty input does not submit on Enter', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await input.fill('');
    const postCountBefore = await page.locator('.post').count();
    await input.press('Enter');
    await page.waitForTimeout(1000);
    const postCountAfter = await page.locator('.post').count();
    // No new post should appear
    expect(postCountAfter).toBe(postCountBefore);
  });

  test('typing reaches the compose input', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await expect(input).toBeFocused();
    await page.keyboard.type('focus test');
    const value = await input.inputValue();
    expect(value).toContain('focus test');
  });

  test('multiple messages can be sent in sequence', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'first message');
    await page.waitForTimeout(3000);
    await sendMessage(page, 'second message');
    await page.waitForTimeout(3000);
    // Should have at least 2 user messages
    const posts = await page.locator('.post').count();
    expect(posts).toBeGreaterThanOrEqual(2);
  });
});

test('native attachment transport displays upload then sending without blocking newer typing',async({page,request})=>{
 const session=await(await request.post(`${BASE_URL}/api/sessions`,{data:{title:'transfer functional',agent_id:'transfer-functional'}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto(BASE_URL);await waitForAppShell(page);
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});
 let releaseUpload,releaseSend,held=false,sending=false;
 const uploadGate=new Promise(r=>releaseUpload=r),sendGate=new Promise(r=>releaseSend=r);
 await page.route(`**/api/sessions/${session.id}/media`,async route=>{const response=await route.fetch({postData:route.request().postDataBuffer()});held=true;await uploadGate;await route.fulfill({response});});
 await page.route(`**/api/sessions/${session.id}/prompt`,async route=>{sending=true;await sendGate;await route.fulfill({response:await route.fetch()});});
 try{
  await input.fill('functional transfer');await page.locator('.compose-box input[type=file]').setInputFiles({name:'transport.txt',mimeType:'text/plain',buffer:Buffer.from('native transfer bytes')});await input.press('Enter');
  await expect.poll(()=>held).toBe(true);await expect(page.getByRole('progressbar',{name:'Attachment upload progress'})).toBeVisible();await expect(page.locator('.gi-compose-sending')).toHaveCount(0);
  await input.fill('newer functional draft');releaseUpload();await expect.poll(()=>sending).toBe(true);
  await expect(page.locator('.gi-compose-upload')).toHaveCount(0);await expect(page.locator('.compose-send-stack .send-btn')).toHaveAttribute('aria-busy','true');await expect(input).toHaveValue('newer functional draft');
  releaseSend();await expect(page.locator('.gi-compose-transfer')).toBeHidden();await expect(input).toHaveValue('newer functional draft');
 }finally{releaseUpload();releaseSend()}
});

test('quick actions prefill native model command without submitting',async({page,request})=>{
 const session=await(await request.post(`${BASE_URL}/api/sessions`,{data:{title:'quick functional',agent_id:'quick-functional'}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto(BASE_URL);await waitForAppShell(page);
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('before quick action');await input.blur();
 await page.locator('.timeline').click({position:{x:160,y:180}});await page.keyboard.type('m');const query=page.locator('.timeline-quick-actions-input');await expect(query).toBeFocused();await query.fill('/model');
 await expect(page.locator('.timeline-quick-actions-item.active .timeline-quick-actions-item-title')).toHaveText('/model');await page.locator('.timeline-quick-actions-item-slash').click();
 await expect(input).toHaveValue('/model ');await expect(input).toBeFocused();expect(await input.evaluate(el=>el.selectionStart)).toBe(7);
 expect((await(await request.get(`${BASE_URL}/api/sessions/${session.id}/turns`)).json()).turns??[]).toEqual([]);
});

test('message copy preserves original Markdown and reports unavailable clipboard honestly',async({page,request})=>{
 const session=await(await request.post(`${BASE_URL}/api/sessions`,{data:{agent_id:'functional-copy',title:'copy'}})).json();
 await request.post(`${BASE_URL}/api/sessions/${session.id}/prompt`,{data:{prompt:'**functional source**',model:'test-model'}});
 let messages;
 await expect.poll(async()=>{messages=(await(await request.get(`${BASE_URL}/api/sessions/${session.id}/messages`)).json()).messages??[];return messages.filter(m=>m.role==='assistant').length;}).toBe(1);
 const stored=messages.find(m=>m.role==='assistant');
 await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);window.__copies=[];document.addEventListener('copy',event=>window.__copies.push({text:event.clipboardData?.getData('text/plain'),trusted:event.isTrusted}));},session.id);
 await page.goto(BASE_URL);await waitForAppShell(page);const post=page.locator(`[id="post-${stored.id}"]`);
 await post.getByRole('button',{name:'Copy message',exact:true}).click();await expect(post.getByRole('button',{name:'Copied',exact:true})).toBeVisible();
 expect(await page.evaluate(()=>window.__copies.at(-1))).toEqual({text:stored.content.trimEnd(),trusted:true});
 await expect(post.getByRole('button',{name:'Copy message',exact:true})).toBeVisible({timeout:5000});
 await page.evaluate(()=>{document.execCommand=()=>false;Object.defineProperty(navigator,'clipboard',{configurable:true,value:undefined});});
 await post.getByRole('button',{name:'Copy message',exact:true}).click();await expect(post.getByRole('button',{name:'Copy failed',exact:true})).toBeVisible();
});

test('unknown native skill fails recoverably without creating a turn',async({page,request})=>{
 const session=await(await request.post('/api/sessions',{data:{agent_id:`unknown-skill-${Date.now()}`,title:'unknown skill'}})).json();await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto('/');const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('/skill:not-loaded retain β');await input.press('Enter');await expect(page.getByRole('alert')).toContainText('unknown or unavailable loaded skill');await expect(input).toHaveValue('/skill:not-loaded retain β');expect((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns||[]).toEqual([]);
});
