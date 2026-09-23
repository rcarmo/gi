import { test, expect } from 'bun:test';
import { readFileSync } from 'node:fs';
import { Parser, AstBuilder, GherkinClassicTokenMatcher } from '@cucumber/gherkin';
import { IdGenerator } from '@cucumber/messages';
import { loadCorpus } from './catalogue.mjs';

test('Gi activation swipe contract is separate from frozen browser and terminal credit', () => {
    const text = readFileSync(new URL('../../features/sessions/gi-swipe.feature', import.meta.url), 'utf8');
    const document = new Parser(new AstBuilder(IdGenerator.incrementing()), new GherkinClassicTokenMatcher()).parse(text);
    const cases = document.feature.children.filter(child => child.scenario).map(child => child.scenario);
    expect(cases).toHaveLength(3);
    expect(cases.map(row => row.tags.find(tag => tag.name.startsWith('@gi-swipe-')).name)).toEqual(['@gi-swipe-001', '@gi-swipe-002', '@gi-swipe-003']);
    expect(text).not.toContain('@ux-'); expect(new Set(loadCorpus().map(row => row.id)).size).toBe(236);
});
