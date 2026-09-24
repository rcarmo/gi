import {test,expect} from 'bun:test';
import {installGiDisplayScale} from '../../../web/src/gi-display-scale';
import {normalizePwaDisplayScalePercent} from '../../../web/src/ui/pwa-display-scale';

function fixture(legacy=false){
 const events=new Map<string,Set<Function>>(),queries=new Map<string,any>(),classes=new Set<string>();let viewport='';
 const runtime:any={navigator:{standalone:true,maxTouchPoints:5},localStorage:{getItem:()=> '85'},document:{documentElement:{classList:{toggle:(name:string,on:boolean)=>on?classes.add(name):classes.delete(name),remove:(name:string)=>classes.delete(name)}},querySelector:()=>({setAttribute:(_k:string,v:string)=>viewport=v})},addEventListener:(name:string,fn:Function)=>{if(!events.has(name))events.set(name,new Set());events.get(name)!.add(fn)},removeEventListener:(name:string,fn:Function)=>events.get(name)?.delete(fn),matchMedia:(query:string)=>{if(!queries.has(query)){const callbacks=new Set<Function>();queries.set(query,{matches:false,callbacks,...(legacy?{addListener:(fn:Function)=>callbacks.add(fn),removeListener:(fn:Function)=>callbacks.delete(fn)}:{addEventListener:(_name:string,fn:Function)=>callbacks.add(fn),removeEventListener:(_name:string,fn:Function)=>callbacks.delete(fn)})});}return queries.get(query)}};
 return {runtime,events,queries,classes,viewport:()=>viewport,focus:()=>{for(const fn of events.get('focus')||[])fn()}};
}
for(const legacy of [false,true])test(`standalone fallback shares native gates and cleans listeners (${legacy?'legacy':'modern'})`,()=>{
 const f=fixture(legacy),close=installGiDisplayScale(f.runtime);
 expect(f.classes.has('gi-standalone-display')).toBe(true);expect(f.viewport()).toContain('initial-scale=0.85');
 f.runtime.navigator.standalone=false;f.focus();expect(f.classes.size).toBe(0);expect(f.viewport()).toContain('initial-scale=1.0');
 const query=f.queries.get('(display-mode: standalone)');query.matches=true;for(const fn of query.callbacks)fn();expect(f.classes.has('gi-standalone-display')).toBe(true);expect(f.viewport()).toContain('initial-scale=0.85');
 f.runtime.navigator.maxTouchPoints=0;f.focus();expect(f.classes.has('gi-standalone-display')).toBe(true);expect(f.viewport()).toContain('initial-scale=1.0');
 close();expect(f.classes.size).toBe(0);for(const set of f.events.values())expect(set.size).toBe(0);for(const query of f.queries.values())expect(query.callbacks.size).toBe(0);
});
test('supplied scale normalization preserves supported bounds and migration inputs',()=>{
 for(const [input,expected]of [[0,20],[19,20],[20,20],[115,115],[116,115],['0.85',85],['90%',90],['bad',100],['',100],[84.7,85]] as const)expect(normalizePwaDisplayScalePercent(input)).toBe(expected);
});
test('unsupported display-mode queries cannot abort startup; navigator fallback and cleanup survive',()=>{
 const f=fixture();f.runtime.matchMedia=()=>{throw new Error('unsupported query')};
 const close=installGiDisplayScale(f.runtime);expect(f.classes.has('gi-standalone-display')).toBe(true);expect(f.viewport()).toContain('initial-scale=0.85');
 f.runtime.navigator.standalone=false;f.focus();expect(f.classes.size).toBe(0);expect(f.viewport()).toContain('initial-scale=1.0');close();
 for(const set of f.events.values())expect(set.size).toBe(0);
});
