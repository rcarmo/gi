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

for(const theme of ['light','dark'])test(`Active session pill uses reference padding and maximum contrast in ${theme} mode`,async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.emulateMedia({colorScheme:theme});await page.addInitScript(theme=>{localStorage.setItem('piclaw_theme',theme);localStorage.setItem('vibes-theme',theme);},theme);
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Contrast retained draft Ω');
  const pill=page.locator('.compose-current-agent-label.active');await expect(pill).toHaveCSS('padding','0px 8px');await expect(pill).toHaveCSS('color','rgb(0, 0, 0)');
  const sample=()=>pill.evaluate(e=>{const c=getComputedStyle(e);const rgb=(s)=>s.match(/\d+(?:\.\d+)?/g).slice(0,3).map(Number);const lum=s=>{const a=rgb(s).map(v=>{v/=255;return v<=.03928?v/12.92:((v+.055)/1.055)**2.4;});return .2126*a[0]+.7152*a[1]+.0722*a[2];};const a=lum(c.color),b=lum(c.backgroundColor);return (Math.max(a,b)+.05)/(Math.min(a,b)+.05);});
  expect(await sample()).toBeGreaterThanOrEqual(4.5);
  await page.keyboard.press('Control+,');const dialog=page.getByRole('dialog');await dialog.getByRole('button',{name:'Appearance',exact:true}).click();
  const tint=dialog.getByRole('textbox',{name:'Custom tint',exact:true}),save=dialog.getByRole('button',{name:'Save appearance',exact:true});
  // Blue requires white, a pale custom tint requires black. The native saved
  // browser appearance changes, never a server write or a screenshot override.
  for(const [value,color]of [['#000066','rgb(255, 255, 255)'],['#ffee00','rgb(0, 0, 0)']]){
   await tint.fill(value);await save.click();await expect(pill).toHaveCSS('color',color);expect(await sample()).toBeGreaterThanOrEqual(4.5);
  }
  await dialog.getByRole('button',{name:'Reset appearance',exact:true}).click();await page.keyboard.press('Escape');await expect(input).toHaveValue('Contrast retained draft Ω');await expect(input).toBeFocused();
  await expect(pill).toHaveCSS('color','rgb(15, 20, 25)');expect(await sample()).toBeGreaterThanOrEqual(4.5);
 }finally{await env.close();}
});

test('Pinned theme text contrast adjusts primary and secondary without changing draft or palette choices',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.emulateMedia({colorScheme:'dark'});await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Palette contrast draft Ω');
  const colours=()=>page.evaluate(()=>{
   const root=getComputedStyle(document.documentElement),keys=['--text-primary','--text-secondary','--bg-primary','--bg-secondary','--bg-hover'];
   const resolve=name=>{const e=document.createElement('span');e.style.color=root.getPropertyValue(name);document.body.append(e);const rgb=getComputedStyle(e).color;e.remove();return rgb;};
   return Object.fromEntries(keys.map(k=>[k,resolve(k)]));
  });
  await expect.poll(async()=> (await colours())['--text-secondary']).toBe('rgb(130, 134, 139)');await expect(page.getByRole('button',{name:'Open model picker',exact:true})).toHaveCSS('color','rgb(130, 134, 139)');
  await page.emulateMedia({colorScheme:'light'});await expect.poll(async()=> (await colours())['--text-secondary']).toBe('rgb(83, 100, 113)');
  const ratio=(a,b)=>{const lum=s=>{const v=s.match(/[\d.]+/g).slice(0,3).map(Number).map(n=>{n/=255;return n<=.04045?n/12.92:((n+.055)/1.055)**2.4;});return .2126*v[0]+.7152*v[1]+.0722*v[2];};const x=lum(a),y=lum(b);return(Math.max(x,y)+.05)/(Math.min(x,y)+.05);};
  await page.keyboard.press('Control+,');const dialog=page.getByRole('dialog');await dialog.getByRole('button',{name:'Appearance',exact:true}).click();
  for(const [preset,tint,mode]of [['monokai','','dark'],['solarized','','light'],['default','#8040aa','dark'],['default','#cc8800','light']]){
   await page.emulateMedia({colorScheme:mode});await dialog.getByRole('combobox',{name:'Theme preset',exact:true}).selectOption(preset);
   if(preset==='default')await dialog.getByRole('textbox',{name:'Custom tint',exact:true}).fill(tint);
   await dialog.getByRole('button',{name:'Save appearance',exact:true}).click();await expect(dialog.getByRole('status')).toHaveText('Appearance saved in this browser.');
   const c=await colours();for(const text of ['--text-primary','--text-secondary'])for(const bg of ['--bg-primary','--bg-secondary','--bg-hover'])expect(ratio(c[text],c[bg]),`${preset}/${mode}/${text}/${bg}`).toBeGreaterThanOrEqual(4.5);
   await expect(dialog.getByRole('button',{name:'Save appearance',exact:true})).toBeFocused();await expect(input).toHaveValue('Palette contrast draft Ω');
  }
  await dialog.getByRole('button',{name:'Reset appearance',exact:true}).click();await page.keyboard.press('Escape');await expect(input).toBeFocused();await expect(input).toHaveValue('Palette contrast draft Ω');await expect.poll(async()=> (await colours())['--text-secondary']).toBe('rgb(83, 100, 113)');
 }finally{await env.close();}
});
