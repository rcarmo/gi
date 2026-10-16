/** @description Long Unicode editor viewport, cursor/menu/resize and exact native submission in six PTYs. */
import {execFileSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
if(!process.env.GI_TEST_RUN_ROOT)throw Error('Use make test-env and source scripts/project-test-env.sh');
const bin=process.env.GI_TUI_BIN||resolve('bin/gi'),out=join(process.env.GI_TEST_RUN_ROOT,'tests/tui-editor-viewport');mkdirSync(out,{recursive:true});
const run=(cmd,args)=>execFileSync(cmd,args,{encoding:'utf8',timeout:15000}),sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(x,m)=>{if(!x)throw Error(m)};
async function wait(fn,label){for(let i=0;i<150;i++){if(fn())return;await sleep(70)}throw Error('timeout: '+label)}
const results=[];
for(const mode of ['fullscreen','regular'])for(const [width,height]of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-editor-view-')),db=join(dir,'state.db'),socket=`gi-editor-${process.pid}-${mode}-${width}`,pane='proof:0.0';
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({defaultProvider:'test',defaultModel:'test-model',enabledModels:['test-model']}));
 const tm=(...args)=>run('tmux',['-L',socket,...args]),keys=(...args)=>tm('send-keys','-t',pane,...args),type=s=>keys('-l',s),cap=()=>tm('capture-pane','-p','-t',pane),clean=()=>cap().replaceAll('▌',''),sql=q=>run('sqlite3',[db,q]).trim();
 const bars=()=>cap().split('\n').map((l,i)=>/─{10}/.test(l)?i:-1).filter(i=>i>=0),editorRows=()=>{const b=bars();assert(mode==='regular'?b.length>=2:b.length===2,'separator count');return b.at(-1)-b.at(-2)-1};
 const shot=name=>{writeFileSync(join(out,`${mode}-${width}-${name}.txt`),cap());writeFileSync(join(out,`${mode}-${width}-${name}.ansi`),tm('capture-pane','-p','-e','-t',pane))};
 try{
  tm('new-session','-d','-s','proof','-x',String(width),'-y',String(height),`cd '${dir}' && HOME='${dir}' TERM=xterm-256color '${bin}' -tui -tui-mode ${mode} -workspace '${dir}' -db '${db}' -model test-model 2>'${dir}/runtime.log'`);tm('set-option','-t','proof','status','off');
  await wait(()=>cap().includes('%/'),'boot');assert(editorRows()===1,'idle rows');
  const lines=Array.from({length:32},(_,i)=>`DRAFT-${String(i).padStart(2,'0')} 中文🙂 e\u0301 `+'x'.repeat(width+11)),draft=lines.join('\n');
  for(let i=0;i<lines.length;i++){if(i)keys('C-j');type(lines[i])}
  await wait(()=>cap().includes('DRAFT-31'),'tail visible');assert(editorRows()<=Math.max(5,Math.floor(height*.3)),'editor took transcript');assert(!cap().includes('DRAFT-00'),'unbounded long draft');shot('tail');
  // Pi Ctrl+A/E move within a line; Up/Down navigate the multiline document.
  keys('C-a');type('X');await wait(()=>clean().includes('XDRAFT-31'),'line-home insertion');keys('BSpace');
  for(let line=0;line<200;line++)keys('Up');keys('C-a');await wait(()=>cap().includes('DRAFT-00'),'first line visible');type('X');await wait(()=>clean().includes('XDRAFT-00'),'first-line insertion');keys('BSpace');
  for(let line=0;line<200;line++)keys('Down');keys('C-e');await wait(()=>cap().includes('DRAFT-31'),'last line visible');type('Y');keys('BSpace');
  const beforeMenu=editorRows();keys('M-m');await wait(()=>cap().includes('Ctrl+S to set as default'),'picker');keys('Escape');await wait(()=>!cap().includes('Ctrl+S to set as default'),'picker close');await wait(()=>editorRows()===beforeMenu,'picker editor restored');
  tm('resize-window','-t','proof','-x','42','-y','14');await sleep(220);for(let line=0;line<200;line++)keys('Up');keys('C-a');await wait(()=>cap().includes('DRAFT-00'),'small first line');assert(editorRows()<=5,'small editor overflow');for(let line=0;line<200;line++)keys('Down');keys('C-e');await wait(()=>cap().includes('DRAFT-31'),'small last line');shot('small');
  tm('resize-window','-t','proof','-x',String(width),'-y',String(height));await sleep(220);assert(editorRows()<=Math.max(5,Math.floor(height*.3)),'restored viewport');
  assert(sql('select count(*) from turns;')==='0','navigation submitted draft');keys('Enter');await wait(()=>sql("select count(*) from turns where status='completed';")==='1','submit once');
  const stored=JSON.parse(sql("select json_quote(content) from messages where role='user' order by rowid desc limit 1;"));assert(stored===draft,'draft bytes changed/truncated');await wait(()=>editorRows()===1,'idle shrink');shot('submitted');
  if(mode==='regular')assert(tm('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag}').trim()==='0 0','regular mouse/history ownership');
  results.push({mode,width,height,maxEditorRows:Math.max(5,Math.floor(height*.3)),exactDraftBytes:true,oneExplicitSubmission:true,cursorMenuResize:true,idleRows:1});
 }catch(error){try{shot('failure');writeFileSync(join(out,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{}throw error}
 finally{try{tm('kill-server')}catch{}rmSync(dir,{recursive:true,force:true})}
}
writeFileSync(join(out,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
