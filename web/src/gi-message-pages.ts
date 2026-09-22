export function mergeMessagePages(current:any[],incoming:any[]) {
 const rows=new Map(current.map(p=>[String(p.id),p]));
 for(const post of incoming)rows.set(String(post.id),post);
 const compare=(a:any,b:any)=>String(a)<String(b)?-1:String(a)>String(b)?1:0;
 return [...rows.values()].sort((a,b)=>compare(a.timestamp,b.timestamp)||compare(a.id,b.id));
}
export function newMessageWindow(){return {before:null as string|null,after:null as string|null,hasMore:false,loaded:false};}
// column-reverse scroll origin is the newest edge (0); history positions are
// negative. Anchor an actual visible message instead of assuming height deltas.
function layoutTop(node:HTMLElement){
 // CSS entry animations translate posts after layout. Ignore those transforms
 // when measuring the scroll anchor so a fade/slide cannot accumulate drift.
 const transform=getComputedStyle(node).transform;
 let y=0;if(transform&&transform!=='none'){try{y=new DOMMatrixReadOnly(transform).m42;}catch{}}
 return node.getBoundingClientRect().top-y;
}
export function captureTimelineAnchor(root:HTMLElement|null,previous?:any){
 if(!root)return null;
 if(previous?.root===root&&previous.id&&document.getElementById(previous.id))return previous;
 const box=root.getBoundingClientRect();
 const node=[...root.querySelectorAll<HTMLElement>('.post[id]')].find(el=>{const r=el.getBoundingClientRect();return r.bottom>box.top&&r.top<box.bottom;});
 return {root,id:node?.id||'',top:node?layoutTop(node):0,scroll:root.scrollTop,height:root.clientHeight};
}
export function restoreTimelineAnchor(anchor:ReturnType<typeof captureTimelineAnchor>){
 if(!anchor?.id||!anchor.root.isConnected)return;
 const node=anchor.root.querySelector<HTMLElement>(`[id="${CSS.escape(anchor.id)}"]`);
 if(node){anchor.root.scrollTop+=layoutTop(node)-anchor.top;anchor.scroll=anchor.root.scrollTop;}
}
