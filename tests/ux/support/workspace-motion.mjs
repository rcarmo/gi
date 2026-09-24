import { expect } from '@playwright/test';

// Observe real layout after trusted pointer/keyboard activation. No CSS clock,
// DOM clicks, class changes or synthetic transition events replace the gesture.
export async function startWorkspaceMotionCapture(page) {
  await page.evaluate(() => {
    window.__workspaceMotion = [];
    window.__workspaceMotionStop = false;
    const capture = () => {
      const rect = (selector) => {
        const r = document.querySelector(selector).getBoundingClientRect();
        return { x: r.x, right: r.right, width: r.width };
      };
      window.__workspaceMotion.push({
        time: performance.now(),
        collapsed: document.querySelector('.app-shell').classList.contains('workspace-collapsed'),
        sidebar: rect('.workspace-sidebar'),
        chat: rect('.app-shell > .container'),
        toggle: rect('.workspace-toggle-tab'),
        overflow: document.documentElement.scrollWidth - innerWidth,
      });
      if (!window.__workspaceMotionStop) requestAnimationFrame(capture);
    };
    capture();
  });
}

export async function stopWorkspaceMotionCapture(page) {
  return page.evaluate(() => {
    window.__workspaceMotionStop = true;
    return window.__workspaceMotion;
  });
}

export function expectAnchoredWorkspaceMotion(frames, open, closed, { animated = true } = {}) {
  expect(frames.length).toBeGreaterThan(2);
  const low = Math.min(open.chat.x, closed.chat.x) - 1;
  const high = Math.max(open.chat.x, closed.chat.x) + 1;
  for (const frame of frames) {
    // A left sidebar must never travel towards the centre while disappearing.
    expect(Math.abs(frame.sidebar.x), `sidebar at ${frame.time}`).toBeLessThanOrEqual(1);
    expect(frame.chat.x, `chat jumped left at ${frame.time}`).toBeGreaterThanOrEqual(low);
    expect(frame.chat.x, `chat jumped right at ${frame.time}`).toBeLessThanOrEqual(high);
    expect(frame.chat.width).toBeGreaterThanOrEqual(Math.min(open.chat.width, closed.chat.width) - 1);
    expect(frame.chat.width).toBeLessThanOrEqual(Math.max(open.chat.width, closed.chat.width) + 1);
    // The toggle follows the sidebar edge rather than teleporting to x=0.
    const progress = frame.sidebar.width / open.sidebar.width;
    const expectedToggle = open.toggle.x * progress;
    expect(Math.abs(frame.toggle.x - expectedToggle)).toBeLessThanOrEqual(2);
    // Check synchronisation, not only the end-state envelope: a chat column
    // that snaps early but stays inside that envelope is still incorrect.
    for (const key of ['x', 'width']) {
      const expected = closed.chat[key] + (open.chat[key] - closed.chat[key]) * progress;
      expect(Math.abs(frame.chat[key] - expected), `chat ${key} out of step`).toBeLessThanOrEqual(2);
    }
    expect(frame.overflow).toBeLessThanOrEqual(1);
  }
  if (animated) {
    const intermediate = frames.filter(f => f.sidebar.width > 10 && f.sidebar.width < open.sidebar.width - 10);
    expect(intermediate.length, 'must observe actual intermediate animation frames').toBeGreaterThan(0);
    expect(intermediate.some(f => f.chat.width > Math.min(closed.chat.width, open.chat.width) + 1 && f.chat.width < Math.max(closed.chat.width, open.chat.width) - 1)).toBe(true);
  }
}

export async function checkWorkspaceMotion(page, info, unchanged = async () => {}) {
  const toggle = page.locator('.workspace-toggle-tab');
  const sidebar = page.locator('.workspace-sidebar');
  const settle = () => page.waitForTimeout(250);
  const snapshot = () => page.evaluate(() => {
    const rect = (s) => { const r = document.querySelector(s).getBoundingClientRect(); return { x: r.x, right: r.right, width: r.width }; };
    return { sidebar: rect('.workspace-sidebar'), chat: rect('.app-shell > .container'), toggle: rect('.workspace-toggle-tab') };
  });
  if (await page.locator('.app-shell').evaluate(el => el.classList.contains('workspace-collapsed'))) await toggle.click();
  await settle();
  const open = await snapshot();
  expect(open.sidebar.x).toBe(0);
  expect(open.sidebar.width).toBeGreaterThan(100);
  expect(open.chat.right).toBe(page.viewportSize().width);
  await startWorkspaceMotionCapture(page);
  await toggle.click();
  await settle();
  const closed = await snapshot();
  const closing = await stopWorkspaceMotionCapture(page);
  await info.attach('workspace-closing-frames', { body: JSON.stringify(closing), contentType: 'application/json' });
  expect(closed.sidebar.width).toBe(0);
  expect(closed.chat.width).toBe(900);
  expect(closed.chat.x).toBe((page.viewportSize().width - 900) / 2);
  expectAnchoredWorkspaceMotion(closing, open, closed);
  await unchanged();

  await startWorkspaceMotionCapture(page);
  await toggle.focus();
  await page.keyboard.press('Enter');
  await settle();
  const opening = await stopWorkspaceMotionCapture(page);
  expectAnchoredWorkspaceMotion(opening, open, closed);
  expect(await snapshot()).toEqual(open);
  await unchanged();

  // Reverse while the native transition is in flight, not after its end state.
  await startWorkspaceMotionCapture(page);
  await page.keyboard.press('Enter');
  await expect.poll(() => sidebar.evaluate(el => el.getBoundingClientRect().width), { intervals: [5] }).toBeLessThan(open.sidebar.width - 1);
  await page.keyboard.press('Enter');
  await settle();
  const reversed = await stopWorkspaceMotionCapture(page);
  await info.attach('workspace-reversal-frames', { body: JSON.stringify(reversed), contentType: 'application/json' });
  expect(reversed.some((f, i) => i && reversed[i-1].collapsed && !f.collapsed && f.sidebar.width > 0 && f.sidebar.width < open.sidebar.width)).toBe(true);
  expectAnchoredWorkspaceMotion(reversed, open, closed, { animated: false });
  expect(await snapshot()).toEqual(open);
  await unchanged();

  await page.emulateMedia({ reducedMotion: 'reduce' });
  for (const collapsed of [true, false]) {
    await page.keyboard.press('Enter');
    await expect(page.locator('.app-shell')).toHaveClass(collapsed ? /workspace-collapsed/ : /^(?!.*workspace-collapsed)/);
    expect(await page.evaluate(() => ['.workspace-sidebar','.workspace-splitter','.workspace-toggle-tab','.app-shell > .container'].flatMap(s => document.querySelector(s).getAnimations()).length)).toBe(0);
    expect(await snapshot()).toEqual(collapsed ? closed : open);
    await unchanged();
  }
  await page.emulateMedia({ reducedMotion: 'no-preference' });
}
