/**
 * @script Stored assistant source-copy acceptance in fullscreen and regular PTYs.
 * @description Checks exact OSC 52 source bytes, fallback, draft, idle geometry and persisted latest-message selection at three Pi sizes.
 */
import {spawnSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {resolve,join} from 'node:path';

const root=process.cwd(),binary=process.env.GI_TUI_BIN||resolve('bin/gi'),artifacts=resolve('test-results/tui-source-copy');
mkdirSync(artifacts,{recursive:true});
const socket=`gi-source-copy-${process.pid}`;
const run=(command,args)=>{const result=spawnSync(command,args,{encoding:'utf8',timeout:15000});if(result.status!==0)throw Error(`${command} ${args.join(' ')}: ${result.stderr}`);return result.stdout;};
const tmux=(...args)=>run('tmux',['-L',socket,...args]);
const sleep=ms=>new Promise(resolve=>setTimeout(resolve,ms));
const wait=async(check,label)=>{const end=Date.now()+15000;while(Date.now()<end){if(check())return;await sleep(80)}throw Error('Timed out: '+label);};
const assert=(condition,label)=>{if(!condition)throw Error(label);};
const source='\n  # Stored 世界  \n\n```js\n  const value = "<tag>";\n```\n\n';
const quote=value=>`'${value.replaceAll("'","''")}'`;
const outcomes=[];
try{tmux('new-session','-d','-s','keeper','sleep 1200');for(const mode of ['fullscreen','regular'])for(const [width,height] of [[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-source-copy-')),db=join(dir,'gi.db'),session=`source-${mode}-${width}`,pane=session+':0.0';
 mkdirSync(join(dir,'.pi'));writeFileSync(join(dir,'.pi/settings.json'),JSON.stringify({model:'test-model',enabledModels:['test-model','bootstrap'],tuiClipboardMode:'off'}));
 const sql=query=>run('sqlite3',['-cmd','.timeout 5000',db,query]).trim();
 const capture=()=>tmux('capture-pane','-p','-t',pane).replace(/\s+$/,'');
 const history=()=>tmux('capture-pane','-p','-S','-','-t',pane);
 const keys=(...args)=>tmux('send-keys','-t',pane,...args),type=text=>keys('-l',text);
 const bars=text=>text.split('\n').map((line,index)=>/^\s*─{10,}\s*$/.test(line)?index:-1).filter(index=>index>=0);
 const shot=name=>writeFileSync(join(artifacts,`${mode}-${width}x${height}-${name}.txt`),capture()+'\n');
 const launch=()=>tmux('new-session','-d','-s',session,'-x',String(width),'-y',String(height),`cd '${dir}' && TERM=xterm-256color COLORTERM=truecolor '${binary}' -tui ${mode==='regular'?'-tui-mode regular ':''}-db '${db}' -workspace '${dir}' -model test-model 2>'${dir}/runtime.log'; printf '\nCOPY EXITED\n'; sleep 60`);
 try{
  launch();tmux('set-option','-s','set-clipboard','on');tmux('set-option','-t',session,'status','off');
  await wait(()=>capture().includes('m0/t0'),'first startup');
  const sessionID=sql('select id from sessions limit 1;');assert(sessionID,'missing persisted session');
  keys('C-d');await wait(()=>capture().includes('COPY EXITED'),'first exit');tmux('kill-session','-t',session);
  // Seed through the actual SQLite store, preserving leading/trailing spaces and
  // the trailing newline. The latest assistant must win over an earlier reply.
  sql(`insert into messages(id,session_id,role,content,payload_json,created_at) values('source-older',${quote(sessionID)},'assistant','older','{}','2026-01-01'),('source-latest',${quote(sessionID)},'assistant',${quote(source)},'{}','2026-01-02');`);
  launch();tmux('set-option','-t',session,'status','off');await wait(()=>capture().includes('m2/t0'),'reopen persisted messages');
  const idleBars=bars(capture());assert(idleBars.length===2,'idle separator count');
  if(mode==='regular')assert(tmux('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag} #{mouse_button_flag}').trim()==='0 0 0','regular captured terminal history');
  // tmux may not adopt application OSC 52 into its own paste buffer. Capture
  // the real pane byte stream and decode the sequence instead of assuming it does.
  const pty=join(artifacts,`${mode}-${width}-pty.bin`);
  const copied=()=>{try{const raw=readFileSync(pty);const marker=Buffer.from('\x1b]52;c;');const start=raw.indexOf(marker);if(start<0)return null;const end=raw.indexOf(7,start+marker.length);if(end<0)return null;return Buffer.from(raw.subarray(start+marker.length,end).toString(),'base64')}catch{return null}};
  tmux('pipe-pane','-o','-t',pane,`cat > '${pty}'`);type('/copy --osc52');keys('Enter');
  await wait(()=>copied()!==null,'native OSC 52 emission');
  assert(copied().equals(Buffer.from(source)),'OSC 52 clipboard changed source bytes');
  assert(sql('select count(*) from messages;')==='2','copy wrote a message');
  assert(JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'copy added idle rows');
  shot('source-copied');
  type('/copy --off');keys('Enter');await wait(()=>history().includes('clipboard unavailable'),'clipboard-off fallback');
  const copiedOnce=readFileSync(pty).toString('latin1').split('\x1b]52;c;').length-1;
  assert(copiedOnce===1,'clipboard-off emitted another OSC 52 sequence');
  type('unsent source draft 世界');keys('Left','Left','Left');await sleep(130);
  assert(capture().replaceAll('▌','').includes('unsent source draft 世界'),'copy changed editor');
  assert(JSON.stringify(bars(capture()))===JSON.stringify(idleBars),'draft/footer grew idle rows');
  shot('fallback-and-draft');
  outcomes.push(`${mode} ${width}x${height}: latest stored assistant OSC52 source bytes, off fallback, draft and zero added idle rows`);
 }catch(error){try{shot('failed')}catch{};try{writeFileSync(join(artifacts,`${mode}-${width}-runtime.log`),readFileSync(join(dir,'runtime.log')))}catch{};throw error;}
 finally{try{tmux('kill-session','-t',session)}catch{};rmSync(dir,{recursive:true,force:true});}
}}finally{try{tmux('kill-server')}catch{}}
writeFileSync(join(artifacts,'summary.txt'),outcomes.join('\n')+'\n');console.log(outcomes.join('\n'));
