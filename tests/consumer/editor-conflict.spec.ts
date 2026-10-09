import { test, expect } from '../fixtures-vibes/profiled-fixtures';
import { shell } from '../../references/fixtures-vibes/suite/keychain';
import { clickVisible, closeControl, editorFile, editorText, openInEditor, openWorkspace, pane, removeFiles, treeRow, typeInEditor, uncoverEditor } from '../../references/fixtures-vibes/suite/workspace';

// Catalogue ID supplies only the existing editor gate. This is a Gi contract,
// not an adapted passing result for the shared workspace018 scenario.
test('@ux-workspace-018 Gi consumer: Reload, reviewed conditional Overwrite and create-only Save Copy', async ({ page, runtime, sel }) => {
  test.setTimeout(120000);
  const session = await runtime.newSession();
  const created: string[] = [];
  page.once('dialog', d => void d.accept());
  await page.goto(session.url);
  await openWorkspace(page);
  const name = await editorFile(page, sel, 'v1', n => created.push(n));
  const read = async () => {
    const response = await page.request.get(new URL(`/api/workspace/file?path=${encodeURIComponent(name)}&mode=edit`, page.url()).href);
    expect(response.status()).toBe(200);
    return response.json();
  };
  const conflict = page.getByRole('button', { name: /^reload$/i }).filter({ visible: true });
  try {
    await uncoverEditor(page, sel);
    await typeInEditor(page, sel, ' local');
    await shell(page, runtime, sel, `printf ' remote-reload' >> ${name}`);
    await expect(conflict).toBeVisible({ timeout: 20000 });
    await conflict.click();
    await expect(editorText(page, sel)).toHaveText('v1 remote-reload');
    await expect(closeControl(page, name)).toHaveAccessibleName(/^close\b/i);

    await typeInEditor(page, sel, ' local-overwrite');
    await shell(page, runtime, sel, `printf ' remote-overwrite' >> ${name}`);
    await expect(conflict).toBeVisible({ timeout: 20000 });
    const reviewed = await read();
    const writes: any[] = [];
    page.on('request', r => {
      if (r.method() === 'PUT' && new URL(r.url()).pathname === '/api/workspace/file') writes.push(JSON.parse(r.postData()!));
    });
    const overwrite = page.getByRole('button', { name: /^overwrite$/i }).filter({ visible: true });
    const review = async (snapshot: any, approve: boolean, beforeApproval?: () => Promise<void>) => {
      const pending = page.waitForEvent('dialog');
      const clicking = overwrite.click();
      const dialog = await pending;
      let accepted = false;
      try {
        expect(dialog.type()).toBe('confirm');
        expect(dialog.message()).toContain(name);
        expect(dialog.message()).toContain(snapshot.text);
        expect(dialog.message()).toContain('Another change will cause a new conflict.');
        await beforeApproval?.();
        accepted = approve;
      } finally {
        await (accepted ? dialog.accept() : dialog.dismiss());
        await clicking;
      }
    };
    // Cancelling a native confirmation cannot authorise a write.
    await review(reviewed, false);
    expect(writes).toHaveLength(0);
    expect((await read()).revision).toBe(reviewed.revision);
    await expect(closeControl(page, name)).toHaveAccessibleName(/unsaved/i);
    await expect(conflict).toBeVisible({ timeout: 20000 });

    // A newer remote change while approval is pending must reject this exact reviewed revision.
    await review(reviewed, true, async () => {
      expect(writes).toHaveLength(0);
      const changed = await page.request.put(new URL('/api/workspace/file', page.url()).href, {
        data: { path: name, content: reviewed.text + ' newest', expected_revision: reviewed.revision },
      });
      expect(changed.status()).toBe(200);
    });
    await expect(conflict).toBeVisible({ timeout: 20000 });
    await expect(closeControl(page, name)).toHaveAccessibleName(/unsaved/i);
    expect(writes).toEqual([{ path: name, content: 'v1 remote-reload local-overwrite', expected_revision: reviewed.revision }]);
    const newer = await read();
    expect(newer.text).toBe(reviewed.text + ' newest');
    expect(newer.revision).not.toBe(reviewed.revision);

    // A fresh, explicit approval uses only its newly displayed snapshot; no automatic retry.
    await review(newer, true);
    await expect(closeControl(page, name)).toHaveAccessibleName(/^close\b/i);
    expect(writes).toEqual([
      { path: name, content: 'v1 remote-reload local-overwrite', expected_revision: reviewed.revision },
      { path: name, content: 'v1 remote-reload local-overwrite', expected_revision: newer.revision },
    ]);
    expect((await read()).text).toBe('v1 remote-reload local-overwrite');

    await uncoverEditor(page, sel);
    await expect(editorText(page, sel)).toHaveText('v1 remote-reload local-overwrite');
    // Focus actual text after the native confirmation, rather than the editor's trailing viewport padding.
    await editorText(page, sel).locator('.cm-line').first().click();
    await expect(editorText(page, sel)).toBeFocused();
    await typeInEditor(page, sel, ' keep-this-copy');
    await shell(page, runtime, sel, `printf ' remote-copy' >> ${name}`);
    await expect(conflict).toBeVisible({ timeout: 20000 });
    const copyPosts: any[] = [];
    page.on('request', r => {
      if (r.method() === 'POST' && new URL(r.url()).pathname === '/api/workspace/file') copyPosts.push(JSON.parse(r.postData()!));
    });
    await page.getByRole('button', { name: /^save copy$/i }).filter({ visible: true }).click();
    await expect.poll(() => copyPosts.length).toBe(1);
    const copy = copyPosts[0].name;
    created.push(copy);
    expect(copyPosts[0]).toEqual({ path: '.', name: copy, content: 'v1 remote-reload local-overwrite keep-this-copy' });
    expect(copy).toMatch(new RegExp(`^${name.replace('.md', '')}\\..+\\.md$`));
    await expect(closeControl(page, name)).toHaveAccessibleName(/unsaved/i);
    expect((await read()).text).toBe('v1 remote-reload local-overwrite remote-copy');
    await openWorkspace(page);
    await pane(page).getByRole('button', { name: /^refresh tree$/i }).click();
    await expect.poll(async () => !!await treeRow(page, copy)).toBe(true);
    await openInEditor(page, sel, copy);
    await expect(editorText(page, sel)).toHaveText('v1 remote-reload local-overwrite keep-this-copy');
  } finally {
    await removeFiles(page, created);
  }
});
