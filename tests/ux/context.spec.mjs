import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';

test('@ux-context-002 Keep unknown usage values visibly unknown',async({page,request},info)=>{
 const scenario=loadCorpus().find(row=>row.id==='@ux-context-002');await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const agent=`context-${info.project.name}-${Date.now()}`;
 const session=await(await request.post('/api/sessions',{data:{agent_id:agent,title:`@${agent}`}})).json();
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);
 await page.goto('/');
 const meter=page.locator('.compose-context-pie');
 await expect(meter).toBeVisible();await expect(meter).toBeDisabled();
 await expect(meter).toHaveAttribute('title','Context: ? / ? tokens (?%) — usage unavailable — Context usage');
 await expect(meter).toHaveAttribute('aria-label','Context: ? / ? tokens (?%)');
 let state=await(await request.get(`/api/sessions/${session.id}/model`)).json();
 expect(state.context_usage).toMatchObject({tokens:null,percent:null,contextWindow:null,source:'unavailable'});
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});
 await input.fill('Native shell has no provider token measurement');await input.press('Enter');
 await expect.poll(async()=>{const{turns}=await(await request.get(`/api/sessions/${session.id}/turns`)).json();return turns?.[0]?.status;}).toBe('completed');
 await page.reload();await expect(meter).toHaveAttribute('title','Context: ? / ? tokens (?%) — usage unavailable — Compact context');
 state=await(await request.get(`/api/sessions/${session.id}/model`)).json();expect(state.context_usage.tokens).toBeNull();
 await input.fill('draft stays untouched');await page.getByRole('button',{name:'Open model picker',exact:true}).click();
 await page.getByRole('menu',{name:'Model picker',exact:true}).getByRole('menuitem').filter({hasText:'test/bootstrap'}).click();
 await expect(input).toHaveValue('draft stays untouched');await expect(meter).toBeEnabled();
});
