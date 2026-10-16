/** @description No-selected-model Return retains exact draft/cursor; explicit picker selection allows one admission, six PTYs. */
import{execFileSync}from'node:child_process';import{mkdtempSync,mkdirSync,writeFileSync,readFileSync,rmSync}from'node:fs';import{tmpdir}from'node:os';import{join,resolve}from'node:path';
if(!process.env.GI_TEST_RUN_ROOT)throw Error('Use make test-env and source scripts/project-test-env.sh');
const bin=process.env.GI_TUI_BIN||resolve('bin/gi'),out=join(process.env.GI_TEST_RUN_ROOT,'tests/tui-unselected-model');mkdirSync(out,{recursive:true});
const run=(cmd,args)=>execFileSync(cmd,args,{encoding:'utf8',timeout:15000}),sleep=ms=>new Promise(r=>setTimeout(r,ms)),assert=(v,m)=>{if(!v)throw Error(m)};
async function wait(fn,label){for(let i=0;i<160;i++){if(fn())return;await sleep(60)}throw Error('timeout: '+label)}
const results=[];
for(const mode of ['fullscreen','regular'])for(const[width,height]of[[60,18],[100,22],[140,36]]){
 const dir=mkdtempSync(join(tmpdir(),'gi-no-model-')),db=join(dir,'state.db'),socket=`gi-no-model-${process.pid}-${mode}-${width}`,pane='proof:0.0';mkdirSync(join(dir,'.pi'));const settings=JSON.stringify({defaultProvider:'test',defaultModel:' ',enabledModels:['test-model']});writeFileSync(join(dir,'.pi/settings.json'),settings);
 const tm=(...a)=>run('tmux',['-L',socket,...a]),keys=(...a)=>tm('send-keys','-t',pane,...a),type=s=>keys('-l',s),cap=()=>tm('capture-pane','-p','-t',pane),all=()=>tm('capture-pane','-p','-S','-','-t',pane),sql=q=>run('sqlite3',['-cmd','.timeout 5000',db,q]).trim(),shot=name=>writeFileSync(join(out,`${mode}-${width}-${name}.txt`),all());
 try{
  tm('new-session','-d','-s','proof','-x',String(width),'-y',String(height),`cd '${dir}' && HOME='${dir}' '${bin}' -tui -tui-mode ${mode} -db '${db}' -workspace '${dir}'; echo EXIT; sleep 60`);await wait(()=>cap().includes('%/'),'ready');
  const draft='unsent 中文🙂 tail';type(draft);keys('Left','Left','Left','Left');await sleep(140);const before=cap().split('\n').find(x=>x.includes('unsent'));assert(before,'draft missing'); // Pi draws the cursor as a reverse-video cell; the X insertion below proves its position.
  keys('Enter');await wait(()=>all().includes('no model selected'),'rejected Return');assert(cap().split('\n').find(x=>x.includes('unsent'))===before,'draft/cursor changed');keys('Enter');await sleep(180);assert(cap().split('\n').find(x=>x.includes('unsent'))===before,'repeat changed draft');assert(sql('SELECT count(*) FROM turns')==='0','rejection admitted');shot('retained');
  keys('M-m');await wait(()=>cap().includes('Ctrl+S to set as default'),'picker');type('test-model');await wait(()=>cap().includes('test-model [test]'),'available option');keys('Enter');await wait(()=>sql("SELECT json_extract(state_json,'$.selected_model') FROM sessions LIMIT 1")==='test-model','model persisted');assert(cap().split('\n').find(x=>x.includes('unsent'))===before,'selection changed draft/cursor');assert(sql('SELECT count(*) FROM turns')==='0','picker submitted');
  type('X');keys('Enter');await wait(()=>sql('SELECT count(*) FROM turns')==='1','explicit admission');await wait(()=>sql("SELECT status FROM turns LIMIT 1")==='completed','completion');assert(sql('SELECT prompt FROM turns LIMIT 1')==='unsent 中文🙂 Xtail','cursor insertion bytes changed');shot('submitted');
  assert(readFileSync(join(dir,'.pi/settings.json'),'utf8')===settings,'global settings changed');if(mode==='regular')assert(tm('display-message','-p','-t',pane,'#{alternate_on} #{mouse_any_flag}').trim()==='0 0','regular ownership');
  const screen=cap().trimEnd().split('\n'),bars=screen.map((x,i)=>/─{10}/.test(x)?i:-1).filter(i=>i>=0);assert(bars.at(-1)-bars.at(-2)===2,'idle editor grew');
  results.push({mode,width,height,repeatedReturnRetained:true,cursorPreserved:true,selectionNoSubmit:true,oneExplicitAdmission:true,exactPrompt:true,globalSettingsUnchanged:true,zeroIdleGrowth:true});
 }catch(error){try{shot('failure')}catch{}throw error}finally{try{tm('kill-server')}catch{}rmSync(dir,{recursive:true,force:true})}
}
writeFileSync(join(out,'summary.json'),JSON.stringify(results,null,2));console.log(JSON.stringify(results,null,2));
