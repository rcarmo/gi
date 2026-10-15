import fs from 'node:fs';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
const root='/workspace/tmp/gi-ux-reaudit-source';
const out='/workspace/tmp/gi-ux-reaudit-20260926';
const {generateMessages}=await import(root+'/node_modules/@cucumber/gherkin/dist/index.js');
const {IdGenerator,SourceMediaType}=await import(root+'/node_modules/@cucumber/messages/dist/index.js');
const {loadCorpus,verifySources}=await import(root+'/tests/ux/support/catalogue.mjs');
verifySources();
const canonical=[...loadCorpus(),...loadCorpus('shared')];
const files=execFileSync('git',['ls-files','*.feature'],{cwd:root,encoding:'utf8'}).trim().split('\n');
const scenarios=[];const expanded=[];const fileSummaries=[];
function group(file){
 if(file.startsWith('features/'))return 'native-tui-search';
 if(file.startsWith('tests/features/'))return 'gi-additions';
 if(file.includes('/additions/'))return 'passkey-addition';
 if(file.endsWith('shared-canonical-ux.feature'))return 'shared';
 if(file.endsWith('canonical-ux.feature'))return 'classic-canonical';
 if(file.includes('/canonical/core-auth')||file.includes('/canonical/core-settings')||file.includes('/settings/'))return 'classic-auth-settings';
 if(file.includes('/canonical/'))return 'classic-core-workspace';
 if(file.includes('/compose/'))return 'classic-compose';
 if(file.includes('/sessions/')||file.includes('/mobile/'))return 'classic-sessions-mobile';
 return 'classic-editor-media';
}
for(const file of files){
 const source=fs.readFileSync(path.join(root,file),'utf8');
 const env=generateMessages(source,file,SourceMediaType.TEXT_X_CUCUMBER_GHERKIN_PLAIN,{newId:IdGenerator.incrementing(),includeGherkinDocument:true,includePickles:true,includeSource:false});
 const errors=env.filter(x=>x.parseError);if(errors.length)throw Error(JSON.stringify(errors));
 const feature=env.find(x=>x.gherkinDocument)?.gherkinDocument?.feature;
 const pickles=env.filter(x=>x.pickle).map(x=>x.pickle);let defs=0;
 function walk(children,background=[]){let bg=[...background];for(const node of children||[]){if(node.background)bg.push(...node.background.steps.map(s=>s.keyword+s.text));if(node.rule)walk(node.rule.children,bg);if(node.scenario){const s=node.scenario;defs++;const relevant=pickles.filter(p=>p.astNodeIds.includes(s.id));const tags=[...(feature.tags||[]),...(s.tags||[])].map(t=>t.name);const existing=canonical.filter(c=>file==='tests/ux/'+c.uri&&relevant.some(p=>p.name===c.name));const ids=[...new Set(existing.map(c=>c.id))];const key=file+':'+s.location.line;const row={key,group:group(file),file,line:s.location.line,name:s.name,keyword:s.keyword,tags,canonical_ids:ids,background:bg,steps:s.steps.map(step=>({text:step.keyword+step.text,line:step.location.line,dataTable:step.dataTable?.rows.map(r=>r.cells.map(c=>c.value)),docString:step.docString?.content})),examples:(s.examples||[]).map(e=>({line:e.location.line,header:e.tableHeader?.cells.map(c=>c.value),rows:e.tableBody.map(r=>r.cells.map(c=>c.value))})),expanded_count:relevant.length};scenarios.push(row);for(const p of relevant)expanded.push({key,file,name:p.name,steps:p.steps.map(s=>s.text),tags:p.tags.map(t=>t.name)});}}}
 walk(feature?.children);fileSummaries.push({file,group:group(file),definitions:defs,expanded:pickles.length});
}
const summary={baseline:execFileSync('git',['rev-parse','HEAD'],{cwd:root,encoding:'utf8'}).trim(),files:files.length,definitions:scenarios.length,expanded:expanded.length,groups:Object.fromEntries([...new Set(scenarios.map(s=>s.group))].map(g=>[g,{definitions:scenarios.filter(s=>s.group===g).length,expanded:scenarios.filter(s=>s.group===g).reduce((n,s)=>n+s.expanded_count,0)}])),fileSummaries};
fs.writeFileSync(out+'/inventory.json',JSON.stringify({summary,scenarios,expanded},null,2));
for(const g of Object.keys(summary.groups))fs.writeFileSync(out+'/'+g+'-input.json',JSON.stringify(scenarios.filter(s=>s.group===g),null,2));
console.log(JSON.stringify(summary,null,2));
