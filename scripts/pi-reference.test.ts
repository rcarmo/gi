import { test, expect } from 'bun:test';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { requirePi110, goldenPath } from './pi-reference.mjs';

test('reference generators refuse other versions or package identities before writing goldens', () => {
    const root = mkdtempSync(join(process.env.TMPDIR!, 'pi-reference-'));
    const name = '@earendil-works/pi-tui';
    const dir = join(root, name); mkdirSync(dir, { recursive: true });
    const write = (version: string, identity = name) => writeFileSync(join(dir, 'package.json'), JSON.stringify({ name: identity, version }));
    try {
        write('1.1.0'); expect(() => requirePi110(root, [name])).not.toThrow();
        for (const version of ['1.0.4', '1.1.1']) {
            write(version); expect(() => requirePi110(root, [name])).toThrow('Expected');
        }
        write('1.1.0', 'different-package'); expect(() => requirePi110(root, [name])).toThrow('Expected');
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

test('README names the go-ai dependency pin and current Pi baseline', async () => {
    const readme = await Bun.file('README.md').text();
    const mod = await Bun.file('go.mod').text();
    const version = mod.match(/github\.com\/rcarmo\/go-ai\s+(\S+)/)![1];
    expect(readme).toContain(version);
    expect(readme).toContain('Pi 1.1.0');
    expect(readme).not.toContain('The terminal follows Pi 1.0.1');
    const links = [
        ...readme.matchAll(/\]\(([^)]+)\)/g),
        ...readme.matchAll(/^\[[^\]]+\]:\s+(\S+)/gm),
    ].map(m => m[1]).filter(url => !url.includes('://') && !url.startsWith('#'));
    for (const link of links) expect(existsSync(link.split('#')[0])).toBe(true);
});
