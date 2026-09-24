import {test,expect} from '@playwright/test';
test.skip(!process.env.GI_UX_RECOVERY_CONTROLS,'requires isolated recovery-control fixture (make test-ux-recovery-controls)');
test('native stored protected recovery control is hidden without losing malformed neighbour or draft',async({page,request})=>{
 const id='recovery-controls-fixture',response=await request.get(`/api/sessions/${id}/messages`);expect(response.ok()).toBe(true);
 const initial=(await response.json()).messages;expect(initial.some(m=>m.id==='recovery-control-valid-typed')).toBe(true);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),id);await page.goto('/');
 await expect(page.locator('#post-recovery-control-partial-typed')).toBeVisible();await expect(page.locator('#post-recovery-control-valid-typed')).toHaveCount(0);
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('protected recovery draft');await page.reload();await expect(input).toHaveValue('protected recovery draft');
 await expect(page.locator('#post-recovery-control-valid-typed')).toHaveCount(0);expect((await(await request.get(`/api/sessions/${id}/messages`)).json()).messages).toEqual(initial);
});
