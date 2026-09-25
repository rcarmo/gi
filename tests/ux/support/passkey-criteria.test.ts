import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {resolve} from 'node:path';
import {generateMessages} from '@cucumber/gherkin';
import {IdGenerator,SourceMediaType} from '@cucumber/messages';
import {loadCorpus,mappedIds,sharedMappedIds} from './catalogue.mjs';

const root=resolve(import.meta.dir,'../../..');
const ledger=JSON.parse(readFileSync(resolve(root,'docs/internal/passkey-criteria.json'),'utf8'));
const text=readFileSync(resolve(root,ledger.source.path),'utf8');
const messages=generateMessages(text,ledger.source.path,SourceMediaType.TEXT_X_CUCUMBER_GHERKIN_PLAIN,{newId:IdGenerator.incrementing(),includeGherkinDocument:true,includePickles:true});
const feature=messages.find(m=>m.gherkinDocument)!.gherkinDocument!.feature!;
const collect=(children:any[]):any[]=>children.flatMap(c=>c.scenario?[c.scenario]:c.rule?collect(c.rule.children):[]);
const states=new Set(['bounded','substituted','partial','gap','manual','unsupported']);
const checkDisposition=(row:any)=>{
 expect(states.has(row.state)).toBe(true);expect(row.note.trim().length).toBeGreaterThan(0);
 expect(Array.isArray(row.evidence)).toBe(true);
 if(['bounded','substituted','partial'].includes(row.state))expect(row.evidence.length).toBeGreaterThan(0);
 for(const ref of row.evidence)expect(ledger.evidence[ref]).toBeDefined();
};

test('additive passkey ledger pins every Background, criterion and example without granting frozen parity',()=>{
 expect(ledger.schemaVersion).toBe(1);
 expect(ledger.source.path).toBe('tests/ux/features/additions/piclaw-2026-09-24/piclaw-single-user-passkey-settings.feature');
 expect(createHash('sha256').update(text).digest('hex')).toBe('bd48cab9126778763cab3ddfd8ee04a89bd8c29da62e3bc4188cf24dc3a55b82');
 expect(ledger.source.sha256).toBe('bd48cab9126778763cab3ddfd8ee04a89bd8c29da62e3bc4188cf24dc3a55b82');
 expect(messages.filter(m=>m.parseError)).toEqual([]);expect(messages.filter(m=>m.pickle)).toHaveLength(56);
 const source=collect(feature.children);expect(source).toHaveLength(26);expect(ledger.scenarios).toHaveLength(26);
 expect(new Set(ledger.scenarios.map((s:any)=>s.id)).size).toBe(26);
 const matchSteps=(actual:any[],expected:any[])=>{
  expect(actual.map(({line,keyword,text})=>({line,keyword,text}))).toEqual(expected.map(s=>({line:s.location.line,keyword:s.keyword.trim(),text:s.text})));
  actual.forEach(checkDisposition);
 };
 matchSteps(ledger.background,feature.children.find(c=>c.background)!.background!.steps);
 for(const [i,s]of source.entries()){
  const row=ledger.scenarios[i];expect(row.id).toBe(s.tags.find((t:any)=>t.name.startsWith('@ux-single-passkeys-')).name);
  expect(row.title).toBe(s.name);expect(row.line).toBe(s.location.line);expect(row.formalMapping).toBe(false);matchSteps(row.criteria,s.steps);
  const examples=s.examples.flatMap((ex:any)=>(ex.tableBody||[]).map((r:any)=>({line:r.location.line,values:Object.fromEntries(ex.tableHeader.cells.map((c:any,j:number)=>[c.value,r.cells[j].value]))})));
  expect(row.examples.map(({line,values}:any)=>({line,values}))).toEqual(examples);row.examples.forEach(checkDisposition);
 }
 expect(ledger.scenarios.reduce((n:number,s:any)=>n+s.criteria.length,0)).toBe(169);
 expect(ledger.scenarios.reduce((n:number,s:any)=>n+s.examples.length,0)).toBe(40);
 expect(ledger.terminal.idleRows).toBe(0);
 // This ledger is additive; it is deliberately not added to existing sets.
 const classic=loadCorpus();expect(classic).toHaveLength(256);expect(new Set(classic.map(c=>c.id)).size).toBe(236);expect(loadCorpus('shared')).toHaveLength(42);
 expect(mappedIds.size).toBe(101);expect(sharedMappedIds.size).toBe(30);
 expect(mappedIds.has('@ux-original-008')).toBe(true); // disputed mapping retained, not re-awarded
 for(const s of ledger.scenarios){expect(mappedIds.has(s.id)).toBe(false);expect(sharedMappedIds.has(s.id)).toBe(false);}
});

test('passkey criterion evidence points to named tests or a declared fixture, not passing-run claims',()=>{
 expect(Object.keys(ledger.evidence).length).toBeGreaterThan(0);
 const anchors=new Set<string>();
 for(const [id,evidence]of Object.entries(ledger.evidence) as [string,any][]){
  const identity=`${evidence.file}:${evidence.anchor}`;expect(anchors.has(identity)).toBe(false);anchors.add(identity);
  expect(evidence.file.startsWith('tests/')||evidence.file.startsWith('internal/')).toBe(true);
  expect(evidence.file.includes('..')).toBe(false);expect(evidence.scope.trim().length).toBeGreaterThan(0);
  const content=readFileSync(resolve(root,evidence.file),'utf8');expect(content.split(evidence.anchor).length-1).toBe(1);
  expect(evidence.anchor.length).toBeGreaterThan(15);
  if(id==='fixture')expect(evidence.anchor.startsWith('export async function ')).toBe(true);
  else if(evidence.file.endsWith('_test.go'))expect(evidence.anchor.startsWith('func Test')).toBe(true);
  else{
   expect(evidence.file.endsWith('.spec.mjs')||evidence.file.endsWith('.spec.ts')).toBe(true);
   expect(["'",'"','`'].some(quote=>content.includes(`test(${quote}${evidence.anchor}${quote},`))).toBe(true);
  }
 }
 // Known gaps must not silently acquire complete-looking dispositions.
 const criterion=(id:string,index:number)=>ledger.scenarios.find((s:any)=>s.id.endsWith(id)).criteria[index-1];
 // Return/reopen evidence closes014's explicit-return gap, without awarding
 // full-scenario credit or changing the inherited HTTPS/skin substitutions.
 expect(criterion('014',3).state).toBe('bounded');
 expect(criterion('014',3).evidence).toEqual(['return-reconcile']);
 expect(criterion('017',2).state).toBe('partial');
 expect(criterion('020',7).state).toBe('gap');
 expect(criterion('021',5).state).toBe('partial');
 expect(criterion('024',7).state).toBe('partial');
});
