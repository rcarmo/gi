/** @script Native OSC8 link/search and selection acceptance, six PTYs. */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync,existsSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const root=process.cwd(),bin=process.env.GI_TUI_BIN||resolve('bin/gi'),out=resolve('test-results/tui-links');mkdirSync(out,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd}: ${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',args),sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(x,m)=>{if(!x)throw Error(m);};
async function wait(fn,label){const end=Date.now()+15000;while(Date.now()<end){if(await fn())return;await sleep(80);}throw Error('Timed out: '+label);}
const url='https://example.invalid/a',sequence=`\x1b]8;;${url}\x1b\\`,results=[];
for(const mode of ['fullscreen','regular'])for(const [width,height]of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-tui-links-')),socket=`gi-links-${process.pid}-${width}-${mode}`,pane='proof:0.0',db=join(dir,'state.db'),rawPath=join(dir,'raw');
 const tm=(...a)=>tmux('-L',socket,...a),keys=(...a)=>tm('send-keys','-t',pane,...a),type=t=>keys('-l',t),cap=()=>tm('capture-pane','-p','-t',pane),history=()=>tm('capture-pane','-p','-S','-','-t',pane),raw=()=>existsSync(rawPath)?readFileSync(rawPath,'utf8'):'';
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim(),clean=s=>s.replaceAll('▌','');
 const sgr=(code,x,y,end='M')=>{const text=`\x1b[<${code};${x+1};${y+1}${end}`;keys('-H',...Buffer.from(text).toString('hex').match(/../g));};
 const footprint=()=>{const s=cap().trimEnd().split('\n'),rows=s.map((l,i)=>/^\s*─{10,}\s*$/.test(l)?i:-1).filter(i=>i>=0);return mode==='regular'?JSON.stringify({editor:rows.at(-1)-rows.at(-2),dock:s.length-rows.at(-2)}):JSON.stringify(rows);};
 const shot=n=>{writeFileSync(join(out,`${mode}-${width}x${height}-${n}.txt`),history());writeFileSync(join(out,`${mode}-${width}x${height}-${n}.ansi`),tm('capture-pane','-p','-e','-t',pane));};
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model'],tuiClipboardMode:'osc52'}));
 try{
  tm('new-session','-d','-s','proof','-x',String(width),'-y',String(height),`cd '${dir}' && HOME='${dir}' TERM=xterm-256color COLORTERM=truecolor '${bin}' -tui -tui-mode ${mode} -workspace '${dir}' -db '${db}' -model test-model 2>'${dir}/runtime.log'`);
  tm('set-option','-t','proof','status','off');tm('set-option','-g','set-clipboard','on');tm('pipe-pane','-t',pane,`cat >> '${rawPath}'`);await wait(()=>cap().includes('m0/t0'),'startup');
  const baseline=footprint();type(`Visit [DOC](${url})`);keys('Enter');await wait(()=>sql("select count(*) from turns where status='completed';")==='1'&&sql('select count(*) from session_active_turns;')==='0','native response');await wait(()=>raw().includes(sequence),'OSC8 target emitted');
  type('kept draft β');keys('Left','Left');await sleep(160);assert(footprint()===baseline,'idle footprint changed');
  if(mode==='fullscreen'){
   const offset=raw().length;keys('-H','1b','5b','31','30','32','3b','36','75');await wait(()=>cap().includes('Search '),'search open');type('example');await wait(()=>cap().includes('Search 1/2'),'search query');await wait(()=>raw().slice(offset).includes(sequence),'OSC8 lost in search');shot('search');keys('Escape');await wait(()=>!cap().includes('Search '),'search closes');
   const rows=cap().split('\n'),y=rows.findIndex(l=>l.includes(url));assert(y>=0,'visible link');const x=[...rows[y].slice(0,rows[y].indexOf(url))].length;
   const prior=history();sgr(0,x,y);sgr(0,x,y,'m');await sleep(140);assert(history()===prior,'stationary link click altered transcript');
   sgr(0,x,y);sgr(32,x+5,y);sgr(0,x+5,y,'m');await wait(()=>{try{return tm('show-buffer')==='https';}catch{return false;}},'link drag visible-text copy');keys('Escape');
  }else{assert(!raw().includes('\x1b[?1006h'),'regular mouse capture enabled');}
  assert(clean(cap()).includes('kept draft β'),'search/click lost draft');type('X');await sleep(100);assert(clean(cap()).includes('kept draftX β'),'cursor moved');keys('BSpace');
  tm('resize-window','-t','proof','-x','80','-y','24');await sleep(180);tm('resize-window','-t','proof','-x',String(width),'-y',String(height));await sleep(180);assert(clean(cap()).includes('kept draft β'),'resize lost draft');assert(footprint()===baseline,'resize grew idle dock');shot('done');
  assert(sql('select count(*) from turns;')==='1','link gesture submitted a turn');assert(sql("select count(*) from messages where role='user';")==='1','draft dispatched');
  // Force native Markdown projection, where long targets used to be split
  // before OSC8 metadata existed. Do not reconstruct targets from fragments.
  keys('C-a','C-k');const longURL='https://example.invalid/'+('wrapped-path-'.repeat(14))+'end';
  const longSequence=`\x1b]8;;${longURL}\x1b\\`,longStart=raw().length;
  type('# Wrapped');keys('C-j','C-j');type(`Read [DOC](${longURL}).`);keys('Enter');
  await wait(()=>sql("select count(*) from turns where status='completed';")==='2'&&sql('select count(*) from session_active_turns;')==='0','wrapped native response');
  if(mode==='fullscreen')keys('End'); // prior selection deliberately stopped following
  await wait(()=>raw().slice(longStart).includes(longSequence),'complete wrapped OSC8 target emitted');
  type('wrapped link draft');await sleep(120);assert(footprint()===baseline,'wrapped link grew idle dock');shot('wrapped');
  if(mode==='fullscreen'){
   const offset=raw().length;keys('-H','1b','5b','31','30','32','3b','36','75');await wait(()=>cap().includes('Search '),'wrapped search open');type('end');await wait(()=>cap().includes('Search 1/2'),'visible wrapped-link suffix occurrences');await wait(()=>raw().slice(offset).includes(longSequence),'complete target survives wrapped search');shot('wrapped-search');keys('Escape');await wait(()=>!cap().includes('Search '),'wrapped search closes');
  }
  const resizedAt=raw().length;tm('resize-window','-t','proof','-x',String(width+8),'-y',String(height+3));await sleep(180);tm('resize-window','-t','proof','-x',String(width),'-y',String(height));if(mode==='fullscreen')keys('End');await sleep(180);
  if(mode==='fullscreen')await wait(()=>raw().slice(resizedAt).includes(longSequence),'wrapped target survives resize');
  assert(clean(cap()).includes('wrapped link draft'),'wrapped search/resize lost draft');assert(footprint()===baseline,'wrapped resize grew idle dock');
  assert(sql('select count(*) from turns;')==='2','wrapped rendering submitted extra work');
  results.push({mode,width,height,osc8:true,wrappedTarget:true,searchPreservesTarget:mode==='fullscreen',dragCopiesVisibleText:mode==='fullscreen',nativeTerminalActivationOnly:true,draftCursorResize:true,zeroIdleRows:true});
 }catch(error){try{shot('failure');writeFileSync(join(out,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')));}catch{}throw error;}
 finally{try{tm('kill-server');}catch{}rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(out,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
