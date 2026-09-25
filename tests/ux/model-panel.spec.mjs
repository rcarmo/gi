import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';

test('Native model panel search, metadata, clear and Models-settings handoff preserve draft and focus',async({page,request},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Panel handoff draft Ω');
  const id=await page.evaluate(()=>localStorage.getItem('gi_session_id')),before=await(await request.get(`${env.origin}/api/sessions/${id}/model`)).json();
  const writes=[];await page.route('**/api/**',route=>{if(!['GET','HEAD','OPTIONS'].includes(route.request().method())){writes.push(route.request().url());return route.abort();}return route.continue();});
  const trigger=page.getByRole('button',{name:'Open model picker',exact:true});await trigger.click();
  const panel=page.locator('.compose-model-catalogue'),search=panel.getByRole('searchbox',{name:'Search models',exact:true});await expect(search).toBeFocused();
  await expect(panel.locator('.compose-model-catalogue-section-heading').first()).toContainText('Current');
  await expect(panel.locator('.current-model')).toHaveCount(1);await expect(panel.locator('.compose-model-catalogue-option-name').first()).not.toHaveText('');
  await search.fill('does-not-exist');await expect(panel.getByRole('menuitem')).toHaveCount(0);await expect(panel.locator('.compose-model-catalogue-summary')).toContainText('0 models');
  await panel.getByRole('button',{name:'Clear model search',exact:true}).click();await expect(search).toBeFocused();await expect(panel.getByRole('menuitem').first()).toBeVisible();
  await search.fill(before.current);await expect(panel.getByRole('menuitem')).toHaveCount(1);await expect(panel.getByRole('menuitem')).toContainText(before.current);
  await panel.getByRole('button',{name:'Open Models settings',exact:true}).click();await expect(panel).toHaveCount(0);
  const dialog=page.getByRole('dialog');await expect(dialog).toBeVisible();await expect(dialog.locator('.settings-nav-item.active')).toHaveText('Models');await expect(dialog.getByRole('searchbox',{name:'Filter models',exact:true})).toBeVisible();
  await page.keyboard.press('Escape');await expect(dialog).toBeHidden();await expect(trigger).toBeFocused();await expect(input).toHaveValue('Panel handoff draft Ω');
  await trigger.click();await expect(search).toBeFocused();await expect(panel.locator('.compose-model-catalogue-summary')).not.toContainText('Refreshing');
  const initial=await panel.locator('[data-model-index].active').getAttribute('data-model-index');
  await search.press('ArrowDown');await expect(search).toBeFocused();await expect(panel.locator('[data-model-index].active')).not.toHaveAttribute('data-model-index',initial);
  await page.keyboard.press('Escape');await expect(trigger).toBeFocused();expect(writes).toEqual([]);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);
 }finally{await env.close();}
});

test('Invalid Settings section requests use General and model panel dismissal stays modal-owned',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Settings fallback draft');
  await page.evaluate(()=>window.dispatchEvent(new CustomEvent('piclaw:open-settings',{detail:{section:'not-a-pane',opener:{}}})));
  const dialog=page.getByRole('dialog');await expect(dialog).toBeVisible();await expect(dialog.locator('.settings-nav-item.active')).toHaveText('General');await page.keyboard.press('Escape');await expect(input).toBeFocused();
  await page.getByRole('button',{name:'Open model picker',exact:true}).click();await page.keyboard.press('Control+,');await expect(dialog).toBeVisible();await expect(dialog.locator('.settings-nav-item.active')).toHaveText('General');
  await page.keyboard.press('Escape');await expect(dialog).toBeHidden();await expect(page.locator('.compose-model-catalogue')).toBeVisible();await page.keyboard.press('Escape');await expect(page.locator('.compose-model-catalogue')).toBeHidden();await expect(input).toHaveValue('Settings fallback draft');
 }finally{await env.close();}
});
