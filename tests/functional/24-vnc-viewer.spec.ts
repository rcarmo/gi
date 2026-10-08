import { test, expect } from '@playwright/test';
import { readFileSync, writeFileSync } from 'node:fs';

test('VNC configured policy, real frame and user-input gates', async ({ page, browserName }) => {
  test.skip(!process.env.GI_VNC_INPUT_COUNTER, 'requires live owned RFB fixture');
  const mode = process.env.GI_VNC_VIEWER_MODE || 'readonly';
  const counter = () => Number(readFileSync(process.env.GI_VNC_INPUT_COUNTER!, 'utf8'));
  const cpu = process.env.PROFILING === '1' && browserName === 'chromium' ? await page.context().newCDPSession(page) : null;
  if (cpu) {
    await cpu.send('Profiler.enable'); await cpu.send('Profiler.start');
    await cpu.send('HeapProfiler.enable'); await cpu.send('HeapProfiler.startSampling', { samplingInterval: 32768 });
  }
  try {
    const policy = await (await page.request.get('/vnc/session')).json();
    expect(policy.direct_connect_enabled).toBe(false);
    expect(policy.targets).toHaveLength(mode === 'empty' ? 0 : 1);
    await page.goto('/');
    await expect(page.locator('.compose-box textarea')).toBeVisible();
    await page.getByRole('button', { name: /^menu$|workspace menu/i }).first().click();
    await page.getByRole('menuitem', { name: /open vnc in tab/i }).click();
    await expect(page.getByText('Direct connections are disabled. Choose a configured target.')).toBeVisible();
    if (mode === 'empty') {
      await expect(page.getByText('No configured targets.', { exact: true })).toBeVisible();
      await expect(page.getByLabel(/^Server$/i)).toHaveCount(0);
      await expect(page.getByRole('button', { name: /^Connect$/i })).toHaveCount(0);
      await expect(page.locator('[data-display-canvas]')).toHaveCount(0);
      return;
    }
    await page.getByRole('button', { name: /Live RFB fixture/ }).click();
    await expect(page.locator('.vnc-pane-shell')).toHaveAttribute('data-vnc-state', 'connected', { timeout: 15000 });
    const canvas = page.locator('[data-display-canvas]');
    await expect(canvas).toBeVisible();
    // Measure decoded framebuffer colour against the independent raw-RFB fixture.
    await expect.poll(() => canvas.evaluate((e: HTMLCanvasElement) => {
      const pixels = e.getContext('2d')!.getImageData(0, 0, e.width, e.height).data;
      let matching = 0;
      for (let i = 0; i < pixels.length; i += 4) {
        if (Math.abs(pixels[i] - 200) <= 1 && Math.abs(pixels[i + 1] - 120) <= 1 && Math.abs(pixels[i + 2] - 40) <= 1) matching++;
      }
      return matching;
    })).toBeGreaterThan(0);
    const before = counter();
    // Use a lower canvas point, clear of the viewer toolbar overlay.
    const box = (await canvas.boundingBox())!;
    await canvas.click({ position: { x: box.width / 2, y: box.height * 0.75 } });
    await page.keyboard.type('unsafe'); await page.keyboard.press('Enter');
    await expect(canvas).toBeFocused();
    if (mode === 'interactive') await expect.poll(counter).toBeGreaterThan(before);
    else {
      // Same action sequence has an interactive positive control in the next run.
      await page.waitForTimeout(300);
      expect(counter()).toBe(before);
    }
    await page.keyboard.press('Control+Alt+Shift+V');
    await page.locator('[data-vnc-session-chrome] details').filter({ hasText: 'Clipboard' }).locator('summary').click();
    const send = page.getByRole('button', { name: 'Send to remote', exact: true });
    await expect(send).toBeVisible();
    if (mode === 'readonly') await expect(send).toBeDisabled();
    else {
      await page.locator('[data-vnc-clipboard]').fill('clipboard control');
      const count = counter(); await send.click(); await expect.poll(counter).toBeGreaterThan(count);
    }
  } finally {
    if (cpu) {
      const p = await cpu.send('Profiler.stop'); const h = await cpu.send('HeapProfiler.stopSampling');
      const cpuPath = test.info().outputPath('browser.cpuprofile');
      const heapPath = test.info().outputPath('browser.heapprofile');
      writeFileSync(cpuPath, JSON.stringify(p.profile));
      writeFileSync(heapPath, JSON.stringify(h.profile));
      await test.info().attach('browser.cpuprofile', { path: cpuPath });
      await test.info().attach('browser.heapprofile', { path: heapPath });
      await cpu.detach();
    }
  }
});
