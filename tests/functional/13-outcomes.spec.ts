import {test,expect} from '@playwright/test';
test('native recovered reply exposes one outcome after its timestamp',async({page,request})=>{
 test.skip(!process.env.GI_UX_OUTCOMES,'Recovery lifecycle fixture: make test-ux-outcomes');
 const response=await request.get('/api/sessions/outcome-fixture/messages');expect(response.status()).toBe(200);const reply=(await response.json()).messages.find(m=>m.role==='assistant');expect(reply.payload.content_blocks[0].recovered).toBe(true);
 await page.addInitScript(()=>localStorage.setItem('gi_session_id','outcome-fixture'));await page.goto('/');const post=page.locator(`[id="post-${reply.id}"]`);await expect(post.locator('.post-recovery-chip')).toHaveText('recovered');expect(await post.locator('.post-recovery-chip').evaluate(el=>el.previousElementSibling?.classList.contains('post-time'))).toBe(true);
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('functional outcome draft');await page.reload();await expect(input).toHaveValue('functional outcome draft');await expect(post.locator('.post-recovery-chip')).toHaveCount(1);
});
