import { defineConfig } from './references/fixtures-vibes/node_modules/@playwright/test/index.mjs';
import base from './playwright.fixtures.config';
import { resolve } from 'node:path';

// Gi-only persistence contracts. These do not replace the pinned shared specs.
const results = resolve(process.env.GI_TEST_RUN_ROOT!, 'results', 'consumer');
export default defineConfig({
  ...base,
  testDir: resolve('tests/consumer'),
  reporter: [['line'], ['json', { outputFile: resolve(results, 'results.json') }]],
  outputDir: resolve(results, 'artifacts'),
});
