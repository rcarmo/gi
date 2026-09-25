import {useEffect,useRef,useState} from './vendor/preact-htm.js';

const storageKey='piclaw_compose_height';
import {composeHeightBounds,clampComposeHeight,readComposeHeight} from './gi-compose-height.js';

// Own only surface size. Content, refs, submission and popup focus remain in
// the supplied component. Cancel/unmount never commits an in-flight drag.
export function useGiComposeSurface(textareaRef:any){
 const manual=useRef<number|null>(null),initialised=useRef(false),stop=useRef<null|((commit:boolean)=>void)>(null);
 const mounted=useRef(true);
 const [,renderHeight]=useState(0);
 const changed=()=>{if(mounted.current)renderHeight(n=>n+1);};
 const ownsSurface=()=>textareaRef.current?.isConnected&&textareaRef.current.getClientRects().length>0&&!document.querySelector('[role="dialog"][aria-modal="true"]');
 if(!initialised.current){initialised.current=true;try{manual.current=readComposeHeight(localStorage.getItem(storageKey));}catch{}}
 const resize=(target?:HTMLTextAreaElement)=>{
  const textarea=target||textareaRef.current;if(!textarea)return;
  const bounds=composeHeightBounds(innerWidth,innerHeight);
  textarea.style.minHeight='';textarea.style.height='auto';
  const next=manual.current==null?Math.max(bounds.min,Math.min(textarea.scrollHeight,bounds.autoMax)):clampComposeHeight(manual.current,innerWidth,innerHeight);
  if(manual.current!=null)textarea.style.minHeight=`${next}px`;
  textarea.style.height=`${next}px`;
  textarea.style.overflowY=textarea.scrollHeight>next?'auto':'hidden';
 };
 const persist=()=>{try{if(manual.current==null)localStorage.removeItem(storageKey);else localStorage.setItem(storageKey,String(manual.current));}catch{}};
 const reset=()=>{if(!ownsSurface())return;stop.current?.(false);manual.current=null;persist();resize();changed();};
 const start=(event:any,touch:boolean)=>{
  const textarea=textareaRef.current,point=touch?event.touches?.[0]:event;
  if(!textarea||!point||!ownsSurface()||event.defaultPrevented||(!touch&&event.button!==0))return;
  event.preventDefault();stop.current?.(false);
  const handle=event.currentTarget,startY=point.clientY,startHeight=textarea.getBoundingClientRect().height,previous=manual.current;
  const cursor=document.body.style.cursor,selection=document.body.style.userSelect;
  handle.classList.add('dragging');document.body.style.cursor='row-resize';document.body.style.userSelect='none';
  const move=(e:any)=>{if(!ownsSurface()){finish(false);return;}const point=touch?e.touches?.[0]:e;if(!point)return;if(touch)e.preventDefault();manual.current=clampComposeHeight(startHeight+startY-point.clientY,innerWidth,innerHeight);resize();};
  const finish=(commit:boolean)=>{
   document.removeEventListener('mousemove',move);document.removeEventListener('mouseup',end);
   document.removeEventListener('touchmove',move);document.removeEventListener('touchend',end);document.removeEventListener('touchcancel',cancel);document.removeEventListener('keydown',key,true);window.removeEventListener('blur',cancel);
   handle.classList.remove('dragging');document.body.style.cursor=cursor;document.body.style.userSelect=selection;
   stop.current=null;if(commit)persist();else manual.current=previous;resize();changed();
  };
  const end=()=>finish(true),cancel=()=>finish(false),key=(e:KeyboardEvent)=>{if(e.key==='Escape'){e.preventDefault();e.stopPropagation();cancel();}};
  stop.current=finish;
  document.addEventListener(touch?'touchmove':'mousemove',move,{passive:false});document.addEventListener(touch?'touchend':'mouseup',end);
  if(touch)document.addEventListener('touchcancel',cancel);
  document.addEventListener('keydown',key,true);window.addEventListener('blur',cancel);
 };
 useEffect(()=>{
  mounted.current=true;
  const onResize=()=>{resize();changed();};window.addEventListener('resize',onResize);resize();
  return ()=>{mounted.current=false;stop.current?.(false);window.removeEventListener('resize',onResize);};
 },[]);
 const onKeyDown=(event:KeyboardEvent)=>{
  if(!ownsSurface()||event.defaultPrevented||event.isComposing||event.altKey||event.ctrlKey||event.metaKey||event.shiftKey||!['ArrowUp','ArrowDown','Home'].includes(event.key))return;
  event.preventDefault();event.stopPropagation();if(event.key==='Home'){reset();return;}
  manual.current=clampComposeHeight((textareaRef.current?.getBoundingClientRect().height||0)+(event.key==='ArrowUp'?10:-10),innerWidth,innerHeight);
  persist();resize();changed();
 };
 const bounds=composeHeightBounds(innerWidth,innerHeight);
 return {min:bounds.min,max:bounds.manualMax,value:clampComposeHeight(textareaRef.current?.getBoundingClientRect().height||manual.current||bounds.min,innerWidth,innerHeight),resize,onMouseDown:(e:any)=>start(e,false),onTouchStart:(e:any)=>start(e,true),onKeyDown,reset};
}
