import {createServer} from 'node:http';
import {readFile,realpath} from 'node:fs/promises';
import {resolve,extname} from 'node:path';
import {createHash} from 'node:crypto';
const mime={'.html':'text/html','.js':'text/javascript','.css':'text/css','.json':'application/json','.svg':'image/svg+xml','.png':'image/png','.ico':'image/x-icon','.woff2':'font/woff2'};
export async function installPixelHost({page,host,root,state,reference}){
 if(!['gi','piclaw'].includes(host))throw Error(`Unknown pixel host: ${host}`);
 const failures=[],calls=[],assets={},streams=new Map(),streamAborts=[];
 const onError=e=>failures.push(`page: ${e.message}`);
 const onFailed=r=>{const error=r.failure()?.errorText,path=new URL(r.url()).pathname;
  // Gi aborts the initial unscoped fetch stream on session activation. Retain
  // that cancellation as evidence and require a live replacement at capture.
  if(host==='gi'&&path==='/sse/stream'&&error==='net::ERR_ABORTED'&&streamAborts.length===0){streamAborts.push({path,error});return;}
  failures.push(`network: ${r.url()} ${error}`);
 };
 page.on('pageerror',onError);page.on('requestfailed',onFailed);
 const server=createServer((req,res)=>{const u=new URL(req.url,'http://fixture');if(req.method!=='GET'||!['/sse/stream','/sse/topics'].includes(u.pathname)){failures.push(`Unexpected native request ${req.method} ${u.pathname}`);res.writeHead(404);res.end();return;}res.writeHead(200,{'Content-Type':'text/event-stream','Cache-Control':'no-store'});res.write(': pixel fixture\n\n');streams.set(res,u.pathname);res.on('close',()=>streams.delete(res));});
 await new Promise(r=>server.listen(0,'127.0.0.1',r));const origin=`http://127.0.0.1:${server.address().port}`;
 const session={id:'main',title:state.sessionLabel,scope:{agent_id:'web'},state:{selected_model:state.model.current},created_at:state.now,updated_at:state.now,status:'idle',parent_session_id:null};
 const context={tokens:0,contextWindow:65536,percent:0,compact_command:'/compact'};
 // Piclaw agent_name is the session handle here, not the global assistant
 // identity; Gi's sessionToActiveChat projects session.title the same way.
 const chat={chat_jid:state.sessionId,name:state.sessionLabel,agent_name:state.sessionLabel,model:state.model.current,parent_chat_jid:null,root_chat_jid:state.sessionId,created_at:state.now,updated_at:state.now,is_archived:false,is_pinned:false,is_running:false,message_count:0,session_kind:'root'};
 const piclaw={
  '/agent/active-chats':{chats:[chat]},'/agent/branches':{chats:[chat]},'/agent/picker-pins':{pins:[]},
  '/agent/roster':{agents:[{id:'default',name:state.agentName}],user:{name:state.userName}},
  '/agent/models':{...state.model,models:[{id:'fixture',provider:'test',name:'fixture',context_window:65536}],available_model_count:1},
  '/agent/context':context,'/agent/queue-state':{items:[],count:0},'/agent/autoresearch/status':{enabled:false},
  '/agent/status':{status:{status:'idle',type:'idle'},model:state.model,context,metrics:state.metrics,agent_name:state.agentName,errors:[]},
  '/agent/commands':{commands:state.commands},'/agent/settings/quick-actions':{ok:true,settings:{workspaceCommands:[],slashCommands:state.commands.map(c=>c.name)}},
  '/agent/addons/web-entries':{entries:[]},'/agent/system-metrics':state.metrics,'/timeline':{posts:[],has_more:false},
  '/workspace/tree':{root:{name:'fixture',path:'.',type:'dir',children:[]},truncated:false},'/workspace/index-status':{state:'ready',roots:['.']},
 };
 const gi={
  '/api/auth/status':{mode:'single-user',enrolled:false,authenticated:false,totp_enabled:false,browser_login_available:true,setup_available:true},
  '/api/runtime/config':{...state.model,user_name:state.userName,assistant_name:state.agentName,workspace_root:'/fixture',default_model:state.model.current,default_thinking_level:'medium',version:'fixture'},
  '/api/sessions':{sessions:[session]},'/api/sessions/main':session,'/api/sessions/main/messages':{messages:[],has_more:false},
  '/api/sessions/main/model':state.model,'/api/sessions/main/branches':{branches:[]},'/api/sessions/main/queue':{items:[],count:0},'/api/sessions/main/activity':{status:'idle',turn_id:null},'/api/sessions/main/compaction':{state:'idle',running:false},
  '/api/quick-actions':{commands:state.commands,settings:{workspaceCommands:[],slashCommands:state.commands.map(c=>c.name)}},'/api/metrics':state.metrics,
  '/api/workspace/tree':{name:'fixture',path:'.',type:'dir',children:[]},'/api/workspace/index':{state:'ready',roots:['.']},'/api/workspace/index/status':{state:'ready',roots:['.']},
 };
 const allowedQueries=new Set(['chat_jid','root_chat_jid','include_archived','limit','before','after','ui','scope','path','depth','show_hidden']);
 await page.addInitScript(s=>{localStorage.setItem('piclaw_theme',s.theme);localStorage.setItem('vibes-theme',s.theme);localStorage.setItem('workspaceOpen','false');localStorage.setItem('piclaw_system_meters_enabled','false');localStorage.setItem('piclaw_compose_height',String(s.composeHeightPreference));localStorage.setItem('gi_session_id','main');},state);
 const handler=async route=>{
  const r=route.request(),u=new URL(r.url());calls.push({method:r.method(),path:u.pathname,query:u.search});
  try{
   if(u.origin!==origin)throw Error(`External request ${u.origin}`);
   if(['/sse/stream','/sse/topics'].includes(u.pathname)){if(r.method()!=='GET')throw Error('Invalid SSE method');for(const [k,v]of u.searchParams){if(!['chat_jid','session_id','patterns','after'].includes(k)||k==='chat_jid'&&v!==state.sessionId||k==='session_id'&&v!=='main')throw Error(`Invalid SSE query ${k}`);}return route.continue();}
   let file;
   if(host==='piclaw'&&reference.files[u.pathname])file=resolve(root,reference.files[u.pathname].relativePath);
   if(host==='gi'){
    if(u.pathname==='/')file=resolve(root,'index.html');
    else if(/^\/(css|dist|fonts|js|editor-vendor)\//.test(u.pathname)||/^\/(favicon(?:-\d+x\d+)?\.(?:png|ico)|manifest\.json)$/.test(u.pathname))file=resolve(root,decodeURIComponent(u.pathname.slice(1)));
   }
   if(file){if(r.method()!=='GET'||!(await realpath(file)).startsWith((await realpath(root))+'/'))throw Error('Invalid asset path/method');for(const k of u.searchParams.keys())if(!['v','chat_jid'].includes(k))throw Error(`Invalid asset query ${k}`);const bytes=await readFile(file),sha256=createHash('sha256').update(bytes).digest('hex');if(host==='piclaw'&&sha256!==reference.files[u.pathname].sha256)throw Error(`Reference hash changed: ${u.pathname}`);if(assets[u.pathname]&&assets[u.pathname].sha256!==sha256)throw Error(`Asset changed during capture: ${u.pathname}`);assets[u.pathname]={path:file,sha256,bytes:bytes.length};return route.fulfill({contentType:mime[extname(file)]||'application/octet-stream',body:bytes});}
   for(const [key,value]of u.searchParams){if(!allowedQueries.has(key))throw Error(`Undeclared query ${key}`);if(['chat_jid','root_chat_jid'].includes(key)&&value!==state.sessionId)throw Error('Wrong session scope');}
   if(host==='piclaw'&&r.method()==='POST'&&['/workspace/visibility','/agent/push/presence'].includes(u.pathname))return route.fulfill({json:{ok:true}});
   const table=host==='piclaw'?piclaw:gi;if(r.method()==='GET'&&Object.hasOwn(table,u.pathname))return route.fulfill({json:table[u.pathname]});
   throw Error(`Undeclared ${r.method()} ${u.pathname}${u.search}`);
  }catch(e){failures.push(e.message);return route.fulfill({status:500,json:{error:e.message}});}
 };
 await page.route('**/*',handler);
 return {origin,assets,calls,failures,streamAborts,
  assert(){if(failures.length)throw Error(failures.join('; '));if(![...streams.values()].includes('/sse/stream'))throw Error('No connected fixture stream');},
  async connected(){for(let n=0;n<100&&!([...streams.values()].includes('/sse/stream'));n++)await page.waitForTimeout(50);this.assert();for(const [s,path]of streams)if(path==='/sse/stream')s.write(`event: connected\ndata: ${JSON.stringify({chat_jid:state.sessionId})}\n\n`);},
  async dispose(){page.off('pageerror',onError);page.off('requestfailed',onFailed);await page.unroute('**/*',handler);for(const s of streams.keys())s.end();server.closeAllConnections();await new Promise(r=>server.close(r));}
 };
}
