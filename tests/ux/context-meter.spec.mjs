import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
const circumference=2*Math.PI*9;
async function fixture(page,request,info){
 const token=`meter-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-child',title:token+'-child'}})).json()).branch.chat_jid.slice(3);
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const pie=page.locator('.compose-context-pie'),arc=pie.locator('circle').last();
 const state=async(id=main.id)=>(await(await request.get(`/api/sessions/${id}/model`)).json());
 const measure=async(tokens,label,fill)=>{
  const sent=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith(`/api/sessions/${main.id}/prompt`));
  await input.fill(`UX meter tokens:${tokens}`);await input.press('Enter');const turn=await(await sent).json();
  await expect.poll(async()=> (await state()).context_usage.measurement?.turn_id).toBe(turn.turn_id);
  const usage=(await state()).context_usage;expect(usage).toMatchObject({tokens,contextWindow:2_000_000,percent:tokens/20000,source:'provider_request'});
  await expect(pie).toHaveAttribute('aria-label',label,{timeout:15000});
  await expect(pie).toHaveAttribute('title',label+' — latest measured provider request');
  await expect(pie).toHaveAttribute('data-tooltip',label+' — latest measured provider request');
  const values=(await arc.getAttribute('stroke-dasharray')).split(' ').map(Number);
  expect(values[0]).toBeCloseTo(circumference*fill/100,8);expect(values[1]).toBeCloseTo(circumference,8);
  await expect(page.getByRole('button',{name:'Send message',exact:true})).toBeVisible();
  return turn;
 };
 const switchTo=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();};
 return{main,child,input,pie,arc,measure,state,switchTo};
}
async function source(info,id){const scenario=loadCorpus().find(x=>x.id===id);await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});}

test('@ux-context-001 Show supplied usage in the context tooltip',async({page,request},info)=>{
 await source(info,'@ux-context-001');const{main,child,input,pie,arc,measure,state,switchTo}=await fixture(page,request,info);
 // Fixed expectations intentionally do not call the production formatter.
 await measure(85000,'Context: 85K / 2.0M tokens (4%)',4.25);
 await measure(1500000,'Context: 1.5M / 2.0M tokens (75%)',75);
 await measure(2500000,'Context: 2.5M / 2.0M tokens (125%)',100);
 await expect(pie).toBeDisabled();await input.fill('preserved over reload');
 await page.reload();await expect(pie).toHaveAttribute('data-tooltip','Context: 2.5M / 2.0M tokens (125%) — latest measured provider request');await expect(input).toHaveValue('preserved over reload');
 await switchTo(child);expect((await state(child)).context_usage.tokens).toBeNull();await expect(pie).toHaveAttribute('aria-label','Context: ? / 2.0M tokens (?%)');await expect(pie).toHaveAttribute('data-tooltip','Context: ? / 2.0M tokens (?%) — usage unavailable');
 await switchTo(main.id);await expect(pie).toHaveAttribute('data-tooltip','Context: 2.5M / 2.0M tokens (125%) — latest measured provider request');await expect(input).toHaveValue('preserved over reload');
 await measure(0,'Context: 0 / 2.0M tokens (0%)',0);
 // Local providers may explicitly report zero input alongside nonzero output.
 expect((await state()).context_usage.tokens).toBe(0);await expect(arc).toHaveAttribute('stroke','var(--context-green, #22c55e)');
});

test('@ux-context-005 Apply the coded usage warning colours',async({page,request},info)=>{
 await source(info,'@ux-context-005');const{measure,arc}=await fixture(page,request,info);
 for(const [tokens,label,fill,color] of [
  [1500000,'Context: 1.5M / 2.0M tokens (75%)',75,'var(--context-green, #22c55e)'],
  [1500200,'Context: 1.5M / 2.0M tokens (75%)',75.01,'var(--context-amber, #f59e0b)'],
  [1800000,'Context: 1.8M / 2.0M tokens (90%)',90,'var(--context-amber, #f59e0b)'],
  [1800200,'Context: 1.8M / 2.0M tokens (90%)',90.01,'var(--context-red, #ef4444)'],
  [2500000,'Context: 2.5M / 2.0M tokens (125%)',100,'var(--context-red, #ef4444)'],
 ]){await measure(tokens,label,fill);await expect(arc).toHaveAttribute('stroke',color);}
});
