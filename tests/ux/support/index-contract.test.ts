import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {generateMessages} from '@cucumber/gherkin';
import {IdGenerator} from '@cucumber/messages';
import {loadCorpus} from './catalogue.mjs';

test('derived index contract parses separately without adding frozen parity credit',()=>{
 const path=resolve(import.meta.dir,'../../../features/search/workspace-index.feature');
 const envelopes=generateMessages(readFileSync(path,'utf8'),path,'text/x.cucumber.gherkin+plain',{newId:IdGenerator.incrementing(),includeGherkinDocument:true,includePickles:true});
 expect(envelopes.filter(e=>e.parseError)).toEqual([]);
 const pickles=envelopes.filter(e=>e.pickle).map(e=>e.pickle!);
 expect(pickles).toHaveLength(22);
 const ids=pickles.map(p=>p.tags.find(t=>t.name.startsWith('@index-derived-'))?.name);
 expect(new Set(ids).size).toBe(22);expect(ids.every(Boolean)).toBe(true);
 expect(pickles.every(p=>p.tags.some(t=>t.name==='@proposal'))).toBe(true);
 expect(pickles.some(p=>p.tags.some(t=>t.name.startsWith('@ux-')))).toBe(false);
 const classic=loadCorpus();expect(classic).toHaveLength(256);expect(new Set(classic.map(row=>row.id)).size).toBe(236);
 expect(loadCorpus('shared')).toHaveLength(42);
});
