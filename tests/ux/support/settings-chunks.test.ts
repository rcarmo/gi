import { test, expect } from 'bun:test';
import { readFileSync, readdirSync } from 'node:fs';
import { join, resolve, dirname } from 'node:path';

// Check committed/generated assets, not a helper-only substitute for browser evidence.
test('hashed settings graph resolves embedded chunks without reinitialising the versioned bootstrap', () => {
  const root = resolve('internal/web/static/dist');
  const bootstrap = readFileSync(join(root, 'app.bundle.js'), 'utf8');
  const app = bootstrap.match(/import\("\.\/chunks\/(app-[a-z0-9]+\.js)"\)/)?.[1];
  expect(app).toBeTruthy();
  const files = readdirSync(join(root, 'chunks')).filter(file => file.endsWith('.js'));
  expect(files.filter(file => file.startsWith('app-'))).toEqual([app!]);
  const appSource = readFileSync(join(root, 'chunks', app!), 'utf8');
  for (const section of ['models', 'appearance', 'compaction', 'providers']) {
    const matches = files.filter(file => file.startsWith(`gi-settings-${section}-`)); expect(matches).toHaveLength(1);
    expect(appSource).toContain(`import("./${matches[0]}")`);
  }
  expect(appSource).not.toContain('function Models(');
  expect(appSource).not.toContain('function Appearance(');
  for (const file of files) {
    const path = join(root, 'chunks', file), source = readFileSync(path, 'utf8');
    expect(source).not.toMatch(/(?:from\s*|import\s*\()"[^"]*app\.bundle\.js[^"]*"/);
    for (const match of source.matchAll(/(?:from\s*|import\s*\()"(\.[^"\n]+\.js)"/g)) {
      const dependency = resolve(dirname(path), match[1]);
      expect(dependency.startsWith(root + '/')).toBe(true);
      expect(readFileSync(dependency).length).toBeGreaterThan(0);
    }
  }
});
