import { test, expect } from 'bun:test';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { resolve } from 'node:path';
import { loadCorpus, verifySources, mappedIds, sharedMappedIds, uxRoot } from './catalogue.mjs';

test('frozen Piclaw sources and shared Vibes/Tau contract retain their exact hashes', () => {
  expect(() => verifySources()).not.toThrow();
});

test('imported Piclaw component sources retain their upstream hashes', () => {
  const root = resolve(uxRoot, '../..');
  for (const name of ['piclaw-menu-70d33bc93.json', 'piclaw-session-picker-70d33bc93.json']) {
    const manifest = JSON.parse(readFileSync(resolve(root, 'web/upstream', name), 'utf8'));
    for (const file of manifest.files) {
      const hash = createHash('sha256').update(readFileSync(resolve(root, file.destination))).digest('hex');
      expect(hash).toBe(file.sha256);
    }
  }
});

test('all frozen scenarios and outline examples are inventoried, not just mapped tests', () => {
  const cases = loadCorpus();
  expect(cases).toHaveLength(256);
  expect(new Set(cases.map(item => item.id)).size).toBe(236);
  for (const id of mappedIds) expect(cases.some(item => item.id === id)).toBe(true);
  expect(cases.find(row => row.id === '@ux-compaction-006')?.name).toBe('Check model context compatibility before switching');
  expect(cases.find(row => row.id === '@ux-compaction-007')?.name).toBe('Refresh model information after an accepted switch');
  expect(cases.find(row => row.id === '@ux-context-001')?.name).toBe('Show supplied usage in the context tooltip');
  expect(cases.find(row => row.id === '@ux-context-005')?.name).toBe('Apply the coded usage warning colours');
  const shared = loadCorpus('shared');
  expect(shared).toHaveLength(42);
  for (const id of sharedMappedIds) expect(shared.some(item => item.id === id)).toBe(true);
  expect(shared.find(row => row.id === '@shared-28')?.name).toBe('Return a queued item to the latest editor draft');
  expect(shared.find(row => row.id === '@shared-30')?.name).toBe('Steer only a matching active run');
});
