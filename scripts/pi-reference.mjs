import { readFileSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';

/** Refuse stale installed packages when regenerating the current Pi goldens. */
export function requirePi110(root, packages) {
    for (const name of packages) {
        const pkg = JSON.parse(readFileSync(join(root, name, 'package.json'), 'utf8'));
        if (pkg.name !== name || pkg.version !== '1.1.0') {
            throw new Error(`Expected ${name}@1.1.0; provide its published node_modules directory`);
        }
    }
}
export function goldenPath(name) {
    const dir = process.env.PI_GOLDEN_OUTPUT_DIR || 'internal/tui/testdata';
    mkdirSync(dir, { recursive: true });
    return join(dir, name);
}
