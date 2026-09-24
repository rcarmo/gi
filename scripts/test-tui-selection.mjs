/**
 * @script Fullscreen drag selection/copy/edge scroll at three real PTY sizes.
 * @description Native shell turns and SGR mouse events; OSC 52 captured by an isolated tmux clipboard, not a mock renderer.
 */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const root=process.cwd(),bin=process.env.GI_TUI_BIN||resolve('bin/gi'),artifacts=resolve('test-results/tui-selection'),socket=`gi-selection-${process.pid}`;mkdirSync(artifacts,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',['-L',socket,...args]),sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(fn,label)=>{const end=Date.now()+15000;while(Date.now()<end){if(await fn())return;await sleep(80)}throw Error('Timed out: '+label);};
const assert=(v,label)=>{if(!v)throw Error(label)};
const results=[];
try{for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-selection-')),session=`gi-selection-${width}`,pane=session+':0.0',db=join(dir,'gi.db');
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model','bootstrap'],tuiClipboardMode:'osc52'}));
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const capture=()=>tmux('capture-pane','-p','-t',pane).replace(/\s+$/,'');
 const ansi=()=>tmux('capture-pane','-e','-p','-t',pane);
 const keys=(...args)=>tmux('send-keys','-t',pane,...args),type=text=>keys('-l',text);
 const sequence=text=>keys('-H',...[...Buffer.from(text)].map(v=>v.toString(16)));
 const mouse=(code,x,y,release=false)=>sequence(`\x1b[<${code};${x+1};${y+1}${release?'m':'M'}`);
 const clip=()=>tmux('show-buffer');
 const shot=name=>{const text=capture();writeFileSync(join(artifacts,`${width}x${height}-${name}.txt`),text+'\n');writeFileSync(join(artifacts,`${width}x${height}-${name}.ansi`),ansi());return text;};
 const bars=text=>text.split('\n').map((line,i)=>/^\s*─{10,}\s*$/.test(line)?i:-1).filter(i=>i>=0);
 const idle=()=>sql('select count(*) from session_active_turns;')==='0';
 const drag=async(x1,y1,x2,y2)=>{mouse(0,x1,y1);await sleep(100);mouse(32,x2,y2);await sleep(100);mouse(0,x2,y2,true);};
 try{
  tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && PATH='${root}/tests/ux/shell':"$PATH" GI_UX_QUEUE_GATES='${dir}' TERM=xterm-256color COLORTERM=truecolor '${bin}' -tui -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'`);
  tmux('set-option','-s','set-clipboard','on');tmux('set-option','-t',session,'status','off');await wait(()=>capture().includes('m0/t0'),'startup');
  for(let i=1;i<=22;i++){type(`SELECT-${String(i).padStart(2,'0')} unicode 中文🙂`);keys('Enter');await wait(()=>idle()&&sql("select count(*) from turns where status='completed';")===String(i),'native history');}
  type('newer selection draft');keys('Left','Left','Left');keys('Home');await wait(()=>capture().includes('you: SELECT-01'),'top');
  const baseline=shot('before'),idleBars=bars(baseline),lines=baseline.split('\n'),row=lines.findIndex(l=>l.includes('you: SELECT-01')),col=lines[row].indexOf('you:'),secondRow=lines.findIndex(l=>l.includes('you: SELECT-02'));
  assert(secondRow>row,'second prompt must be visible for padded drag');
  tmux('set-buffer','sentinel');await drag(col,row,col+24,secondRow);await wait(()=>capture().includes('Selection copied'),'release copy');
  const copied=clip();assert(copied.includes('SELECT-01')&&copied.includes('SELECT-02'),'copy missing native rows');assert(!copied.includes('\x1b'),'ANSI in copied selection');
  assert(ansi().includes('48;2;212;212;212'),'highlight absent');shot('selected');
  tmux('set-buffer','reset');keys('C-c');await wait(()=>clip()===copied,'Ctrl-C selection copy');keys('C-x');await wait(()=>clip()===copied,'Ctrl-X selection copy');
  keys('Escape');await wait(()=>!capture().includes('Selection'),'Escape clears');
  assert(JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'selection added idle rows');
  type('X');await sleep(120);assert(capture().replaceAll('▌','').includes('selection drXaft'),'selection changed cursor');keys('BSpace');
  // Drag in reverse copies the same visible rows.
  await drag(col+24,secondRow,col,row);await wait(()=>capture().includes('Selection copied'),'reverse copy');assert(clip()===copied,'reverse differs');keys('Escape');await sleep(120);
  // Hold below the viewport. The timer scrolls without more mouse motion.
  mouse(0,col,row+1);await sleep(100);mouse(32,col+24,idleBars[0]-1);await sleep(2400);mouse(0,col+24,idleBars[0]-1,true);
  await wait(()=>capture().includes('Selection copied'),'edge copy');assert(clip().includes('SELECT-03'),'edge scroll failed to select offscreen output');shot('edge-scroll');
  keys('Escape');await sleep(120);keys('Home');await wait(()=>capture().includes('you: SELECT-01'),'top before resize');
  await drag(col,row,col+12,row+1);await wait(()=>capture().includes('Selection copied'),'resize selection');const beforeResize=clip();
  tmux('resize-window','-t',session,'-x','80','-y','24');await sleep(180);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(180);
  assert(!capture().includes('Selection'),'resize retained invalid selection');assert(clip()===beforeResize,'resize copied unexpected data');
  // Current output invalidates a held selection before a delayed release.
  keys('End','C-a','C-k');type('UX queue gate:select');keys('Enter');await wait(()=>!idle(),'gate');type('newer active draft');keys('Home');await wait(()=>capture().includes('you: SELECT-01'),'history during run');
  mouse(0,col,row);await sleep(90);mouse(32,col+18,row+2);await sleep(90);tmux('set-buffer','no stale copy');writeFileSync(join(dir,'select'),'go');
  await wait(()=>idle(),'native completion');await sleep(250);mouse(0,col+18,row+2,true);await sleep(150);
  assert(clip()==='no stale copy','changed transcript copied stale selection');assert(!capture().includes('Selection'),'arrival retained stale selection');
  assert(capture().replaceAll('▌','').includes('newer active draft'),'arrival lost editor');
  assert(sql("select count(*) from messages where role='user';")==='23','selection submitted draft');shot('completed');
  // A normal click still expands a tool, without clipboard writes.
  keys('End','C-a','C-k');type("!!for i in $(seq 1 24); do printf 'TOOL-%02d\\n' $i; done");keys('Enter');await wait(()=>capture().includes('F8 expand'),'collapsed tool');
  const toolRow=capture().split('\n').findIndex(l=>l.includes('F8 expand'));tmux('set-buffer','click unchanged');
  mouse(0,col+3,toolRow);await sleep(150);mouse(0,col+3,toolRow,true);await wait(()=>capture().includes('F8 collapse'),'plain click expands');assert(clip()==='click unchanged','click copied text');
  // Clipboard opt-out is respected by drag release too.
  type('/copy --off --persist');keys('Enter');await wait(()=>capture().includes('clipboard unavailable'),'clipboard off');
  keys('Home');await wait(()=>capture().includes('you: SELECT-01'),'top off');await drag(col,row,col+12,row+1);await wait(()=>capture().includes('Clipboard off'),'opt-out feedback');assert(clip()==='click unchanged','opt-out emitted OSC52');
  keys('Escape');await sleep(120);type('final unsent draft');
  // View/session transitions discard selections; returning preserves the draft.
  await drag(col,row,col+12,row+1);await wait(()=>capture().includes('Clipboard off'),'selection before picker');
  keys('M-s');await wait(()=>capture().includes('Select session'),'session picker');keys('Escape');await wait(()=>!capture().includes('Select session'),'picker closed');assert(!capture().includes('Selection'),'picker retained selection');
  assert(capture().replaceAll('▌','').includes('final unsent draft'),'picker lost editor');
  results.push(`${width}x${height}: native SGR forward/reverse drag, OSC52 release/Ctrl-C/Ctrl-X clipboard, highlight/Escape/editor preservation, held edge autoscroll, resize/new-output invalidation; zero idle rows`);
 }catch(error){try{shot('failure')}catch{};try{writeFileSync(join(artifacts,`${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{};throw error;}
 finally{try{tmux('kill-session','-t',session)}catch{};rmSync(dir,{recursive:true,force:true});}
}}finally{try{tmux('kill-server')}catch{}}
writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
