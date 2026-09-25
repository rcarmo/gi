import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';

test('Native session panel exposes row metadata and leading pin without selecting another session',async({page,request},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Session panel retained draft Ω');
  const main=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
  const created=await request.post(`${env.origin}/api/sessions`,{data:{agent_id:'panel',title:'Another session'}});expect(created.status()).toBe(201);const other=(await created.json()).id;
  // Reload reads native catalogue and restores this session's draft.
  await page.reload();await expect(input).toHaveValue('Session panel retained draft Ω');
  const trigger=page.locator('.compose-session-trigger-top button');await trigger.click();const panel=page.locator('.compose-session-popup'),search=panel.getByRole('searchbox',{name:'Search sessions',exact:true});await expect(search).toBeFocused();
  await expect(panel.locator('.compose-session-popup-header')).toContainText('Search sessions');
  const row=panel.locator(`[data-session-jid="gi:${other}"]`),current=panel.locator(`[data-session-jid="gi:${main}"]`);
  await expect(row.locator('.compose-session-row-label')).toHaveText('@another-session');await expect(row.locator('.compose-session-row-jid')).toHaveText(`gi:${other}`);await expect(current.locator('.compose-session-status-pill.current')).toHaveText('current');
  const pin=row.getByRole('button',{name:'Pin @Another session',exact:true});await expect(pin).toHaveClass(/compose-session-row-pin/);await expect(pin).toHaveAttribute('aria-pressed','false');
  await pin.click();await expect(row.getByRole('button',{name:'Unpin @Another session',exact:true})).toHaveAttribute('aria-pressed','true');
  expect((await(await request.get(`${env.origin}/api/sessions/${other}`)).json()).state.pinned).toBe(true);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main);
  await search.fill(`gi:${other}`);await expect(current).toHaveCount(0);await expect(row).toBeVisible();
  await panel.getByRole('button',{name:'Close session picker',exact:true}).click();await expect(panel).toBeHidden();await expect(trigger).toBeFocused();await expect(input).toHaveValue('Session panel retained draft Ω');
  await trigger.click();await expect(search).toBeFocused();await page.keyboard.press('Escape');await expect(trigger).toBeFocused();
 }finally{await env.close();}
});

test('Session search initial selection settles before first-frame Arrow navigation',async({page,request},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();
  const main=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
  expect((await request.patch(`${env.origin}/api/sessions/${main}`,{data:{action:'rename',title:'boundary-main'}})).ok()).toBe(true);
  const response=await request.post(`${env.origin}/api/sessions`,{data:{agent_id:'panel',title:'boundary-other'}});const other=(await response.json()).id;
  expect((await request.post(`${env.origin}/api/sessions`,{data:{agent_id:'panel',title:'unmatched'}})).ok()).toBe(true);
  await page.reload();await expect(input).toBeFocused();await page.locator('.compose-session-trigger-top button').click();
  const search=page.getByRole('searchbox',{name:'Search sessions',exact:true});await expect(search).toBeFocused();
  await expect(page.locator('#compose-session-results [data-session-entry-key]')).toHaveCount(3);
  const result=await page.evaluate(()=>new Promise(resolve=>{
   const search=document.querySelector('.compose-session-search'),menu=document.querySelector('#compose-session-results');
   const observer=new MutationObserver(()=>{
    if(search.value!=='boundary-'||menu.querySelectorAll('[data-session-entry-key]').length!==2)return;
    observer.disconnect();
    search.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowDown',bubbles:true,cancelable:true}));
    setTimeout(()=>resolve(menu.querySelector('[data-session-entry-key].active')?.getAttribute('data-session-entry-key')),160);
   });observer.observe(menu,{childList:true,subtree:true,attributes:true});setTimeout(()=>{observer.disconnect();resolve('observer did not fire');},2000);
   search.value='boundary-';search.dispatchEvent(new Event('input',{bubbles:true}));
  }));
  expect(result).toBe(`session:gi:${other}`);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(main);
 }finally{await env.close();}
});
