/** @description Explicit durable queue list/remove/steer, restart and native guards in six PTYs. */
import{execFileSync}from'node:child_process';
import{mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync}from'node:fs';
import{tmpdir}from'node:os';import{join,resolve}from'node:path';
const bin=process.env.GI_TUI_BIN||resolve('bin/gi'),out=resolve('test-results/tui-queue-commands');mkdirSync(out,{recursive:true});
const run=(cmd,args)=>execFileSync(cmd,args,{encoding:'utf8',timeout:15000}),sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(v,m)=>{if(!v)throw Error(m)};
async function wait(fn,label){for(let i=0;i<160;i++){if(fn())return;await sleep(60)}throw Error('timeout: '+label)}
const results=[];
for(const mode of ['fullscreen','regular'])for(const[width,height]of[[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-queue-cli-')),db=join(dir,'state.db'),socket=`gi-queue-${process.pid}-${mode}-${width}`,pane='proof:0.0';mkdirSync(join(dir,'.pi'));const settings=JSON.stringify({defaultProvider:'test',defaultModel:'test-model',enabledModels:['test-model']});writeFileSync(join(dir,'.pi/settings.json'),settings);
 const tm=(...a)=>run('tmux',['-L',socket,...a]),keys=(...a)=>tm('send-keys','-t',pane,...a),type=s=>keys('-l',s),cap=()=>tm('capture-pane','-p','-t',pane),all=()=>tm('capture-pane','-p','-S','-','-t',pane),sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const command=async text=>{type(text);await wait(()=>cap().replaceAll('▌','').includes(text),'command input '+text);keys('Enter');await sleep(140)},shot=name=>{writeFileSync(join(out,`${mode}-${width}-${name}.txt`),all());writeFileSync(join(out,`${mode}-${width}-${name}.ansi`),tm('capture-pane','-p','-e','-t',pane))};
 const launch=()=>tm('new-session','-d','-s','proof','-x',String(width),'-y',String(height),`cd '${dir}' && HOME='${dir}' '${bin}' -tui -tui-mode ${mode} -db '${db}' -workspace '${dir}' -model test-model; echo EXIT; sleep 60`);
 try{
  launch();await wait(()=>cap().includes('m0/t0'),'ready');const id=sql('SELECT id FROM sessions LIMIT 1');
  for(let i=0;i<8;i++)sql(`INSERT INTO turns(id,session_id,status,phase,prompt,metadata_json,queue_position,created_at,updated_at) VALUES('q${i}','${id}','queued','queued','native fixture ${i}','{"custom":"retained"}',${i+1},datetime('now'),datetime('now'))`);
  await command('/queue');await wait(()=>all().includes('8 queued'),'durable list');shot('list');
  await command('/queue 2');await wait(()=>all().includes('page 2/2'),'second page');shot('page2');
  keys('C-c');await wait(()=>cap().includes('EXIT'),'exit');tm('kill-session','-t','proof');launch();await wait(()=>cap().includes('test-model'),'reopen');
  await command('/queue');await wait(()=>all().includes('8 queued'),'restart queue');
  await command('/queue remove q0');await wait(()=>sql("SELECT status FROM turns WHERE id='q0'")==='cancelled','remove');await command('/queue remove q0');await wait(()=>all().includes('remove failed'),'stale removal');
  sql(`INSERT INTO turns(id,session_id,status,phase,prompt,metadata_json,created_at,updated_at) VALUES('run','${id}','running','requesting_model','active','{}',datetime('now'),datetime('now')); INSERT INTO session_active_turns(session_id,turn_id,worker_id,claim_token,claimed_at,updated_at) VALUES('${id}','run','fixture','fixture-token',datetime('now'),datetime('now'));`);
  await command('/queue steer q1 stale');await wait(()=>all().includes('steer failed'),'stale steer');assert(sql("SELECT status FROM turns WHERE id='q1'")==='queued','failed steer consumed');
  await command('/queue steer q1 run');await wait(()=>sql("SELECT count(*) FROM steering_queue WHERE session_id='"+id+"' AND json_extract(payload_json,'$.source_queue_id')='q1'")==='1','native steer');
  assert(sql('SELECT count(*) FROM turns')==='9','command created prompt');assert(sql("SELECT json_extract(payload_json,'$.custom') FROM steering_queue WHERE session_id='"+id+"'")==='retained','metadata lost');
  await command('/queue steer q1 run');await wait(()=>all().includes('steer failed'),'duplicate steer');assert(sql('SELECT count(*) FROM steering_queue')==='1','duplicate admission');
  type('unsent queue draft 中文🙂');await sleep(140);shot('after');assert(cap().replaceAll('▌','').includes('unsent queue draft 中文🙂'),'draft lost');assert(readFileSync(join(dir,'.pi/settings.json'),'utf8')===settings,'settings mutated');
  const screen=cap().trimEnd().split('\n'),bars=screen.map((x,i)=>/─{10}/.test(x)?i:-1).filter(i=>i>=0);assert(bars.at(-1)-bars.at(-2)===2,'extra editor rows');
  if(mode==='regular')assert(tm('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag}').trim()==='0 0','regular ownership');
  results.push({mode,width,height,restartQueue:true,nativeRemoveSteer:true,staleRejected:true,noNewTurns:true,metadataPreserved:true,zeroIdleRows:true});
 }catch(error){try{shot('failure')}catch{}throw error}finally{try{tm('kill-server')}catch{}rmSync(dir,{recursive:true,force:true})}
}
writeFileSync(join(out,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
