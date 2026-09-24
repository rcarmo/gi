#!/usr/bin/env bun
/** @description Six-PTY Alt-S resize/selection/cancel acceptance without idle chrome. */
import {execFileSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
const binary=process.env.GI_TUI_BIN||resolve('bin/gi'),artifacts=resolve('test-results/tui-session-picker');mkdirSync(artifacts,{recursive:true});
const socket=`gi-sessions-${process.pid}`,tmux=(...args)=>execFileSync('tmux',['-L',socket,...args],{encoding:'utf8',stdio:['ignore','pipe','pipe']}),sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(v,m)=>{if(!v)throw Error(m)};
const modes=(process.env.GI_SESSION_PICKER_MODES||'fullscreen,regular').split(','),results=[];
for(const mode of modes)for(const [width,height]of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-session-picker-')),db=join(dir,'gi.db'),session=`${mode}-${width}`,pane=session+':0.0';mkdirSync(join(dir,'.pi'));
 const settings=JSON.stringify({defaultProvider:'test',defaultModel:'test-model',enabledModels:['test/test-model','test/bootstrap']});writeFileSync(join(dir,'.pi/settings.json'),settings);
 const sql=q=>execFileSync('sqlite3',[db,q],{encoding:'utf8'}).trim(),capture=()=>tmux('capture-pane','-p','-t',pane),history=()=>tmux('capture-pane','-p','-S','-','-t',pane),keys=(...k)=>tmux('send-keys','-t',pane,...k),type=s=>tmux('send-keys','-t',pane,'-l',s);
 const wait=async(fn,label)=>{for(let i=0;i<170;i++){try{if(fn())return;}catch{}await sleep(60)}throw Error(`timeout ${label}\n${capture()}`)};
 const command=async s=>{type(s);keys('Enter');await sleep(170)};
 const bars=()=>capture().split('\n').map((l,i)=>/─{5}/.test(l)?i:-1).filter(i=>i>=0),rows=()=>capture().split('\n').filter(l=>/^\s*[›*×]?\s*\d+\. /.test(l)),choice=()=>rows().find(l=>l.trimStart().startsWith('›'))||'';
 const open=async()=>{keys('Escape','s');await wait(()=>capture().includes('Select session'),'open');};
 const close=async()=>{keys('Escape');await wait(()=>!capture().includes('Select session'),'close');await sleep(120)};
 const select=async query=>{await open();type(query);await wait(()=>rows().length===1,'one matching session');keys('Enter');await wait(()=>!capture().includes('Select session'),'selection close');await sleep(170)};
 const shot=label=>{for(const [ext,args]of [['txt',[]],['ansi',['-e']]])writeFileSync(join(artifacts,`${mode}-${width}x${height}-${label}.${ext}`),tmux('capture-pane','-p',...args,'-t',pane))};
 tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && '${binary}' -tui -db '${db}' -workspace '${dir}' -tui-mode ${mode} 2>'${dir}/runtime.log'; printf '\\nselector-exit-complete\\n'; sleep 30`);
 try{
  const raw=join(artifacts,`${mode}-${width}x${height}-raw.ansi`);tmux('pipe-pane','-O','-t',pane,`cat > '${raw}'`);
  await wait(()=>capture().includes('test-model'),'startup');await command('/fork @other');await wait(()=>sql('select count(*) from sessions')==='2','fork');
  for(let i=0;i<8;i++)await command(`/fork @extra${i}`);await command('/switch @agent');
  for(let i=0;i<4;i++){await command(`Session history ${i}`);await wait(()=>sql("select count(*) from turns where status='completed'")===String(i+1)&&sql('select count(*) from session_active_turns')==='0','history completion');}
  await sleep(250);const count=sql('select count(*) from turns'),beforeHistory=history();
  type('A draft 中文🙂');keys('Left','Left');await sleep(120);const idle=bars();shot('before');
  await open();assert(rows().length>0&&rows().length<=6,'unbounded selector');shot('open');
  const selected=choice();keys('Down');await wait(()=>choice()!==selected,'down');keys('Up');await wait(()=>choice()===selected,'up');
  keys('End');await wait(()=>choice().includes('10.'),'end');keys('Home');await wait(()=>choice().includes('1.'),'home');keys('NPage');await wait(()=>choice().includes('6.'),'page');keys('PPage');await wait(()=>choice().includes('1.'),'page up');keys('End');await wait(()=>choice().includes('10.'),'end before resize');
  tmux('resize-window','-t',session,'-x',String(width+8),'-y',String(height+3));await sleep(170);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(170);assert(rows().length<=6&&choice().includes('10.'),'resize selected row');
  type('no-such-session');await wait(()=>capture().includes('no matching sessions'),'empty');keys('Enter');assert(capture().includes('Select session'),'empty accepted');await close();
  assert(JSON.stringify(bars())===JSON.stringify(idle),'idle rows changed');type('X');await sleep(100);assert(capture().replaceAll('▌','').includes('A draft 中X文🙂'),'A cursor changed');keys('BSpace');
  await select('@other');assert(!capture().includes('A draft'),'A leaked');type('B draft');keys('Left','Left');
  await select('@agent');await wait(()=>capture().includes('A draft'),'A restored');type('Y');await sleep(100);assert(capture().replaceAll('▌','').includes('A draft 中Y文🙂'),'A cursor not restored');keys('BSpace');
  await select('@other');type('Z');await sleep(100);assert(capture().replaceAll('▌','').includes('B draZft'),'B cursor not restored');keys('BSpace');await select('@agent');
  keys('C-e','C-j');type('second line');await sleep(100);const multiline=bars();await open();type('@other');await close();assert(capture().includes('second line')&&JSON.stringify(bars())===JSON.stringify(multiline),'multiline restoration');
  assert(sql('select count(*) from turns')===count,'selector submitted turn');assert(readFileSync(join(dir,'.pi/settings.json'),'utf8')===settings,'settings changed');
  if(mode==='regular'){
   assert(tmux('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag}').trim()==='0 0','regular flags');
   const after=history();assert(!after.includes('Select session'),'selector contaminated scrollback');for(let i=0;i<4;i++)assert(beforeHistory.includes(`Session history ${i}`)&&after.includes(`Session history ${i}`),'history lost');
  }
  shot('after');await open();shot('exit-open');keys('C-c');await wait(()=>capture().includes('selector-exit-complete'),'exit');assert(tmux('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag}').trim()==='0 0','exit flags');
  if(mode==='regular'){await sleep(100);assert(!readFileSync(raw,'utf8').includes('\x1b[3J'),'scrollback erase');assert(!history().includes('Select session'),'exit left selector history');}
  results.push(`${mode} ${width}x${height}: ≤6 results, all navigation/resize/cancel/select, A/B Unicode draft+cursor, multiline/idle rows/history, no turn/settings writes, exit with selector open`);
 }catch(error){try{shot('failure');writeFileSync(join(artifacts,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{}throw error}
 finally{try{tmux('kill-session','-t',session)}catch{}rmSync(dir,{recursive:true,force:true})}
}
writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
