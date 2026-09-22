// Mutable ownership precedes Preact renders, so an in-flight normal timeline
// response cannot win the frame in which search opens (or closes).
export function createSearchView(){
 let view={active:false,query:'',scope:'current',generation:0};
 return {capture:()=>({...view}),isCurrent:(value:any)=>value.generation===view.generation,
  enter(){view={...view,active:true,generation:view.generation+1};return {...view};},
  close(){view={active:false,query:'',scope:'current',generation:view.generation+1};return {...view};},
  query(query:string){view={...view,query:query.trim(),generation:view.generation+1};return {...view};},
  scope(scope:string){if(!['current','root','all'].includes(scope))return {...view};view={...view,scope,generation:view.generation+1};return {...view};}
 };
}
