import { test, expect } from '../fixtures-vibes/profiled-fixtures';
import { shell } from '../../references/fixtures-vibes/suite/keychain';
import { editorFile, editorText, openWorkspace, removeFiles, save, typeInEditor, uncoverEditor } from '../../references/fixtures-vibes/suite/workspace';

// Separate consumer contract for the lazy-loader SSE refresh boundary.
test('@ux-workspace-019 Gi consumer: clean external refresh retains revision for the next save', async ({ page, runtime, sel }) => {
  const session = await runtime.newSession();
  await page.goto(session.url);
  await openWorkspace(page);
  const created: string[] = [];
  const name = await editorFile(page, sel, 'baseline', n => created.push(n));
  page.on('dialog', d => void d.accept());
  try {
    // Agent shell writes generate the actual workspace-update SSE.
    await shell(page, runtime, sel, `printf ' externally-refreshed' >> ${name}`);
    await uncoverEditor(page, sel);
    await expect(editorText(page, sel)).toHaveText('baseline externally-refreshed');
    const response = await page.request.get(new URL(`/api/workspace/file?path=${encodeURIComponent(name)}&mode=edit`, page.url()).href);
    expect(response.status()).toBe(200);
    const snapshot = await response.json();
    const writes: any[] = [];
    page.on('request', r => {
      if (r.method() === 'PUT' && new URL(r.url()).pathname === '/api/workspace/file') writes.push(JSON.parse(r.postData()!));
    });
    await typeInEditor(page, sel, ' local-after-refresh');
    // The loader must forward the complete snapshot revision, not just text/mtime.
    await expect(page.getByRole('button', { name: /^save$/i }).filter({ visible: true }).first()).toBeEnabled();
    await save(page, name);
    expect(writes).toEqual([{ path: name, content: 'baseline externally-refreshed local-after-refresh', expected_revision: snapshot.revision }]);
  } finally {
    await removeFiles(page, created);
  }
});
