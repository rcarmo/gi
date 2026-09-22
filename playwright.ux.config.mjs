import { defineConfig } from '@playwright/test';

const sizes = { phone: { width: 390, height: 844 }, tablet: { width: 820, height: 1180 }, desktop: { width: 1440, height: 900 } };
export default defineConfig({
  testDir: './tests/ux', testMatch: ['classic.spec.mjs', 'session.spec.mjs', 'drafts.spec.mjs', 'queue.spec.mjs'], timeout: 30000,
  expect: { timeout: 4000 }, retries: 0, workers: 1,
  outputDir: 'test-results/ux-parity/artifacts',
  reporter: [['line'], ['json', { outputFile: 'test-results/ux-parity/results.json' }]],
  use: {
    baseURL: process.env.GI_TEST_URL || 'http://127.0.0.1:19091',
    serviceWorkers: 'block', headless: true,
    trace: 'retain-on-failure', screenshot: 'only-on-failure',
  },
  projects: ['chromium', 'webkit'].flatMap(browserName => Object.entries(sizes).map(([size, viewport]) => ({
    name: `${browserName}-${size}`, use: { browserName, viewport },
  }))),
});
