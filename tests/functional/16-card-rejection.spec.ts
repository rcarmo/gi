import {test,expect} from '@playwright/test';
test.skip(!process.env.GI_UX_CARD_REJECTION,'requires isolated native card fixture (make test-ux-card-rejection)');
test('unsupported Submit shows native renderer error and does not persist a submission',async({page,request})=>{
 const id='card-rejection-main',get=async()=> (await(await request.get(`/api/sessions/${id}/messages`)).json()).messages,before=await get();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),id);await page.goto('/');const post=page.locator(`#post-${id}-post`),answer=post.getByRole('textbox',{name:'Card answer',exact:true});await answer.fill('retained functional answer');await post.getByRole('button',{name:'Submit answer',exact:true}).click();
 await expect(post.locator('.adaptive-card-notice-error')).toContainText('Your inputs have not been submitted.');await expect(answer).toHaveValue('retained functional answer');await expect(post.locator('.adaptive-card-status,.adaptive-card-submission-receipt')).toHaveCount(0);expect(await get()).toEqual(before);
});
