import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {patchTimelineMenu} from '../../../scripts/patch-timeline-menu.mjs';
import {bindMenuDismissal} from '../../../web/src/gi-menu-dismissal';
const path='web/src/components/timeline-menu.ts';
test('menu build adapter changes dismissal only and leaves supplied source unchanged',()=>{
 const original=readFileSync(path,'utf8'),out=patchTimelineMenu(original);
 expect(out).toContain("import { bindMenuDismissal }");expect(out).not.toContain("document.addEventListener('mousedown', onClick, true)");
 expect(out.indexOf('render(content, portalRef.current)')).toBeLessThan(out.indexOf('return bindMenuDismissal('));
 const content=(s:string)=>s.slice(s.indexOf('    const content = html`'),s.indexOf('    useLayoutEffect(() => {\n        if (portalRef.current) render'));
 expect(content(out)).toBe(content(original));expect(readFileSync(path,'utf8')).toBe(original);
});
test('menu build adapter fails closed on missing, duplicate or already-patched anchors',()=>{
 const original=readFileSync(path,'utf8');
 for(const source of ['',original+original,original.replace('const onKey =','const changed ='),patchTimelineMenu(original)])expect(()=>patchTimelineMenu(source)).toThrow('anchor changed');
});
function fixture(){
 const listeners=new Map<string,Function>();
 const doc={addEventListener:(name:string,fn:Function)=>listeners.set(name,fn),removeEventListener:(name:string)=>listeners.delete(name)};let closed=0,focused=0;
 const menu={ownerDocument:doc},trigger={isConnected:true,hasAttribute:()=>false,focus:()=>focused++};
 const cleanup=bindMenuDismissal(menu as any,trigger as any,()=>closed++);
 const fire=(type:string,path:any[]=[],props:any={})=>{
  // Unit callback fixture, not a DOM event: isTrusted is nonconfigurable on
  // real events. Browser tests supply trusted mouse/touch events separately.
  const e={type,isTrusted:props.trusted!==false,composedPath:()=>path,defaultPrevented:false,preventDefault(){this.defaultPrevented=true;},stopImmediatePropagation(){},...props};listeners.get(type)?.(e);return e;
 };
 return{menu,trigger,cleanup,fire,counts:()=>[closed,focused]};
}
test('menu consumes outside mouse click once, preserves touch defaults and cleans up',()=>{
 const f=fixture();
 expect(f.fire('pointerdown',[],{pointerType:'touch'}).defaultPrevented).toBe(false);
 f.fire('pointercancel');expect(f.counts()).toEqual([0,0]);
 expect(f.fire('mousedown').defaultPrevented).toBe(true);expect(f.counts()).toEqual([0,0]);
 expect(f.fire('click',[],{trusted:false}).defaultPrevented).toBe(false);expect(f.counts()).toEqual([0,0]);
 expect(f.fire('click').defaultPrevented).toBe(true);expect(f.counts()).toEqual([1,1]);
 f.cleanup();expect(f.fire('mousedown').defaultPrevented).toBe(false);expect(f.fire('click').defaultPrevented).toBe(false);expect(f.counts()).toEqual([1,1]);
});
test('menu preserves inside/trigger events and composing keys; Escape restores connected trigger',()=>{
 const f=fixture();for(const node of [f.menu,f.trigger])for(const type of ['pointerdown','mousedown','click'])expect(f.fire(type,[node]).defaultPrevented).toBe(false);
 expect(f.fire('keydown',[],{key:'Escape',isComposing:true}).defaultPrevented).toBe(false);expect(f.counts()).toEqual([0,0]);
 expect(f.fire('keydown',[],{key:'Escape'}).defaultPrevented).toBe(true);expect(f.counts()).toEqual([1,1]);
 f.trigger.isConnected=false;f.fire('keydown',[],{key:'Escape'});expect(f.counts()).toEqual([2,1]);f.cleanup();
});
test('background menu leaves Settings keys and pointer events alone',()=>{
 const doc=new EventTarget() as any;doc.querySelector=()=>({});let closed=0;
 const clean=bindMenuDismissal({ownerDocument:doc} as any,{isConnected:true,hasAttribute:()=>false,focus:()=>{throw Error('stole modal focus');}} as any,()=>closed++);
 for(const type of ['mousedown','pointerdown','click','keydown']){const e=new Event(type,{cancelable:true});Object.assign(e,{key:'Escape'});doc.dispatchEvent(e);expect(e.defaultPrevented).toBe(false);}
 expect(closed).toBe(0);clean();
});
