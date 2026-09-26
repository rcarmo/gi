import {useLayoutEffect,useRef,useState} from './vendor/preact-htm.js';
import {createLocalNotifications,notificationCapability} from './gi-notifications-state.js';
import {focusWindowBestEffort} from './ui/notification-focus.js';

export function useGiNotifications(chatJid:string,onSelect:(chat:string)=>void){
 const [state,setState]=useState({supported:!notificationCapability(window),enabled:false,permission:'default',notice:'',requesting:false});
 const controller=useRef<any>(null),select=useRef(onSelect);select.current=onSelect;
 useLayoutEffect(()=>{
  const c=createLocalNotifications({win:window,doc:document,notify:setState,focus:()=>focusWindowBestEffort(window),selectChat:(chat:string)=>select.current(chat)});controller.current=c;c.setChat(chatJid);
  return ()=>{controller.current=null;c.dispose();};
 },[]);
 useLayoutEffect(()=>{controller.current?.setChat(chatJid);},[chatJid]);
 return {...state,toggle:()=>controller.current?.toggle(),event:(type:string,data:any)=>controller.current?.event(type,data),dismiss:()=>controller.current?.dismiss()};
}
