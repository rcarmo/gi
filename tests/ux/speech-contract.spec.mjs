import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
import {speechRuntime,fixture} from './support/speech-fixture.mjs';

test.skip(!process.env.GI_UX_SPEECH,'Requires isolated native empty-assistant seed; use make test-ux-speech-contract');
async function evidence(info,kind,id){const s=loadCorpus(kind).find(s=>s.id===id);await info.attach('gherkin',{body:s.steps.join('\n'),contentType:'text/plain'});}
async function emptyGate(context,request){
 const messages=(await(await request.get('/api/sessions/speech-empty-fixture/messages')).json()).messages;
 expect(messages).toHaveLength(2);for(const m of messages){expect(m.role).toBe('assistant');expect(m.content.trim()).toBe('');}
 const page=await context.newPage();try{
  await speechRuntime(page);await page.addInitScript(()=>localStorage.setItem('gi_session_id','speech-empty-fixture'));await page.goto('/');
  for(const m of messages){const post=page.locator(`[id="post-${m.id}"]`);await expect(post).toBeVisible();await expect(post.locator('.post-speak-btn')).toHaveCount(0);}
  expect((await(await request.get('/api/sessions/speech-empty-fixture/messages')).json()).messages).toEqual(messages);
 }finally{await page.close();}
}
async function capability(page,request,context,info){
 await emptyGate(context,request);await speechRuntime(page,false);const f=await fixture(page,request,info);
 for(const m of f.agents)await expect(f.post(m.id)).toBeVisible();await expect(page.locator('.post-speak-btn')).toHaveCount(0);
 await page.evaluate(()=>sessionStorage.setItem('speech-fixture-enabled','true'));await page.reload();
 await expect(page.getByRole('button',{name:'Read aloud',exact:true})).toHaveCount(2);
 for(const m of f.agents)await expect(f.post(m.id).getByRole('button',{name:'Read aloud',exact:true})).toHaveCount(1);
 for(const m of f.messages.filter(m=>m.role==='user'))await expect(f.post(m.id).locator('.post-speak-btn')).toHaveCount(0);
 await expect(f.input).toHaveValue('unsent speech draft β');await expect(page.locator('.compose-box')).toContainText('speech-ref.txt');
 return f;
}

test('@ux-timeline-027 assistant speech controls require native text and supported browser APIs',async({page,request,context},info)=>{
 await evidence(info,'classic','@ux-timeline-027');await capability(page,request,context,info);
});

for(const [kind,id]of [['classic','@ux-original-028'],['shared','@shared-42']])
test(`${id} copy native code and transfer speech without stale owners or draft changes`,async({page,request,context},info)=>{
 await evidence(info,kind,id);const f=await capability(page,request,context,info),first=f.post(f.agents[0].id),second=f.post(f.agents[1].id);
 const code=f.agents[0].content.match(/```js\n([\s\S]*?)```/)[1];expect(code).toContain('"<b>literal</b>"');
 await page.evaluate(()=>{window.__copies=[];document.addEventListener('copy',e=>window.__copies.push({trusted:e.isTrusted,text:e.clipboardData?.getData('text/plain'),html:e.clipboardData?.getData('text/html')}));});
 const copy=first.locator('button.post-code-copy-btn');await copy.click();await expect(copy).toHaveClass(/is-success/);
 expect(await page.evaluate(()=>window.__copies.at(-1))).toEqual({trusted:true,text:code,html:''});
 await first.getByRole('button',{name:'Read aloud',exact:true}).focus();await page.keyboard.press('Enter');await expect(first.getByRole('button',{name:'Stop reading aloud',exact:true})).toHaveAttribute('aria-pressed','true');
 expect(await page.evaluate(()=>window.__speech.utterances[0].text)).toContain('Code block omitted.');
 await second.getByRole('button',{name:'Read aloud',exact:true}).click();await expect(first.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');
 expect(await page.evaluate(()=>window.__speech.calls.map(c=>c[0]))).toEqual(['speak','cancel','speak']);
 await page.evaluate(()=>{window.__speech.utterances[0].onend();window.__speech.utterances[0].onerror();});
 await expect(second.getByRole('button',{name:'Stop reading aloud',exact:true})).toHaveAttribute('aria-pressed','true');
 await copy.click();expect(await page.evaluate(()=>window.__copies.at(-1))).toEqual({trusted:true,text:code,html:''});
 await expect(second.getByRole('button',{name:'Stop reading aloud',exact:true})).toHaveAttribute('aria-pressed','true');
 await second.getByRole('button',{name:'Stop reading aloud',exact:true}).press('Space');await expect(page.locator('.post-speak-btn.is-active')).toHaveCount(0);
 await expect(f.input).toHaveValue('unsent speech draft β');await expect(page.locator('.compose-box')).toContainText('speech-ref.txt');
 expect((await(await request.get(`/api/sessions/${f.main.id}/messages`)).json()).messages).toEqual(f.messages);
 expect((await(await request.get(`/api/sessions/${f.other}/turns`)).json()).turns||[]).toEqual([]);
 await page.reload();await expect(f.input).toHaveValue('unsent speech draft β');await expect(page.locator('.compose-box')).toContainText('speech-ref.txt');await expect(page.locator('.post-speak-btn.is-active')).toHaveCount(0);
});
