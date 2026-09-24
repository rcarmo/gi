import {test,expect} from 'bun:test';
import {bindQuickActionsFocus,quickActionsOpener} from '../../../web/src/gi-quick-actions-focus';
function fixture(){
 const doc=new EventTarget() as any;let frame:()=>void=()=>{},canceled=0,closed=0,focused=0,restored=0,modal=false,inert=false;
 doc.querySelector=()=>modal?{}:null;doc.defaultView={requestAnimationFrame:(fn:()=>void)=>{frame=fn;return 7;},cancelAnimationFrame:(id:number)=>{expect(id).toBe(7);canceled++;}};
 const button=new EventTarget();const root={ownerDocument:doc,closest:()=>null,querySelector:()=>button},input={isConnected:true,focus:()=>focused++};
 const opener={isConnected:true,closest:()=>inert?{}:null,hasAttribute:()=>false,getClientRects:()=>[{}],focus:()=>restored++};
 const cleanup=bindQuickActionsFocus(root as any,input as any,opener as any,()=>closed++);
 const escape=()=>{const e=new Event('keydown',{cancelable:true});Object.assign(e,{key:'Escape'});doc.dispatchEvent(e);return e;};
 return {cleanup,escape,button,frame:()=>frame(),opener,setInert:()=>inert=true,setModal:()=>modal=true,counts:()=>({canceled,closed,focused,restored})};
}
test('palette dismissal and unmount fence a pending animation-frame focus',()=>{
 const f=fixture();expect(f.escape().defaultPrevented).toBe(true);f.frame();expect(f.counts()).toEqual({canceled:1,closed:1,focused:0,restored:1});f.cleanup();f.frame();expect(f.counts().focused).toBe(0);
 const unmounted=fixture();unmounted.cleanup();unmounted.frame();expect(unmounted.counts()).toEqual({canceled:1,closed:0,focused:0,restored:0});
});
test('palette restores only connected usable opener and does not focus beneath Settings',()=>{
 for(const mode of ['removed','inert']){const f=fixture();f.frame();expect(f.counts().focused).toBe(1);if(mode==='removed')f.opener.isConnected=false;else f.setInert();f.escape();expect(f.counts().restored).toBe(0);f.cleanup();}
 const f=fixture();f.setModal();f.frame();expect(f.counts().focused).toBe(0);expect(f.escape().defaultPrevented).toBe(false);f.cleanup();
});
test('palette uses the real focused opener or falls back to the conversation region',()=>{
 const region={},body={},html={},active={closest:()=>null};
 const doc={body,documentElement:html,activeElement:active,querySelector:(s:string)=>{expect(s).toBe('.container[aria-label="Conversation"]');return region;}} as any;
 expect(quickActionsOpener(doc)).toBe(active);doc.activeElement=body;expect(quickActionsOpener(doc)).toBe(region);doc.activeElement=html;expect(quickActionsOpener(doc)).toBe(region);doc.activeElement={closest:()=>({})};expect(quickActionsOpener(doc)).toBe(region);
});
test('close control shares one-shot dismissal, respects modal ownership and detaches on cleanup',()=>{
 const f=fixture();f.button.dispatchEvent(new Event('click'));f.button.dispatchEvent(new Event('click'));f.frame();expect(f.counts()).toEqual({canceled:1,closed:1,focused:0,restored:1});f.cleanup();f.button.dispatchEvent(new Event('click'));expect(f.counts().closed).toBe(1);
 const modal=fixture();modal.setModal();modal.button.dispatchEvent(new Event('click'));expect(modal.counts().closed).toBe(0);modal.cleanup();
 const unmounted=fixture();unmounted.cleanup();unmounted.button.dispatchEvent(new Event('click'));expect(unmounted.counts().closed).toBe(0);
});
