import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  retries: 0,
  use: {
    baseURL: process.env.GI_TEST_URL || 'http://127.0.0.1:8090',
    headless: true,
  },
});
