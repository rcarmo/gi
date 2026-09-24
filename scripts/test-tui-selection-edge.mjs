/** @script Fullscreen right-edge clipboard proof with real SGR/OSC52 at three sizes. */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
const bin=process.env.GI_TUI_BIN||resolve('bin/gi'),out=resolve('test-results/tui-selection-edge');mkdirSync(out,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd}: ${r.stderr}`);return r.stdout;};
const sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(v,m)=>{if(!v)throw Error(m);};async function wait(fn,label){const end=Date.now()+15000;while(Date.now()<end){if(await fn())return;await sleep(80);}throw Error('Timed out: '+label);}
const results=[];
for(const [width,height]of [[60,18],[100,22],[140,36]])for(const overflow of [false,true]){
 const dir=mkdtempSync(join(tmpdir(),'gi-edge-')),socket=`gi-edge-${process.pid}-${width}-${overflow}`,db=join(dir,'state.db'),pane='proof:0.0';
 const tm=(...a)=>run('tmux',['-L',socket,...a]),keys=(...a)=>tm('send-keys','-t',pane,...a),type=s=>keys('-l',s),cap=()=>tm('capture-pane','-p','-t',pane),sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const sgr=(b,x,y,end='M')=>keys('-H',...Buffer.from(`\x1b[<${b};${x+1};${y+1}${end}`).toString('hex').match(/../g));
 const clipboard=()=>{try{return tm('show-buffer');}catch{return '';}};
 const bands=()=>cap().split('\n').map((s,i)=>/^\s*─{10,}\s*$/.test(s)?i:-1).filter(i=>i>=0);
 const shot=name=>{writeFileSync(join(out,`${width}x${height}-${overflow?'scroll':'no-scroll'}-${name}.txt`),cap());writeFileSync(join(out,`${width}x${height}-${overflow?'scroll':'no-scroll'}-${name}.ansi`),tm('capture-pane','-p','-e','-t',pane));};
 const launch=()=>{tm('new-session','-d','-s','proof','-x',String(width),'-y',String(height),`cd '${dir}' && HOME='${dir}' TERM=xterm-256color COLORTERM=truecolor '${bin}' -tui -workspace '${dir}' -db '${db}' -model test-model 2>'${dir}/runtime.log'`);tm('set-option','-t','proof','status','off');tm('set-option','-g','set-clipboard','on');};
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model'],tuiClipboardMode:'osc52'}));
 try{
  launch();await wait(()=>cap().includes('m0/t0'),'bootstrap');tm('kill-session','-t','proof');await sleep(200);
  const session=sql('select id from sessions limit 1;'),pad=width<80||height<20?0:1,contentWidth=width-2*pad-(overflow?1:0),suffix=overflow?'界':'e\u0301';
  const suffixWidth=overflow?2:1,body='EDGE'+ 'x'.repeat(contentWidth-5-4-suffixWidth)+suffix,expected='sys: '+body;
  // Isolated persisted history is seeded only after Gi has created its schema
  // and session. No live database, input handler or copy API is replaced.
  for(let i=0;i<(overflow?32:1);i++)sql(`insert into messages(id,session_id,role,content,payload_json,created_at) values('edge-${i}','${session}','system','${body}','{}','2026-01-01T00:00:${String(i).padStart(2,'0')}Z');`);
  launch();await wait(()=>cap().includes('sys: EDGE'),'seeded native history');type('draft β middle');keys('Left','Left','Left');await sleep(120);const baseline=JSON.stringify(bands());
  let lines=cap().split('\n'),y=lines.findIndex(l=>l.includes('sys: EDGE')),left=pad,right=pad+contentWidth-1;assert(y>=0,'row absent');
  tm('set-buffer','scrollbar-sentinel');
  if(overflow){sgr(0,pad+contentWidth,y);sgr(32,left,y);sgr(0,left,y,'m');await sleep(100);assert(clipboard()==='scrollbar-sentinel','scrollbar-origin drag copied text');}
  // Scrollbar may move the view; find a fresh visible identical row.
  y=cap().split('\n').findIndex(l=>l.includes('sys: EDGE'));assert(y>=0,'row after scrollbar');
  tm('set-buffer','forward-sentinel');sgr(0,left,y);sgr(32,right,y);sgr(0,right,y,'m');await wait(()=>clipboard()===expected,'forward final cell copy');shot('selected');keys('Escape');await sleep(100);
  tm('set-buffer','reverse-sentinel');sgr(0,right,y);sgr(32,left,y);sgr(0,left,y,'m');await wait(()=>clipboard()===expected,'reverse final cell copy');keys('Escape');await sleep(100);
  assert(cap().replaceAll('▌','').includes('draft β middle'),'selection lost draft');type('X');await sleep(100);assert(cap().replaceAll('▌','').includes('draft β midXdle'),'selection moved cursor');keys('BSpace');
  tm('resize-window','-t','proof','-x','80','-y','24');await sleep(180);tm('resize-window','-t','proof','-x',String(width),'-y',String(height));await sleep(180);assert(cap().replaceAll('▌','').includes('draft β middle'),'resize lost draft');assert(JSON.stringify(bands())===baseline,'extra idle rows');shot('done');
  assert(sql('select count(*) from turns;')==='0','selection submitted draft');results.push({width,height,overflow,forward:true,reverse:true,grapheme:suffix,scrollbarPressSafe:overflow?true:null,draftCursorResize:true,zeroIdleRows:true});
 }catch(e){try{shot('failure');writeFileSync(join(out,`${width}-${overflow}-runtime.log`),readFileSync(join(dir,'runtime.log')));}catch{}throw e;}
 finally{try{tm('kill-server');}catch{}rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(out,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
