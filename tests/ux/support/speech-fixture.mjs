import {expect} from '@playwright/test';
import {loadCorpus} from './catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
// Only the OS speech boundary is deterministic. Posts, controls, selection,
// ownership and persisted messages all use Gi's real browser/API paths.
export async function speechRuntime(page,available=true){
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
export async function fixture(page,request,info){
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
export async function evidence(info,id){const s=loadCorpus().find(s=>s.id===id);await info.attach('gherkin',{body:s.steps.join('\n'),contentType:'text/plain'});}
