import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {mountWorkspaceTab} from '../../../web/src/gi-workspace-tab-lifecycle';
import {patchWorkspaceReadonly} from '../../../scripts/patch-workspace-readonly.mjs';
function fixture(){
 const states:any[]=[],mounts:any[]=[],reads:any[]=[];let disposed=0,cleared=0;
 const container={replaceChildren:()=>cleared++} as any;
 const registry={resolve:(context:any)=>({mount:(host:any,ctx:any)=>{mounts.push({host,ctx});return{dispose:()=>disposed++}}})};
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
 for(const registry of [{resolve:()=>null},{resolve:()=>({mount:()=>{throw new Error('renderer')}})}]){
  const stop=mountWorkspaceTab(f.container,'broken',s=>f.states.push(s),async()=>({kind:'text'}),registry);await Promise.resolve();expect(f.states.at(-1).loading).toBe(false);expect(f.states.at(-1).error).toBeTruthy();stop();
 }
});
test('read-only label adapter rejects source drift without altering supplied bytes',()=>{
 const path='web/src/components/workspace-explorer.ts',source=readFileSync(path,'utf8'),adapted=patchWorkspaceReadonly(source);
 expect(adapted).toContain("? 'Open read-only tab'");expect(adapted).not.toContain('>Open in editor</button>');expect(readFileSync(path,'utf8')).toBe(source);
 for(const changed of ['',source+source,adapted,source.replace('Open in editor','Changed')])expect(()=>patchWorkspaceReadonly(changed)).toThrow('anchor changed');
});
