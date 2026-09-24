import {test,expect} from '@playwright/test';
test.skip(!process.env.GI_UX_RECOVERY_PLACEHOLDERS,'requires isolated placeholder fixture (make test-ux-recovery-placeholders)');
test('empty info recovery placeholder is hidden without removing stored row or visible attachments and cards',async({page,request})=>{
 const id='recovery-placeholders-fixture',response=await request.get(`/api/sessions/${id}/messages`);expect(response.ok()).toBe(true);const initial=(await response.json()).messages;expect(initial.some(m=>m.id==='recovery-placeholder-empty-info')).toBe(true);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),id);await page.goto('/');
 await expect(page.locator('#post-recovery-placeholder-authored-text')).toContainText('Recovery authored text stays visible');await expect(page.locator('#post-recovery-placeholder-empty-info')).toHaveCount(0);
 await expect(page.locator('#post-recovery-placeholder-file')).toContainText('recovery-retained.txt');await expect(page.locator('#post-recovery-placeholder-card')).toContainText('Recovery retained card');
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('placeholder functional draft');await page.reload();await expect(input).toHaveValue('placeholder functional draft');await expect(page.locator('#post-recovery-placeholder-empty-info')).toHaveCount(0);
 expect((await(await request.get(`/api/sessions/${id}/messages`)).json()).messages).toEqual(initial);
});
