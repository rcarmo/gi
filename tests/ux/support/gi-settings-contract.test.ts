import { test, expect } from 'bun:test';
import { readFileSync } from 'node:fs';
import { Parser, AstBuilder, GherkinClassicTokenMatcher } from '@cucumber/gherkin';
import { IdGenerator } from '@cucumber/messages';
import { loadCorpus } from './catalogue.mjs';

test('Gi settings scenarios stay separate from the immutable parity corpus', () => {
  const source = readFileSync(new URL('../../features/settings/gi-settings.feature', import.meta.url), 'utf8');
  const ast = new Parser(new AstBuilder(IdGenerator.incrementing()), new GherkinClassicTokenMatcher()).parse(source);
  const cases = ast.feature!.children.flatMap(child => child.scenario ? [child.scenario] : []);
  expect(cases).toHaveLength(14);
  const ids = cases.flatMap(row => row.tags.map(tag => tag.name).filter(tag => tag.startsWith('@gi-settings-')));
  expect(new Set(ids).size).toBe(14);
  expect(cases.filter(row => row.tags.some(tag => tag.name === '@proposal'))).toHaveLength(1);
  expect(source).not.toMatch(/@ux-/);
  expect(new Set(loadCorpus().map(row => row.id)).size).toBe(236);
  expect(loadCorpus('shared')).toHaveLength(42);
});
