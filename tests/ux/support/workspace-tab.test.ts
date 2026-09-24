import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {mountWorkspaceTab, observeWorkspaceTab} from '../../../web/src/gi-workspace-tab-lifecycle';
import {patchWorkspaceReadonly} from '../../../scripts/patch-workspace-readonly.mjs';
function fixture(){
 const states:any[]=[],mounts:any[]=[],reads:any[]=[];let disposed=0,cleared=0;
 const container={replaceChildren:()=>cleared++} as any;
 const registry={resolve:(context:any)=>({placement:'tabs',capabilities:['readonly','preview'],mount:(host:any,ctx:any)=>{mounts.push({host,ctx});return{dispose:()=>disposed++}}})};
 const read=(path:string,limit:number)=>new Promise((resolve,reject)=>reads.push({path,limit,resolve,reject}));
 return {states,mounts,reads,container,registry,read,counts:()=>({disposed,cleared}),mount:(path:string)=>mountWorkspaceTab(container,path,state=>states.push(state),read,registry)};
}
test('tab occurrence fences reads and errors after close/switch; mount/dispose exactly once',async()=>{
 const f=fixture(),closeA=f.mount('a.md');expect(f.reads[0]).toMatchObject({path:'a.md',limit:20000});closeA();const closeB=f.mount('b.txt');
 f.reads[0].resolve({path:'a.md',kind:'text',text:'old'});await Promise.resolve();expect(f.mounts).toHaveLength(0);
 f.reads[1].resolve({path:'b.txt',kind:'text',text:'new'});await Promise.resolve();expect(f.mounts).toHaveLength(1);expect(f.mounts[0].ctx).toMatchObject({path:'b.txt',mode:'view',preview:{text:'new'}});expect(f.states.at(-1)).toEqual({loading:false,error:''});closeB();closeB();expect(f.counts().disposed).toBe(1);
 const before=f.states.length,close=f.mount('a.md');close();f.reads[2].reject(new Error('late'));await Promise.resolve();expect(f.states).toHaveLength(before+1);
});
test('native failure is explicit; remount retries and rejects missing/throwing renderer',async()=>{
 const f=fixture();const close=f.mount('missing');f.reads[0].reject(new Error('404'));await Promise.resolve();expect(f.states.at(-1)).toEqual({loading:false,error:'404'});close();
 const retry=f.mount('missing');f.reads[1].resolve({kind:'text',text:'recovered'});await Promise.resolve();expect(f.mounts[0].ctx.preview.text).toBe('recovered');retry();
 for(const registry of [{resolve:()=>null},{resolve:()=>({placement:'tabs',capabilities:['preview'],mount:()=>{throw new Error('renderer')}})}]){
  const stop=mountWorkspaceTab(f.container,'broken',s=>f.states.push(s),async()=>({kind:'text'}),registry);await Promise.resolve();expect(f.states.at(-1).loading).toBe(false);expect(f.states.at(-1).error).toBeTruthy();stop();
 }
});
test('read-only label adapter rejects source drift without altering supplied bytes',()=>{
 const path='web/src/components/workspace-explorer.ts',source=readFileSync(path,'utf8'),adapted=patchWorkspaceReadonly(source);
 expect(adapted).toContain("? 'Open read-only tab'");expect(adapted).not.toContain('>Open in editor</button>');expect(readFileSync(path,'utf8')).toBe(source);
 for(const changed of ['',source+source,adapted,source.replace('Open in editor','Changed')])expect(()=>patchWorkspaceReadonly(changed)).toThrow('anchor changed');
});

