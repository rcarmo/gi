import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
// Only the OS speech boundary is deterministic. Posts, controls, selection,
// ownership and persisted messages all use Gi's real browser/API paths.
async function speechRuntime(page,available=true){
 await page.addInitScript(available=>{
  available = available || sessionStorage.getItem('speech-fixture-enabled') === 'true';
  const log={calls:[],utterances:[],fail:false};window.__speech=log;
  Object.defineProperty(window,'SpeechSynthesisUtterance',{configurable:true,value:available?class{constructor(text){this.text=text;}}:undefined});
  Object.defineProperty(window,'speechSynthesis',{configurable:true,value:available?{
   speak(u){log.calls.push(['speak',u.text]);log.utterances.push(u);if(log.fail)throw Error('audio denied');},
   cancel(){log.calls.push(['cancel']);log.utterances.at(-1)?.onend?.();},
  }:undefined});
 },available);
}
async function fixture(page,request,info){
 const token=`speech-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const fork=await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:token+'-other',title:token+'-other'}})).json();
 for(const prompt of ['# First native speech\n\nHello **world**.\n\n```js\nconst value = "<b>literal</b>";\n```','Second native speech with [a link](https://example.invalid).']){
  const response=await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt,model:'test-model'}});expect(response.status()).toBe(202);const {turn_id}=await response.json();
  await expect.poll(async()=>((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).find(t=>t.id===turn_id)?.status).toBe('completed');
 }
 const messages=(await(await request.get(`/api/sessions/${main.id}/messages`)).json()).messages,agents=messages.filter(m=>m.role==='assistant');expect(agents).toHaveLength(2);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('unsent speech draft β');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'speech-ref.txt',mimeType:'text/plain',buffer:Buffer.from('unchanged speech attachment')});
 return{main,other:fork.branch.chat_jid.slice(3),messages,agents,input,post:id=>page.locator(`.timeline [id="post-${id}"]`)};
}
async function evidence(info,id){const s=loadCorpus().find(s=>s.id===id);await info.attach('gherkin',{body:s.steps.join('\n'),contentType:'text/plain'});}

test('Gi read-aloud capability is absent without speech APIs and present for native assistant text',async({page,request},info)=>{
 await speechRuntime(page,false);const f=await fixture(page,request,info);
 await expect(page.getByRole('button',{name:'Read aloud',exact:true})).toHaveCount(0);
 await page.evaluate(()=>sessionStorage.setItem('speech-fixture-enabled','true'));
 // Remount with a supported boundary on a fresh page, not a fake Post.
 await page.reload();await expect(page.getByRole('button',{name:'Read aloud',exact:true})).toHaveCount(2);
 for(const m of f.messages.filter(m=>m.role==='user'))await expect(f.post(m.id).locator('.post-speak-btn')).toHaveCount(0);
 await expect(f.input).toHaveValue('unsent speech draft β');await expect(page.locator('.compose-box')).toContainText('speech-ref.txt');
});

test('@ux-timeline-028 native speech transfers ownership and fences old same-post callbacks',async({page,request},info)=>{
 await evidence(info,'@ux-timeline-028');await speechRuntime(page);const f=await fixture(page,request,info),first=f.post(f.agents[0].id),second=f.post(f.agents[1].id);
 await first.getByRole('button',{name:'Read aloud',exact:true}).focus();await page.keyboard.press('Enter');await expect(first.getByRole('button',{name:'Stop reading aloud',exact:true})).toHaveAttribute('aria-pressed','true');
 const text=await page.evaluate(()=>window.__speech.utterances[0].text);expect(text).toContain('Code block omitted.');expect(text).not.toContain('const value');expect(text.length).toBeLessThanOrEqual(1600);
 await second.getByRole('button',{name:'Read aloud',exact:true}).click();await expect(first.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');await expect(second.getByRole('button',{name:'Stop reading aloud',exact:true})).toHaveAttribute('aria-pressed','true');
 expect(await page.evaluate(()=>window.__speech.calls.map(c=>c[0]))).toEqual(['speak','cancel','speak']);
 await page.evaluate(()=>{window.__speech.utterances[0].onend();window.__speech.utterances[0].onerror();});await expect(second.getByRole('button',{name:'Stop reading aloud',exact:true})).toBeVisible();
 await second.getByRole('button',{name:'Stop reading aloud',exact:true}).press('Space');await expect(second.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');
 await second.getByRole('button',{name:'Read aloud',exact:true}).click();await page.evaluate(()=>window.__speech.utterances[1].onend());await expect(second.getByRole('button',{name:'Stop reading aloud',exact:true})).toBeVisible();
 await page.evaluate(()=>window.__speech.utterances[2].onend());await expect(second.getByRole('button',{name:'Read aloud',exact:true})).toBeVisible();
 await expect(f.input).toHaveValue('unsent speech draft β');await expect(page.locator('.compose-box')).toContainText('speech-ref.txt');expect((await(await request.get(`/api/sessions/${f.main.id}/messages`)).json()).messages).toEqual(f.messages);
 await info.attach('speech',{body:await page.screenshot(),contentType:'image/png'});
});

test('Gi speech errors, selection changes and page hide stop only owned playback without changing drafts',async({page,request},info)=>{
 await speechRuntime(page);const f=await fixture(page,request,info),post=f.post(f.agents[0].id);
 await page.evaluate(()=>window.__speech.fail=true);await post.getByRole('button',{name:'Read aloud',exact:true}).click();await expect(post.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');
 await page.evaluate(()=>window.__speech.fail=false);await post.getByRole('button',{name:'Read aloud',exact:true}).click();await expect(post.getByRole('button',{name:'Stop reading aloud',exact:true})).toBeVisible();
 const switchTo=async id=>{await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);};
 await switchTo(f.other);await f.input.fill('other speech draft');await page.evaluate(()=>{for(const u of window.__speech.utterances){u.onend();u.onerror();}});await expect(page.locator('.post-speak-btn.is-active')).toHaveCount(0);await expect(f.input).toHaveValue('other speech draft');
 await switchTo(f.main.id);await expect(f.input).toHaveValue('unsent speech draft β');await expect(page.locator('.compose-box')).toContainText('speech-ref.txt');await post.getByRole('button',{name:'Read aloud',exact:true}).click();
 await page.evaluate(()=>window.dispatchEvent(new Event('pagehide')));await expect(post.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');
 await post.getByRole('button',{name:'Read aloud',exact:true}).click();await page.evaluate(()=>{Object.defineProperty(document,'hidden',{configurable:true,value:true});document.dispatchEvent(new Event('visibilitychange'));delete document.hidden;});await expect(post.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');
 await post.getByRole('button',{name:'Read aloud',exact:true}).click();await page.evaluate(()=>window.__speech.utterances.at(-1).onerror());await expect(post.getByRole('button',{name:'Read aloud',exact:true})).toHaveAttribute('aria-pressed','false');
 expect((await(await request.get(`/api/sessions/${f.main.id}/messages`)).json()).messages).toEqual(f.messages);expect((await(await request.get(`/api/sessions/${f.other}/turns`)).json()).turns||[]).toEqual([]);
});
