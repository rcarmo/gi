import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
test.skip(!process.env.GI_UX_OUTCOMES,'Requires real stale-claim recovery; use make test-ux-outcomes');
const inputName='Message (Enter to send, Shift+Enter for newline)...';
test('@ux-timeline-026 native recovered outcome follows timestamp on the same metadata row',async({page,request},info)=>{
 const s=loadCorpus().find(s=>s.id==='@ux-timeline-026');await info.attach('gherkin',{body:s.steps.join('\n'),contentType:'text/plain'});
 await expect.poll(async()=>((await(await request.get('/api/sessions/outcome-fixture/turns')).json()).turns||[])[0]?.status).toBe('completed');
 const before=(await(await request.get('/api/sessions/outcome-fixture/messages')).json()).messages,reply=before.find(m=>m.role==='assistant');expect(reply).toBeTruthy();expect(reply.payload.content_blocks).toEqual([{type:'recovery_marker',recovered:true,recovery_kind:'native_interrupted_turn',attempts_used:2}]);
 await page.addInitScript(()=>localStorage.setItem('gi_session_id','outcome-fixture'));await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await input.fill('outcome draft β');await page.locator('.compose-box input[type=file]').setInputFiles({name:'outcome.txt',mimeType:'text/plain',buffer:Buffer.from('retained outcome bytes')});
 const post=page.locator(`[id="post-${reply.id}"]`),chip=post.locator('.post-recovery-chip'),time=post.locator('.post-time');
 async function check(){await expect(chip).toHaveText('recovered');await expect(chip).toHaveAttribute('title','Recovered after 2 attempts');expect(await chip.evaluate(el=>el.previousElementSibling?.classList.contains('post-time'))).toBe(true);const a=await time.boundingBox(),b=await chip.boundingBox();expect(Math.abs(a.y+a.height/2-b.y-b.height/2)).toBeLessThanOrEqual(2);expect(b.x).toBeGreaterThanOrEqual(a.x+a.width);}
 await check();await page.reload();await check();await expect(input).toHaveValue('outcome draft β');await expect(page.locator('.compose-box')).toContainText('outcome.txt');
 await page.getByRole('button',{name:'Search',exact:true}).click();const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});await search.fill('recovered outcome');const found=page.waitForResponse(r=>r.url().includes('/search?')&&r.status()===200);await search.press('Enter');await found;await check();await page.getByRole('button',{name:'Close search',exact:true}).click();await expect(input).toHaveValue('outcome draft β');
 // A fresh successful turn has no native recovery event and must not inherit a chip.
 const session=await(await request.post('/api/sessions',{data:{agent_id:`plain-${info.project.name}-${Date.now()}`,title:'Plain'}})).json();await request.post(`/api/sessions/${session.id}/prompt`,{data:{prompt:'plain outcome',model:'test-model'}});await expect.poll(async()=>((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns||[])[0]?.status).toBe('completed');const ordinary=(await(await request.get(`/api/sessions/${session.id}/messages`)).json()).messages.find(m=>m.role==='assistant');expect(ordinary.payload.content_blocks).toBeUndefined();
 expect((await(await request.get('/api/sessions/outcome-fixture/messages')).json()).messages).toEqual(before);await info.attach('native-outcome',{body:await page.screenshot(),contentType:'image/png'});
});
