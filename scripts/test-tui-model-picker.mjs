#!/usr/bin/env bun
/** @description Verify bounded Alt-M metadata filtering and enabled navigation in six native PTYs. */
import {execFileSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
const binary=process.env.GI_TUI_MODEL_BIN||resolve('bin/gi-tui-model-picker'),artifacts=resolve('test-results/tui-model-picker');mkdirSync(artifacts,{recursive:true});
const socket=`gi-model-${process.pid}`,tmux=(...args)=>execFileSync('tmux',['-L',socket,...args],{encoding:'utf8'}),sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(v,m)=>{if(!v)throw Error(m)};
const wait=async(fn,label)=>{for(let i=0;i<150;i++){if(fn())return;await sleep(60)}throw Error(`timeout: ${label}`)};
const results=[];
for(const mode of ['fullscreen','regular'])for(const [width,height]of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-model-picker-')),db=join(dir,'gi.db'),session=`${mode}-${width}`,pane=session+':0.0';mkdirSync(join(dir,'.pi'));
 const enabled=['test/test-model','opencode-zen/picker-small','opencode-zen/picker-pine','test/unavailable','opencode-zen/picker-piper',...Array.from({length:6},(_,i)=>`opencode-zen/picker-oak${i+1}`),'test/bootstrap'];
 const settings=JSON.stringify({defaultProvider:'test',defaultModel:'test-model',enabledModels:enabled});writeFileSync(join(dir,'.pi/settings.json'),settings);
 const sql=q=>execFileSync('sqlite3',[db,q],{encoding:'utf8'}).trim(),capture=()=>tmux('capture-pane','-p','-t',pane),history=()=>tmux('capture-pane','-p','-S','-','-t',pane),keys=(...k)=>tmux('send-keys','-t',pane,...k),type=s=>tmux('send-keys','-t',pane,'-l',s);
 const bars=()=>capture().split('\n').map((l,i)=>/─{5}/.test(l)?i:-1).filter(i=>i>=0),choice=()=>capture().split('\n').find(l=>l.trimStart().startsWith('› '))||'',open=async()=>{keys('Escape','m');await wait(()=>capture().includes('Select model'),'open');},close=async()=>{keys('Escape');await wait(()=>!capture().includes('Select model'),'close');await sleep(150)};
 const filter=async text=>{await close();await open();type(text);await wait(()=>capture().includes(`search: ${text}`),'filter');await sleep(80)};
 const shot=name=>{writeFileSync(join(artifacts,`${mode}-${width}x${height}-${name}.txt`),capture());writeFileSync(join(artifacts,`${mode}-${width}x${height}-${name}.ansi`),tmux('capture-pane','-p','-e','-t',pane));};
 tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && '${binary}' -db '${db}' -workspace '${dir}' -tui-mode ${mode} 2>'${dir}/runtime.log'`);
 try{
  await wait(()=>capture().includes('test-model'),'boot');await sleep(250);
  type('unsent model draft');keys('Left','Left','Left');await sleep(150);const idle=bars(),historyBefore=history();shot('before');
  await open();await wait(()=>capture().includes('× 2. opencode-zen/picker-small'),'disabled fit');const rows=capture().split('\n').filter(l=>/^\s*(?:›|\*|×)?\s*\d+\. /.test(l));assert(rows.length>0&&rows.length<=6,'selector bound');if(mode==='fullscreen')assert(bars().length===2,'additional idle separators');else assert(tmux('display-message','-p','-t',pane,'#{alternate_on}').trim()==='1','regular selector not isolated');
  keys('Home','Down');await wait(()=>choice().includes('picker-pine'),'skip small');keys('Down');await wait(()=>choice().includes('picker-piper'),'skip unavailable');keys('Home','PPage');await wait(()=>choice().includes('test/test-model'),'page up clamp');keys('NPage');await wait(()=>choice().includes('picker-oak3'),'page enabled');keys('End');await wait(()=>choice().includes('test/bootstrap'),'end');keys('Down');await wait(()=>choice().includes('test/test-model'),'wrap');keys('Up');await wait(()=>choice().includes('test/bootstrap'),'reverse wrap');
  await filter('forest pine');assert(choice().includes('picker-pine'),'display-name filter');assert(!capture().includes('picker-oak'),'nonmatching rows');
  await filter('32k ctx reasoning');assert(choice().includes('picker-pine')&&!capture().includes('picker-piper'),'capability search');shot('filtered');
  tmux('resize-window','-t',session,'-x',String(width+8),'-y',String(height+3));await sleep(130);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(130);assert(capture().includes('32k ctx reasoning'),'resize query');
  await filter('picker-small');assert(!choice(),'disabled-only selected');keys('Enter');await wait(()=>capture().includes('context too small'),'blocked reason');assert(sql("select coalesce(json_extract(state_json,'$.selected_model'),'') from sessions where id='picker-main'")==='','blocked mutation');shot('blocked');
  await filter('no-match-xyz');await wait(()=>capture().includes('no matching models'),'empty');keys('Enter');assert(capture().includes('Select model'),'empty accepted');
  await close();assert(JSON.stringify(bars())===JSON.stringify(idle),'idle footprint');// The renderer hides the hardware cursor; verify the logical insertion point below.
  type('X');await sleep(100);assert(capture().replaceAll('▌','').includes('unsent model drXaft'),'draft/cursor bytes');keys('BSpace');
  await open();type('forest pine');await wait(()=>choice().includes('picker-pine'),'select native');keys('Enter');await wait(()=>!capture().includes('Select model'),'accepted');assert(sql("select json_extract(state_json,'$.selected_model') from sessions where id='picker-main'")==='opencode-zen/picker-pine','native selected model');
  assert(sql("select coalesce(json_extract(state_json,'$.selected_model'),'') from sessions where id='picker-other'")==='','other-session write');assert(sql('select count(*) from turns')==='1','draft submitted');assert(readFileSync(join(dir,'.pi/settings.json'),'utf8')===settings,'global settings changed');
  await open();type('bootstrap');await wait(()=>choice().includes('test/bootstrap'),'restore');keys('Enter');await wait(()=>!capture().includes('Select model'),'restore selected');shot('after');assert(JSON.stringify(bars())===JSON.stringify(idle),'post-selection rows');
  keys('C-e','C-j');type('second line 中文🙂');await sleep(150);const multilineBars=bars();await open();type('forest pine');await wait(()=>choice().includes('picker-pine'),'multiline filter');await close();assert(capture().includes('second line 中文🙂')&&JSON.stringify(bars())===JSON.stringify(multilineBars),'multiline draft/dock');
  if(mode==='regular'){assert(tmux('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag}').trim()==='0 0','regular restored');assert(!history().includes('Select model'),'selector entered terminal history');for(let i=0;i<12;i++){const line=`Picker history ${String(i).padStart(2,'0')}`;assert(historyBefore.includes(line)&&history().includes(line),`history lost ${i}`);}}
  assert(sql('select count(*) from turns')==='1','multiline submitted');shot('multiline');
  results.push(`${mode} ${width}x${height}: metadata/filter/empty/disabled/navigation/resize/native persistence, draft/cursor/idle rows and other-session/settings preserved`);
 }catch(error){try{shot('failure');writeFileSync(join(artifacts,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{}throw error}
 finally{try{tmux('kill-session','-t',session)}catch{}rmSync(dir,{recursive:true,force:true})}
}
// Killing the final session also terminates this isolated tmux server.
writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
