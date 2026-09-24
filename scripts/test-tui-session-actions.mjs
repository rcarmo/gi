/** @script Native Alt-S action submenu; six fullscreen/regular PTYs, no idle rows. */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';import {join,resolve} from 'node:path';
const bin=process.env.GI_TUI_BIN||resolve('bin/gi'),out=resolve('test-results/tui-session-actions');mkdirSync(out,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd}: ${r.stderr}`);return r.stdout;},sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(v,m)=>{if(!v)throw Error(m)};
async function wait(fn,label){const end=Date.now()+15000;while(Date.now()<end){if(await fn())return;await sleep(80)}throw Error('Timed out: '+label)}
const results=[];
for(const mode of ['fullscreen','regular'])for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-session-actions-')),socket=`gi-actions-${process.pid}-${mode}-${width}`,db=join(dir,'state.db'),pane='proof:0.0';
 const tm=(...a)=>run('tmux',['-L',socket,...a]),keys=(...a)=>tm('send-keys','-t',pane,...a),type=s=>keys('-l',s),cap=()=>tm('capture-pane','-p','-t',pane),hist=()=>tm('capture-pane','-p','-S','-','-t',pane),sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const clean=s=>s.replaceAll('▌',''),select=label=>{const lines=cap().split('\n').map(s=>s.trim());return lines.some(l=>l.startsWith('› ')&&l.slice(2).replace(/^\d+\. /,'')===label)};
 const shot=n=>{writeFileSync(join(out,`${mode}-${width}x${height}-${n}.txt`),cap());writeFileSync(join(out,`${mode}-${width}x${height}-${n}.ansi`),tm('capture-pane','-p','-e','-t',pane));};
 const footprint=()=>{const rows=cap().trimEnd().split('\n'),bars=rows.map((s,i)=>/^\s*─{10,}\s*$/.test(s)?i:-1).filter(i=>i>=0);return JSON.stringify(mode==='regular'?{editor:bars.at(-1)-bars.at(-2),dock:rows.length-bars.at(-2)}:bars)};
 const open=async id=>{await wait(()=>!cap().includes('Select session')&&!cap().includes('Session actions'),'picker closed');keys('M-s');await wait(()=>cap().includes('Select session'),'picker');type(id);await wait(()=>cap().includes('(1 match)'),'filtered picker');await sleep(100);keys('Right');await wait(()=>cap().includes('Session actions'),'actions');};
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model']}));
 try{
  tm('new-session','-d','-s','proof','-x',String(width),'-y',String(height),`cd '${dir}' && HOME='${dir}' TERM=xterm-256color COLORTERM=truecolor '${bin}' -tui -tui-mode ${mode} -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'`);tm('set-option','-t','proof','status','off');await wait(()=>cap().includes('m0/t0'),'startup');
  const main=sql('select id from sessions limit 1;');type('retained native history');keys('Enter');await wait(()=>sql("select count(*) from turns where status='completed';")==='1','history');
  type('/fork actions-child');keys('Enter');await wait(()=>sql('select count(*) from sessions;')==='2'&&cap().includes('@actions-child'),'fork');
  const child=sql(`select id from sessions where parent_session_id='${main}';`);assert(child,'fork not child');
  for(let i=0;i<7;i++){type('/new');keys('Enter');await wait(()=>sql('select count(*) from sessions;')===String(i+3),'extra native session');await sleep(100);}
  type(`/resume ${main}`);keys('Enter');await wait(()=>cap().includes(main),'return root');await sleep(200);
  type('keep draft 中文 β');keys('Left','Left');await sleep(150);const before=footprint(),history=hist(),settings=readFileSync(join(dir,'.pi/settings.json'),'utf8');
  const counts=sql("select (select count(*) from turns)||':'||(select count(*) from messages);");
  keys('M-s');await wait(()=>cap().includes('Select session'),'tall picker');await sleep(100);keys('Right');await wait(()=>cap().includes('Session actions'),'tall-to-short actions');
  const actionRows=cap().split('\n').filter(l=>/^\s*[›*×]?\s*\d+\. /.test(l));assert(actionRows.length<=2&&!actionRows.some(l=>l.includes('@')),'stale picker rows survived submenu shrink');shot('tall-to-short');keys('Escape');await wait(()=>cap().includes('Select session'),'restore tall picker');keys('Escape');
  await open(main);assert(select('Pin')&&!cap().includes('Archive'),'main archive advertised');keys('Escape');await wait(()=>cap().includes('Select session'),'back');assert(cap().includes(main.slice(0,15)),'picker target lost');keys('Escape');
  await open(child);assert(select('Pin')&&cap().includes('Archive'),'child actions absent');shot('actions');
  tm('resize-window','-t','proof','-x','80','-y','24');await sleep(180);tm('resize-window','-t','proof','-x',String(width),'-y',String(height));await sleep(180);assert(select('Pin'),'resize lost selected action');
  keys('Enter');await wait(()=>sql(`select json_extract(state_json,'$.pinned') from sessions where id='${child}';`)==='1','pin');await wait(()=>cap().includes('Select session'),'return after pin');assert(cap().includes('pinned'),'pin label stale');
  keys('Right');await wait(()=>select('Unpin'),'unpin action');keys('Enter');await wait(()=>sql(`select json_extract(state_json,'$.pinned') from sessions where id='${child}';`)==='0','unpin');
  keys('Right');await wait(()=>cap().includes('Session actions'),'archive actions');keys('Down');await wait(()=>select('Archive'),'archive selection');keys('Enter');await wait(()=>sql(`select coalesce(json_extract(state_json,'$.archived_at'),'') from sessions where id='${child}';`)!=='','archive');
  keys('Right');await wait(()=>select('Restore'),'restore only');assert(!cap().includes('Unpin')&&!cap().split('\n').some(l=>/^\s*[› ]+\d+\. Pin\s*$/.test(l)),'archived edit advertised');shot('restore');keys('Enter');await wait(()=>sql(`select coalesce(json_extract(state_json,'$.archived_at'),'') from sessions where id='${child}';`)==='','restore');keys('Escape');await sleep(150);
  assert(clean(cap()).includes('keep draft 中文 β'),'draft lost');type('X');await sleep(120);assert(clean(cap()).includes('keep draft 中文X β'),'cursor moved');keys('BSpace');assert(footprint()===before,'idle footprint grew');
  assert(sql("select (select count(*) from turns)||':'||(select count(*) from messages);")===counts,'action submitted draft/history');assert(readFileSync(join(dir,'.pi/settings.json'),'utf8')===settings,'global settings changed');
  if(mode==='regular')assert(hist().includes('retained native history')&&hist().includes(history.split('\n').find(l=>l.includes('retained native history'))),'native history lost');
  await open(child);keys('Down');await wait(()=>select('Archive'),'archive not available');
  // Native store race: queued work arrives after menu open; source remains idle.
  sql(`insert into turns(id,session_id,status,phase,prompt,metadata_json,created_at,updated_at,queue_position) values('busy-fixture','${child}','queued','queued','later','{}','2026-01-01','2026-01-01',1);`);
  keys('Enter');await wait(()=>cap().includes('action unavailable'),'busy race rejected');assert(sql(`select coalesce(json_extract(state_json,'$.archived_at'),'') from sessions where id='${child}';`)==='','busy archive wrote');keys('Right');await wait(()=>cap().includes('Session actions'),'refresh capabilities');assert(!cap().includes('Archive'),'busy action still offered');keys('Escape');await wait(()=>cap().includes('Select session'),'error back');keys('Escape');await wait(()=>!cap().includes('Select session'),'error close');
  sql("delete from turns where id='busy-fixture';");await sleep(120);assert(clean(cap()).includes('keep draft 中文 β'),'error lost draft');shot('done');
  results.push({mode,width,height,pinUnpinArchiveRestore:true,rootBusyGates:true,draftCursorResize:true,zeroIdleRows:true,historyPreserved:true});
 }catch(e){try{shot('failure');writeFileSync(join(out,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{}throw e;}
 finally{try{tm('kill-server')}catch{}rmSync(dir,{recursive:true,force:true})}
}
writeFileSync(join(out,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
