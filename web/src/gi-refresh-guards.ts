// Activation and transport readiness may occur in either order. Exactly one
// owns their initial refresh; a real disconnect opens a new readiness epoch.
// No response fetched before native SSE subscription is reused as fresh state.
export function createActivationRefreshGate() {
 let selection:unknown=null,connected=false,claimed=false;
 const select=(key:unknown)=>{if(key!==selection){selection=key;connected=false;claimed=false;}};
 const claim=()=>{if(!connected||claimed)return false;claimed=true;return true;};
 return {
  select,
  activate(key:unknown){select(key);return claim();},
  status(key:unknown,status:string){select(key);if(status!=='connected'){connected=false;claimed=false;return false;}connected=true;return claim();},
  ready(key:unknown){return key===selection&&connected;},
 };
}

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
