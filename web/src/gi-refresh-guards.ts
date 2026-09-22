// HTTP responses have their own generation in addition to session/SSE scope.
export function createTimelineRevision() {
 let generation=0;
 return {begin:()=>++generation,invalidate:()=>++generation,accepts:(value:number)=>value===generation};
}
export function createAssetVersionGuard(initial: string|null) {
 let baseline=initial?.trim()||'';const seen=new Set<string>();
 return {observe(value:unknown){
  if(typeof value!=='string'||!value.trim())return false;
  const version=value.trim();
  if(!baseline){baseline=version;return false;}
  if(version===baseline||seen.has(version))return false;
  seen.add(version);return true;
 }};
}
export function loadedAssetVersion(doc:Document):string|null {
 const script=doc.querySelector('script[src*="/dist/app.bundle.js"]');
 const src=script?.getAttribute('src');
 if(!src)return null;
 try{return new URL(src,doc.baseURI).searchParams.get('v');}catch{return null;}
}
