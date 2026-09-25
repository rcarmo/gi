import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';

test('Compose surface uses persisted bounded height and retains draft across picker and reload',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.addInitScript(()=>{if(localStorage.getItem('piclaw_compose_height')===null)localStorage.setItem('piclaw_compose_height','80');});
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();
  await input.fill('Retained surface draft Ω\nSecond line');
  const handle=page.getByRole('separator',{name:'Resize message input'}),trigger=page.locator('.compose-session-trigger-top button');
  await expect(trigger).toHaveCount(1);await expect(page.locator('.compose-actions .compose-session-trigger')).toHaveCount(0);
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(80);
  await expect(input).toHaveCSS('padding-top','2px');
  await handle.focus();await page.keyboard.press('ArrowUp');await page.keyboard.press('ArrowUp');
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(100);
  expect(await page.evaluate(()=>localStorage.getItem('piclaw_compose_height'))).toBe('100');
  await trigger.click();await expect(page.locator('.compose-session-popup')).toBeVisible();await page.keyboard.press('Escape');await expect(trigger).toBeFocused();
  await page.locator('.compose-model-hint-btn').click();await expect(page.locator('.compose-model-popup')).toBeVisible();await page.keyboard.press('Escape');
  await expect(input).toHaveValue('Retained surface draft Ω\nSecond line');
  await page.reload();await expect(input).toHaveValue('Retained surface draft Ω\nSecond line');
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(100);
  await expect(input).toBeFocused();
 }finally{await env.close();}
});

test('Compose surface mouse resize commits, Escape cancels, viewport caps and long draft scrolls',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();
  const draft='Long surface draft Ω\n'.repeat(90);await input.fill(draft);
  const autoMax=Math.max(page.viewportSize().width>=1024?70:50,Math.min(Math.floor(page.viewportSize().height*.4),300,...(page.viewportSize().width<=639?[Math.floor(page.viewportSize().height*.3),200]:[])));
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(autoMax);
  expect(await input.evaluate(e=>e.scrollHeight>e.clientHeight&&getComputedStyle(e).overflowY==='auto')).toBe(true);
  const handle=page.getByRole('separator',{name:'Resize message input'}),box=await handle.boundingBox();
  await page.mouse.move(box.x+box.width/2,box.y+box.height/2);await page.mouse.down();await page.mouse.move(box.x+box.width/2,box.y+box.height/2-40);await page.mouse.up();
  const manual=Math.min(autoMax+40,Math.floor(page.viewportSize().height*.5),520);
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(manual);
  expect(await page.evaluate(()=>localStorage.getItem('piclaw_compose_height'))).toBe(String(manual));
  const moved=await handle.boundingBox();await page.mouse.move(moved.x+moved.width/2,moved.y+7);await page.mouse.down();await page.mouse.move(moved.x+moved.width/2,moved.y-30);await page.keyboard.press('Escape');await page.mouse.up();
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(manual);
  expect(await page.evaluate(()=>document.body.style.cursor)).toBe('');
  await page.setViewportSize({width:390,height:400});await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(200);
  await expect(input).toHaveValue(draft);
  await handle.focus();await page.keyboard.press('Home');expect(await page.evaluate(()=>localStorage.getItem('piclaw_compose_height'))).toBeNull();
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(120);
  await page.reload();await expect(input).toHaveValue(draft);
 }finally{await env.close();}
});

test.describe('Touch-capable surface',()=>{
 test.use({hasTouch:true});
 test('Touch resize cancellation, modal ownership and window blur restore page state',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Owned surface Ω');
  const handle=page.getByRole('separator',{name:'Resize message input'});await expect(handle).toBeVisible();await page.waitForTimeout(100);const height=await input.evaluate(e=>e.getBoundingClientRect().height);
  await handle.evaluate(h=>{const e=new Event('touchstart',{bubbles:true,cancelable:true});Object.defineProperty(e,'touches',{value:[{clientY:300}]});h.dispatchEvent(e);});
  await page.evaluate(()=>{const e=new Event('touchmove',{bubbles:true,cancelable:true});Object.defineProperty(e,'touches',{value:[{clientY:260}]});document.dispatchEvent(e);});
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(height+40);
  await page.evaluate(()=>document.dispatchEvent(new Event('touchcancel')));
  await expect.poll(()=>input.evaluate(e=>e.getBoundingClientRect().height)).toBe(height);
  expect(await page.evaluate(()=>document.body.style.userSelect)).toBe('');
  await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Settings',exact:true}).click();await expect(page.getByRole('dialog')).toBeVisible();
  await handle.dispatchEvent('keydown',{key:'ArrowUp'});await handle.dispatchEvent('mousedown',{button:0,clientY:300});
  expect(await page.evaluate(()=>document.body.style.cursor)).toBe('');expect(await page.evaluate(()=>localStorage.getItem('piclaw_compose_height'))).toBeNull();
  await page.keyboard.press('Escape');await expect(page.getByRole('dialog')).toBeHidden();await expect(input).toHaveValue('Owned surface Ω');
  await handle.evaluate(h=>{const e=new Event('touchstart',{bubbles:true,cancelable:true});Object.defineProperty(e,'touches',{value:[{clientY:300}]});h.dispatchEvent(e);});
  await page.evaluate(()=>window.dispatchEvent(new Event('blur')));
  expect(await page.evaluate(()=>document.body.style.cursor)).toBe('');
  await expect(input).toHaveValue('Owned surface Ω');
 }finally{await env.close();}
});

});
