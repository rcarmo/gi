/**
 * @script Pi-style regular TUI native terminal scrollback acceptance.
 * @description Checks main-screen output, tmux-native history/selection, no mouse capture, immutable outcomes, bounded idle dock and editor preservation.
 */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const root=process.cwd(),artifacts=resolve('test-results/tui-regular');mkdirSync(artifacts,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',args),sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(fn,label)=>{const end=Date.now()+15000;while(Date.now()<end){if(await fn())return;await sleep(80)}throw Error('Timed out: '+label);};
const assert=(v,label)=>{if(!v)throw Error(label)};
const results=[];
for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-regular-')),session=`gi-regular-${process.pid}-${width}`,pane=session+':0.0',db=join(dir,'gi.db');
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model','bootstrap']}));
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const capture=()=>tmux('capture-pane','-p','-t',pane).replace(/\s+$/,'');
 const history=()=>tmux('capture-pane','-p','-S','-','-t',pane);
 const ansi=()=>tmux('capture-pane','-e','-p','-S','-','-t',pane);
 const flags=()=>tmux('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag} #{mouse_button_flag}').trim();
 const keys=(...args)=>tmux('send-keys','-t',pane,...args),type=text=>keys('-l',text);
 const shot=name=>{const text=capture();writeFileSync(join(artifacts,`${width}x${height}-${name}.txt`),text+'\n');writeFileSync(join(artifacts,`${width}x${height}-${name}.history`),history());return text;};
 const bars=text=>text.split('\n').map((line,i)=>/^\s*─{10,}\s*$/.test(line)?i:-1).filter(i=>i>=0);
 const idle=()=>sql('select count(*) from session_active_turns;')==='0';
 const intact=label=>{const text=history().replace(/\s+/g,' ');assert(JSON.stringify([...text.matchAll(/Gi received: Native regular (\d+)\b/g)].map(m=>Number(m[1])))===JSON.stringify(Array.from({length:12},(_,i)=>i+1)),`${label}: lost, duplicated or reordered native history`);assert(text.includes('Gi received: UX queue gate:regular'),`${label}: lost gate response`);};
 const launch=()=>tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && printf 'PREEXISTING SHELL OUTPUT\\n' && PATH='${root}/tests/ux/shell':"$PATH" GI_UX_QUEUE_GATES='${dir}' TERM=xterm-256color COLORTERM=truecolor '${root}/bin/gi' -tui -tui-mode regular -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'; printf '\\nREGULAR EXITED\\n'; sleep 60`);
 try{
  launch();tmux('set-option','-t',session,'status','off');await wait(()=>capture().includes('m0/t0'),'startup');
  assert(flags()==='0 0 0','regular enabled alternate screen or mouse capture');assert(history().includes('PREEXISTING SHELL OUTPUT'),'startup erased shell history');
  const baseline=shot('idle'),idleBars=bars(baseline);assert(idleBars.length===2,'idle separators');assert(idleBars[1]-idleBars[0]===2,'idle editor height');
  for(let i=1;i<=12;i++) {type(`Native regular ${i}`);keys('Enter');await wait(()=>idle()&&sql("select count(*) from turns where status='completed';")===String(i),'native turn');await wait(()=>history().includes(`Gi received: Native regular ${i}`),'printed turn');}
  const before=history();assert(Number(tmux('display-message','-p','-t',pane,'#{history_size}'))>0,'no native scrollback');
  assert((before.match(/you: Native regular 1\n/g)||[]).length===1,'printed prompt duplicated');
  type('UX queue gate:regular');keys('Enter');await wait(()=>!idle(),'gate');
  type('newer draft');keys('Left','Left','Left');
  tmux('copy-mode','-t',pane);tmux('send-keys','-t',pane,'-X','history-top');
  await wait(()=>Number(tmux('display-message','-p','-t',pane,'#{scroll_position}'))>0,'native copy-mode history');
  // Host-native selection/copy works because Gi never captures mouse or switches screen.
  tmux('send-keys','-t',pane,'-X','start-of-line');tmux('send-keys','-t',pane,'-X','begin-selection');tmux('send-keys','-t',pane,'-X','end-of-line');tmux('send-keys','-t',pane,'-X','copy-selection');
  assert(tmux('show-buffer').includes('PREEXISTING SHELL OUTPUT'),'native selection copy');
  writeFileSync(join(dir,'regular'),'release');await wait(()=>idle(),'complete while reader scrolled');await sleep(200);
  assert(tmux('display-message','-p','-t',pane,'#{pane_in_mode}').trim()==='1','output exited native history');
  assert(Number(tmux('display-message','-p','-t',pane,'#{scroll_position}'))>0,'completion moved native reader');shot('native-history');
  tmux('send-keys','-t',pane,'-X','cancel');await wait(()=>capture().includes('newer dr'),'newer editor preserved');
  await wait(()=>history().replace(/\s+/g,' ').includes('Gi received: UX queue gate:regular'),'accepted result printed');
  type('X');await sleep(120);assert(capture().replaceAll('▌','').includes('newer drXaft'),'completion moved cursor');keys('BSpace');
  keys('Home');type('X');await sleep(120);assert(capture().replaceAll('▌','').includes('Xnewer draft'),'Home should edit in regular mode');keys('BSpace');keys('End');
  intact('before multiline');keys('-H','1b','5b','31','33','3b','32','75');type('second line');await sleep(150);assert(capture().includes('second line'),'multiline editor hidden');
  assert(bars(capture())[1]-bars(capture())[0]===3,'multiline editor did not grow by one row');
  tmux('resize-window','-t',session,'-x','80','-y','24');await sleep(180);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(200);
  intact('after resize');assert(capture().includes('second line'),'resize lost multiline tail');
  const lines=capture().replaceAll('▌','').split('\n');assert(lines.some((line,i)=>line.trim()==='newer draft'&&lines[i+1]?.trim()==='second line'),'multiline rows merged after resize');
  shot('multiline-resized');keys('C-a','C-k');type('newer draft');await sleep(100);
  assert(history().includes('PREEXISTING SHELL OUTPUT'),'resize lost native history');assert(capture().replaceAll('▌','').includes('newer draft'),'resize lost draft');
  intact('before selector');keys('M-m');await wait(()=>capture().includes('Select model'),'temporary selector');keys('Escape');await wait(()=>!capture().includes('Select model')&&capture().replaceAll('▌','').includes('newer draft'),'selector restores editor');
  intact('after selector');keys('C-a','C-k');type('!!printf "regular failed outcome\\n"; exit 7');keys('Enter');await wait(()=>history().includes('exit status 7'),'local error printed');
  assert(ansi().includes('48;2;60;40;40'),'error band absent from native history');assert(ansi().includes('48;2;40;50;40'),'success band absent');
  type("!!for i in $(seq 1 40); do printf 'NATIVE-LONG-%02d\\n' $i; done");keys('Enter');await wait(()=>history().includes('NATIVE-LONG-40'),'full tool output');
  assert(history().includes('NATIVE-LONG-01'),'printed collapsed/truncated preview instead of full retained tool output');
  await sleep(1250);intact('completed redraw');const end=shot('completed');assert(JSON.stringify(bars(end))===JSON.stringify(idleBars),'idle dock gained rows');
  assert(flags()==='0 0 0','regular state changed terminal mode');
  assert(sql("select count(*) from messages where role='user';")==='13','editor operations submitted extra work');
  // Session changes label append-only history; editor snapshots stay local.
  type('/fork regular-child');keys('Enter');await wait(()=>sql('select count(*) from sessions;')==='2','child created');
  type('child draft');keys('M-s');await wait(()=>capture().includes('Select session'),'session selector');type('@agent');await sleep(150);keys('Enter');
  await wait(()=>capture().replaceAll('▌','').includes('m26/t13'),'origin selected');
  assert(!capture().includes('child draft'),'child editor leaked into origin');assert(history().includes('sys: session '),'session boundary missing');
  keys('C-d');await wait(()=>capture().includes('REGULAR EXITED'),'normal exit');assert(history().includes('NATIVE-LONG-01'),'exit erased scrollback');shot('exit');
  tmux('kill-session','-t',session);launch();tmux('set-option','-t',session,'status','off');
  await wait(()=>history().includes('Gi received: Native regular 1'),'reopened durable transcript');
  assert(flags()==='0 0 0','reopen changed terminal mode');shot('reopened');
  results.push(`${width}x${height}: main screen, no mouse capture, ordered/deduplicated native history and host copy/selection during completion; full retained outcome bands; multiline editor/cursor/resize/selector/session/exit/reopen preservation, five-row idle dock`);
 }catch(error){try{shot('failure')}catch{};try{writeFileSync(join(artifacts,`${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{};throw error;}
 finally{try{tmux('kill-session','-t',session)}catch{};rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
