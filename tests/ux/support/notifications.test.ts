import {test,expect} from 'bun:test';
import {createLocalNotifications,notificationCandidate,notificationCapability,NOTIFICATION_PREFERENCE,NOTIFICATION_SEEN} from '../../../web/src/gi-notifications-state.ts';
function shared(){const data=new Map<string,string>();let chain=Promise.resolve();return {data,store:{get length(){return data.size;},key:(i:number)=>[...data.keys()][i],getItem:(k:string)=>data.get(k)??null,setItem:(k:string,v:string)=>data.set(k,v),removeItem:(k:string)=>data.delete(k)},locks:{request:(_:string,options:any,callback?:any)=>{const fn=callback||options;const next=chain.then(fn);chain=next.catch(()=>{});return next;}}};}
function fixture(s=shared(),id='a',timerOverride:any={}){
 let clock=10000,requests=0;const notices:any[]=[],deliveries:any[]=[],selected:any[]=[];const win:any=new EventTarget(),doc:any=new EventTarget();doc.visibilityState='hidden';doc.hasFocus=()=>false;
 class Notification {static permission='granted';static requestPermission=async()=>{requests++;return 'granted';};onclick:any;onclose:any;closed=false;constructor(public title:string,public options:any){deliveries.push(this);}close(){this.closed=true;}}
 Object.assign(win,{Notification,isSecureContext:true,localStorage:s.store,crypto:{randomUUID:()=>id},navigator:{locks:s.locks}});
 const c=createLocalNotifications({win,doc,now:()=>clock,notify:(n:any)=>notices.push(n),selectChat:(chat:string)=>selected.push(chat),focus:()=>{},timers:{setInterval:()=>1,clearInterval:()=>{},setTimeout,clearTimeout,...timerOverride}});c.setChat('gi:a');
 return {s,c,win,doc,notices,deliveries,selected,advance:(n:number)=>clock+=n,requests:()=>requests,event:(id='reply')=>({id,chat_jid:'gi:a',timestamp:new Date(clock).toISOString(),sender:'agent',is_bot_message:true,is_from_me:false,data:{type:'agent_response',content:'private text'}})};
}
test('fresh assistant-only candidate rejects history, wrong chats and control posts',()=>{
 const f=fixture(),e=f.event();expect(notificationCandidate('new_post',e,'gi:a',10000,10000)).not.toBeNull();
 for(const bad of [{...e,is_from_me:true},{...e,sender:'system'},{...e,chat_jid:'gi:b'},{...e,timestamp:'bad'},{...e,timestamp:new Date(0).toISOString()},{...e,data:{...e.data,content_blocks:[{type:'recovery'}]}},{...e,data:{...e.data,kind:'compaction'}},{...e,data:{...e.data,content:' '}}])expect(notificationCandidate('new_post',bad,'gi:a',10000,10000)).toBeNull();
 expect(notificationCandidate('agent_response',e,'gi:a',10000,10000)).toBeNull();f.c.dispose();
});
test('visible candidate suppresses, hidden lexicographic leader claims once across concurrent callbacks',async()=>{
 const s=shared();s.store.setItem(NOTIFICATION_PREFERENCE,'true');const b=fixture(s,'b'),a=fixture(s,'a');a.doc.visibilityState='visible';a.doc.dispatchEvent(new Event('visibilitychange'));
 expect(await b.c.event('new_post',b.event())).toBe(false);a.doc.visibilityState='hidden';a.doc.dispatchEvent(new Event('visibilitychange'));
 await Promise.all([a.c.event('new_post',a.event('hidden')),b.c.event('new_post',b.event('hidden')),a.c.event('new_post',a.event('hidden'))]);expect(a.deliveries).toHaveLength(1);expect(b.deliveries).toHaveLength(0);expect(a.deliveries[0].options.body).toBe('An assistant reply is ready.');expect(JSON.stringify([...s.data.values()])).not.toContain('private text');
 a.deliveries[0].onclick();expect(a.selected).toEqual(['gi:a']);a.c.dispose();expect(await b.c.event('new_post',b.event('next'))).toBe(true);b.c.dispose();expect([...s.data.keys()].filter(k=>k.includes('.presence.'))).toHaveLength(0);
});
test('permission is gesture-only, pending grant cannot survive logout, pagehide or session change',async()=>{
 for(const retire of [(f:any)=>f.c.cleanup(),(f:any)=>f.win.dispatchEvent(new Event('pagehide')),(f:any)=>f.c.setChat('gi:b')]){
  const f=fixture();f.win.Notification.permission='default';let release:any;f.win.Notification.requestPermission=()=>new Promise(r=>release=r);expect(f.requests()).toBe(0);const p=f.c.toggle();retire(f);release('granted');await p;expect(f.s.store.getItem(NOTIFICATION_PREFERENCE)).not.toBe('true');expect(f.notices.at(-1).requesting).toBe(false);f.c.dispose();
 }
});
test('queued delivery retires on session/logout, disable closes existing notifications',async()=>{
 const f=fixture();await f.c.toggle();await f.c.event('new_post',f.event());expect(f.deliveries).toHaveLength(1);await f.c.toggle();expect(f.deliveries[0].closed).toBe(true);expect(await f.c.event('new_post',f.event('off'))).toBe(false);
 await f.c.toggle();const p=f.c.event('new_post',f.event('queued'));f.c.cleanup();expect(await p).toBe(false);expect(f.s.store.getItem(NOTIFICATION_PREFERENCE)).toBe('false');f.c.dispose();
});
test('storage denial, native constructor failure and ledger saturation fail closed',async()=>{
 const f=fixture();await f.c.toggle();f.s.store.setItem(NOTIFICATION_SEEN,'bad');expect(await f.c.event('new_post',f.event())).toBe(false);expect(f.notices.at(-1).notice).toContain('failed');f.c.dispose();
 const g=fixture();await g.c.toggle();g.s.store.setItem(NOTIFICATION_SEEN,JSON.stringify(Array.from({length:256},(_,i)=>({key:String(i),timestamp:10000}))));expect(await g.c.event('new_post',g.event())).toBe(false);expect(g.deliveries).toHaveLength(0);g.c.dispose();
 const h=fixture();h.s.store.setItem=()=>{throw Error('denied');};await h.c.toggle();expect(await h.c.event('new_post',h.event())).toBe(false);expect(h.notices.at(-1).enabled).toBe(false);h.c.dispose();
 const j=fixture();await j.c.toggle();j.win.Notification=class{static permission='granted';constructor(){throw Error('constructor');}};expect(await j.c.event('new_post',j.event())).toBe(false);expect(j.notices.at(-1).notice).toContain('Web Push');j.c.dispose();
});
test('secure context, Notification and Web Locks are required',()=>{
 expect(notificationCapability({isSecureContext:false})).toContain('HTTPS');expect(notificationCapability({isSecureContext:true})).toContain('API');expect(notificationCapability({isSecureContext:true,Notification:()=>{}})).toContain('locking');
});

