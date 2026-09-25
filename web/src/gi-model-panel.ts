// Native catalogue presentation only: no fabricated pricing, pins or thinking
// mutation endpoints. The supplied picker retains selection/write ownership.
export function modelPanelRows<T extends {label?:string}>(rows:T[],current:string):T[]{
 return [...rows.filter(row=>row.label===current),...rows.filter(row=>row.label!==current)];
}
export function modelPanelName(option:any):string{
 return String(option?.name||option?.id||option?.label||'');
}
export function modelPanelContext(value:unknown):string{
 const n=Number(value);if(!Number.isFinite(n)||n<=0)return '';
 if(n>=1_000_000)return `${(n/1_000_000).toFixed(n%1_000_000===0?0:1).replace(/\.0$/,'')}M context`;
 if(n>=1000)return `${(n/1000).toFixed(n%1000===0?0:1).replace(/\.0$/,'')}K context`;
 return `${Math.round(n)} context`;
}
export function openModelSettings(opener:HTMLElement|null){
 window.dispatchEvent(new CustomEvent('piclaw:open-settings',{detail:{section:'models',opener}}));
}
