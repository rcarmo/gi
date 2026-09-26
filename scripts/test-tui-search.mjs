/**
 * @script Fullscreen rendered-transcript search and prompt jumps with no idle-row growth.
 * @description Real tmux, native shell turns, separate search editor, query/resize/output/session ownership at three sizes.
 */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';
const root=process.cwd(),bin=process.env.GI_TUI_BIN||resolve('bin/gi'),artifacts=resolve('test-results/tui-search');mkdirSync(artifacts,{recursive:true});
const run=(cmd,args)=>{const r=spawnSync(cmd,args,{encoding:'utf8',timeout:15000});if(r.status!==0)throw Error(`${cmd} ${args.join(' ')}\n${r.stderr}`);return r.stdout;};
const tmux=(...args)=>run('tmux',args),sleep=ms=>new Promise(r=>setTimeout(r,ms));
const wait=async(fn,label)=>{const end=Date.now()+15000;while(Date.now()<end){if(await fn())return;await sleep(80)}throw Error('Timed out: '+label);};
const assert=(v,label)=>{if(!v)throw Error(label)};
const results=[];
for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-search-')),session=`gi-search-${process.pid}-${width}`,pane=session+':0.0',db=join(dir,'gi.db');
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model','bootstrap']}));
 const sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim();
 const capture=()=>tmux('capture-pane','-p','-t',pane).replace(/\s+$/,'');
 const ansi=()=>tmux('capture-pane','-e','-p','-t',pane);
 const keys=(...args)=>tmux('send-keys','-t',pane,...args),type=text=>keys('-l',text);
 const sequence=text=>keys('-H',...[...Buffer.from(text)].map(v=>v.toString(16)));
 const search=()=>sequence('\x1b[102;6u'),previous=()=>sequence('\x1b[1;6A'),next=()=>sequence('\x1b[1;6B');
 const shot=name=>{const text=capture();writeFileSync(join(artifacts,`${width}x${height}-${name}.txt`),text+'\n');writeFileSync(join(artifacts,`${width}x${height}-${name}.ansi`),ansi());return text;};
 const bars=text=>text.split('\n').map((line,i)=>/^\s*─{10,}\s*$/.test(line)?i:-1).filter(i=>i>=0);
 const idle=()=>sql('select count(*) from session_active_turns;')==='0';
 try{
  tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && PATH='${root}/tests/ux/shell':"$PATH" GI_UX_QUEUE_GATES='${dir}' TERM=xterm-256color COLORTERM=truecolor '${bin}' -tui -db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'`);
  tmux('set-option','-t',session,'status','off');await wait(()=>capture().includes('m0/t0'),'startup');
  const wrapNeedle='WrapProbe'+ 'z'.repeat(width+12)+'EndProbe';
  for(let i=1;i<=24;i++){type(`Prompt ${String(i).padStart(2,'0')} nebula nebula${i===12?' '+wrapNeedle:''}`);keys('Enter');await wait(()=>idle()&&sql("select count(*) from turns where status='completed';")===String(i),'history turn');}
  type('unsent editor draft');keys('Left','Left','Left');keys('PageUp','PageUp');await sleep(180);
  const original=shot('before'),idleBars=bars(original),anchor=original.split('\n').find(line=>line.includes('Prompt '))?.replace(/[│█]\s*$/,'').trim();
  search();await wait(()=>capture().includes('Search 0/0'),'search opens');
  assert(!capture().includes('unsent editor'),'main draft visible in search');
  type(wrapNeedle);await wait(()=>capture().includes('Search 1/2'),'two cross-soft-wrap occurrences');shot('cross-wrap');keys('Enter');await wait(()=>capture().includes('Search 2/2'),'next cross-wrap occurrence');keys('C-u');
  type('NEBULA');await wait(()=>capture().includes('Search 1/96'),'all rendered occurrences');
  const active=shot('matches');assert(active.split('\n').length<=height,'search grew terminal footprint');
  assert(ansi().includes('48;2;212;212;212'),'current match not highlighted');
  const textOnly=s=>s.replace(/\x1b\[[0-9;]*m/g,'');
  const matchedRow=()=>ansi().split('\n').find(row=>textOnly(row).includes('Prompt 01'));
  const firstOccurrence=matchedRow();assert(firstOccurrence,'first prompt absent');
  keys('Enter');await wait(()=>capture().includes('Search 2/96'),'same-row next occurrence');
  const secondOccurrence=matchedRow();assert(secondOccurrence&&secondOccurrence!==firstOccurrence,'same-row active highlight did not move');
  assert(textOnly(firstOccurrence)===textOnly(secondOccurrence),'next occurrence changed visible text');
  shot('second-occurrence');sequence('\x1b[13;2u');await wait(()=>capture().includes('Search 1/96'),'previous occurrence');
  assert(matchedRow()===firstOccurrence,'previous occurrence did not restore styles');
  sequence('\x1b[13;2u');await wait(()=>capture().includes('Search 96/96'),'reverse occurrence wrap');keys('Enter');await wait(()=>capture().includes('Search 1/96'),'forward occurrence wrap');
  keys('C-g');await wait(()=>capture().includes('Search 2/96'),'next shortcut');
  keys('C-a','C-k');type('absent-query');await wait(()=>capture().includes('Search: no matches'),'no results');
  keys('Enter');assert(sql('select count(*) from turns;')==='24','search submitted a turn');
  keys('C-a','C-k');type('Prompt 07');await wait(()=>capture().includes('Search 1/2'),'specific query');
  tmux('resize-window','-t',session,'-x','80','-y','24');await sleep(200);tmux('resize-window','-t',session,'-x',String(width),'-y',String(height));await sleep(200);
  assert(capture().includes('Search 1/2'),'resize lost query/matches');
  keys('Escape');await wait(()=>capture().replaceAll('▌','').includes('unsent editor draft'),'Escape restores editor');
  assert(JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'search changed idle rows');
  assert(capture().split('\n').some(line=>line.replace(/[│█]\s*$/,'').trim()===anchor),'Escape lost history anchor');
  type('X');await sleep(120);assert(capture().replaceAll('▌','').includes('unsent editor drXaft'),'search moved saved cursor');keys('BSpace');
  // Reopen tests the persistent component slot, not only initial construction.
  search();await wait(()=>capture().includes('Search 0/0'),'reopen empty');type('Prompt 03');await wait(()=>capture().includes('Search 1/2'),'reopened input works');search();await wait(()=>capture().replaceAll('▌','').includes('unsent editor draft'),'toggle closes');
  keys('Home');await wait(()=>capture().includes('you: Prompt 01'),'Home before prompt jump');next();await sleep(150);assert(capture().includes('you: Prompt 02'),'next marked prompt');previous();await sleep(150);assert(capture().includes('you: Prompt 01'),'previous marked prompt');
  keys('End','C-a','C-k');type('UX queue gate:search');keys('Enter');await wait(()=>!idle(),'gate');type('newer active draft');
  search();await wait(()=>capture().includes('Search 0/0'),'active search');type('Gi received: UX');
  writeFileSync(join(dir,'search'),'release');await wait(()=>idle(),'native completion');await wait(()=>/Search [1-9]\/[1-9]/.test(capture()),'search refreshes on native output');shot('live-match');
  keys('Escape');await wait(()=>capture().includes('newer active draft'),'active query restores newer draft');
  assert(sql("select count(*) from messages where role='user';")==='25','search/navigation sent editor content');
  shot('completed');keys('C-a','C-k');type("!!for i in $(seq 1 24); do printf 'TOOL-LINE-%02d 中文🙂\\n' $i; done");keys('Enter');await wait(()=>capture().includes('TOOL-LINE-24'),'native tool output');
  type('tool reader draft');search();await wait(()=>capture().includes('Search 0/0'),'tool search');type('TOOL-LINE-01');await wait(()=>capture().includes('no matches'),'collapsed text excluded');keys('Escape');await wait(()=>!capture().includes('Search:')&&capture().includes('tool reader draft'),'close before expansion');
  keys('C-o');await wait(()=>capture().includes('F8 collapse'),'tool expanded');search();await wait(()=>capture().includes('Search 0/0'),'expanded search');type('TOOL-LINE-01');await wait(()=>capture().includes('Search 1/1'),'expanded text found');
  keys('C-a','C-k');type('中文🙂');await wait(()=>capture().includes('Search 1/25'),'unicode rendered matches');shot('unicode-tool');keys('Escape');
  await wait(()=>capture().includes('tool reader draft'),'tool editor restored');assert(sql('select count(*) from turns;')==='25','tool search submitted query');
  results.push(`${width}x${height}: per-occurrence literal/case-insensitive/Unicode and cross-soft-wrap search, same-row/multi-row precise highlight/next/previous/wrap/no-match, collapsed/expanded tools, live output/resize/reopen, draft/cursor/history restoration, marked-prompt jumps; no idle-row growth or query submissions`);
 }catch(error){try{shot('failure')}catch{};try{writeFileSync(join(artifacts,`${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{};throw error;}
 finally{try{tmux('kill-session','-t',session)}catch{};rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(artifacts,'summary.txt'),results.join('\n')+'\n');console.log(results.join('\n'));
