/**
 * @script Verify bounded explicit terminal index actions in fullscreen/regular modes.
 * @description Native tmux/SQLite acceptance at three sizes, with draft/cursor/reader preservation and no idle rows.
 */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,renameSync,rmSync,readFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const artifacts=resolve('test-results/tui-index');mkdirSync(artifacts,{recursive:true});
const bin=resolve(process.env.GI_TUI_BIN||'bin/gi');
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',['-L',`gi-index-harness-${process.pid}`,...args]),sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(fn,label)=>{for(let end=Date.now()+12000;Date.now()<end;){if(await fn())return;await sleep(70)}throw Error('Timed out: '+label)};
const assert=(v,s)=>{if(!v)throw Error(s)};const results=[];
for(const mode of (process.env.GI_INDEX_TEST_MODE?[process.env.GI_INDEX_TEST_MODE]:['fullscreen','regular']))for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-tui-index-')),session=`gi-index-${process.pid}-${mode}-${width}`,pane=session+':0.0',db=join(dir,'gi.db');
 for(const path of ['.pi/skills','notes'])mkdirSync(join(dir,path),{recursive:true});
 writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model'}));writeFileSync(join(dir,'notes/a.md'),'indexorchid source');
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const capture=()=>tmux('capture-pane','-p','-t',pane).replace(/\s+$/,'');
 const keys=(...args)=>tmux('send-keys','-t',pane,...args),type=s=>keys('-l',s);
 const open=()=>keys('Escape','i');
 const shot=name=>{const txt=capture();writeFileSync(join(artifacts,`${mode}-${width}x${height}-${name}.txt`),txt+'\n');writeFileSync(join(artifacts,`${mode}-${width}x${height}-${name}.ansi`),tmux('capture-pane','-e','-p','-t',pane));return txt};
 const bars=txt=>txt.split('\n').map((s,i)=>/^\s*─{10,}\s*$/.test(s)?i:-1).filter(i=>i>=0);
 const count=()=>sql("select count(*) from messages where role='user'");
 const terminalHistory=()=>tmux('capture-pane','-p','-S','-','-t',pane);
 try{
  tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && TERM=xterm-256color COLORTERM=truecolor '${bin}' -tui -tui-mode ${mode} -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'`);
  tmux('pipe-pane','-t',pane,'-o',`cat > '${join(artifacts,`${mode}-${width}x${height}-raw.ansi`)}'`);tmux('set-option','-s','exit-empty','off');tmux('set-option','-t',session,'status','off');await wait(()=>capture().includes('m0/t0'),'startup');
  assert(sql('select count(*) from workspace_index_workspaces')==='0','startup scanned');
  for(let i=1;i<=12;i++){type(`Index-history-${String(i).padStart(2,'0')}`);keys('Enter');await wait(()=>sql("select count(*) from turns where status='completed'")===String(i)&&sql('select count(*) from session_active_turns')==='0','history');}
  type('unsent index draft');keys('Left','Left','Left');if(mode==='fullscreen')keys('PageUp','PageUp');await sleep(170);
  if(mode==='regular'){const history=terminalHistory();writeFileSync(join(artifacts,`${mode}-${width}x${height}-history-before.txt`),history);for(let n=1;n<=12;n++)assert(history.includes(`Gi: Gi received: Index-history-${String(n).padStart(2,'0')}`),`regular initial history absent: ${n}`);}
  const before=shot('before'),idleBars=bars(before),anchor=before.split('\n').find(l=>l.includes('Index-history-'))?.replace(/[│█]\s*$/,'').trim();
  open();await wait(()=>capture().includes('State: never_indexed'),'status opens');
  type('ordinary typing must not edit');keys('Enter');await sleep(180);await wait(()=>capture().includes('State: never_indexed'),'Enter defaults to status');
  assert(sql('select count(*) from workspace_index_workspaces')==='0','read action scanned');
  keys('Down','Enter');await wait(()=>capture().includes('State: ready'),'refresh');
  assert(sql("select committed_generation from workspace_index_scopes where scope='all'")==='1','native generation');
  assert(sql("select count(*) from workspace_index_fts where workspace_index_fts match 'indexorchid'")==='1','native bytes missing');
  const active=shot('ready');assert(active.split('\n').length<=height,'panel overflow');
  assert(bars(active).length===(mode==='regular'?0:2),'temporary dock separators');
  tmux('resize-window','-t',session,'-x','80','-y','24');await sleep(180);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(180);
  assert(capture().includes('State: ready'),'resize lost panel');keys('Escape');await wait(()=>!capture().includes('Index ·'),'closed');
  assert(JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'idle row growth');
  if(mode==='fullscreen')assert(capture().split('\n').some(l=>l.replace(/[│█]\s*$/,'').trim()===anchor),'reader anchor lost');
  else {const history=terminalHistory();writeFileSync(join(artifacts,`${mode}-${width}x${height}-history.txt`),history);for(let n=1;n<=12;n++)assert(history.includes(`Gi: Gi received: Index-history-${String(n).padStart(2,'0')}`),`regular history lost: ${n}`);assert(!history.includes('Index · all'),'temporary panel entered scrollback');}
  type('X');await sleep(150);assert(capture().replaceAll('▌','').includes('unsent index drXaft'),'draft/cursor changed');keys('BSpace');
  open();await wait(()=>capture().includes('State: ready'),'reopen');keys('Right');await wait(()=>capture().includes('Index · notes')&&capture().includes('never_indexed'),'scope select');
  keys('Left');await wait(()=>capture().includes('Index · all')&&capture().includes('State: ready'),'scope back');
  renameSync(join(dir,'notes'),join(dir,'held'));keys('Down','Enter');await wait(()=>capture().includes('State: error'),'native failure');shot('failure');
  assert(sql("select committed_generation from workspace_index_scopes where scope='all'")==='1','failure changed generation');
  renameSync(join(dir,'held'),join(dir,'notes'));keys('Enter');await wait(()=>capture().includes('State: ready'),'retry');
  assert(sql("select committed_generation from workspace_index_scopes where scope='all'")==='2','retry generation');keys('Escape');await sleep(150);
  assert(count()==='12','index action submitted draft');assert(JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'retry idle rows');shot('after');
  // Peer-held refresh proves immediate close/reopen while pending; no duplicate
  // admission and no late error transcript after the bounded batch finishes.
  sql("insert into workspace_index_leases select id,'held-peer',cast((julianday('now')-2440587.5)*86400000 as integer)+60000 from workspace_index_workspaces");
  open();await wait(()=>capture().includes('State: ready'),'open before busy');keys('Down','Enter');await wait(()=>capture().includes('State: working'),'held refresh');keys('Escape');await wait(()=>!capture().includes('Index ·'),'escape busy');
  await sleep(2100);assert(!capture().includes('refresh remains pending'),'late error leaked');sql("delete from workspace_index_leases where owner_token='held-peer'");
  open();await wait(()=>capture().includes('State: ready'),'open after busy');keys('Escape');await sleep(100);
  assert(count()==='12','busy close sent draft');
  keys('Escape','s');await wait(()=>capture().includes('Select session'),'session selector after panel');await sleep(180);keys('Escape');await wait(()=>!capture().includes('Select session'),'close session selector');await sleep(180);
  assert(capture().replaceAll('▌','').includes('unsent index draft'),'selector lost draft');
  if(mode==='regular')writeFileSync(join(artifacts,`${mode}-${width}-pre-multiline.txt`),terminalHistory());
  keys('C-e','C-j');type('second line');await sleep(150);
  const multilineBars=bars(capture());open();await wait(()=>capture().includes('State: ready'),'multiline open');keys('Escape');await sleep(150);
  assert(capture().includes('second line')&&JSON.stringify(bars(capture()))===JSON.stringify(multilineBars),'multiline dock restoration');
  if(mode==='regular')writeFileSync(join(artifacts,`${mode}-${width}-post-multiline.txt`),terminalHistory());
  keys('C-a','C-k','BSpace','C-a','C-k');type('Index-history-13');keys('Enter');
  await wait(()=>sql("select count(*) from turns where status='completed'")==='13'&&sql('select count(*) from session_active_turns')==='0','post-panel completion');await sleep(250);
  if(mode==='regular'){
   const history=terminalHistory();writeFileSync(join(artifacts,`${mode}-${width}-post-submit.txt`),history);for(let n=1;n<=13;n++)assert(history.includes(`Gi: Gi received: Index-history-${String(n).padStart(2,'0')}`),`post-panel history missing ${n}`);
   assert(!history.includes('Index · all'),'post-panel history contaminated');
   assert(tmux('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag}').trim()==='0 0','regular mode not restored');
   assert(!readFileSync(join(artifacts,`${mode}-${width}x${height}-raw.ansi`),'utf8').includes('\x1b[3J'),'scrollback clear emitted');
  }
  results.push(`${mode} ${width}x${height}: status/refresh/bytes/failure/retry/scope/resize/reopen/busy-close; draft/cursor/reader and idle rows retained; no chat submission`);
 }catch(e){try{shot('failed-acceptance');writeFileSync(join(artifacts,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{}throw e}
 finally{try{tmux('kill-session','-t',session)}catch{}rmSync(dir,{recursive:true,force:true})}
}
tmux('kill-server');writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
