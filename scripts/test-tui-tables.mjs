import {spawnSync} from 'node:child_process';
import fs from 'node:fs/promises';
import path from 'node:path';
import {generateMessages} from '@cucumber/gherkin';
import {IdGenerator,SourceMediaType} from '@cucumber/messages';
const featurePath=path.resolve('features/gi/tui/markdown-tables.feature');
const definitions=generateMessages(await fs.readFile(featurePath,'utf8'),featurePath,SourceMediaType.TEXT_X_CUCUMBER_GHERKIN_PLAIN,{newId:IdGenerator.incrementing(),includePickles:true});
if(definitions.some(x=>x.parseError)||definitions.filter(x=>x.pickle).length!==6)throw Error('table feature inventory changed');
const binary=path.resolve(process.env.GI_TABLE_PTY_BIN||'bin/gi-tui-table-test');
if(!process.env.GI_TEST_RUN_ROOT)throw Error('Use make test-env and source scripts/project-test-env.sh');
const root=path.join(process.env.GI_TEST_RUN_ROOT,'tests/tui-tables');await fs.mkdir(root,{recursive:true});
function sh(cmd,ok=false){const p=spawnSync('bash',['-lc',cmd],{encoding:'utf8',timeout:12000});if(p.status&&!ok)throw Error(p.stderr||p.stdout||cmd);return p.stdout||'';}
const q=s=>`'${String(s).replaceAll("'","'\\''")}'`,sleep=ms=>new Promise(r=>setTimeout(r,ms));
async function wait(f,label){for(let i=0;i<140;i++){try{if(await f())return}catch{}await sleep(60)}throw Error('timeout '+label)}
const chunks=['| Command | Result |\n','| --- | --- |\n| `go test ./internal/', 'tui` | streamed complete row |\n','| 界面 🙂 | é final row |\n'];
const results=[];
for(const {pickle} of definitions.filter(x=>x.pickle)){
 const mode=pickle.steps[0].text.includes('regular')?'regular':'fullscreen';const size=pickle.steps[0].text.match(/(\d+)x(\d+)/);if(!size)throw Error('missing table fixture size');const cols=+size[1],rows=+size[2];
 const dir=path.join(root,`${mode}-${cols}`),socket=`gi-table-${mode}-${cols}-${process.pid}`;await fs.mkdir(path.join(dir,'.pi'),{recursive:true});await fs.writeFile(path.join(dir,'.pi/settings.json'),JSON.stringify({defaultProvider:'table-local',defaultModel:'table-fixture',tui:{}}));await fs.writeFile(path.join(dir,'chunks.json'),JSON.stringify(chunks));
 const db=path.join(dir,'state.db'),sql=s=>sh(`sqlite3 ${q(db)} ${q(s)}`).trim();const tm=c=>sh(`tmux -L ${socket} ${c}`),pane=()=>tm('capture-pane -pt gi:0 -S -400').replaceAll('\u00a0',' ');
 const launch=()=>tm(`new-session -d -s gi -x ${cols} -y ${rows} "env -u PI_CODING_AGENT_DIR -u GI_CODING_AGENT_DIR TERM=xterm-256color GI_TABLE_PTY_DIR=${q(dir)} GI_TABLE_PTY_MODE=${mode} ${q(binary)} -test.run=^TestMarkdownTablePTYFixture$ 2>${q(path.join(dir,'stderr.log'))}"`);
 const type=text=>tm(`send-keys -t gi:0 -l ${q(text)}`);
 try{
  launch();await wait(()=>pane().includes('─')&&sql('select count(*) from sessions')==='1','ready');type('table response please');tm('send-keys -t gi:0 Enter');await wait(()=>sql("select count(*) from turns where status='running'")==='1','running');type('unsent table draft');
  for(let i=0;i<chunks.length;i++){
   await fs.writeFile(path.join(dir,`gate-${i}`),'go');await wait(()=>fs.stat(path.join(dir,`sent-${i}`)).then(()=>true),'provider chunk');await sleep(240);await fs.writeFile(path.join(dir,`stream-${i}.txt`),pane());
   if(i>=2)await wait(()=>pane().includes('│'),'stream grid');
  }
  // Active resize must recompute grid widths rather than wrap old borders.
  for(const width of [38,cols]){tm(`resize-window -t gi:0 -x ${width} -y ${rows}`);await sleep(240);const view=pane();if(!view.includes('unsent table draft'))throw Error('draft lost on streaming resize');await fs.writeFile(path.join(dir,`resize-${width}.txt`),view);
   if(mode==='fullscreen') {const visible=view.split('\n').slice(-rows);for(const line of visible.filter(l=>/^[ ]*[┌├└│]/.test(l)))if(!/[┐┤┘│][ ]*$/.test(line))throw Error('split table border after resize: '+line);}
   }
  await fs.writeFile(path.join(dir,'finish'),'go');await wait(()=>sql("select count(*) from turns where status='completed'")==='1','completed');await sleep(220);tm('send-keys -t gi:0 Home');await sleep(150);const final=pane();await fs.writeFile(path.join(dir,'final.txt'),final);
  if(!final.includes('┌')||!final.includes('└')||!final.includes('界面')||!final.includes('streamed complete row'))throw Error('completed table missing cells/borders');
  if(final.includes('gi-code-')||final.includes('⟦gi:block:')||final.includes('Command:'))throw Error('raw metadata or stacked fallback leaked');
  if(sql("select content from messages where role='assistant' limit 1")!==chunks.join('').trim())throw Error('raw persisted Markdown changed');
  // Reload through native store projection, not stream state.
  tm('kill-session -t gi');launch();await wait(()=>pane().includes('unsent table draft'),'reload draft');tm('send-keys -t gi:0 Home');await sleep(200);const reloaded=pane();await fs.writeFile(path.join(dir,'reload.txt'),reloaded);if(!reloaded.includes('界面')||!reloaded.includes('└'))throw Error('reload table missing');
  results.push({mode,cols,rows,result:'pass',dir});
 }catch(e){await fs.writeFile(path.join(dir,'failure.txt'),pane()).catch(()=>{});results.push({mode,cols,rows,result:'fail',error:String(e),dir});}
 finally{sh(`tmux -L ${socket} kill-server`,true)}
}
await fs.writeFile(path.join(root,'results.json'),JSON.stringify({scope:'Native TUI engine/provider/SSE deltas/storage, disposable tmux regular/fullscreen and resize. No live provider or production process.',results},null,2));console.log(JSON.stringify({root,results},null,2));if(results.some(r=>r.result!=='pass'))process.exitCode=1;
