import { test, expect } from 'bun:test';
import { randomClientId } from '../../../web/src/gi-random-id';
import { createDraftRepository } from '../../../web/src/gi-drafts';
import { patchComposeRandomId } from '../../../scripts/patch-compose-random-id.mjs';

test('UUID v4 uses getRandomValues without secure-context randomUUID', () => {
 let calls=0;
 const source={getRandomValues(bytes:Uint8Array){calls++;bytes.fill(255);return bytes;}};
 expect(randomClientId(source as any)).toBe('ffffffff-ffff-4fff-bfff-ffffffffffff');
 expect(calls).toBe(1);
 expect(new Set(Array.from({length:1000},()=>randomClientId())).size).toBe(1000);
 expect(()=>randomClientId({} as any)).toThrow('message not sent');
});
test('draft capture uses random values when randomUUID is unavailable', async () => {
 const original=globalThis.crypto;
 Object.defineProperty(globalThis,'crypto',{configurable:true,value:{getRandomValues:original.getRandomValues.bind(original)}});
 try {
  const memory:any={async load(){return []},async put(){}};
  const repo=createDraftRepository(memory);await repo.load();
  const capture=repo.begin('A',{text:'typed',media:[],fileRefs:[],messageRefs:[]});await capture.ready;
  expect(capture.token).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  expect(repo.get('A').text).toBe('');
  repo.failed('A',capture.token,'rejected');
  expect(repo.get('A').text).toBe('typed');
 } finally {Object.defineProperty(globalThis,'crypto',{configurable:true,value:original});}
});
test('immutable composer UUID adapter fails closed on changed anchors',async()=>{
 const source=await Bun.file('web/src/components/compose-box.ts').text();
 const patched=patchComposeRandomId(source);
 expect(patched).toContain('randomClientId()');expect(patched).not.toContain('crypto.randomUUID()');
 expect(()=>patchComposeRandomId('')).toThrow('anchor changed');
 expect(()=>patchComposeRandomId('crypto.randomUUID(); crypto.randomUUID()')).toThrow('anchor changed');
});
