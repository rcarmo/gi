import { test, expect } from '../fixtures-vibes/profiled-fixtures';
import { shell } from '../../references/fixtures-vibes/suite/keychain';
import { clickVisible, closeControl, editorFile, editorText, openInEditor, openWorkspace, pane, removeFiles, treeRow, typeInEditor, uncoverEditor } from '../../references/fixtures-vibes/suite/workspace';

// Catalogue ID supplies only the existing editor gate. This is a Gi contract,
// not an adapted passing result for the shared workspace018 scenario.
test('@ux-workspace-018 Gi consumer: Reload, reviewed conditional Overwrite and create-only Save Copy', async ({ page, runtime, sel }) => {
  test.setTimeout(120000);
  const session = await runtime.newSession();
  const created: string[] = [];
  page.on('dialog', d => void d.accept());
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
    await overwrite.click();
    const review = page.getByRole('dialog', { name: 'Review overwrite' });
    await expect(review).toBeVisible();
    await expect(review.getByLabel('Current saved content')).toHaveValue(reviewed.text);
    expect(writes).toHaveLength(0);
    await expect(closeControl(page, name)).toHaveAccessibleName(/unsaved/i);
    // Cancelling a review cannot authorise a write.
    await review.getByRole('button', { name: 'Cancel', exact: true }).click();
    expect(writes).toHaveLength(0);
    expect((await read()).revision).toBe(reviewed.revision);
    await expect(conflict).toBeVisible({ timeout: 20000 });
    await overwrite.click();
    await expect(review.getByLabel('Current saved content')).toHaveValue(reviewed.text);
    await review.getByRole('button', { name: 'Overwrite reviewed revision', exact: true }).click();
    await expect(closeControl(page, name)).toHaveAccessibleName(/^close\b/i);
    expect(writes).toEqual([{ path: name, content: 'v1 remote-reload local-overwrite', expected_revision: reviewed.revision }]);
    expect((await read()).text).toBe('v1 remote-reload local-overwrite');

    await uncoverEditor(page, sel);
    await expect(editorText(page, sel)).toHaveText('v1 remote-reload local-overwrite');
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
    // Failure cleanup must not be blocked by an unanswered review modal.
    await page.getByRole('dialog', { name: 'Review overwrite' }).getByRole('button', { name: 'Cancel', exact: true }).click({ timeout: 1000 }).catch(() => {});
    await removeFiles(page, created);
  }
});
