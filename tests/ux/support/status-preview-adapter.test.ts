import { test, expect } from 'bun:test';
import { readFileSync } from 'node:fs';
import { patchStatusPreview } from '../../../scripts/patch-status-preview.mjs';

const source = readFileSync('web/src/components/status.ts', 'utf8');
test('status adapter patches only disclosure measurement and leaves line metadata/renderer intact', () => {
    const patched = patchStatusPreview(source);
    expect(patched).toContain('usePreviewOverflow(draft, thought, expandedPanels)');
    expect(patched).toContain('ref=${previewOverflow.refs[panelKey]}');
    expect(patched).toContain("'▸ Show more'");
    expect(patched).toContain('truncated.omitted > 0 ? `▸ ${truncated.omitted} more lines`');
    for (const start of ['    const normalizePreview =', '    const truncateLines =', '    const toggleExpand =']) {
        const from = source.indexOf(start), to = source.indexOf('\n\n', from);
        expect(patched).toContain(source.slice(from, to));
    }
    expect(patched).toContain('renderThinkingMarkdown(sourceText)');
    expect(readFileSync('web/src/components/status.ts', 'utf8')).toBe(source);
});

test('status adapter fails closed on missing, duplicate or already-patched upstream anchors', () => {
    expect(() => patchStatusPreview('')).toThrow('anchor changed');
    expect(() => patchStatusPreview(source.replace('const bodyClass =', 'const renamed ='))).toThrow('anchor changed');
    expect(() => patchStatusPreview(source + '\n' + source)).toThrow('anchor changed');
    expect(() => patchStatusPreview(patchStatusPreview(source))).toThrow('anchor changed');
});
