import {test,expect} from '@playwright/test';

test('stored remote metadata uses safe text-only cards without changing the message or draft',async({page,request})=>{
 test.skip(!process.env.GI_UX_LINKS,'Native metadata seed belongs to make test-ux-links isolated server');
 const stored=(await(await request.get('/api/sessions/links-fixture/messages')).json()).messages;
 await page.addInitScript(()=>localStorage.setItem('gi_session_id','links-fixture'));const outbound:string[]=[];
 await page.route(/^https?:\/\/[^/]+\.example\.invalid\//,route=>{outbound.push(route.request().url());return route.abort();});
 await page.goto('/');const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('functional link draft');const post=page.locator('[id="post-links-message"]');
 await expect(post.locator('.resource-link')).toHaveAttribute('target','_blank');await expect(post.locator('.resource-link')).toHaveAttribute('rel','noopener noreferrer');await expect(post.locator('.link-preview')).toHaveAttribute('href','http://preview.example.invalid/article');await expect(post.locator('.link-preview')).toHaveAttribute('rel','noopener noreferrer');await expect(post).not.toContainText('Unsafe');await expect(post.locator('.link-preview-image')).toHaveCount(0);expect(outbound).toEqual([]);
 await page.reload();await expect(input).toHaveValue('functional link draft');expect((await(await request.get('/api/sessions/links-fixture/messages')).json()).messages).toEqual(stored);
});
