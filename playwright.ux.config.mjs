import { defineConfig } from '@playwright/test';

const sizes = { phone: { width: 390, height: 844 }, tablet: { width: 820, height: 1180 }, desktop: { width: 1440, height: 900 } };
export default defineConfig({
  testDir: './tests/ux', testMatch: process.env.GI_UX_INDEX_CONFIG ? ['workspace-index-config.spec.mjs'] : process.env.GI_UX_RECONNECT ? ['reconnect.spec.mjs'] : process.env.GI_UX_COMPACTION ? ['compaction.spec.mjs'] : process.env.GI_UX_METER ? ['context-meter.spec.mjs'] : process.env.GI_UX_CONTEXT ? ['context-fit.spec.mjs'] : process.env.GI_UX_STEER ? ['queue-steer.spec.mjs'] : ['classic.spec.mjs', 'session.spec.mjs', 'drafts.spec.mjs', 'queue.spec.mjs', 'models.spec.mjs', 'context.spec.mjs', 'queue-return.spec.mjs', 'rendering.spec.mjs', 'lightbox.spec.mjs', 'workspace-preview.spec.mjs'], timeout: 30000,
  expect: { timeout: 4000 }, retries: 0, workers: 1,
  outputDir: 'test-results/ux-parity/artifacts',
  reporter: [['line'], ['json', { outputFile: process.env.GI_UX_INDEX_CONFIG ? 'test-results/ux-parity/index-config-results.json' : process.env.GI_UX_RECONNECT ? 'test-results/ux-parity/reconnect-results.json' : process.env.GI_UX_COMPACTION ? 'test-results/ux-parity/compaction-results.json' : process.env.GI_UX_METER ? 'test-results/ux-parity/context-meter-results.json' : process.env.GI_UX_CONTEXT ? 'test-results/ux-parity/context-fit-results.json' : process.env.GI_UX_STEER ? 'test-results/ux-parity/steer-results.json' : 'test-results/ux-parity/results.json' }]],
  use: {
    baseURL: process.env.GI_TEST_URL || 'http://127.0.0.1:19091',
    serviceWorkers: 'block', headless: true,
    trace: 'retain-on-failure', screenshot: 'only-on-failure',
  },
  projects: ['chromium', 'webkit'].flatMap(browserName => Object.entries(sizes).map(([size, viewport]) => ({
    name: `${browserName}-${size}`, use: { browserName, viewport },
  }))),
});
