import { expect, test } from 'bun:test';
import { relocateSpecImports } from './profiled-spec-imports';

test('relocates static and literal dynamic imports without changing assertions', () => {
  const source = `import { test } from '@playwright/test';
import { expect } from '../fixtures';
import { helper } from './helper';
export { api } from '../api';
const { png } = await import('../png');
const local = await import( "./helper" );
const runtime = await import(path);
const external = await import('node:fs');
expect(png(16,16)).toBeDefined();`;
  expect(relocateSpecImports(source, '/suite', '/pw/index.mjs', '/owned/fixtures.ts')).toBe(`import { test } from "/pw/index.mjs";
import { expect } from "/owned/fixtures.ts";
import { helper } from "/suite/specs/helper";
export { api } from "/suite/api";
const { png } = await import("/suite/png");
const local = await import("/suite/specs/helper");
const runtime = await import(path);
const external = await import('node:fs');
expect(png(16,16)).toBeDefined();`);
});