test('visible suppression is durable across later hidden replay; concurrent identity changes converge',async()=>{
 const s=shared();s.store.setItem(NOTIFICATION_PREFERENCE,'true');const a=fixture(s,'a');a.doc.visibilityState='visible';a.doc.dispatchEvent(new Event('visibilitychange'));expect(await a.c.event('new_post',a.event())).toBe(false);
 a.doc.visibilityState='hidden';a.doc.dispatchEvent(new Event('visibilitychange'));expect(await a.c.event('new_post',a.event())).toBe(false);expect(a.deliveries).toHaveLength(0);
 s.store.setItem('piclaw.notifications.deviceId','device-replaced');const b=fixture(s,'b');b.doc.visibilityState='visible';b.doc.dispatchEvent(new Event('visibilitychange'));expect(await a.c.event('new_post',a.event('identity-race'))).toBe(false);expect(a.deliveries).toHaveLength(0);a.c.dispose();b.c.dispose();
});

test('permission wait times out without enabling; late grants cannot clear a later pending request',async()=>{
 const pending=new Map<number,any>();let id=0;const f=fixture(shared(),'a',{setTimeout:(fn:any,ms:number)=>{pending.set(++id,{fn,ms});return id;},clearTimeout:(id:number)=>pending.delete(id)});
 f.win.Notification.permission='default';const grants:any[]=[];f.win.Notification.requestPermission=()=>new Promise(r=>grants.push(r));
 const first=f.c.toggle();[...pending.values()].find(t=>t.ms===15000).fn();await first;expect(f.notices.at(-1).requesting).toBe(false);expect(f.s.store.getItem(NOTIFICATION_PREFERENCE)).toBe('false');
 const second=f.c.toggle();grants[0]('granted');await Promise.resolve();expect(f.notices.at(-1).requesting).toBe(true);await f.c.cleanup();grants[1]('granted');await second;expect(f.s.store.getItem(NOTIFICATION_PREFERENCE)).toBe('false');f.c.dispose();
});
test('logout waits for an entered sibling delivery lock before reporting cleanup complete',async()=>{
 const s=shared(),f=fixture(s);await f.c.toggle();let release:any;const blocker=s.locks.request('gi-notification-delivery',()=>new Promise(r=>release=r));await Promise.resolve();
 let settled=false;const barrier=f.c.cleanup().then(()=>settled=true);await Promise.resolve();expect(settled).toBe(false);expect(s.store.getItem(NOTIFICATION_PREFERENCE)).toBe('false');release();await blocker;await barrier;expect(settled).toBe(true);f.c.dispose();
});

test('revoked storage getter cannot throw from disposal or block logout cleanup',async()=>{
 const f=fixture();await f.c.toggle();await f.c.event('new_post',f.event());Object.defineProperty(f.win,'localStorage',{get(){throw Error('revoked');}});
 await expect(f.c.cleanup()).resolves.toBeUndefined();expect(f.deliveries[0].closed).toBe(true);expect(()=>f.c.dispose()).not.toThrow();
});
