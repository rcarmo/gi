/**
 * 03-chat-flow.spec.ts — Verify the core chat experience.
 *
 * Tests the complete send → process → display cycle including compose box
 * interaction, message persistence, timeline rendering, and content visibility.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, getComposeInput, sendMessage, waitForPostCount, apiGet, findSessionForMessage } from './helpers';

test.describe('Chat flow', () => {

  test('compose box is visible and accepts input', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await expect(input).toBeVisible({ timeout: 5000 });
    await input.fill('test input text');
    await expect(input).toHaveValue('test input text');
  });

  test('Enter submits a message', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'Enter submit test');
    // User message should appear as a post
    await waitForPostCount(page, 1);
  });

  test('assistant responds to a message', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'Hello from functional test');
    // Wait for both user + assistant posts
    await waitForPostCount(page, 2, 15000);
  });

  test('user message content is visible in post-content', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'visible content check');
    await page.waitForTimeout(3000);
    await expect(
      page.locator('.post-content').filter({ hasText: 'visible content check' }).first()
    ).toBeVisible({ timeout: 10000 });
  });

  test('assistant response content is visible in post-content', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'assistant content check');
    // The test instance shell stub responds with "Gi received: ..."
    await expect(
      page.locator('.post-content').filter({ hasText: 'Gi received' }).first()
    ).toBeVisible({ timeout: 15000 });
  });

  test('messages are persisted in the database', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'persistence check');
    await page.waitForTimeout(5000);
    // Fetch sessions and messages via API
    const session = await findSessionForMessage(request, 'persistence check');
    const msgs = await apiGet(request, `/api/sessions/${session.id}/messages`);
    const contents = msgs.messages.map((m: any) => m.content);
    expect(contents.some((c: string) => c.includes('persistence check'))).toBeTruthy();
  });

  test('posts have avatar elements', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'avatar check');
    await waitForPostCount(page, 2, 15000);
    const avatars = page.locator('.post-avatar');
    expect(await avatars.count()).toBeGreaterThan(0);
  });

  test('agent posts have agent-avatar class', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'agent avatar class check');
    await waitForPostCount(page, 2, 15000);
    await expect(page.locator('.post-avatar.agent-avatar').first()).toBeVisible({ timeout: 5000 });
  });

  test('compose box clears after submit', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await input.fill('will be cleared');
    await input.press('Enter');
    await page.waitForTimeout(500);
    // Input should be empty after submit
    const value = await input.inputValue();
    expect(value).toBe('');
  });
});

test('horizontal table scrolling keeps selected session and draft even with Safari swipe enabled',async({page,request})=>{
 const session=await(await request.post('/api/sessions',{data:{title:'scroll owner',agent_id:`scroll-${Date.now()}`}})).json();await request.post('/api/sessions',{data:{title:'scroll neighbour',agent_id:`neighbour-${Date.now()}`}});
 const source='| '+Array.from({length:12},(_,i)=>`Column ${i}`).join(' | ')+' |\n| '+Array(12).fill('---').join(' | ')+' |\n| '+Array(12).fill('wide_column_'+ 'abcdefgh'.repeat(8)).join(' | ')+' |';
 const sent=await(await request.post(`/api/sessions/${session.id}/prompt`,{data:{prompt:source,model:'test-model'}})).json();await expect.poll(async()=>((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns||[]).find(t=>t.id===sent.turn_id)?.status).toBe('completed');
 await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);Object.defineProperty(navigator,'userAgent',{configurable:true,value:'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 Version/17.0 Safari/605.1.15'});},session.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('scroll draft');const table=page.locator('.post:not(.agent-post) table').first();await expect(table).toBeVisible();await table.locator('td').first().hover();await page.mouse.wheel(10000,0);await expect.poll(()=>table.evaluate(el=>el.closest('.post-content')!.scrollLeft)).toBeGreaterThan(0);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(session.id);await expect(input).toHaveValue('scroll draft');
});

test('native assistant read-aloud toggles and clears on audio failure without submitting draft', async ({page,request})=>{
 await page.addInitScript(()=>{
  (window as any).__utterances=[];
  Object.defineProperty(window,'SpeechSynthesisUtterance',{configurable:true,value:class {text:string;constructor(text:string){this.text=text;}}});
  Object.defineProperty(window,'speechSynthesis',{configurable:true,value:{speak(u:any){(window as any).__utterances.push(u);},cancel(){}}});
 });
 const session=await(await request.post('/api/sessions',{data:{agent_id:`speech-${Date.now()}`,title:'speech functional'}})).json();
 const accepted=await request.post(`/api/sessions/${session.id}/prompt`,{data:{prompt:'Native spoken message',model:'test-model'}});expect(accepted.status()).toBe(202);
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns||[])[0]?.status).toBe('completed');
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto(BASE_URL);const input=getComposeInput(page);await input.fill('unsent draft');
 await page.getByRole('button',{name:'Read aloud',exact:true}).click();await expect(page.getByRole('button',{name:'Stop reading aloud',exact:true})).toHaveAttribute('aria-pressed','true');
 expect(await page.evaluate(()=>(window as any).__utterances[0].text)).toContain('Native spoken message');await page.evaluate(()=>(window as any).__utterances[0].onerror());
 await expect(page.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');await expect(input).toHaveValue('unsent draft');expect((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns).toHaveLength(1);
});
