import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  retries: 0,
  workers: 1, // serial — tests share the same server instance
  use: {
    baseURL: process.env.GI_TEST_URL || 'http://127.0.0.1:19090',
    headless: true,
    // Capture traces on failure for debugging
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  // Global setup could start the test server, but we use make test-ux for that
  reporter: [['line'], ['json', { outputFile: 'test-results/results.json' }]],
});
