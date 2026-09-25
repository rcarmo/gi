import {test,expect} from '@playwright/test';
import {journeyEnvironment} from './support/journey-environment.mjs';
async function engine(page){await page.addInitScript(()=>{
 window.__voice=[];
 class Recognition{
  constructor(){this.starts=0;this.stops=0;this.aborts=0;window.__voice.push(this);}
  start(){this.starts++;}stop(){this.stops++;}abort(){this.aborts++;}
 }
 window.SpeechRecognition=Recognition;window.webkitSpeechRecognition=Recognition;
 window.__voiceResult=(text,final=true,index=0)=>{const r=window.__voice[index];r.onresult?.({results:[Object.assign([{transcript:text}],{isFinal:final})],resultIndex:0});};
});}

test('Voice gesture merges interim/final draft, stop settles and manual typing retires callbacks without sending',async({page,request},info)=>{
 const env=await journeyEnvironment(info);await engine(page);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Keep draft Ω');
  const id=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
  await page.locator('.compose-box input[type=file]').setInputFiles({name:'voice-draft.txt',mimeType:'text/plain',buffer:Buffer.from('preserve')});
  const writes=[];await page.route('**/api/**',r=>{if(r.request().method()!=='GET'){writes.push(r.request().url());return r.abort();}return r.continue();});
  expect(await page.evaluate(()=>window.__voice.length)).toBe(0);await page.getByRole('button',{name:'Start voice input',exact:true}).click();
  await expect(page.getByRole('status').filter({hasText:'Allow microphone'})).toBeVisible();await page.evaluate(()=>window.__voice[0].onstart());
  await page.evaluate(()=>window.__voiceResult('hello',false));await expect(input).toHaveValue('Keep draft Ω hello');await page.evaluate(()=>window.__voiceResult('hello world',true));await expect(input).toHaveValue('Keep draft Ω hello world');
  await page.getByRole('button',{name:'Stop voice input',exact:true}).click();expect(await page.evaluate(()=>window.__voice[0].stops)).toBe(1);await page.evaluate(()=>window.__voice[0].onend());await expect(page.getByRole('button',{name:'Start voice input',exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Start voice input',exact:true}).click();await page.evaluate(()=>{window.__voice[1].onstart();window.__lateVoice=window.__voice[1].onresult;});
  await input.fill('Manual edit wins');await page.evaluate(()=>window.__lateVoice({results:[Object.assign([{transcript:'late'}],{isFinal:true})]}));await expect(input).toHaveValue('Manual edit wins');expect(await page.evaluate(()=>window.__voice[1].aborts)).toBe(1);
  await expect(page.locator('.compose-file-pill[title="voice-draft.txt"]')).toBeVisible();expect(writes).toEqual([]);expect((await(await request.get(`${env.origin}/api/sessions/${id}/turns`)).json()).turns||[]).toHaveLength(0);
 }finally{await env.close();}
});

test('Voice permission failure is recoverable and Settings/search/pagehide exclude late recognition',async({page},info)=>{
 const env=await journeyEnvironment(info);await engine(page);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Retain origin');
  await page.getByRole('button',{name:'Start voice input',exact:true}).click();await page.evaluate(()=>window.__voice[0].onerror({error:'not-allowed'}));await expect(page.getByRole('alert').filter({hasText:'permission was denied'})).toBeVisible();await expect(input).toHaveValue('Retain origin');
  for(const action of ['settings','search','pagehide']){
   await page.getByRole('button',{name:'Start voice input',exact:true}).click();await page.evaluate(()=>{const r=window.__voice.at(-1);r.onstart();window.__lateVoice=r.onresult;});
   if(action==='settings'){await page.keyboard.press('Control+,');await expect(page.getByRole('dialog')).toBeVisible();}
   if(action==='search'){await page.getByRole('button',{name:'Search',exact:true}).click();}
   if(action==='pagehide')await page.evaluate(()=>window.dispatchEvent(new Event('pagehide')));
   await expect.poll(()=>page.evaluate(()=>window.__voice.at(-1).aborts)).toBe(1);await page.evaluate(()=>window.__lateVoice({results:[Object.assign([{transcript:'retired'}],{isFinal:true})]}));
   if(action==='settings')await page.keyboard.press('Escape');if(action==='search')await page.getByRole('button',{name:'Close search',exact:true}).click();await expect(input).toHaveValue('Retain origin');
  }
 }finally{await env.close();}
});

test.describe('Touch recognition',()=>{test.use({hasTouch:true});test('Push-to-talk release before permission stops on start and suppresses synthetic click',async({page},info)=>{
 const env=await journeyEnvironment(info);await engine(page);
 try{
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeFocused();const button=page.locator('.voice-input-btn');
  await button.dispatchEvent('pointerdown',{pointerId:7,pointerType:'touch',button:0});await button.dispatchEvent('pointerup',{pointerId:7,pointerType:'touch',button:0});await button.dispatchEvent('click');
  expect(await page.evaluate(()=>window.__voice.length)).toBe(1);await page.evaluate(()=>window.__voice[0].onstart());expect(await page.evaluate(()=>window.__voice[0].stops)).toBe(1);await page.evaluate(()=>window.__voice[0].onend());
  await expect(button).toHaveAttribute('aria-pressed','false');
 }finally{await env.close();}
});});

test('Unsupported recognition stays absent; iOS standalone gives keyboard dictation without native start',async({page},info)=>{
 const env=await journeyEnvironment(info);
 try{
  await page.addInitScript(()=>{delete window.SpeechRecognition;delete window.webkitSpeechRecognition;Object.defineProperty(navigator,'userAgent',{configurable:true,value:'Desktop'});Object.defineProperty(navigator,'maxTouchPoints',{configurable:true,value:0});});
  await page.goto(env.origin);await expect(page.locator('.compose-box textarea')).toBeFocused();await expect(page.locator('.voice-input-btn')).toHaveCount(0);
  await page.addInitScript(()=>{Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone'});Object.defineProperty(navigator,'standalone',{configurable:true,value:true});window.SpeechRecognition=class{constructor(){throw Error('fallback must never instantiate recognition');}};});
  await page.reload();const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();await input.fill('Fallback draft');await page.getByRole('button',{name:'Use keyboard dictation',exact:true}).click();await expect(input).toBeFocused();await expect(page.getByRole('status').filter({hasText:'keyboard dictation'})).toBeVisible();await expect(input).toHaveValue('Fallback draft');
 }finally{await env.close();}
});

test('Session switch retires recognition, persists its origin draft and ignores late results after unmount',async({page,request},info)=>{
 const env=await journeyEnvironment(info);await engine(page);
 try{
  await page.goto(env.origin);const input=page.locator('.compose-box textarea');await expect(input).toBeFocused();const main=await page.evaluate(()=>localStorage.getItem('gi_session_id'));
  const response=await request.post(`${env.origin}/api/sessions`,{data:{agent_id:'voice-test',title:'Voice other'}});const other=(await response.json()).id;await page.reload();await expect(input).toBeFocused();await input.fill('Origin draft');
  await page.getByRole('button',{name:'Start voice input',exact:true}).click();await page.evaluate(()=>{window.__voice[0].onstart();window.__voiceResult('dictation');window.__lateVoice=window.__voice[0].onresult;});await expect(input).toHaveValue('Origin draft dictation');
  const open=()=>page.locator('.compose-session-trigger-top button').click();await open();await page.locator(`[data-session-jid="gi:${other}"] [role=menuitem]`).click();await expect(input).toHaveValue('');expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(other);
  expect(await page.evaluate(()=>window.__voice[0].aborts)).toBe(1);await input.fill('Other draft');await page.evaluate(()=>window.__lateVoice({results:[Object.assign([{transcript:'late retired'}],{isFinal:true})]}));await expect(input).toHaveValue('Other draft');
  await open();await page.locator(`[data-session-jid="gi:${main}"] [role=menuitem]`).click();await expect(input).toHaveValue('Origin draft dictation');await page.reload();await expect(input).toHaveValue('Origin draft dictation');expect((await(await request.get(`${env.origin}/api/sessions/${main}/turns`)).json()).turns||[]).toHaveLength(0);
 }finally{await env.close();}
});
