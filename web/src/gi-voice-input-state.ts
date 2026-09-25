// Browser speech-recognition contract derived from pinned Classic 0afe5366ced9.
// No server endpoint, audio recording, permission preflight or automatic send.
export function voiceSupport(win:any,nav:any){
 const ios=/iPad|iPhone/.test(String(nav?.userAgent||''))||(nav?.platform==='MacIntel'&&Number(nav?.maxTouchPoints)>1);
 const standalone=Boolean(nav?.standalone||win?.matchMedia?.('(display-mode: standalone)').matches);
 const ctor=win?.SpeechRecognition||win?.webkitSpeechRecognition;
 if(!win?.isSecureContext)return {mode:'unavailable',ctor:null,detail:'Voice input requires HTTPS or localhost.'};
 if(ios&&(standalone||!ctor))return {mode:'fallback',ctor:null,detail:'Focus the message box and use the keyboard dictation microphone. In-page recognition is unavailable or unreliable in this iPhone/iPad mode.'};
 if(!ctor)return {mode:'unavailable',ctor:null,detail:'This browser does not expose speech recognition.'};
 return {mode:'native',ctor,detail:'Browser speech recognition may use an external service. Start only when you want to dictate.'};
}
export function voiceText(results:any){
 let final='',interim='';
 for(let i=0;i<Number(results?.length||0);i++){
  const result=results[i],text=String(result?.[0]?.transcript||'').trim();if(!text)continue;
  if(result.isFinal)final=[final,text].filter(Boolean).join(' ');else interim=[interim,text].filter(Boolean).join(' ');
 }
 return [final,interim].filter(Boolean).join(' ');
}
export function mergeVoice(base:string,speech:string){return speech?base+((base&&!/\s$/.test(base))?' ':'')+speech:base;}
export function voiceError(code:string){
 return ({'not-allowed':'Microphone or speech-recognition permission was denied.','service-not-allowed':'Microphone or speech-recognition permission was denied.','no-speech':'No speech was detected. Try again after the listening indicator appears.','audio-capture':'The browser could not access a microphone.','network':'The browser speech-recognition service reported a network or service failure.','aborted':'Voice input was stopped.'} as Record<string,string>)[code]||'Voice input failed.';
}

export function createVoiceInput({read,apply,notify,language='en-US',timers=globalThis}:any){
 let run:any=null,disposed=false;
 const clear=(r:any)=>{timers.clearTimeout(r.startTimer);timers.clearTimeout(r.stopTimer);timers.clearTimeout(r.limitTimer);};
 const detach=(r:any)=>{for(const key of ['onstart','onresult','onerror','onend'])r.recognition[key]=null;};
 const cancel=(message='',publish=true)=>{
  const r=run;run=null;if(r){clear(r);detach(r);try{r.recognition.abort();}catch{}}
  if(publish&&!disposed)notify({phase:'idle',message});
 };
 const valid=(r:any)=>{const state=read();return !disposed&&r===run&&state.allowed&&state.owner===r.owner&&state.text===r.last;};
 const owns=(r:any)=>{if(valid(r))return true;if(run===r)cancel();return false;};
 const stop=()=>{
  const r=run;if(!r||r.stopping)return;if(!owns(r))return;
  r.stopping=true;notify({phase:'stopping',message:'Finishing voice input…'});
  // stop() before onstart is unreliable in WebKit; remember the release and
  // stop immediately when the browser confirms start instead of capturing on.
  r.stopTimer=timers.setTimeout(()=>{if(run===r)cancel('Voice input stopped; browser completion timed out.');},5000);
  if(r.started)try{r.recognition.stop();}catch{cancel('Voice input could not finish. Draft retained.');}
 };
 return {
  start(ctor:any){
   if(disposed||run||!ctor)return false;const state=read();if(!state.allowed)return false;
   let recognition;try{recognition=new ctor();}catch{notify({phase:'error',message:'Voice input could not start.'});return false;}
   try{recognition.lang=language;recognition.continuous=false;recognition.interimResults=true;if('maxAlternatives' in recognition)recognition.maxAlternatives=1;}
   catch{try{recognition.abort();}catch{}notify({phase:'error',message:'Voice input could not be configured.'});return false;}
   const r:any={recognition,owner:state.owner,base:state.text,last:state.text,started:false,stopping:false};run=r;
   recognition.onstart=()=>{if(!owns(r))return;r.started=true;timers.clearTimeout(r.startTimer);if(r.stopping){try{recognition.stop();}catch{cancel('Voice input could not finish. Draft retained.');}}else notify({phase:'listening',message:'Listening… Speak now.'});};
   recognition.onresult=(event:any)=>{
    if(!owns(r))return;
    // Web Speech results are cumulative: rebuild, never append a changed
    // interim/resultIndex to text that already contains an earlier hypothesis.
    const next=mergeVoice(r.base,voiceText(event.results));r.last=next;
    if(!apply(next,r.owner)){cancel();return;}
   };
   recognition.onerror=(event:any)=>{if(!owns(r))return;const message=voiceError(String(event.error||''));cancel('',false);notify({phase:'error',message});};
   recognition.onend=()=>{if(!owns(r))return;run=null;clear(r);detach(r);notify({phase:'idle',message:r.last===r.base?'Voice input ended without a transcript.':'Voice input added to draft.'});};
   notify({phase:'starting',message:'Allow microphone or speech recognition in the browser prompt.'});
   r.startTimer=timers.setTimeout(()=>{if(run===r&&!r.started)cancel('Voice input timed out waiting for browser permission.');},15000);
   r.limitTimer=timers.setTimeout(()=>{if(run===r)cancel('Voice input reached its time limit. Draft retained.');},90000);
   try{recognition.start();return true;}catch{cancel('',false);notify({phase:'error',message:'Voice input could not start. Check browser permission and try again.'});return false;}
  },stop,cancel,
  reconcile(){if(run&&!valid(run))cancel();},
  active:()=>Boolean(run),
  dispose(){disposed=true;cancel('',false);},
 };
}