function lifecycle(overrides:any = {}) {
 const events:string[] = [], states:any[] = [], contexts:any[] = [];
 let close:()=>void, resize:()=>void, closed=0;
 const instance={dispose(){events.push('dispose');},onClose(cb:()=>void){close=cb;events.push('bind-close');},resize(){events.push('resize');},focus(){events.push('focus');},...overrides};
 const container={replaceChildren(){events.push('clear');}} as any;
 const stop=mountWorkspaceTab(container,'bounded.md',s=>states.push(s),async()=>({kind:'text',text:'αβ',mtime:'2026-09-24T00:00:00Z',size:90000,truncated:true}),{resolve(c:any){contexts.push(c);return{placement:'tabs',capabilities:['preview'],mount(){events.push('mount');return instance;}};}}, {
  close(){closed++;},observeResize(_host:any,cb:()=>void){resize=cb;events.push('observe');return()=>events.push('disconnect');}
 });
 return{events,states,contexts,stop,close:()=>close(),resize:()=>resize(),closed:()=>closed};
}
test('read-only contract carries bounded text/mtime/full size and resize/close are occurrence-owned',async()=>{
 const f=lifecycle();await Promise.resolve();
 expect(f.contexts[0]).toMatchObject({path:'bounded.md',mode:'view',content:'αβ',mtime:'2026-09-24T00:00:00Z',size:90000,preview:{truncated:true}});
 expect(f.events).toEqual(['mount','bind-close','observe','resize']);
 f.resize();expect(f.events.at(-1)).toBe('resize');f.close();f.close();expect(f.closed()).toBe(1);
 expect(f.events.slice(-3)).toEqual(['disconnect','dispose','clear']);
 const before=[...f.events];f.resize();f.close();f.stop();expect(f.events).toEqual(before);
});
test('unsupported capability combinations fail before mount; binary context never invents text',async()=>{
 for(const [placement,capabilities] of [['dock',['readonly']],['tabs',['edit']],['tabs',['terminal']],['tabs',['readonly','edit']],['tabs',[]],['tabs',['unknown']]]){
  let mounted=0;const states:any[]=[];
  const stop=mountWorkspaceTab({replaceChildren(){}} as any,'bad',s=>states.push(s),async()=>({kind:'text',text:'x'}),{resolve:()=>({placement,capabilities,mount(){mounted++;}})});
  await Promise.resolve();expect(mounted).toBe(0);expect(states.at(-1).error).toContain('read-only');stop();
 }
 let context:any;const stop=mountWorkspaceTab({replaceChildren(){}} as any,'image.png',()=>{},async()=>({kind:'image',size:80,mtime:'stamp',url:'/raw'}),{resolve(c:any){context=c;return{placement:'tabs',capabilities:['readonly'],mount:()=>({dispose(){}})};}});
 await Promise.resolve();expect(context).toMatchObject({size:80,mtime:'stamp',preview:{kind:'image'}});expect(context.content).toBeUndefined();stop();
});
test('reentrant close registration and failing lifecycle hooks always fence and dispose once',async()=>{
 let close:()=>void;
 const synchronous=lifecycle({onClose(cb:()=>void){close=cb;cb();}});await Promise.resolve();
 expect(synchronous.closed()).toBe(1);expect(synchronous.events).toEqual(['mount','dispose','clear']);close!();synchronous.stop();expect(synchronous.closed()).toBe(1);
 for(const hook of ['onClose','resize']){
  const f=lifecycle({[hook](){throw new Error('hook failed');},dispose(){throw new Error('dispose failed');}});await Promise.resolve();
  expect(f.states.at(-1)).toEqual({loading:false,error:'hook failed'});expect(f.events.at(-1)).toBe('clear');const before=[...f.events];f.stop();expect(f.events).toEqual(before);
 }
 const f=lifecycle({dispose(){throw new Error('dispose failed');}});await Promise.resolve();expect(()=>f.stop()).not.toThrow();expect(f.events.slice(-2)).toEqual(['disconnect','clear']);
});
test('named Piclaw subset contract and supplied preview bytes stay pinned',()=>{
 // Piclaw bfc34e4ebfe9b0ce0deefa6202a9d9a4780a5322, runtime/web/src/panes/.
 const hashes={'pane-types.ts':'7cea01913387fcd2eb487bf44baa802d53f393b92d7db4127c7713f0737b39be','pane-registry.ts':'261d351b0fb76e449f4c9bb883adc2b16e2d348dbffd93e621e5dfc796021ad9','workspace-preview-pane.ts':'7cb6cf7f67042e490caebb7e391c3f0d94107fecb70539c02e6cf83102dacd78'};
 for(const [file,hash] of Object.entries(hashes))expect(createHash('sha256').update(readFileSync(`web/src/panes/${file}`)).digest('hex')).toBe(hash);
});
test('container ResizeObserver and window fallback detach the exact listener',()=>{
 const events:any[]=[];let observerCallback:any,listener:any;
 class RO{constructor(cb:any){observerCallback=cb;}observe(host:any){events.push(host);}disconnect(){events.push('disconnect');}}
 const container:any={ownerDocument:{defaultView:{ResizeObserver:RO}}};
 const stop=observeWorkspaceTab(container,()=>events.push('resize'));observerCallback();stop();expect(events).toEqual([container,'resize','disconnect']);
 container.ownerDocument.defaultView={addEventListener(name:any,cb:any){events.push(name);listener=cb;},removeEventListener(name:any,cb:any){expect(name).toBe('resize');expect(cb).toBe(listener);events.push('remove');}};
 const fallback=observeWorkspaceTab(container,()=>events.push('fallback'));listener();fallback();expect(events.slice(-3)).toEqual(['resize','fallback','remove']);
});
