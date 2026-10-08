// Local acceptance instrumentation. The selected shared specs keep their
// assertions; only their fixture import changes in disposable owned copies.
import { test as base, expect } from '../../references/fixtures-vibes/suite/fixtures';
import { writeFileSync } from 'node:fs';

export const test = base.extend<{ browserCapture: void }>({
  browserCapture: [async ({ page, browserName }, use, info) => {
    if (process.env.PROFILING !== '1' || browserName !== 'chromium') {
      await use();
      return;
    }
    const cdp = await page.context().newCDPSession(page);
    await cdp.send('Profiler.enable');
    await cdp.send('Profiler.start');
    await cdp.send('HeapProfiler.enable');
    await cdp.send('HeapProfiler.startSampling', { samplingInterval: 32768 });
    try { await use(); }
    finally {
      const cpu = await cdp.send('Profiler.stop');
      const heap = await cdp.send('HeapProfiler.stopSampling');
      writeFileSync(info.outputPath('browser.cpuprofile'), JSON.stringify(cpu.profile));
      writeFileSync(info.outputPath('browser.heapprofile'), JSON.stringify(heap.profile));
      await cdp.detach();
    }
  }, { auto: true }],
});
export { expect };
