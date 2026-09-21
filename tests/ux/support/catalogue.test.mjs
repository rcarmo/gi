import { test, expect } from 'bun:test';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { resolve } from 'node:path';
import { loadCorpus, verifySources, mappedIds, uxRoot } from './catalogue.mjs';

test('frozen Piclaw sources and shared Vibes/Tau contract retain their exact hashes', () => {
  expect(() => verifySources()).not.toThrow();
});

test('imported Piclaw component sources retain their upstream hashes', () => {
  const root = resolve(uxRoot, '../..');
  const manifest = JSON.parse(readFileSync(resolve(root, 'web/upstream/piclaw-menu-70d33bc93.json'), 'utf8'));
  for (const file of manifest.files) {
    const hash = createHash('sha256').update(readFileSync(resolve(root, file.destination))).digest('hex');
    expect(hash).toBe(file.sha256);
  }
});

test('all frozen scenarios and outline examples are inventoried, not just mapped tests', () => {
  const cases = loadCorpus();
  expect(cases).toHaveLength(256);
  expect(new Set(cases.map(item => item.id)).size).toBe(236);
  for (const id of mappedIds) expect(cases.some(item => item.id === id)).toBe(true);
  expect(loadCorpus('shared')).toHaveLength(42);
});
