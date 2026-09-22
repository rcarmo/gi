/**
 * @script Native PTY editor/submit/progress reading-position acceptance.
 * @description Exercises 60x18, 100x22, 140x36 using the production shell provider and a process gate.
 */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const root=process.cwd(),artifacts=resolve('test-results/tui-reading');mkdirSync(artifacts,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',args),sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(fn,label)=>{const end=Date.now()+15000;while(Date.now()<end){if(await fn())return;await sleep(80)}throw Error('Timed out: '+label);};
const assert=(v,label)=>{if(!v)throw Error(label)};
const results=[];
for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-tui-reading-')),session=`gi-reading-${process.pid}-${width}`,pane=session+':0.0',db=join(dir,'gi.db');
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model','bootstrap']}));
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const capture=()=>tmux('capture-pane','-p','-t',pane).replace(/\s+$/,'');
 const keys=(...args)=>tmux('send-keys','-t',pane,...args),type=text=>keys('-l',text);
 const shot=name=>{const text=capture();writeFileSync(join(artifacts,`${width}x${height}-${name}.txt`),text+'\n');return text;};
 const bars=text=>text.split('\n').map((line,i)=>/^\s*─{10,}\s*$/.test(line)?i:-1).filter(i=>i>=0);
 const top=text=>{const lines=text.split('\n'),row=lines.findIndex(line=>line.includes('History record'));return row<0?null:JSON.stringify({row,text:lines[row].replace(/[│█]\s*$/,'').trim()});};
 const idle=()=>sql("select count(*) from session_active_turns;")==='0';
 const completed=n=>sql("select count(*) from turns where status='completed';")===String(n)&&idle();
 const launch=()=>tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && PATH='${root}/tests/ux/shell':"$PATH" GI_UX_QUEUE_GATES='${dir}' TERM=xterm-256color COLORTERM=truecolor '${root}/bin/gi' -tui -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'`);
 try{
  launch();tmux('set-option','-t',session,'status','off');await wait(()=>capture().includes('m0/t0'),'startup');
  for(let i=1;i<=22;i++){type(`History record ${String(i).padStart(2,'0')}`);keys('Enter');await wait(()=>completed(i),`history ${i}`);}
  await sleep(250);const baseline=shot('baseline'),idleBars=bars(baseline);assert(idleBars.length===2,'idle separators');
  type('UX queue gate:reading');keys('Enter');await wait(()=>!idle()&&capture().includes('UX queue gate:reading'),'native provider blocked');
  keys('PageUp','PageUp');await sleep(180);const reading=shot('reading'),anchor=top(reading);assert(Boolean(anchor),'history anchor absent');
  type('newer draft');keys('Left','Left','Left');await sleep(150);
  assert(top(capture())===anchor,'typing moved the history reader');
  assert(capture().replaceAll('▌','').includes('newer draft'),'new typing lost');
  writeFileSync(join(dir,'reading'),'release');await wait(()=>completed(23),'native provider completes');await sleep(200);
  const done=shot('completed');assert(top(done)===anchor,'completion moved history anchor');assert(JSON.stringify(bars(done))===JSON.stringify(idleBars),'completion changed idle rows');
  type('X');await sleep(120);assert(capture().replaceAll('▌','').includes('newer drXaft'),'completion moved cursor');keys('BSpace');
  tmux('resize-window','-t',session,'-x','80','-y','24');await sleep(200);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(200);
  assert(top(capture())===anchor,'resize round trip moved history anchor');assert(JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'resize added rows');
  type('X');await sleep(120);assert(capture().replaceAll('▌','').includes('newer drXaft'),'resize moved cursor');keys('BSpace');
  keys('C-a','C-k');type('submitted while reading');keys('Enter');await wait(()=>completed(24),'submit from history');await sleep(180);
  assert(top(capture())===anchor,'submission moved history reader');assert(sql("select count(*) from messages where role='user' and content='submitted while reading';")==='1','submission duplicated');
  // Explicit navigation resumes following. No new button, status row or panel.
  for(let i=0;i<12;i++)keys('PageDown');await sleep(200);
  type('following newest edge');keys('Enter');await wait(()=>completed(25),'follow submission');await wait(()=>capture().includes('Gi received: following newest edge'),'newest response visible');
  const following=shot('following');assert(JSON.stringify(bars(following))===JSON.stringify(idleBars),'following added rows');
  assert(sql("select count(*) from messages where role='user';")==='25','newer unsent draft leaked into store');
  results.push(`${width}x${height}: native delayed provider completion, newer draft/cursor, submission/history anchor, resize round trip, explicit newest-edge following; zero added idle rows`);
 }catch(error){try{shot('failure')}catch{};try{writeFileSync(join(artifacts,`${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{};throw error;}
 finally{try{tmux('kill-session','-t',session)}catch{};rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
