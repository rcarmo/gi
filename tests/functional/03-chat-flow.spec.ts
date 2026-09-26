/**
 * 03-chat-flow.spec.ts — Verify the core chat experience.
 *
 * Tests the complete send → process → display cycle including compose box
 * interaction, message persistence, timeline rendering, and content visibility.
 */
import { test, expect, type Page } from '@playwright/test';
import { randomUUID } from 'node:crypto';
import { BASE_URL, waitForAppShell, getComposeInput, sendMessage, apiGet } from './helpers';

// Existing history must not satisfy a send assertion. Capture this unique
// request and correlate native admission, stored prompt and rendered response.
async function sendExact(page: Page, label: string) {
  const prompt = `${label} ${randomUUID()}`;
  const sessionId = await page.evaluate(() => localStorage.getItem('gi_session_id'));
  expect(sessionId).toBeTruthy();
  const before = await apiGet(page.request, `/api/sessions/${sessionId}/turns`);
  const response = page.waitForResponse(r => r.request().method() === 'POST' && r.url().endsWith(`/api/sessions/${sessionId}/prompt`));
  await sendMessage(page, prompt);
  const accepted = await response;
  expect(accepted.status()).toBe(202);
  const admission = await accepted.json();
  expect(admission.session_id).toBe(sessionId);
  expect(admission.turn_id).toBeTruthy();
  await expect.poll(async () => {
    const data = await apiGet(page.request, `/api/sessions/${sessionId}/turns`);
    return (data.turns || []).find((turn: any) => turn.id === admission.turn_id)?.status;
  }).toBe('completed');
  const turns = (await apiGet(page.request, `/api/sessions/${sessionId}/turns`)).turns || [];
  expect(turns).toHaveLength((before.turns || []).length + 1);
  expect(turns.filter((turn: any) => turn.prompt === prompt)).toHaveLength(1);
  expect(turns.find((turn: any) => turn.id === admission.turn_id)?.prompt).toBe(prompt);
  const user = page.locator('.post:not(.agent-post)').filter({ has: page.locator('.post-content').filter({ hasText: prompt }) });
  const assistant = page.locator('.post.agent-post').filter({ has: page.locator('.post-content').filter({ hasText: `Gi received: ${prompt}` }) });
  await expect(user).toHaveCount(1); await expect(assistant).toHaveCount(1);
  return { prompt, sessionId, admission, user, assistant };
}

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
    await sendExact(page, 'Enter submit test');
  });

  test('assistant responds to a message', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendExact(page, 'Hello from functional test');
  });

  test('user message content is visible in post-content', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const { user } = await sendExact(page, 'visible content check');
    await expect(user.locator('.post-content')).toBeVisible();
  });

  test('assistant response content is visible in post-content', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const { assistant } = await sendExact(page, 'assistant content check');
    await expect(assistant.locator('.post-content')).toBeVisible();
  });

  test('messages are persisted in the database', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const { prompt, sessionId } = await sendExact(page, 'persistence check');
    const msgs = await apiGet(request, `/api/sessions/${sessionId}/messages`);
    expect(msgs.messages.filter((m: any) => m.role === 'user' && m.content === prompt)).toHaveLength(1);
    expect(msgs.messages.filter((m: any) => m.role === 'assistant' && m.content.includes(`Gi received: ${prompt}`))).toHaveLength(1);
    await page.reload(); await waitForAppShell(page);
    await expect(page.locator('.post.agent-post .post-content').filter({ hasText: `Gi received: ${prompt}` })).toHaveCount(1);
  });

  test('posts have avatar elements', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const { user, assistant } = await sendExact(page, 'avatar check');
    await expect(user.locator('.post-avatar')).toBeVisible();
    await expect(assistant.locator('.post-avatar')).toBeVisible();
  });

  test('agent posts have agent-avatar class', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const { assistant } = await sendExact(page, 'agent avatar class check');
    await expect(assistant.locator('.post-avatar.agent-avatar')).toBeVisible();
  });

  test('compose box clears after submit', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendExact(page, 'will be cleared');
    await expect(getComposeInput(page)).toHaveValue('');
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

test('native completed shell status has persisted identity and frozen duration',async({page,request})=>{
 const session=await(await request.post('/api/sessions',{data:{agent_id:`tool-status-${Date.now()}`,title:'tool status'}})).json();
 const response=await request.post(`/api/sessions/${session.id}/prompt`,{data:{prompt:'tool timing proof',model:'test-model'}});expect(response.status()).toBe(202);const {turn_id}=await response.json();
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${session.id}/activity`)).json()).tool?.state)).toBe('completed');
 const activity=await(await request.get(`/api/sessions/${session.id}/activity`)).json();expect(activity.tool.turn_id).toBe(turn_id);expect(activity.tool.tool_call_id).toBeTruthy();expect(activity.tool.duration_ms).toBeGreaterThanOrEqual(0);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto(BASE_URL);await getComposeInput(page).fill('unsent tool draft');const status=page.locator('.gi-tool-activity');await expect(status).toHaveAttribute('data-tool-call-id',activity.tool.tool_call_id);await expect(status.getByLabel('shell: Completed')).toBeVisible();await expect(status.getByLabel('Tool duration')).toHaveText(`${Math.floor(activity.tool.duration_ms/1000)}s`);await expect(status.locator('.spinner')).toHaveCount(0);await expect(getComposeInput(page)).toHaveValue('unsent tool draft');
});
