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

test('cancelled composer upload retains native file and draft without message dispatch',async({page,request})=>{
 const created=await(await request.post('/api/sessions',{data:{agent_id:`cancel-${Date.now()}`,title:'Cancel upload'}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),created.id);let release,held=false;const gate=new Promise<void>(r=>release=r);
 await page.route(`**/api/sessions/${created.id}/media`,async route=>{const response=await route.fetch({postData:route.request().postDataBuffer()});expect(response.status()).toBe(201);held=true;await gate;try{await route.fulfill({response});}catch{}});
 try{
  await page.goto('/');const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('retained cancellation draft');await page.locator('.compose-box input[type=file]').setInputFiles({name:'cancel.txt',mimeType:'text/plain',buffer:Buffer.from('native cancellation bytes')});await input.press('Enter');await expect.poll(()=>held).toBe(true);
  await page.getByRole('button',{name:'Cancel uploads',exact:true}).click();await expect(input).toHaveValue('retained cancellation draft');await expect(page.locator('.compose-box').getByText('cancel.txt',{exact:true})).toBeVisible();await expect(page.locator('.gi-compose-transfer')).toBeHidden();
  expect((await(await request.get(`/api/sessions/${created.id}/turns`)).json()).turns||[]).toEqual([]);
 }finally{release();}
});
