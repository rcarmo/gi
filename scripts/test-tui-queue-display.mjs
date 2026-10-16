/** Disposable PTY + SQLite proof of Gi's read-only pending panel. */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
if(!process.env.GI_TEST_RUN_ROOT)throw Error('Use make test-env and source scripts/project-test-env.sh');
const binary=process.env.GI_TUI_BIN||resolve('bin/gi'),shell=resolve('tests/ux/shell'),out=join(process.env.GI_TEST_RUN_ROOT,'tests/tui-queue-display');mkdirSync(out,{recursive:true});
const run=(cmd,args)=>{const result=spawnSync(cmd,args,{encoding:'utf8',timeout:12000});if(result.status!==0)throw Error(`${cmd}: ${result.stderr}`);return result.stdout};
const sleep=ms=>new Promise(r=>setTimeout(r,ms));const wait=async(fn,label)=>{const end=Date.now()+15000;while(Date.now()<end){if(fn())return;await sleep(80)}throw Error(`Timed out ${label}`)};const quote=s=>`'${s.replaceAll("'","''")}'`;const results=[];
for(const mode of ['fullscreen','regular']){
 const dir=mkdtempSync(join(tmpdir(),'gi-queue-display-')),db=join(dir,'gi.db'),gates=join(dir,'gates'),socket=`gi-queue-display-${process.pid}-${mode}`,pane='proof:0.0';mkdirSync(gates);mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model']}));
 const tm=(...args)=>run('tmux',['-L',socket,...args]);const cap=()=>tm('capture-pane','-p','-t',pane).replaceAll('\u00a0',' ');const sql=query=>run('sqlite3',['-cmd','.timeout 5000',db,query]).trim();const send=async(text,key='Enter')=>{tm('send-keys','-t',pane,'C-a','C-k');tm('send-keys','-t',pane,'-l',text);await wait(()=>cap().replace(/\s+/g,'').includes(text.replace(/\s+/g,'')),'input '+text);if(text.startsWith('/'))tm('send-keys','-t',pane,'Escape');tm('send-keys','-t',pane,key);await sleep(150)};const altEnter=()=>tm('send-keys','-t',pane,'-H','1b','5b','31','33','3b','33','75');
 try{
  tm('new-session','-d','-s','proof','-x','100','-y','32',`cd '${dir}' && HOME='${dir}' PATH='${shell}':"$PATH" GI_UX_QUEUE_GATES='${gates}' '${binary}' -tui -tui-mode ${mode} -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'; sleep 20`);tm('set-option','-t','proof','status','off');await wait(()=>cap().includes('%/'),'startup');
  // A different process writes committed queue state while this TUI is idle.
  // This checks the render poll, not the external engine's admission semantics.
  const sid=sql('select id from sessions limit 1');
  sql(`insert into turns(id,session_id,status,phase,prompt,metadata_json,created_at,updated_at,queue_position) values('external-follow',${quote(sid)},'queued','queued','external follow-up','{}',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'),1);`);
  sql(`insert into steering_queue(session_id,turn_id,role,content,payload_json,media_json,queue_mode,status,created_at,updated_at) values(${quote(sid)},null,'user','external steering','{}','[]','one-at-a-time','queued',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'));`);
  await wait(()=>{const screen=cap();return screen.includes('Steering: external steering')&&screen.includes('Follow-up: external follow-up')},'other process queue visible without TUI input');
  writeFileSync(join(out,mode+'-external-pending.txt'),cap());
  sql(`update turns set status='cancelled',phase='aborted' where id='external-follow' and session_id=${quote(sid)} and status='queued';`);
  sql(`update steering_queue set status='returned' where session_id=${quote(sid)} and content='external steering' and status='queued';`);
  await wait(()=>!cap().includes('↳ alt+up to edit all queued messages'),'other process queue cleared without TUI input');
  writeFileSync(join(out,mode+'-external-cleared.txt'),cap());
  sql(`insert into turns(id,session_id,status,phase,prompt,metadata_json,created_at,updated_at,queue_position) values('external-media',${quote(sid)},'queued','queued','',json_object('media',json_array(json_object('media_id',7,'session_id',${quote(sid)}))),strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'),1);`);
  await wait(()=>cap().includes('Follow-up: [attachment]'),'media-only pending item visible');
  writeFileSync(join(out,mode+'-media-only.txt'),cap());
  await send('/queue');await wait(()=>cap().includes('external-media  [attachment]'),'media-only inspectable by ID');
  if(sql("select status from turns where id='external-media';")!=='queued')throw Error('media-only display changed queued turn');
  sql(`update turns set status='cancelled',phase='aborted' where id='external-media' and status='queued';`);
  await wait(()=>!cap().includes('Follow-up: [attachment]'),'media-only pending display cleared');
  // Switching changes only the selected session's pending view and editor.
  // Use Gi's session creation command: a bare SQL session lacks its required
  // runtime identity, so /switch correctly rejects it.
  await send('/new');await wait(()=>sql('select count(*) from sessions;')==='2','new session with identity');
  const other=sql(`select id from sessions where id!=${quote(sid)} limit 1;`);
  await send(`/switch ${sid}`);await wait(()=>cap().includes(sid),'origin after new session');
  sql(`insert into turns(id,session_id,status,phase,prompt,metadata_json,created_at,updated_at,queue_position) values('other-follow',${quote(other)},'queued','queued','other session follow','{}',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'),1);`);
  sql(`insert into turns(id,session_id,status,phase,prompt,metadata_json,created_at,updated_at,queue_position) values('origin-follow',${quote(sid)},'queued','queued','origin session follow','{}',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'),2);`);
  await wait(()=>cap().includes('Follow-up: origin session follow'),'origin pending after external writes');
  await send(`/switch ${other}`);await wait(()=>{const screen=cap();return screen.includes('Follow-up: other session follow')&&!screen.includes('Follow-up: origin session follow')},'switch to other pending panel');
  writeFileSync(join(out,mode+'-other-session.txt'),cap());
  await send(`/switch ${sid}`);await wait(()=>{const screen=cap();return screen.includes('Follow-up: origin session follow')&&!screen.includes('Follow-up: other session follow')},'return to origin pending panel');
  if(sql("select count(*) from turns where id in ('other-follow','origin-follow') and status='queued';")!=='2')throw Error('switch changed other session queue');
  sql(`update turns set status='cancelled',phase='aborted' where id in ('other-follow','origin-follow') and status='queued';`);
  await wait(()=>!cap().includes('↳ alt+up to edit all queued messages'),'session switch queue cleanup');
  await send('UX queue gate:held');await wait(()=>sql("select count(*) from turn_events where event_type='tool.started';")==='1','held shell');
  await send('steer at boundary');await wait(()=>sql("select count(*) from steering_queue where content='steer at boundary' and status='queued';")==='1','steer');
  tm('send-keys','-t',pane,'-l','follow after completion');altEnter();await wait(()=>sql("select count(*) from turns where prompt='follow after completion' and status='queued';")==='1','follow-up');
  await wait(()=>{const text=cap();return text.includes('Steering: steer at boundary')&&text.includes('Follow-up: follow after completion')&&text.includes('alt+up to edit all queued messages')},'visible pending labels/hint');
  let text=cap(),steering=text.indexOf('Steering: steer at boundary'),followUp=text.indexOf('Follow-up: follow after completion');
  if(steering<0||followUp<=steering)throw Error('pending ordering reversed');
  writeFileSync(join(out,mode+'-pending.txt'),text);
  await send('/queue');await wait(()=>{const screen=cap();return screen.includes('1 queued · 1 steering')&&screen.includes('Steering: steer at boundary')&&screen.includes('follow after completion')},'read-only /queue inspection');
  if(sql("select count(*) from turns where prompt='follow after completion' and status='queued';")!=='1'||sql("select count(*) from steering_queue where content='steer at boundary' and status='queued';")!=='1')throw Error('/queue inspection changed delivery');
  writeFileSync(join(out,mode+'-inspection.txt'),cap());
  tm('send-keys','-t',pane,'-l','unsubmitted draft');await wait(()=>sql("select count(*) from kv_store where namespace='tui_text_draft_v1' and cast(value as text) like '%unsubmitted draft%';")==='1','durable draft');
  if(!cap().includes('Follow-up: follow after completion'))throw Error('typing draft hid queue');
  tm('resize-window','-t','proof','-x','64','-y','12');await wait(()=>cap().includes('queue: 2 pending; /queue to inspect'),'compact pending summary');
  writeFileSync(join(out,mode+'-compact.txt'),cap());
  tm('resize-window','-t','proof','-x','100','-y','32');await wait(()=>{const screen=cap();return screen.includes('Steering: steer at boundary')&&screen.includes('Follow-up: follow after completion')},'pending after resize');
  writeFileSync(join(gates,'held'),'go');await wait(()=>sql("select count(*) from turns where status in ('queued','running','steering');")==='0','queue drain');
  await wait(()=>{const screen=cap();return !screen.includes('↳ alt+up to edit all queued messages')&&!screen.includes('queue: 2 pending; /queue to inspect')},'pending panel clear');
  if(sql("select count(*) from messages where role='user' and content in ('steer at boundary','follow after completion');")!=='2')throw Error('pending display disrupted delivery');
  if(sql("select count(*) from kv_store where namespace='tui_text_draft_v1' and cast(value as text) like '%unsubmitted draft%';")!=='1')throw Error('pending display changed draft');
  results.push({mode,result:'pass',mediaOnlyVisible:true,sessionSwitchIsolated:true,externalUpdateWithoutInput:true,steeringBeforeFollowUp:true,draftPreserved:true,clearedAfterDelivery:true});
 }catch(e){results.push({mode,result:'fail',error:String(e)});try{writeFileSync(join(out,mode+'-failure.txt'),cap());writeFileSync(join(out,mode+'-runtime.log'),readFileSync(join(dir,'runtime.log')))}catch{}}
 finally{writeFileSync(join(gates,'held'),'go');try{tm('kill-server')}catch{}rmSync(dir,{recursive:true,force:true})}
}
writeFileSync(join(out,'results.json'),JSON.stringify({scope:'Gi disposable PTY/SQLite pending display, no Pi PTY or live provider',results},null,2));console.log(JSON.stringify({out,results},null,2));if(results.some(r=>r.result==='fail'))process.exitCode=1;
