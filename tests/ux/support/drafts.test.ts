import { expect, test } from 'bun:test';
import { createDraftRepository, emptyDraft, mergeDrafts } from '../../../web/src/gi-drafts';

function storage() {
  const rows = new Map();
  return { rows, load: async () => structuredClone([...rows.values()]), put: async (row: any) => { rows.set(row.sessionId, structuredClone(row)); } };
}

test('failed capture merges newer origin text, files and references once', async () => {
  const disk = storage(), repo = createDraftRepository(disk);
  const file = new File(['image'], 'one.png', { type: 'image/png', lastModified: 1 });
  const captured = { text: 'first', media: [file], fileRefs: ['one'], messageRefs: [1] };
  const send = repo.begin('A', captured);
  await send.ready;
  repo.update('A', { text: 'newer', media: [file, new File(['x'], 'two.txt')], fileRefs: ['one', 'two'], messageRefs: [1, 2] });
  repo.update('B', { text: 'other session' });
  repo.failed('A', send.token, 'network failed');
  repo.failed('A', send.token, 'network failed');
  expect(repo.get('A')).toMatchObject({ text: 'first\n\nnewer', fileRefs: ['one', 'two'], messageRefs: [1,2] });
  expect(repo.get('A').media).toHaveLength(2);
  expect(repo.get('B').text).toBe('other session');
  await repo.flush();
  const recovered = createDraftRepository(disk); await recovered.load();
  expect(recovered.get('A').text).toBe('first\n\nnewer');
  expect(recovered.error('A')).toBe('network failed');
});

test('reload recovers unacknowledged sends before new text and does not duplicate on another reload', async () => {
  const disk = storage(), repo = createDraftRepository(disk);
  const first = repo.begin('A', { ...emptyDraft(), text: 'first' }); await first.ready;
  const second = repo.begin('A', { ...emptyDraft(), text: 'second' }); await second.ready;
  repo.update('A', { text: 'newest' }); await repo.flush();
  const reloaded = createDraftRepository(disk); await reloaded.load();
  expect(reloaded.get('A').text).toBe('first\n\nsecond\n\nnewest');
  expect(reloaded.error('A')).toContain('Delivery is unknown');
  const twice = createDraftRepository(disk); await twice.load();
  expect(twice.get('A').text).toBe(reloaded.get('A').text);
});

test('acknowledgement removes only its capture, preserving subsequent edits and writes', async () => {
  const disk = storage(), repo = createDraftRepository(disk);
  const send = repo.begin('A', { ...emptyDraft(), text: 'sent' }); await send.ready;
  repo.update('A', { text: 'keep' });
  await repo.accepted('A', send.token);
  const reloaded = createDraftRepository(disk); await reloaded.load();
  expect(reloaded.get('A').text).toBe('keep');
  expect(reloaded.error('A')).toBe('');
});

test('capture persistence failure rejects readiness; recovery remains in memory and reports error', async () => {
  const errors: Error[] = [];
  const repo = createDraftRepository({ load: async () => [], put: async () => { throw new Error('quota'); } }, error => errors.push(error));
  const send = repo.begin('A', { ...emptyDraft(), text: 'keep me' });
  await expect(send.ready).rejects.toThrow('quota');
  repo.failed('A', send.token, 'not sent');
  await expect(repo.flush()).rejects.toThrow('quota');
  expect(repo.get('A').text).toBe('keep me');
  expect(errors.length).toBeGreaterThan(0);
});

test('merge recognises already-restored prefix without losing current references', () => {
  expect(mergeDrafts({ ...emptyDraft(), text: 'hello', fileRefs: ['a'] }, { ...emptyDraft(), text: 'hello\n\nnewer', fileRefs: ['b'] }))
    .toMatchObject({ text: 'hello\n\nnewer', fileRefs: ['a','b'] });
});
