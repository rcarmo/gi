import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function fixture(page,request,info) {
  const main=await (await request.post('/api/sessions',{data:{title:`lightbox ${info.project.name}`,agent_id:'lightbox'}})).json();
  await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);
  await page.goto('/');
  const input=page.getByRole('textbox',{name:inputName,exact:true}); await expect(input).toBeVisible();
  const base64=await page.evaluate(()=>{const c=document.createElement('canvas');c.width=320;c.height=160;const x=c.getContext('2d');x.fillStyle='#325979';x.fillRect(0,0,320,160);x.fillStyle='#eff6ff';x.font='22px sans-serif';x.fillText('Native stored image',35,85);return c.toDataURL('image/png').split(',')[1]});
  const bytes=Buffer.from(base64,'base64');
  await input.fill('lightbox attachment');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'native-image.png',mimeType:'image/png',buffer:bytes});
  const sent=page.waitForRequest(r=>r.method()==='POST'&&r.url().endsWith(`/api/sessions/${main.id}/prompt`));
  await input.press('Enter'); const body=(await sent).postDataJSON();
  await expect.poll(async()=>((await (await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).some(t=>t.status==='completed')).toBe(true);
  const stored=(await (await request.get(`/api/sessions/${main.id}/messages`)).json()).messages.find(m=>m.role==='user');
  const post=page.locator(`#post-${stored.id}`), image=post.locator('.media-preview img');
  await expect(image).toBeVisible();
  await expect.poll(()=>image.evaluate(el=>el.complete&&el.naturalWidth===320)).toBe(true);
  const id=body.media[0].media_id;
  expect(await (await request.get(`/api/media/${id}/raw`)).body()).toEqual(bytes);
  await input.fill('unsent image review draft');
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'keep.txt',mimeType:'text/plain',buffer:Buffer.from('retained')});
  let submits=0; page.on('request',r=>{if(r.method()==='POST'&&/\/(prompt|queue)$/.test(new URL(r.url()).pathname))submits++});
  const open=async()=>{await image.click();await expect(page.locator('.image-modal')).toBeVisible();await expect.poll(()=>page.locator('.image-modal img').evaluate(el=>el.complete&&el.naturalWidth===320)).toBe(true)};
  const preserved=async()=>{await expect(input).toHaveValue('unsent image review draft');await expect(page.locator('.compose-file-pill').filter({hasText:'keep.txt'})).toBeVisible();expect(submits).toBe(0);};
  return {main,input,post,image,id,open,preserved};
}
async function evidence(info,id) {await info.attach('gherkin',{body:loadCorpus().find(row=>row.id===id).steps.join('\n'),contentType:'text/plain'});}

test('@ux-timeline-013 Escape dismisses the stored-image lightbox and preserves the draft',async({page,request},info)=>{
  await evidence(info,'@ux-timeline-013'); const f=await fixture(page,request,info);
  await f.open(); await page.keyboard.press('Escape'); await expect(page.locator('.image-modal')).toHaveCount(0);await expect(f.post).toBeVisible();await f.preserved();
  await page.reload();await expect(f.image).toBeVisible();await f.open();await page.keyboard.press('Escape');await f.preserved();
});
test('@ux-timeline-014 Non-Escape keys leave the lightbox open without sending or changing the draft',async({page,request},info)=>{
  await evidence(info,'@ux-timeline-014'); const f=await fixture(page,request,info);await f.open();
  for(const key of ['Space','Enter','x','ArrowDown','ArrowLeft']){await page.keyboard.press(key);await expect(page.locator('.image-modal')).toBeVisible();}
  await page.keyboard.press('Escape');await f.preserved();
});
test('@ux-timeline-015 Both backdrop and image clicks dismiss the lightbox',async({page,request},info)=>{
  await evidence(info,'@ux-timeline-015');const f=await fixture(page,request,info);await f.open();
  await page.locator('.image-modal').click({position:{x:5,y:5}});await expect(page.locator('.image-modal')).toHaveCount(0);
  await f.open();await page.locator('.image-modal img').click();await expect(page.locator('.image-modal')).toHaveCount(0);await f.preserved();
});
test.describe('touch surface',()=>{
  test.use({hasTouch:true});
  test('@ux-timeline-016 Native touch taps dismiss the stored-image lightbox',async({page,request},info)=>{
    await evidence(info,'@ux-timeline-016');const f=await fixture(page,request,info);
    // Playwright WebKit on Linux can report maxTouchPoints=0 despite touch
    // emulation. Require delivered trusted touch events, not a navigator stub.
    await page.evaluate(()=>{window.__touches=[];document.addEventListener('touchstart',e=>window.__touches.push({trusted:e.isTrusted,count:e.touches.length}),true)});
    await f.image.tap();await expect(page.locator('.image-modal')).toBeVisible();
    await page.screenshot({path:info.outputPath('native-lightbox.png')});await info.attach('native-lightbox',{path:info.outputPath('native-lightbox.png'),contentType:'image/png'});
    await page.locator('.image-modal').tap({position:{x:5,y:5}});await expect(page.locator('.image-modal')).toHaveCount(0);
    await f.image.tap();await page.locator('.image-modal img').tap();await expect(page.locator('.image-modal')).toHaveCount(0);await f.preserved();
    expect(await page.evaluate(()=>window.__touches)).toEqual(Array.from({length:4},()=>({trusted:true,count:1})));
  });
});

test('Gi stored images survive search, session changes and a native lookup failure without draft loss',async({page,request},info)=>{
  const f=await fixture(page,request,info);
  const child=(await (await request.post(`/api/sessions/${f.main.id}/fork`,{data:{title:'image-other',agent_id:'image-other'}})).json()).branch.chat_jid.slice(3);
  const select=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();};
  await f.open();await page.keyboard.press('Escape');await select(child);await expect(f.post).toHaveCount(0);await expect(page.locator('.image-modal')).toHaveCount(0);
  await f.input.fill('other session draft');await select(f.main.id);await f.preserved();await expect(f.image).toBeVisible();
  await page.getByRole('button',{name:'Search',exact:true}).click();const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});
  await search.fill('lightbox attachment');await search.press('Enter');await f.open();await page.keyboard.press('Escape');
  await page.getByRole('button',{name:'Close search',exact:true}).click();await f.preserved();
  const pattern=`**/api/media/${f.id}/raw`;
  await page.route(pattern,route=>route.continue({url:route.request().url().replace(`/media/${f.id}/raw`,'/media/999999999/raw')}));
  const failed=page.waitForResponse(r=>r.url().includes('/api/media/')&&r.status()===404);
  await page.reload();await failed;await f.preserved();
  await page.unroute(pattern);await page.reload();await expect.poll(()=>f.image.evaluate(el=>el.complete&&el.naturalWidth===320)).toBe(true);await f.preserved();
  await select(child);await expect(f.input).toHaveValue('other session draft');
});
