import { defineConfig } from './references/fixtures-vibes/node_modules/@playwright/test/index.mjs';
import shared from './references/fixtures-vibes/suite/playwright.config';
import { resolve } from 'node:path';

if (!process.env.GI_TEST_RUN_ROOT) throw new Error('Run fixture checks through the project Makefile');
const results = resolve(process.env.GI_TEST_RUN_ROOT, 'results', 'fixtures');
export default defineConfig({
  ...shared,
  testDir: process.env.GI_PROFILED_SPECS || resolve('references/fixtures-vibes/suite/specs'),
  globalSetup: resolve('references/fixtures-vibes/suite/global-setup.ts'),
  reporter: [['line'], ['json', { outputFile: resolve(results, 'results.json') }]],
  outputDir: resolve(results, 'artifacts'),
});
