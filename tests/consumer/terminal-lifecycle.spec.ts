import { test, expect } from '../fixtures-vibes/profiled-fixtures';
import { editorFile, tab, removeFiles } from '../../references/fixtures-vibes/suite/workspace';

// Shared keyboard assertions remain upstream; this checks Gi's real teardown
// boundary and the late xterm viewport callback seen in independent acceptance.
test('@ux-terminal-008 Gi consumer: repeated dock teardown leaves no late renderer callbacks', async ({ page, runtime, sel }) => {
  test.setTimeout(90_000);
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await page.goto((await runtime.newSession()).url);
  const file = await editorFile(page, sel, 'terminal lifecycle');
  try {
    for (let i = 0; i < 10; i++) {
      await tab(page, file).click();
      await page.keyboard.press('Control+Backquote');
      const terminal = page.locator(sel('terminal')).filter({ visible: true }).first();
      await expect(terminal).toHaveAccessibleName(/connected/i);
      await terminal.click();
      await page.keyboard.type(`printf 'cycle:%s\\n' "$(( ${i}+1 ))"`);
      await page.keyboard.press('Enter');
      await expect(terminal).toContainText(`cycle:${i + 1}`);
      await tab(page, file).click();
      await page.keyboard.press('Control+Backquote');
      await expect(page.locator(sel('terminal')).filter({ visible: true })).toHaveCount(0);
      // Let already-queued viewport/render work run after the unmount.
      await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(errors).toEqual([]);
    }
  } finally {
    await removeFiles(page, [file]);
  }
  expect(errors).toEqual([]);
});
