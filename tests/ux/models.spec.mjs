import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function setup(page,request,info){
 const agent=`models-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:agent,title:`@${agent}`}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:`${agent}-child`,title:`${agent}-child`}})).json()).branch.chat_jid.slice(3);
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const modelButton=page.getByRole('button',{name:'Open model picker',exact:true});
 const menu=page.getByRole('menu',{name:'Model picker',exact:true});
 const switchTo=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);};
 const option=name=>menu.getByRole('menuitem').filter({hasText:name});
 return{main,child,input,modelButton,menu,switchTo,option};
}
async function model(request,id){return(await(await request.get(`/api/sessions/${id}/model`)).json()).current;}

test('@ux-original-020 Select a model for the captured chat without submitting its draft',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-original-020');expect(scenario).toBeTruthy();
 await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,modelButton,option}=await setup(page,request,info);
 await input.fill('unsent model draft');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'draft.txt',mimeType:'text/plain',buffer:Buffer.from('draft')});
 const accepted=page.waitForResponse(res=>res.url().endsWith(`/api/sessions/${main.id}/model`)&&res.request().method()==='PATCH');
 await modelButton.click();await option('test/bootstrap').click();
 const chosen=await accepted;expect(chosen.status()).toBe(200);
 const payload=await chosen.json();expect(payload.current).toBe('test/bootstrap');
 expect(payload).toHaveProperty('context_window');expect(payload).toHaveProperty('context_usage');
 await expect(modelButton).toHaveText('test/bootstrap');await expect(input).toHaveValue('unsent model draft');
 await expect(page.locator('.compose-file-pill[title="draft.txt"]')).toBeVisible();
 expect(await model(request,child)).toBe('test/test-model');
 await page.reload();await expect(modelButton).toHaveText('test/bootstrap');await expect(input).toHaveValue('unsent model draft');
 // Native keyboard selection follows the same mutation path.
 await modelButton.focus();await page.keyboard.press('Enter');await option('test/test-model').focus();await page.keyboard.press('Enter');
 await expect(modelButton).toHaveText('test/test-model');expect(await model(request,main.id)).toBe('test/test-model');
 const rejected=page.waitForResponse(res=>res.url().endsWith(`/api/sessions/${main.id}/model`)&&res.request().method()==='PATCH');
 await modelButton.click();await option('test/unavailable-model').click();
 expect((await rejected).status()).toBe(400);
 await expect(page.getByRole('alert')).toContainText('unavailable or lacks credentials');
 expect(await model(request,main.id)).toBe('test/test-model');await expect(input).toHaveValue('unsent model draft');
 const{turns}=await(await request.get(`/api/sessions/${main.id}/turns`)).json();expect(turns||[]).toHaveLength(0);
});

test('@ux-original-021 Do not apply stale model responses to another chat',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-original-021');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,child,input,modelButton,option,switchTo}=await setup(page,request,info);
 await input.fill('main draft');
 let release,held=false,done;const gate=new Promise(resolve=>{release=resolve;});const delivered=new Promise(resolve=>{done=resolve;});
 await page.route(`**/api/sessions/${main.id}/model`,async route=>{
  if(route.request().method()!=='PATCH')return route.continue();const response=await route.fetch();held=true;await gate;await route.fulfill({response});done();
 });
 try{
  await modelButton.click();await option('test/bootstrap').click();await expect.poll(()=>held).toBe(true);
  await switchTo(child);await input.fill('child draft');release();await delivered;
  await expect(modelButton).toHaveText('test/test-model');await expect(input).toHaveValue('child draft');
  expect(await model(request,main.id)).toBe('test/bootstrap');expect(await model(request,child)).toBe('test/test-model');
  await page.unroute(`**/api/sessions/${main.id}/model`);
  await switchTo(main.id);await expect(modelButton).toHaveText('test/bootstrap');await expect(input).toHaveValue('main draft');
 }finally{release();}
});

test('Gi superseded model catalogue cannot undo an accepted choice on an A-B-A revisit',async({page,request},info)=>{
 const{main,child,modelButton,option,switchTo}=await setup(page,request,info);
 let release,held=false,done;const gate=new Promise(resolve=>{release=resolve;});const delivered=new Promise(resolve=>{done=resolve;});
 await page.route(`**/api/sessions/${main.id}/model`,async route=>{
  if(held||route.request().method()!=='GET')return route.continue();const response=await route.fetch();held=true;await gate;await route.fulfill({response});done();
 });
 try{
  await modelButton.click();await expect.poll(()=>held).toBe(true);
  await switchTo(child);await switchTo(main.id);
  await modelButton.click();await option('test/bootstrap').click();await expect(modelButton).toHaveText('test/bootstrap');
  release();await delivered;await expect(modelButton).toHaveText('test/bootstrap');
 }finally{release();}
});

test('Gi failed model mutation after switching sessions reports no false success in the target',async({page,request},info)=>{
 const{main,child,input,modelButton,option,switchTo}=await setup(page,request,info);
 let release,held=false,done;const gate=new Promise(resolve=>{release=resolve;});const delivered=new Promise(resolve=>{done=resolve;});
 await page.route(`**/api/sessions/${main.id}/model`,async route=>{
  if(route.request().method()!=='PATCH')return route.continue();const response=await route.fetch();expect(response.status()).toBe(400);held=true;await gate;await route.fulfill({response});done();
 });
 try{
  await input.fill('keep origin draft');await modelButton.click();await option('test/unavailable-model').click();await expect.poll(()=>held).toBe(true);
  await switchTo(child);await input.fill('target draft');release();await delivered;
  await expect(page.getByRole('alert')).toHaveCount(0);await expect(modelButton).toHaveText('test/test-model');await expect(input).toHaveValue('target draft');
  await switchTo(main.id);await expect(input).toHaveValue('keep origin draft');expect(await model(request,main.id)).toBe('test/test-model');
 }finally{release();}
});

test('@ux-compaction-008 Handle a model command using the configured provider catalogue',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-compaction-008');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const{main,input,modelButton}=await setup(page,request,info);
 await input.fill('/model test/bootstrap');await input.press('Enter');await expect(modelButton).toHaveText('test/bootstrap');
 await input.fill('/model nonexistent-model');await input.press('Enter');
 await expect(page.getByRole('alert')).toContainText('unknown model');
 await expect(input).toHaveValue('/model nonexistent-model');expect(await model(request,main.id)).toBe('test/bootstrap');
 const{turns}=await(await request.get(`/api/sessions/${main.id}/turns`)).json();expect(turns||[]).toHaveLength(0);
});
