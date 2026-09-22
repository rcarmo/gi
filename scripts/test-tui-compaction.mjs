/**
 * @script Native PTY manual compaction acceptance at 60x18, 100x22, 140x36.
 * @description Uses the production TUI/engine with an opt-in native hook gate.
 */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const root=process.cwd(),artifacts=resolve('test-results/tui-compaction');mkdirSync(artifacts,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',args),sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(fn,label)=>{const end=Date.now()+10000;while(Date.now()<end){if(await fn())return;await sleep(70)}throw Error('Timed out: '+label);};
const assert=(v,label)=>{if(!v)throw Error(label)};
const results=[];
for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-tui-compact-')),session=`gi-compact-${process.pid}-${width}`,pane=session+':0.0',db=join(dir,'gi.db');
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model','bootstrap']}));
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const capture=()=>tmux('capture-pane','-p','-t',pane).replace(/\s+$/,'');
 const keys=(...args)=>tmux('send-keys','-t',pane,...args),type=text=>keys('-l',text);
 const shot=name=>{const text=capture();writeFileSync(join(artifacts,`${width}x${height}-${name}.txt`),text+'\n');return text;};
 const bars=text=>text.split('\n').map((line,i)=>/^\s*─{10,}\s*$/.test(line)?i:-1).filter(i=>i>=0);
 const launch=()=>tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && GI_TUI_COMPACTION_FIXTURE='${dir}' TERM=xterm-256color COLORTERM=truecolor '${root}/bin/gi-tui-compaction-test' -test.run '^TestTerminalCompactionPTYFixture$' 2>'${dir}/runtime.log'`);
 try{
  launch();await wait(()=>capture().includes('m0/t0'),'startup');
  type('untouched draft');keys('Left','Left');keys('M-c');await wait(()=>capture().includes('Compact unavailable'),'empty rejection');assert(sql('select count(*) from turns;')==='0','empty compaction created a turn');assert(capture().includes('untouched dra'),'rejected shortcut lost draft');keys('C-u','C-k');
  type('first native prompt');keys('Enter');await wait(()=>sql("select count(*) from turns where status='completed';")==='1','first turn');
  type('second native prompt');keys('Enter');await wait(()=>sql("select count(*) from turns where status='completed';")==='2','second turn');
  await sleep(4200);type('preserved compaction draft');keys('Left','Left','Left');await sleep(180);
  const idle=shot('before'),idleBars=bars(idle);assert(idleBars.length===2,'idle separators');
  keys('M-c');await wait(()=>sql("select count(*) from turns where phase='compacting';")==='1','compaction started');await wait(()=>capture().includes('Compacting'),'active footer');
  const active=shot('active');assert(JSON.stringify(bars(active))===JSON.stringify(idleBars),'active operation added rows');assert(active.includes('preserved compaction'),'active lost editor');
  keys('M-c');await wait(()=>capture().includes('Compact unavailable'),'busy rejection');assert(sql("select count(*) from turns where json_extract(metadata_json,'$.operation')='manual_compaction';")==='1','duplicate admission');
  tmux('resize-window','-t',session,'-x','80','-y','24');await sleep(200);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(200);
  keys('Escape');await wait(()=>sql("select count(*) from turns where status='cancelled' and json_extract(metadata_json,'$.operation')='manual_compaction';")==='1','cancel');await wait(()=>!capture().includes('Compacting'),'inactive footer');
  assert(sql('select count(*) from context_checkpoints;')==='0','cancel wrote checkpoint');await wait(()=>JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'cancellation notice expired');
  const cancelled=shot('cancelled');assert(cancelled.includes('preserved compaction'),'cancel lost draft');assert(JSON.stringify(bars(cancelled))===JSON.stringify(idleBars),'cancel idle rows changed');
  // Cursor remains three positions from end across compaction and resize.
  type('X');await sleep(120);assert(capture().replaceAll('▌','').includes('drXaft'),'cursor position changed');keys('BSpace');
  keys('M-c');await wait(()=>sql("select count(*) from turns where phase='compacting';")==='1','retry started');writeFileSync(join(dir,'release'),'go');
  await wait(()=>sql('select count(*) from context_checkpoints;')==='1','checkpoint committed');await wait(()=>sql("select count(*) from turns where status='completed' and json_extract(metadata_json,'$.operation')='manual_compaction';")==='1','maintenance completed');await wait(()=>capture().includes('Context compacted'),'completion notice');shot('success-notice');await wait(()=>JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'completion notice expired');
  const success=shot('completed');assert(success.includes('preserved compaction'),'success lost editor');assert(JSON.stringify(bars(success))===JSON.stringify(idleBars),'success added idle rows');assert(sql("select count(*) from messages where role='user';")==='2','maintenance sent draft');
  // Explicit command invokes the same operation; info remains diagnostics.
  keys('C-a','C-k');type('/compact info');keys('Enter');await wait(()=>capture().includes('threshold_tokens'),'info');
  type('/compact');keys('Enter');await wait(()=>sql("select count(*) from turns where phase='compacting';")==='1','command');writeFileSync(join(dir,'release'),'go');await wait(()=>sql("select count(*) from turns where status='completed' and json_extract(metadata_json,'$.operation')='manual_compaction';")==='2','command completed');
  keys('C-c');await sleep(200);try{tmux('kill-session','-t',session)}catch{}
  launch();await wait(()=>capture().includes('m'),'reopen');type('post checkpoint prompt');keys('Enter');await wait(()=>sql("select count(*) from turns where status='completed' and json_extract(metadata_json,'$.operation') is null;")==='3','post-reopen native prompt');
  assert(sql('select count(*) from context_checkpoints;')==='1','checkpoint not durable');shot('restart');
  results.push(`${width}x${height}: native Alt-C/command, empty/busy rejection, Escape cancellation, draft/cursor/resize preservation, success/reopen checkpoint; zero added idle rows`);
 }catch(error){try{shot('failure')}catch{};try{writeFileSync(join(artifacts,`${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{};throw error;}
 finally{try{tmux('kill-session','-t',session)}catch{};rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
