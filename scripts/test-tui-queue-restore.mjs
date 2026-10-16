/** Disposable PTY/SQLite proof: Alt+Up returns two text-only queued turns to the editor without redelivery. */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
if(!process.env.GI_TEST_RUN_ROOT)throw Error('Use make test-env and source scripts/project-test-env.sh');
const binary=process.env.GI_TUI_BIN||resolve('bin/gi'),shell=resolve('tests/ux/shell'),out=join(process.env.GI_TEST_RUN_ROOT,'tests/tui-queue-restore');mkdirSync(out,{recursive:true});
const run=(c,a)=>{const r=spawnSync(c,a,{encoding:'utf8',timeout:12000});if(r.status!==0)throw Error(`${c}: ${r.stderr}`);return r.stdout};
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(f,label)=>{const end=Date.now()+12000;while(Date.now()<end){if(f())return;await sleep(80)}throw Error(`Timed out ${label}`)};
const quote=s=>`'${s.replaceAll("'","''")}'`;
const results=[];
for(const mode of ['fullscreen','regular']){
 const dir=mkdtempSync(join(tmpdir(),'gi-queue-restore-')),db=join(dir,'gi.db'),gates=join(dir,'gates'),socket=`gi-queue-restore-${process.pid}-${mode}`,pane='proof:0.0';mkdirSync(gates);mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model']}));
 const tm=(...a)=>run('tmux',['-L',socket,...a]);const cap=()=>tm('capture-pane','-p','-t',pane).replaceAll('\u00a0',' ');const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const send=(text,key='Enter')=>{tm('send-keys','-t',pane,'-l',text);tm('send-keys','-t',pane,key)};
 const altEnter=()=>tm('send-keys','-t',pane,'-H','1b','5b','31','33','3b','33','75');
 const altUp=()=>tm('send-keys','-t',pane,'-H','1b','5b','31','3b','33','41');
 try{
  tm('new-session','-d','-s','proof','-x','100','-y','32',`cd '${dir}' && HOME='${dir}' PATH='${shell}':"$PATH" GI_UX_QUEUE_GATES='${gates}' '${binary}' -tui -tui-mode ${mode} -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'; sleep 20`);
  tm('set-option','-t','proof','status','off');await wait(()=>cap().includes('%/'),'startup');
  send('UX queue gate:held');await wait(()=>sql("select count(*) from turn_events where event_type='tool.started';")==='1','held active turn');
  const sid=sql('select id from sessions limit 1');
  for(const text of ['restore first','restore second']){tm('send-keys','-t',pane,'-l',text);altEnter();await wait(()=>sql(`select count(*) from turns where prompt=${quote(text)} and status='queued';`)==='1',`queue ${text}`)}
  tm('send-keys','-t',pane,'-l','newer draft');
  await wait(()=>sql(`select json_extract(value,'$.text') from kv_store where namespace='tui_text_draft_v1' and key=${quote(sid)};`)==='newer draft','unsent draft journal');
  altUp();
  const expected='restore first\n\nrestore second\n\nnewer draft';
  await wait(()=>sql(`select json_extract(value,'$.text') from kv_store where namespace='tui_text_draft_v1' and key=${quote(sid)};`)==='restore first\n\nrestore second\n\nnewer draft','restored journal');
  if(sql("select count(*) from turns where prompt in ('restore first','restore second') and status='cancelled' and phase='aborted';")!=='2')throw Error('queued turns were not cancelled');
  if(sql("select count(*) from turns where status='queued';")!=='0')throw Error('durable queue not empty');
  if(sql(`select json_extract(state_json,'$.queue_count') from sessions where id=${quote(sid)};`)!=='0')throw Error('queue_count not cleared');
  if(!cap().includes('newer draft')||!cap().includes('restore first'))throw Error('restored editor not visible');
  writeFileSync(join(out,mode+'-restored.txt'),cap());
  writeFileSync(join(gates,'held'),'go');await wait(()=>sql("select count(*) from turns where status='running';")==='0','active completion');
  if(sql("select count(*) from messages where role='user' and content in ('restore first','restore second');")!=='0')throw Error('restored prompt was redelivered');
  results.push({mode,result:'pass',restored:expected});
 }catch(e){results.push({mode,result:'fail',error:String(e)});try{writeFileSync(join(out,mode+'-failure.txt'),cap());writeFileSync(join(out,mode+'-runtime.log'),readFileSync(join(dir,'runtime.log')))}catch{}}
 finally{writeFileSync(join(gates,'held'),'go');try{tm('kill-server')}catch{}rmSync(dir,{recursive:true,force:true})}
}
writeFileSync(join(out,'results.json'),JSON.stringify({scope:'Actual tmux Alt+Up and SQLite on disposable Gi TUI, shell fixture, no live provider or production TUI',results},null,2));console.log(JSON.stringify({out,results},null,2));if(results.some(r=>r.result==='fail'))process.exitCode=1;
