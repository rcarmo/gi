import { defineConfig, devices } from '@playwright/test';
import base from './playwright.config';

export default defineConfig({
  ...base,
  testMatch: '**/24-vnc-viewer.spec.ts',
  outputDir: `${process.env.GI_TEST_RUN_ROOT}/vnc-viewer/${process.env.GI_VNC_VIEWER_MODE || 'readonly'}/results/artifacts`,
  reporter: [['line'], ['json', { outputFile: `${process.env.GI_TEST_RUN_ROOT}/vnc-viewer/${process.env.GI_VNC_VIEWER_MODE || 'readonly'}/results/results.json` }]],
  projects: [
    { name: 'chromium-phone', use: { ...devices['Pixel 7'], browserName: 'chromium' } },
    { name: 'chromium-tablet', use: { ...devices['iPad Mini'], browserName: 'chromium' } },
    { name: 'chromium-desktop', use: { browserName: 'chromium', viewport: { width: 1440, height: 900 } } },
    { name: 'webkit-phone', use: { ...devices['iPhone 13'], browserName: 'webkit' } },
    { name: 'webkit-tablet', use: { ...devices['iPad Mini'], browserName: 'webkit' } },
    { name: 'webkit-desktop', use: { browserName: 'webkit', viewport: { width: 1440, height: 900 } } },
  ],
});
