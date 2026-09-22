import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
const file={name:'draft.txt',mimeType:'text/plain',buffer:Buffer.from('keep these bytes')};
async function fixture(page,request,info){
 const token=`context-fit-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-child',title:token+'-child'}})).json()).branch.chat_jid.slice(3);
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const state=async (id=main.id)=>(await(await request.get(`/api/sessions/${id}/model`)).json());
 expect((await state()).context_usage.tokens).toBeNull();
 const response=page.waitForResponse(r=>r.url().endsWith(`/api/sessions/${main.id}/prompt`)&&r.request().method()==='POST');
 await input.fill('Measure a real local-provider request');await input.press('Enter');const turn=await(await response).json();
 await expect.poll(async()=> (await state()).context_usage.tokens).toBe(100);
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 32K tokens (0%)',{timeout:15000});
 // Measurement provenance is exposed by the real model API, never seeded in DB.
 expect((await state()).context_usage.measurement).toMatchObject({turn_id:turn.turn_id,model:'ux-local/gate',iteration:1});
 const modelButton=page.getByRole('button',{name:'Open model picker',exact:true});const menu=page.getByRole('menu',{name:'Model picker',exact:true});
 const option=name=>menu.getByRole('menuitem').filter({hasText:name});
 const switchTo=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();};
 return{main,child,input,state,turn,modelButton,menu,option,switchTo};
}

test('@ux-compaction-006 Check model context compatibility before switching',async({page,request},info)=>{
 const scenario=loadCorpus().find(x=>x.id==='@ux-compaction-006');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,state,modelButton,menu,option,switchTo}=await fixture(page,request,info);
 await input.fill('unsent retained');await page.locator('.compose-box input[type=file]').setInputFiles(file);
 let mutations=0;page.on('request',r=>{if(r.method()==='PATCH'&&r.url().endsWith(`/api/sessions/${main.id}/model`))mutations++;});
 await modelButton.click();const small=option('ux-local/small');await expect(small).toBeDisabled();await expect(small).toHaveAttribute('title','Blocked: ux-local/small context window is smaller than latest measured request');
 const box=await small.boundingBox();await page.mouse.click(box.x+box.width/2,box.y+box.height/2);expect(mutations).toBe(0);
 await page.keyboard.press('Escape');await expect(modelButton).toHaveText('ux-local/gate');await expect(input).toHaveValue('unsent retained');
 const denied=await request.patch(`/api/sessions/${main.id}/model`,{data:{model:'ux-local/small'}});expect(denied.status()).toBe(400);expect((await denied.json()).error).toContain('context window 80 is smaller than latest measured request (100 tokens)');
 expect((await state()).current).toBe('ux-local/gate');
 await modelButton.click();await expect(option('ux-local/equal')).toBeEnabled();await option('ux-local/equal').click();await expect(menu).toHaveCount(0);await expect(modelButton).toHaveText('ux-local/equal');
 expect((await state()).current).toBe('ux-local/equal');
 await page.reload();await expect(input).toHaveValue('unsent retained');await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);
 // Known capacity with unknown usage must not inherit another session's gate.
 await switchTo(child);expect((await state(child)).context_usage.tokens).toBeNull();await modelButton.click();await expect(option('ux-local/small')).toBeEnabled();await option('ux-local/small').click();
 expect((await state(child)).current).toBe('ux-local/small');expect((await state()).current).toBe('ux-local/equal');
 await switchTo(main.id);await expect(input).toHaveValue('unsent retained');await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);
});

test('@ux-compaction-007 Refresh model information after an accepted switch',async({page,request},info)=>{
 const scenario=loadCorpus().find(x=>x.id==='@ux-compaction-007');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,state,turn,modelButton,menu,option,switchTo}=await fixture(page,request,info);
 await input.fill('retained after accepted switch');await page.locator('.compose-box input[type=file]').setInputFiles(file);
 await modelButton.click();await option('ux-local/equal').click();await expect(menu).toHaveCount(0);await expect(modelButton).toHaveText('ux-local/equal');
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 100 tokens (100%)');
 let selected=await state();expect(selected.current).toBe('ux-local/equal');expect(selected.context_usage).toMatchObject({tokens:100,contextWindow:100,percent:100,source:'provider_request'});
 await modelButton.click();await option('ux-local/large').click();await expect(modelButton).toHaveText('ux-local/large');await expect(menu).toHaveCount(0);
 await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 200 tokens (50%)');
 selected=await state();expect(selected.context_usage.measurement.turn_id).toBe(turn.turn_id);expect(selected.context_usage.tokens).toBe(100);
 expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns).toHaveLength(1);
 await page.reload();await expect(modelButton).toHaveText('ux-local/large');await expect(input).toHaveValue('retained after accepted switch');await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toHaveCount(1);
 await switchTo(child);expect((await state(child)).context_usage.tokens).toBeNull();await switchTo(main.id);await expect(page.locator('.compose-context-pie')).toHaveAttribute('aria-label','Context: 100 / 200 tokens (50%)');
});
