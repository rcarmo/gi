import {test,expect} from '@playwright/test';
import {waitForAppShell} from './helpers';

test('stored raster media renders and opens through authenticated native adapters',async({page,request})=>{
 const s=await (await request.post('/api/sessions',{data:{title:'native-media',agent_id:'native-media'}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),s.id);
 await page.goto('/');await waitForAppShell(page);
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});
 const raw=await page.evaluate(()=>{const c=document.createElement('canvas');c.width=80;c.height=40;c.getContext('2d')!.fillRect(0,0,80,40);return c.toDataURL('image/png').split(',')[1]});
 await input.fill('functional stored image');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'functional.png',mimeType:'image/png',buffer:Buffer.from(raw,'base64')});
 await input.press('Enter');const image=page.locator('.media-preview img').first();await expect(image).toBeVisible();
 await expect.poll(()=>image.evaluate(el=>(el as HTMLImageElement).naturalWidth)).toBe(80);
 await input.fill('preserved');await image.click();await expect(page.locator('.image-modal img')).toBeVisible();
 await page.keyboard.press('Escape');await expect(page.locator('.image-modal')).toHaveCount(0);await expect(input).toHaveValue('preserved');
 await page.reload();await expect(image).toBeVisible();await expect(input).toHaveValue('preserved');
});
