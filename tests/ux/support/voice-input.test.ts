import {test,expect} from 'bun:test';
import {createVoiceInput,voiceSupport,voiceText,mergeVoice,voiceError} from '../../../web/src/gi-voice-input-state.ts';
function fixture(){
 let current={allowed:true,owner:'a',text:'draft'},applied:any[]=[],notices:any[]=[],instances:any[]=[],timers=new Map<number,any>(),id=0;
 class Recognition{onstart:any;onresult:any;onerror:any;onend:any;starts=0;stops=0;aborts=0;constructor(){instances.push(this);}start(){this.starts++;}stop(){this.stops++;}abort(){this.aborts++;}}
 const c=createVoiceInput({read:()=>current,apply:(text:string,owner:string)=>{if(owner!==current.owner)return false;current.text=text;applied.push(text);return true;},notify:(n:any)=>notices.push(n),timers:{setTimeout:(fn:any,ms:number)=>{timers.set(++id,{fn,ms});return id;},clearTimeout:(id:number)=>timers.delete(id)}});
 return {c,Recognition,current,instances,applied,notices,timers};
}
const result=(text:string,final:boolean)=>Object.assign([{transcript:text}],{isFinal:final});
test('speech support gates secure context, unavailable engines and iOS fallback',()=>{
 const ctor=()=>{};expect(voiceSupport({isSecureContext:false,SpeechRecognition:ctor},{}).mode).toBe('unavailable');
 expect(voiceSupport({isSecureContext:true},{}).mode).toBe('unavailable');expect(voiceSupport({isSecureContext:true,webkitSpeechRecognition:ctor},{}).mode).toBe('native');
 expect(voiceSupport({isSecureContext:true,SpeechRecognition:ctor},{platform:'MacIntel',maxTouchPoints:5,standalone:true}).mode).toBe('fallback');
 expect(voiceSupport({isSecureContext:true},{userAgent:'iPhone'}).mode).toBe('fallback');
});
test('cumulative final/interim replaces hypotheses without duplicate text or implicit send',()=>{
 const f=fixture();expect(f.instances).toHaveLength(0);f.c.start(f.Recognition);const r=f.instances[0];r.onstart();
 r.onresult({resultIndex:0,results:[result('hello',false)]});expect(f.current.text).toBe('draft hello');
 r.onresult({resultIndex:0,results:[result('hello world',true)]});expect(f.current.text).toBe('draft hello world');
 r.onresult({resultIndex:1,results:[result('hello world',true),result('again',false)]});expect(f.current.text).toBe('draft hello world again');
 r.onend();expect(f.c.active()).toBe(false);expect(f.timers.size).toBe(0);
 expect(mergeVoice('draft\n','hello')).toBe('draft\nhello');expect(voiceText([result(' a ',true),result('b',false)])).toBe('a b');
});
test('manual edits, session changes and exclusion reject captured stale callbacks',()=>{
 for(const change of [(s:any)=>s.text='new edit',(s:any)=>s.owner='b',(s:any)=>s.allowed=false]){
  const f=fixture();f.c.start(f.Recognition);const r=f.instances[0],late=r.onresult;change(f.current);f.c.reconcile();late({results:[result('late',true)]});expect(f.applied).toHaveLength(0);expect(r.aborts).toBe(1);expect(f.c.active()).toBe(false);
 }
});
test('early touch release waits for native start, then stops once and rejects retired events',()=>{
 const f=fixture();f.c.start(f.Recognition);const r=f.instances[0];f.c.stop();f.c.stop();expect(r.stops).toBe(0);r.onstart();expect(r.stops).toBe(1);
 const late=r.onresult;f.c.cancel();f.c.start(f.Recognition);late({results:[result('late',true)]});expect(f.applied).toHaveLength(0);expect(f.instances[1].aborts).toBe(0);f.c.dispose();expect(f.timers.size).toBe(0);
});
test('permission/start timeout, stop timeout and errors retain draft and settle once',()=>{
 const f=fixture();f.c.start(f.Recognition);const timer=[...f.timers.values()].find(t=>t.ms===15000);timer.fn();expect(f.c.active()).toBe(false);expect(f.current.text).toBe('draft');expect(f.timers.size).toBe(0);
 f.c.start(f.Recognition);const r=f.instances[1];r.onstart();f.c.stop();[...f.timers.values()].find(t=>t.ms===5000).fn();expect(r.aborts).toBe(1);expect(f.timers.size).toBe(0);
 f.c.start(f.Recognition);f.instances[2].onerror({error:'not-allowed'});expect(f.notices.at(-1).phase).toBe('error');expect(f.c.active()).toBe(false);expect(f.current.text).toBe('draft');expect(voiceError('network')).toContain('service');
});

test('construction, configuration and start errors release ownership and permit retry',()=>{
 const f=fixture();for(const ctor of [class{constructor(){throw Error('construct');}},class{set lang(value){throw Error('config');}abort(){}},class{start(){throw Error('start');}abort(){}}]){
  expect(f.c.start(ctor)).toBe(false);expect(f.c.active()).toBe(false);expect(f.timers.size).toBe(0);
 }
 expect(f.c.start(f.Recognition)).toBe(true);f.c.dispose();expect(f.c.start(f.Recognition)).toBe(false);expect(f.current.text).toBe('draft');
});

test('recognition absolute deadline aborts a browser that never ends',()=>{
 const f=fixture();f.c.start(f.Recognition);f.instances[0].onstart();f.instances[0].onresult({results:[result('kept',true)]});
 [...f.timers.values()].find(t=>t.ms===90000).fn();expect(f.c.active()).toBe(false);expect(f.instances[0].aborts).toBe(1);expect(f.current.text).toBe('draft kept');expect(f.timers.size).toBe(0);
});
