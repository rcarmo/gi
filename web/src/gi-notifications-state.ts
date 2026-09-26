import {createLocalNotificationPresenceSnapshot,publishLocalNotificationPresence,withdrawLocalNotificationPresence,shouldNotifyLocallyForChat,listLiveLocalNotificationPresence} from './ui/notification-delivery-coordinator.ts';
export const NOTIFICATION_PREFERENCE='gi_notifications_v1';
export const NOTIFICATION_SEEN='gi_notifications_seen_v1';
const LOCK='gi-notification-delivery',MAX_AGE=120000,MAX_SEEN=256;
// Settings waits for registered delivery barriers before sending logout. The
// local controller retires immediately; acquiring the same browser lock also
// drains an already-entered sibling callback before browser authority changes.
export async function cleanupLocalNotifications(win:any){
 const pending:Promise<unknown>[]=[];
 win.dispatchEvent(new win.CustomEvent('gi-notification-cleanup',{detail:{waitUntil:(promise:Promise<unknown>)=>pending.push(Promise.resolve(promise))}}));
 await Promise.all(pending);
}
export function notificationCandidate(type:string,data:any,chat:string,now:number,since:number){
 if(type!=='new_post'||data?.chat_jid!==chat||!data?.id||data?.is_bot_message!==true||data?.is_from_me===true||data?.sender==='system'||data?.data?.type!=='agent_response')return null;
 const timestamp=Date.parse(data.timestamp);if(!Number.isFinite(timestamp)||timestamp<since||timestamp<now-MAX_AGE||timestamp>now+5000)return null;
 const payload=data.data;
 // Recovery, compaction and tool/control posts do not represent a normal
 // assistant completion. Local notification text contains no reply content.
 if(payload.content_blocks?.length||data.content_blocks?.length||payload.kind||data.kind||payload.suppress_notification||data.suppress_notification)return null;
 const text=typeof payload.content==='string'?payload.content:'';if(!text.trim())return null;
 return {key:JSON.stringify([chat,String(data.id)]),timestamp};
}
export function notificationCapability(win:any){
 if(!win?.isSecureContext)return 'Local notifications require HTTPS or localhost.';
 if(typeof win.Notification!=='function')return 'This browser has no local notification API.';
 if(!win.navigator?.locks?.request)return 'This browser lacks the locking required for duplicate-safe local notifications.';
 return '';
}
export function createLocalNotifications({win,doc,now=()=>Date.now(),notify,selectChat,focus,timers=globalThis}:any){
 let disposed=false,paused=false,generation=0,chat='',enabled=false,permission='default',notice='',requesting=false,interval:any;
 let started=now(),device='',client='',permissionCancel:null|(()=>void)=null;const open=new Set<any>(),pendingLocks=new Set<AbortController>();
 const retire=()=>{generation++;requesting=false;permissionCancel?.();permissionCancel=null;for(const controller of pendingLocks)controller.abort();pendingLocks.clear();};
 const capability=notificationCapability(win);
 const storage=()=>win.localStorage;
 const report=()=>{if(!disposed)notify({supported:!capability,enabled,permission,notice,requesting});};
 const close=()=>{for(const n of open){n.onclick=null;n.onclose=null;try{n.close();}catch{}}open.clear();};
 const withdraw=()=>{try{if(device&&client)withdrawLocalNotificationPresence({deviceId:device,clientId:client},win);}catch{/* Revoked storage must not interrupt retirement or logout. */}};
 const fail=(message:string)=>{enabled=false;notice=message;close();try{storage().setItem(NOTIFICATION_PREFERENCE,'false');}catch{}report();};
 const publish=()=>{
  if(disposed||paused||capability||!chat)return false;
  try{
   // Concurrent first tabs may create different device IDs before the shared
   // write settles. Adopt the stored identity before every presence/delivery.
   const storedDevice=storage().getItem('piclaw.notifications.deviceId');
   if(!storedDevice)throw Error('device identity missing');
   if(storedDevice!==device){withdraw();device=storedDevice;}
   const snapshot=createLocalNotificationPresenceSnapshot({runtimeWindow:win,runtimeDocument:doc,deviceId:device,clientId:client,chatJid:chat,updatedAtMs:now()});
   publishLocalNotificationPresence(snapshot,win);
   // The supplied publisher is best-effort; explicitly verify storage so a
   // blocked store can never turn coordination into duplicate delivery.
   const saved=JSON.parse(storage().getItem(`piclaw.notifications.presence.${device}:${client}`)||'null');
   if(saved?.chatJid!==chat||saved?.updatedAtMs!==snapshot.updatedAtMs)throw Error('presence write');
   return true;
  }catch{fail('Local notifications unavailable: browser storage could not retain coordination state.');return false;}
 };
 const readPreference=()=>{try{const value=storage().getItem(NOTIFICATION_PREFERENCE);if(value!==null&&value!=='true'&&value!=='false')throw Error('invalid preference');return value==='true';}catch{fail('Local notifications unavailable: browser storage could not be read.');return false;}};
 const sync=()=>{if(disposed)return;permission=win.Notification?.permission||'default';if(paused){enabled=false;report();return;}enabled=!capability&&permission==='granted'&&readPreference();if(!enabled)close();publish();report();};
 if(!capability){
  try{
   device=storage().getItem('piclaw.notifications.deviceId')||`device-${win.crypto.randomUUID()}`;
   storage().setItem('piclaw.notifications.deviceId',device);device=storage().getItem('piclaw.notifications.deviceId');
   // Per-document identity, never copied sessionStorage from a duplicated tab.
   client=`client-${win.crypto.randomUUID()}`;
  }catch{notice='Local notifications unavailable: browser storage or identity creation failed.';}
 }
 const available=()=>!capability&&!!device&&!!client;
 const stateChange=()=>sync(),hide=()=>{paused=true;retire();withdraw();close();report();},show=()=>{paused=false;started=now();sync();};
 const storageChange=(event:any)=>{if(event.key==='piclaw.notifications.deviceId'){publish();return;}if(event.key===NOTIFICATION_PREFERENCE||event.key===null){retire();sync();}};
 const cleanup=async()=>{
  retire();enabled=false;paused=true;close();withdraw();try{storage().setItem(NOTIFICATION_PREFERENCE,'false');}catch{}report();
  if(!available())return;
  const controller=new AbortController(),timeout=timers.setTimeout(()=>controller.abort(),5000);
  try{await win.navigator.locks.request(LOCK,{signal:controller.signal},()=>{});}finally{timers.clearTimeout(timeout);}
 };
 const cleanupEvent=(event:any)=>{const barrier=cleanup();if(typeof event.detail?.waitUntil==='function')event.detail.waitUntil(barrier);else void barrier.catch(()=>{});};
 if(available()){
  permission=win.Notification.permission;enabled=permission==='granted'&&readPreference();
  interval=timers.setInterval(()=>publish(),15000);
  doc.addEventListener('visibilitychange',stateChange);win.addEventListener('focus',stateChange);win.addEventListener('pageshow',show);win.addEventListener('pagehide',hide);win.addEventListener('storage',storageChange);win.addEventListener('gi-notification-cleanup',cleanupEvent);
 }
 report();
 return {
  setChat(value:string){if(chat===value)return;retire();withdraw();close();chat=value;started=now();publish();report();},
  async toggle(){
   if(disposed||requesting)return;
   // Explicit opt-in may recover after a failed logout; no automatic resume.
   if(paused){paused=false;started=now();}
   if(!available()){notice=capability||notice||'Local notifications unavailable.';report();return;}
   if(win.Notification.permission==='denied'){permission='denied';fail('Notifications are blocked. Change browser permission to enable them.');return;}
   const version=++generation;requesting=true;notice='';report();
   try{
    let nextPermission=win.Notification.permission;
    if(nextPermission==='default')nextPermission=await new Promise<string>((resolve,reject)=>{
     let settled=false;
     const finish=(value:string,error?:unknown)=>{if(settled)return;settled=true;timers.clearTimeout(timeout);permissionCancel=null;if(error)reject(error);else resolve(value);};
     const timeout=timers.setTimeout(()=>finish('default',Error('permission timeout')),15000);
     permissionCancel=()=>finish('default');
     // Invoke synchronously inside the user-gesture call stack.
     try{Promise.resolve(win.Notification.requestPermission()).then(value=>finish(value||'default'),error=>finish('default',error));}
     catch(error){finish('default',error);}
    });
    if(disposed||paused||version!==generation)return;
    permission=nextPermission||'default';const next=permission==='granted'&&!enabled;
    storage().setItem(NOTIFICATION_PREFERENCE,String(next));if(storage().getItem(NOTIFICATION_PREFERENCE)!==String(next))throw Error('preference not retained');
    enabled=next;started=now();if(!enabled)close();if(!publish())return;
    notice=next?'Local notifications enabled for this browser. Matching chat tabs must remain open; Web Push and closed-tab delivery are not available.':permission==='denied'?'Notifications were denied. Change browser permission to retry.':permission==='default'?'Permission was not granted. Local notifications remain off.':'Local notifications disabled.';
   }catch{if(!disposed&&version===generation)fail('Local notification permission or storage failed. Retry explicitly; no delivery was enabled.');}
   finally{if(!disposed&&version===generation){requesting=false;report();}}
  },
  async event(type:string,data:any){
   if(disposed||paused||!available()||!enabled||win.Notification.permission!=='granted')return false;
   const candidate=notificationCandidate(type,data,chat,now(),started);if(!candidate)return false;
   const version=generation,target=chat,controller=new AbortController();pendingLocks.add(controller);
   const timeout=timers.setTimeout(()=>controller.abort(),5000);
   try{return await win.navigator.locks.request(LOCK,{signal:controller.signal},async()=>{
    if(disposed||paused||version!==generation||target!==chat||win.Notification.permission!=='granted'||!readPreference()||!notificationCandidate(type,data,chat,now(),started)||!publish())return false;
    const visible=doc.visibilityState!=='hidden'||listLiveLocalNotificationPresence({runtimeWindow:win,deviceId:device,nowMs:now()}).some((entry:any)=>entry.chatJid===chat&&entry.visibilityState==='visible');
    if(!visible&&!shouldNotifyLocallyForChat({runtimeWindow:win,runtimeDocument:doc,chatJid:chat,deviceId:device,clientId:client,updatedAtMs:now()}))return false;
    const raw=storage().getItem(NOTIFICATION_SEEN);if(raw&&raw.length>300000)throw Error('oversized dedupe state');const parsed=raw===null?[]:JSON.parse(raw);
    if(!Array.isArray(parsed)||parsed.some(row=>!row||typeof row.key!=='string'||!Number.isFinite(row.timestamp)))throw Error('invalid dedupe state');
    const live=parsed.filter(row=>row.timestamp>=now()-MAX_AGE&&row.timestamp<=now()+5000);
    if(live.some(row=>row.key===candidate.key)||live.length>=MAX_SEEN)return false;
    // Claim under the lock before invoking the browser API. A failed native
    // constructor is not replayed by another tab; the user gets an error.
    live.push(candidate);const serialised=JSON.stringify(live);storage().setItem(NOTIFICATION_SEEN,serialised);if(storage().getItem(NOTIFICATION_SEEN)!==serialised)throw Error('dedupe write');
    if(visible||!readPreference()||version!==generation)return false; // Suppressed visible replies cannot replay later.
    const n=new win.Notification('Gi',{body:'An assistant reply is ready.',tag:`gi-${String(data.id)}`});open.add(n);
    n.onclick=()=>{if(disposed||paused||version!==generation)return;focus();selectChat(target);try{n.close();}catch{}open.delete(n);};n.onclose=()=>open.delete(n);
    if(open.size>8){const oldest=open.values().next().value;try{oldest.onclick=null;oldest.close();}catch{}open.delete(oldest);}
    return true;
   });}catch{if(!disposed&&version===generation)fail('Local notification delivery failed. The browser may require Web Push, which Gi does not provide.');return false;}
   finally{timers.clearTimeout(timeout);pendingLocks.delete(controller);}
  },
  dismiss(){notice='';report();},cleanup,
  dispose(){if(disposed)return;disposed=true;retire();timers.clearInterval(interval);withdraw();close();doc.removeEventListener('visibilitychange',stateChange);win.removeEventListener('focus',stateChange);win.removeEventListener('pageshow',show);win.removeEventListener('pagehide',hide);win.removeEventListener('storage',storageChange);win.removeEventListener('gi-notification-cleanup',cleanupEvent);},
 };
}
