#!/usr/bin/env bun
/** @description Six PTYs: session media refs, rejected admission/retry, exact native metadata and no idle chrome. */
import {execFileSync} from 'node:child_process';
import {mkdirSync,readFileSync,writeFileSync} from 'node:fs';
import {join,resolve} from 'node:path';
const bin=resolve(process.env.GI_TUI_BIN||'bin/gi'),out=resolve('test-results/tui-pending-media'),socket=`gi-media-${process.pid}`;mkdirSync(out,{recursive:true});
const tm=(...args)=>execFileSync('tmux',['-L',socket,...args],{encoding:'utf8'}).trimEnd(),sleep=ms=>new Promise(r=>setTimeout(r,ms)),quote=s=>`'${s.replaceAll("'","'\\''")}'`;
const capture=target=>tm('capture-pane','-p','-t',target).replaceAll('▌',''),history=target=>tm('capture-pane','-p','-S','-','-t',target),send=(target,...keys)=>tm('send-keys','-t',target,...keys),type=(target,text)=>tm('send-keys','-t',target,'-l',text),command=async(target,text)=>{send(target,'C-e','C-u');type(target,text);send(target,'Enter');await sleep(180);};
const sql=(db,q)=>execFileSync('sqlite3',[db,q],{encoding:'utf8'}).trim();
async function wait(target,predicate,label){for(let i=0;i<180;i++){if(predicate(capture(target)))return;await sleep(50);}throw Error(`timeout ${label}`);}
const separators=text=>text.split('\n').slice(-5).filter(line=>line.trim().replace(/\x1b\[[0-9;]*m/g,'').match(/^[─━]{3,}$/)).length;
const summary=[];
try{for(const mode of ['fullscreen','regular'])for(const[cols,rows]of [[60,18],[100,22],[140,36]]){
 const key=`${mode}-${cols}x${rows}`,root=join(out,key),db=join(root,'gi.db');mkdirSync(join(root,'.pi'),{recursive:true});mkdirSync(join(root,'.piclaw'),{recursive:true});
 const settings='{"defaultModel":"bootstrap"}\n';writeFileSync(join(root,'.pi/settings.json'),settings);writeFileSync(join(root,'.piclaw/config.json'),'{}\n');
 writeFileSync(join(root,'a.txt'),'native αβ attachment\n');writeFileSync(join(root,'b.txt'),'other session bytes\n');
 // Fresh unique database each run, without deleting retained evidence.
 const liveDB=db.replace('.db',`-${Date.now()}.db`),name=`m-${mode}-${cols}`,target=`${name}:0.0`;
 const launch=`HOME=${quote(root)} ${quote(bin)} -tui -db ${quote(liveDB)} -workspace ${quote(root)} -model bootstrap -tui-mode ${mode} 2>${quote(join(root,'runtime.log'))}; echo media-exit-complete; sleep 60`;
 tm('new-session','-d','-s',name,'-x',String(cols),'-y',String(rows),'bash','-c',launch);
 try{
 await wait(target,t=>t.includes('bootstrap'),'ready');await command(target,'/new');await sleep(200);await command(target,'/new');await sleep(200);
 const ids=sql(liveDB,"select id from sessions order by created_at").split('\n');if(ids.length<2)throw Error('two sessions');const a=ids.at(-2),b=ids.at(-1);
 await command(target,`/switch ${a}`);await sleep(180);const baseline=separators(capture(target));
 await command(target,'/attach a.txt');await wait(target,t=>t.includes('staged for prompt admission'),'stage a');await command(target,'/attachments');await wait(target,t=>t.includes('1 pending'),'list a');
 type(target,'A draft 中文🙂');send(target,'Left');send(target,'Left');
 send(target,'Escape','s');await wait(target,t=>t.includes('Select session'),'switch b picker');type(target,b);await wait(target,t=>t.includes(b.slice(-12))||t.includes('1.'),'filtered b');send(target,'Enter');await wait(target,t=>!t.includes('Select session'),'switched b');
 await command(target,'/attachments');await wait(target,t=>t.includes('0 pending'),'b empty');await command(target,'/attach b.txt');await wait(target,t=>t.includes('staged for prompt admission'),'stage b');type(target,'B draft β');
 send(target,'Escape','s');await wait(target,t=>t.includes('Select session'),'switch a picker');type(target,a);await sleep(100);send(target,'Enter');await wait(target,t=>t.includes('A draft 中文🙂')&&!t.includes('Select session'),'a draft');type(target,'X');await wait(target,t=>t.includes('A draft 中X文🙂'),'cursor');
 // Bracketed paste preserves a real multiline draft while media stays pending.
 send(target,'C-e','C-u');type(target,'\x1b[200~first 中文\nsecond β\x1b[201~');await wait(target,t=>t.includes('first 中文')&&t.includes('second β'),'multiline');
 tm('resize-window','-t',name,'-x',String(cols+12),'-y',String(rows+4));await sleep(130);tm('resize-window','-t',name,'-x',String(cols),'-y',String(rows));await sleep(600);await wait(target,t=>t.includes('second β'),'resize draft');writeFileSync(join(root,'draft.txt'),capture(target));
 send(target,'C-u','BSpace','C-u');await command(target,'/attachments');await wait(target,t=>t.includes('1 pending'),'origin refs');
 sql(liveDB,`CREATE TRIGGER reject_media_admission BEFORE INSERT ON turns BEGIN SELECT RAISE(ABORT,'media admission fixture'); END;`);
 await command(target,'echo rejected media');await wait(target,t=>t.includes('admission rejected; references retained'),'rejected recovery');if(sql(liveDB,'SELECT COUNT(*) FROM turns')!=='0')throw Error('rejection created turn');
 await command(target,'/attachments');await wait(target,t=>t.includes('1 pending'),'recovered refs');writeFileSync(join(root,'rejected.txt'),capture(target));
 sql(liveDB,'DROP TRIGGER reject_media_admission');await command(target,'echo accepted media');await wait(target,()=>Number(sql(liveDB,"SELECT COUNT(*) FROM turns WHERE finished_at IS NOT NULL"))===1,'accepted native');
 await command(target,'/attachments');await wait(target,t=>t.includes('0 pending'),'consumed');
 const turn=JSON.parse(sql(liveDB,`SELECT metadata_json FROM turns WHERE session_id='${a}' LIMIT 1`));if(turn.media?.length!==1||turn.media[0].filename!=='a.txt'||!turn.tui_media_claim)throw Error('native media metadata');
 const bytes=sql(liveDB,`SELECT hex(content) FROM media WHERE id=${turn.media[0].media_id}`);if(bytes!==Buffer.from('native αβ attachment\n').toString('hex').toUpperCase())throw Error('native bytes');
 await command(target,'echo no reuse');await wait(target,()=>Number(sql(liveDB,"SELECT COUNT(*) FROM turns WHERE finished_at IS NOT NULL"))===2,'second native');if(sql(liveDB,"SELECT COUNT(*) FROM turns WHERE json_array_length(json_extract(metadata_json,'$.media'))>0")!=='1')throw Error('media reused');
 await command(target,`/switch ${b}`);await wait(target,t=>t.includes('B draft β'),'b draft');await command(target,'/attachments');await wait(target,t=>t.includes('1 pending'),'b retained');await command(target,'/detach all');await wait(target,t=>t.includes('removed 1 pending'),'detach');await command(target,'/attachments');await wait(target,t=>t.includes('0 pending'),'detached');
 type(target,'final draft β');await sleep(100);if(separators(capture(target))!==baseline)throw Error('idle footprint');if(readFileSync(join(root,'.pi/settings.json'),'utf8')!==settings)throw Error('settings mutated');
 if(sql(liveDB,'SELECT COUNT(*) FROM media')!=='2')throw Error('detach deleted stored media');if(sql(liveDB,`SELECT COUNT(*) FROM turns WHERE session_id='${b}'`)!=='0')throw Error('other session submitted');
 writeFileSync(join(root,'after.txt'),capture(target));writeFileSync(join(root,'after.ansi'),tm('capture-pane','-e','-p','-t',target));writeFileSync(join(root,'native.json'),JSON.stringify({turn,bytes},null,2));
 await command(target,'/attach a.txt');await wait(target,t=>t.includes('staged for prompt admission'),'stage restart');
 send(target,'C-c');await wait(target,t=>t.includes('media-exit-complete'),'exit');tm('kill-session','-t',name);
 const relaunch=async()=>{tm('new-session','-d','-s',name,'-x',String(cols),'-y',String(rows),'bash','-c',launch);await wait(target,t=>t.includes('bootstrap'),'restart ready');await command(target,`/switch ${b}`);};
 await relaunch();await command(target,'/attachments');await wait(target,t=>t.includes('1 pending')&&t.includes('survives restart'),'durable staged refs');
 // Simulate a crash after durable claim acquisition but before admission.
 send(target,'C-c');await wait(target,t=>t.includes('media-exit-complete'),'second exit');tm('kill-session','-t',name);
 sql(liveDB,`UPDATE kv_store SET value=json_object('pending',json('[]'),'claim',json_object('token','pty-unresolved','refs',json_extract(value,'$.pending'))) WHERE namespace='tui_pending_media_v1' AND key='${b}'`);
 await relaunch();await command(target,'/attachments');await wait(target,t=>t.includes('unresolved admission'),'unresolved held');
 type(target,'must not resend');send(target,'Enter');await wait(target,t=>t.includes('must not resend')&&t.includes('retained'),'held claim blocks send');if(sql(liveDB,'SELECT COUNT(*) FROM turns')!=='2')throw Error('restart duplicated send');
 await command(target,'/detach unresolved');await wait(target,t=>t.includes('discarded 1 unresolved'),'explicit discard');await command(target,'/attachments');await wait(target,t=>t.includes('0 pending'),'discarded refs');
 if(sql(liveDB,'SELECT COUNT(*) FROM media')!=='3')throw Error('discard deleted media');writeFileSync(join(root,'restart.txt'),capture(target));
 send(target,'C-c');await wait(target,t=>t.includes('media-exit-complete'),'final exit');summary.push({mode,cols,rows,sessionRefs:true,rejectedRecovery:true,nativeBytes:true,noDuplicate:true,restartStaged:true,unresolvedHeld:true,explicitDiscard:true,zeroIdleRows:true});tm('kill-session','-t',name);
 }catch(error){try{writeFileSync(join(root,'failure.txt'),history(target));}catch{}throw error;}
}}finally{try{execFileSync('tmux',['-L',socket,'kill-server'],{stdio:'ignore'});}catch{}}
writeFileSync(join(out,'summary.json'),JSON.stringify(summary,null,2));console.log(JSON.stringify(summary,null,2));
