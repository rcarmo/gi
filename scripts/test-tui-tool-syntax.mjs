/** Native read/write highlighting, expand, resize and persisted reopen. */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {resolve,join} from 'node:path';
if(!process.env.GI_TEST_RUN_ROOT)throw Error('Use make test-env and source scripts/project-test-env.sh');
const root=process.cwd(),artifacts=join(process.env.GI_TEST_RUN_ROOT,'tests/tui-tool-syntax'),socket=`gi-tool-syntax-${process.pid}`;
mkdirSync(artifacts,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',['-L',socket,'-f','/dev/null',...args]);
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(fn,label)=>{const end=Date.now()+20000;while(Date.now()<end){if(fn())return;await sleep(80);}throw Error(`Timed out: ${label}`);};
const assert=(v,label)=>{if(!v)throw Error(label);};
const results=[];
try { for(const mode of ['fullscreen','regular']) for(const width of [60,100,140]) {
 const dir=mkdtempSync(resolve('.gi-tool-syntax-')),session=`syntax-${mode}-${width}`,pane=`${session}:0.0`,db=join(dir,'state.db');
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({defaultProvider:'syntax-local',defaultModel:'syntax-fixture',enabledModels:['syntax-local/syntax-fixture']}));
 writeFileSync(join(dir,'read.go'),'package reader\n\nfunc reader() {\n\tprintln("READ-SENTINEL")\n}\n');
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const capture=()=>tmux('capture-pane','-p','-S','-','-t',pane),ansi=()=>tmux('capture-pane','-e','-p','-S','-','-t',pane);
 const keys=(...args)=>tmux('send-keys','-t',pane,...args);
 const shot=name=>{writeFileSync(join(artifacts,`${mode}-${width}-${name}.txt`),capture());writeFileSync(join(artifacts,`${mode}-${width}-${name}.ansi`),ansi());};
 const launch=()=>{tmux('new-session','-d','-s',session,'-x',String(width),'-y','40',`cd '${dir}' && GI_TOOL_SYNTAX_PTY_DIR='${dir}' GI_TOOL_SYNTAX_PTY_MODE='${mode}' TERM=xterm-256color COLORTERM=truecolor '${process.env.GI_TOOL_SYNTAX_TEST_BIN||root+'/bin/gi-tool-syntax-test'}' -test.run '^TestToolSyntaxPTYFixture$' 2>'${dir}/runtime.log'`);tmux('set-option','-s','exit-empty','off');tmux('set-option','-t',session,'status','off');};
 const keyword=word=>new RegExp(`\\x1b\\[[0-9;]*38;2;105;173;208[0-9;]*m${word}`).test(ansi());
 const packageCount=()=>[...ansi().matchAll(/\x1b\[[0-9;]*38;2;105;173;208[0-9;]*mpackage/g)].length;
 try {
  launch();await wait(()=>capture().includes('%/'),'startup');
  keys('-l','highlight fixture');keys('Enter');
  await wait(()=>sql("select count(*) from turns where status='completed';")==='1','completion');
  await wait(()=>capture().includes('WRITE-SENTINEL'),'write preview');
  assert(readFileSync(join(dir,'write.go'),'utf8').includes('WRITE-SENTINEL'),'write did not execute');
  assert(keyword('package'),'write keyword colour absent');
  if(mode==='fullscreen') {
   assert(!capture().includes('READ-SENTINEL'),'successful read not collapsed');shot('collapsed');
   keys('C-o');await wait(()=>capture().includes('READ-SENTINEL'),'expanded read');
  } else {await wait(()=>capture().includes('READ-SENTINEL'),'regular static read');}
  assert(packageCount()>=2,'expanded read and write keyword colours absent');shot('expanded');
  tmux('resize-window','-t',session,'-x','38','-y','40');await sleep(200);
  tmux('resize-window','-t',session,'-x',String(width),'-y','40');await sleep(200);
  assert(capture().includes('WRITE-SENTINEL'),'resize lost source');shot('resized');
  keys('C-c');await sleep(250);try{tmux('kill-session','-t',session);}catch{}
  // Changed filesystem must not change historical displayed source.
  writeFileSync(join(dir,'read.go'),'FILESYSTEM CHANGED');writeFileSync(join(dir,'write.go'),'FILESYSTEM CHANGED');
  launch();await wait(()=>capture().includes('%/'),'reopen');
  if(mode==='fullscreen') {keys('C-o');await wait(()=>capture().includes('READ-SENTINEL'),'reopened expanded read');}
  await wait(()=>capture().includes('WRITE-SENTINEL'),'reopened write');
  assert(packageCount()>=2,'reopened read and write keyword colours absent');assert(!capture().includes('FILESYSTEM CHANGED'),'history reread filesystem');shot('reopened');
  results.push({mode,width,live:true,reload:true,keywordColour:true,resize:true});
 } catch(error) {try{shot('failure');writeFileSync(join(artifacts,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')));}catch{}throw error;}
 finally {try{tmux('kill-session','-t',session);}catch{}rmSync(dir,{recursive:true,force:true});}
}} finally {try{tmux('kill-server');}catch{}}
writeFileSync(join(artifacts,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
