import { test, expect } from 'bun:test';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { requirePi130, goldenPath } from './pi-reference.mjs';

test('reference generators refuse other versions or package identities before writing goldens', () => {
    const root = mkdtempSync(join(process.env.TMPDIR!, 'pi-reference-'));
    const name = '@earendil-works/pi-tui';
    const dir = join(root, name); mkdirSync(dir, { recursive: true });
    const write = (version: string, identity = name) => writeFileSync(join(dir, 'package.json'), JSON.stringify({ name: identity, version }));
    try {
        write('1.3.0'); expect(() => requirePi130(root, [name])).not.toThrow();
        for (const version of ['1.0.4', '1.3.1']) {
            write(version); expect(() => requirePi130(root, [name])).toThrow('Expected');
        }
        write('1.3.0', 'different-package'); expect(() => requirePi130(root, [name])).toThrow('Expected');
    } finally { rmSync(root, { recursive: true }); }
});

test('scratch golden output leaves the source fixture directory unchanged', () => {
    const root = mkdtempSync(join(process.env.TMPDIR!, 'pi-goldens-'));
    const before = process.env.PI_GOLDEN_OUTPUT_DIR;
    try {
        process.env.PI_GOLDEN_OUTPUT_DIR = root;
        expect(goldenPath('navigation.json')).toBe(join(root, 'navigation.json'));
    } finally {
        if (before === undefined) delete process.env.PI_GOLDEN_OUTPUT_DIR;
        else process.env.PI_GOLDEN_OUTPUT_DIR = before;
        rmSync(root, { recursive: true });
    }
});


test('README records the pinned dependency and current TUI target', () => {
 const readme=readFileSync(join(import.meta.dir,'../README.md'),'utf8');
 const mod=readFileSync(join(import.meta.dir,'../go.mod'),'utf8');
 const version=mod.match(/github\.com\/rcarmo\/go-ai\s+(\S+)/)?.[1];
 expect(version).toBeDefined();expect(readme).toContain(version);expect(readme).toContain('Pi TUI 1.3.0');
 expect(readme).toContain('alignment acceptance in progress');
 expect(readme).not.toContain('The TUI is not yet considered stable');
 expect(/\]\(\/workspace\//.test(readme)).toBe(false);
});
