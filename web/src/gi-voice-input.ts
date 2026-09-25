import {html,useLayoutEffect,useRef,useState} from './vendor/preact-htm.js';
import {createVoiceInput,voiceSupport} from './gi-voice-input-state.js';

export function useGiVoiceInput(readProps:()=>any,textareaRef:any){
 const latest=useRef(readProps);latest.current=readProps;
 const [state,setState]=useState({phase:'idle',message:''});
 const active=['starting','listening','stopping'].includes(state.phase);
 const support=voiceSupport(window,navigator),controller=useRef<any>(null),pointer=useRef<number|null>(null),suppressClick=useRef(0);
 const read=()=>{
  const props=latest.current(),textarea=textareaRef.current;
  return {owner:props.owner,text:textarea?.value??props.content,
   allowed:!props.searchMode&&!props.disabled&&document.visibilityState!=='hidden'&&textarea?.isConnected&&textarea.getClientRects().length>0&&getComputedStyle(textarea).visibility==='visible'&&!document.querySelector('[role="dialog"][aria-modal="true"]')};
 };
 if(!controller.current)controller.current=createVoiceInput({read,language:navigator.language||'en-US',notify:setState,apply:(text:string,owner:string)=>{
  const props=latest.current();if(read().owner!==owner||!read().allowed)return false;
  props.setContent(text);textareaRef.current.value=text;props.resize(textareaRef.current);return true;
 }});
 const voice=controller.current;
 const cancel=()=>{pointer.current=null;voice.cancel();};
 const props=readProps();
 useLayoutEffect(()=>{voice.reconcile();},[props.owner,props.searchMode,props.disabled,props.content]);
 useLayoutEffect(()=>()=>voice.dispose(),[]);
 useLayoutEffect(()=>{
  if(!active)return;
  const hide=()=>{if(document.visibilityState==='hidden')cancel();},exit=()=>cancel(),resize=()=>voice.reconcile();
  const observer=new MutationObserver(()=>voice.reconcile());observer.observe(document.body,{childList:true,subtree:true,attributes:true,attributeFilter:['hidden','inert','aria-modal','style','class']});
  const releaseOutside=(event:PointerEvent)=>{if(pointer.current===event.pointerId){pointer.current=null;suppressClick.current=Date.now()+800;if(event.type==='pointercancel')voice.cancel();else voice.stop();}};
  document.addEventListener('visibilitychange',hide);window.addEventListener('pagehide',exit);window.addEventListener('resize',resize);document.addEventListener('pointerup',releaseOutside);document.addEventListener('pointercancel',releaseOutside);
  return ()=>{observer.disconnect();document.removeEventListener('visibilitychange',hide);window.removeEventListener('pagehide',exit);window.removeEventListener('resize',resize);document.removeEventListener('pointerup',releaseOutside);document.removeEventListener('pointercancel',releaseOutside);};
 },[active]);
 const toggle=()=>{
  if(!read().allowed)return;
  if(voice.active()){voice.stop();return;}
  if(support.mode==='fallback'){setState({phase:'idle',message:support.detail});textareaRef.current?.focus();return;}
  if(support.mode==='native'){latest.current().onStart();voice.start(support.ctor);}
 };
 const release=(event:any)=>{
  if(pointer.current==null||event.pointerId!==pointer.current)return;
  pointer.current=null;suppressClick.current=Date.now()+800;
  if(event.type==='pointercancel')voice.cancel();else voice.stop();
 };
 const title=active?'Stop voice input':support.mode==='fallback'?'Use keyboard dictation':'Start voice input';
 const button=!props.searchMode&&!props.disabled&&support.mode!=='unavailable'?html`
  <button class=${`icon-btn voice-input-btn${active?' active':''}${support.mode==='fallback'?' fallback':''}`} type="button"
   title=${active?title:support.mode==='native'?`${title}. Browser recognition may use an external service.`:title} aria-label=${title} aria-pressed=${active?'true':'false'}
   onClick=${()=>{if(Date.now()<suppressClick.current){suppressClick.current=0;return;}toggle();}}
   onPointerDown=${(event:any)=>{if(event.pointerType==='mouse'||event.button!==0||support.mode!=='native')return;event.preventDefault();pointer.current=event.pointerId;suppressClick.current=Date.now()+800;toggle();}}
   onPointerUp=${release} onPointerCancel=${release} onPointerLeave=${release}>
   <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <path d="M12 3a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V6a3 3 0 0 0-3-3z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><path d="M12 19v3"/>
   </svg>
  </button>`:null;
 const status=!props.searchMode&&state.message?html`<div class=${`compose-inline-status compose-speech-status ${state.phase==='error'?'compose-speech-status-error':''}`} role=${state.phase==='error'?'alert':'status'} aria-live="polite">
  <div class="compose-inline-status-row"><span class="compose-inline-status-title">${state.message}</span><button type="button" class="gi-voice-dismiss" aria-label=${active?'Cancel voice input':'Dismiss voice input status'} onClick=${()=>{voice.cancel('',false);setState({phase:'idle',message:''});}}>×</button></div>
 </div>`:null;
 return {button,status,cancel,beforeKey:(event:KeyboardEvent)=>{if(event.key==='Escape'&&voice.active()){event.preventDefault();event.stopPropagation();voice.cancel();return true;}return false;}};
}
