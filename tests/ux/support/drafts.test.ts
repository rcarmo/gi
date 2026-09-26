import { expect, test } from 'bun:test';
import { createDraftRepository, emptyDraft, mergeDrafts } from '../../../web/src/gi-drafts';

const draft = (text: string) => ({ ...emptyDraft(), text });
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

test('queue return persistence gates deletion and retry never merges the durable ID twice', async () => {
  const disk=storage();let fail=true;const writes:any[]=[];
  const repo=createDraftRepository({load:disk.load,put:async row=>{writes.push(structuredClone(row));if(fail)throw Error('quota');await disk.put(row);}});
  repo.update('A',{text:'latest',fileRefs:['new']});
  const capture={...emptyDraft(),text:'queued',fileRefs:['old'],media:[new File(['bytes'],'queued.txt')]};
  const first=repo.prepareQueueReturn('A','turn1',capture);
  await expect(first.ready).rejects.toThrow('quota');
  expect(repo.get('A').text).toBe('queued\n\nlatest');
  repo.update('A',{text:'queued\n\nlatest\nconcurrent'});
  fail=false;await repo.prepareQueueReturn('A','turn1',capture).ready;await repo.flushStable();
  expect(repo.get('A').text).toBe('queued\n\nlatest\nconcurrent');
  expect(disk.rows.get('A').queueReturns.turn1.state).toBe('prepared');
  const loaded=createDraftRepository(disk);await loaded.load();
  await loaded.prepareQueueReturn('A','turn1',capture).ready;
  expect(loaded.get('A').media).toHaveLength(1);expect(loaded.get('A').fileRefs).toEqual(['old','new']);
  await loaded.completeQueueReturn('A','turn1');expect(disk.rows.get('A').queueReturns.turn1.state).toBe('removed');
  expect(repo.get('B').text).toBe('');
});

test('distinct queue IDs with identical text remain distinct; retry keys do not duplicate', async()=>{
 const repo=createDraftRepository(storage());
 await repo.prepareQueueReturn('A','one',{...emptyDraft(),text:'same'}).ready;
 await repo.prepareQueueReturn('A','two',{...emptyDraft(),text:'same'}).ready;
 await repo.prepareQueueReturn('A','two',{...emptyDraft(),text:'same'}).ready;
 expect(repo.get('A').text).toBe('same\n\nsame');
});

test('merge recognises already-restored prefix without losing current references', () => {
  expect(mergeDrafts({ ...emptyDraft(), text: 'hello', fileRefs: ['a'] }, { ...emptyDraft(), text: 'hello\n\nnewer', fileRefs: ['b'] }))
    .toMatchObject({ text: 'hello\n\nnewer', fileRefs: ['a','b'] });
});

test('reopen retires only confirmed exact session tokens and keeps newer draft/media',async()=>{
 const {pendingSendKey}=await import('../../../web/src/gi-drafts');
 const db=storage();let repo=createDraftRepository(db);await repo.load();
 const accepted=repo.begin('A',draft('already delivered'));await accepted.ready;
 const unknown=repo.begin('A',draft('unknown send'));await unknown.ready;
 repo.update('A',draft('newer'));await repo.flushStable();
 let requests:any;
 repo=createDraftRepository(db,()=>{},async pending=>{requests=pending;return new Set([pendingSendKey('A',accepted.token),pendingSendKey('B',unknown.token)])});await repo.load();
 expect(requests).toEqual([{sessionId:'A',token:accepted.token},{sessionId:'A',token:unknown.token}]);
 expect(repo.get('A').text).toBe('unknown send\n\nnewer');expect(repo.error('A')).toContain('Delivery is unknown');expect(db.rows.get('A').pending).toEqual([]);
 repo=createDraftRepository(db);await repo.load();expect(repo.get('A').text).toBe('unknown send\n\nnewer');
});
test('recovery failure preserves legacy unknown draft; confirmed cleanup failure blocks load rather than replay',async()=>{
 const db=storage();let repo=createDraftRepository(db);await repo.load();const accepted=repo.begin('A',draft('pending'));await accepted.ready;
 repo=createDraftRepository(db,()=>{},async()=>{throw Error('offline')});await repo.load();expect(repo.get('A').text).toBe('pending');expect(repo.error('A')).toContain('unknown');
 const next=repo.begin('A',draft('next'));await next.ready;
 const {pendingSendKey}=await import('../../../web/src/gi-drafts');
 repo=createDraftRepository({...db,put:async()=>{throw Error('disk')}},()=>{},async()=>new Set([pendingSendKey('A',next.token)]));await expect(repo.load()).rejects.toThrow('disk');expect(repo.get('A').text).toBe('');
});
