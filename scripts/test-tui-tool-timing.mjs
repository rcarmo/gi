/** @script Native tool timing acceptance in six fullscreen/regular PTYs. */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const root=process.cwd(),bin=process.env.GI_TUI_BIN||resolve('bin/gi'),artifacts=resolve('test-results/tui-tool-timing');mkdirSync(artifacts,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',args),sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(v,m)=>{if(!v)throw Error(m);};
async function wait(fn,label){const end=Date.now()+16000;while(Date.now()<end){if(await fn())return;await sleep(80);}throw Error('Timed out: '+label);}
const results=[];
for(const mode of ['fullscreen','regular'])for(const [width,height]of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-tool-timing-')),session=`gi-tools-${process.pid}-${mode}-${width}`,pane=session+':0.0',db=join(dir,'state.db');
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model','bootstrap']}));
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim(),cap=()=>tmux('capture-pane','-p','-t',pane),hist=()=>tmux('capture-pane','-p','-S','-','-t',pane),ansi=()=>tmux('capture-pane','-p','-e','-S','-','-t',pane);
 const keys=(...v)=>tmux('send-keys','-t',pane,...v),type=t=>keys('-l',t),plain=t=>t.replaceAll('▌','');
 const shot=name=>{writeFileSync(join(artifacts,`${mode}-${width}x${height}-${name}.txt`),hist());writeFileSync(join(artifacts,`${mode}-${width}x${height}-${name}.ansi`),ansi());};
 const idle=()=>sql('select count(*) from session_active_turns;')==='0';
 const barRows=()=>cap().split('\n').map((l,i)=>/^\s*─{10,}\s*$/.test(l)?i:-1).filter(i=>i>=0);
 const footprint=()=>{const rows=barRows();if(mode==='fullscreen')return JSON.stringify(rows);const end=cap().trimEnd().split('\n').length;return JSON.stringify({editor:rows.at(-1)-rows.at(-2),dock:end-rows.at(-2)});};
 const line=()=>hist().split('\n').filter(l=>l.includes('shell')&&/\d+(?:\.\d+)?(?:ms|s)/.test(l)).at(-1)?.trim();
 try{
  tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && HOME='${dir}' PATH='${root}/tests/ux/shell':"$PATH" GI_UX_QUEUE_GATES='${dir}' TERM=xterm-256color COLORTERM=truecolor '${bin}' -tui -tui-mode ${mode} -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'`);tmux('set-option','-t',session,'status','off');
  await wait(()=>cap().includes('m0/t0'),'startup');const rows=barRows();assert(rows.length===2,'initial editor bands');const idleFootprint=footprint();writeFileSync(join(artifacts,`${mode}-${width}-baseline.txt`),cap());
  type('UX queue gate:timed');keys('Enter');await wait(()=>!idle()&&sql("select count(*) from turn_events where event_type='tool.started';")==='1','tool start');
  type('draft β middle');keys('Left','Left','Left');await sleep(1200);
  if(mode==='fullscreen'){await wait(()=>Boolean(line()),'live tool elapsed');const first=line();await wait(()=>line()!==first,'elapsed tick');}else{assert(!hist().includes('tool running'),'regular mode printed mutable tool block');}
  shot('running');tmux('resize-window','-t',session,'-x','80','-y','24');await sleep(180);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(180);
  assert(plain(cap()).includes('draft β middle'),'resize lost draft');type('X');await sleep(100);assert(plain(cap()).includes('draft β midXdle'),'resize lost cursor');keys('BSpace');
  writeFileSync(join(dir,'timed'),'release');await wait(()=>idle()&&sql("select count(*) from turns where status='completed';")==='1','completion');await wait(()=>Boolean(line())&&hist().includes('Gi received:'),'final tool output');
  const final=line();await sleep(1250);assert(line()===final,'terminal duration drift');assert(plain(cap()).includes('draft β middle'),'completion lost draft');assert(footprint()===idleFootprint,`completion footprint ${mode}-${width}: ${footprint()} != ${idleFootprint}`);
  const eventTimes=JSON.parse(run('sqlite3',['-json',db,"select event_type,created_at from turn_events where event_type in ('tool.started','tool.finished') order by seq;"]));const ms=Date.parse(eventTimes[1].created_at)-Date.parse(eventTimes[0].created_at);assert(ms>=1000,'native event interval absent');
  const match=final.match(/([\d.]+)(ms|s)/);assert(match,'no frozen elapsed');const rendered=Number(match[1])*(match[2]==='s'?1000:1);assert(Math.abs(rendered-ms)<500,'elapsed differs from event timestamps');shot('completed');
  keys('C-a','C-k');type('UX tool fail:terminal');keys('Enter');await wait(()=>idle()&&sql("select count(*) from turns where status='failed';")==='1','native failure');await wait(()=>/shell.*error|error.*shell/.test(hist()),'failed block');
  type('after failure β');const failed=line();await sleep(1200);assert(line()===failed,'failure duration drift');assert(plain(cap()).includes('after failure β'),'failed block consumes editor');assert(footprint()===idleFootprint,'failed block added idle rows');shot('failed');
  assert(sql("select count(*) from turn_events where event_type='tool.started';")==='2','start count');assert(sql("select count(*) from turn_events where event_type in ('tool.finished','tool.failed');")==='2','terminal count');
  if(mode==='regular'){assert(!hist().includes('tool running'),'mutable tool text leaked into scrollback');assert((hist().match(/Gi received:/g)||[]).length>=1,'immutable output missing');}
  results.push({mode,width,height,inlineTiming:true,nativeTimestampMatch:true,frozenCompletion:true,frozenFailure:true,resizeCursorDraft:true,zeroIdleRows:true});
 }catch(error){try{shot('failure');writeFileSync(join(artifacts,`${mode}-${width}-screen.txt`),cap());writeFileSync(join(artifacts,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')));}catch{}throw error;}
 finally{try{tmux('kill-session','-t',session);}catch{}rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(artifacts,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
