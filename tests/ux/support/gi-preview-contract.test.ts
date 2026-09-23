import { test, expect } from 'bun:test';
import { readFileSync } from 'node:fs';
import { Parser, AstBuilder, GherkinClassicTokenMatcher } from '@cucumber/gherkin';
import { IdGenerator } from '@cucumber/messages';
import { loadCorpus } from './catalogue.mjs';

test('native preview lifecycle criteria do not expand the frozen corpus', () => {
    const text = readFileSync(new URL('../../features/sessions/gi-preview.feature', import.meta.url), 'utf8');
    const document = new Parser(new AstBuilder(IdGenerator.incrementing()), new GherkinClassicTokenMatcher()).parse(text);
    const cases = document.feature.children.filter(child => child.scenario).map(child => child.scenario);
    expect(cases).toHaveLength(3); expect(text).not.toContain('@ux-');
    expect(cases.map(row => row.tags.find(tag => tag.name.startsWith('@gi-preview-')).name)).toEqual(['@gi-preview-001','@gi-preview-002','@gi-preview-003']);
    expect(new Set(loadCorpus().map(row=>row.id)).size).toBe(236);
});
