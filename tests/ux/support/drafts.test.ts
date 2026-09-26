import { expect, test } from 'bun:test';
import { createDraftRepository, emptyDraft, mergeDrafts, DraftConflictError } from '../../../web/src/gi-drafts';

const draft = (text: string) => ({ ...emptyDraft(), text });
function storage() {
  const rows = new Map();
  return { rows, load: async () => structuredClone([...rows.values()]), put: async (row: any, expected: number) => {
    if ((rows.get(row.sessionId)?.revision ?? 0) !== expected) throw new DraftConflictError();
    const revision = expected + 1; rows.set(row.sessionId, structuredClone({ ...row, revision })); return revision;
  } };
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
  const repo=createDraftRepository({load:disk.load,put:async (row,expected)=>{writes.push(structuredClone(row));if(fail)throw Error('quota');return disk.put(row,expected);}});
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

test('stale second repository cannot erase a foreign pending capture or dispatch its own', async () => {
 const db=storage(), errors:Error[]=[];
 const a=createDraftRepository(db),b=createDraftRepository(db,e=>errors.push(e));await a.load();await b.load();
 const pending=a.begin('A',draft('foreign pending'));await pending.ready;
 b.update('A',draft('local edit'));await expect(b.flushStable()).rejects.toThrow('another tab');
 expect(db.rows.get('A').pending.map(p=>p.id)).toEqual([pending.token]);
 const rejected=b.begin('A',draft('local edit'));await expect(rejected.ready).rejects.toThrow('another tab');
 b.failed('A',rejected.token,'not sent');await expect(b.flushStable()).rejects.toThrow('another tab');
 expect(b.get('A').text).toBe('local edit');expect(errors.length).toBeGreaterThan(0);
 const reopened=createDraftRepository(db);await reopened.load();expect(reopened.get('A').text).toBe('foreign pending');
});

test('stale unknown recovery cannot publish after confirmed cleanup commits', async () => {
 const {pendingSendKey}=await import('../../../web/src/gi-drafts');
 const db=storage(),seed=createDraftRepository(db);const send=seed.begin('A',draft('accepted'));await send.ready;
 let entered!:()=>void,release!:()=>void;const started=new Promise<void>(r=>entered=r),gate=new Promise<void>(r=>release=r);
 const stale=createDraftRepository(db,()=>{},async()=>{entered();await gate;return new Set()});
 const loading=stale.load();await started;
 const confirmed=createDraftRepository(db,()=>{},async()=>new Set([pendingSendKey('A',send.token)]));await confirmed.load();
 release();await expect(loading).rejects.toThrow('another tab');
 expect(stale.get('A').text).toBe('');expect(db.rows.get('A').draft.text).toBe('');expect(db.rows.get('A').pending).toEqual([]);
});

test('queued same-tab writes use last committed revision, not snapshot revision',async()=>{
 const db=storage(),repo=createDraftRepository(db);await repo.load();
 for(let i=0;i<30;i++)repo.update('A',draft(`edit ${i}`));
 const send=repo.begin('A',draft('sent'));repo.update('A',draft('later'));await send.ready;await repo.accepted('A',send.token);await repo.flushStable();
 expect(db.rows.get('A').revision).toBe(33);expect(db.rows.get('A').draft.text).toBe('later');expect(db.rows.get('A').pending).toEqual([]);
});

test('a conflicted session stays frozen but a distinct session can still persist',async()=>{
 const db=storage(),a=createDraftRepository(db),b=createDraftRepository(db);await a.load();await b.load();
 a.update('A',draft('winner'));await a.flushStable();b.update('A',draft('loser'));await expect(b.flushStable()).rejects.toThrow('another tab');
 for(let i=0;i<3;i++)b.update('A',draft(`local ${i}`));await expect(b.flushStable()).rejects.toThrow('another tab');
 b.update('B',draft('independent'));await b.flushStable();expect(db.rows.get('A').draft.text).toBe('winner');expect(db.rows.get('B').draft.text).toBe('independent');expect(b.get('A').text).toBe('local 2');
 await expect(b.begin('A',b.get('A')).ready).rejects.toThrow('another tab');
});

test('stale queue return cannot delete foreign capture; stale completion cannot overwrite newer draft',async()=>{
 const db=storage(),a=createDraftRepository(db),b=createDraftRepository(db);await a.load();await b.load();
 const pending=a.begin('A',draft('foreign send'));await pending.ready;
 await expect(b.prepareQueueReturn('A','queue',draft('queued')).ready).rejects.toThrow('another tab');
 expect(db.rows.get('A').pending[0].id).toBe(pending.token);expect(db.rows.get('A').queueReturns?.queue).toBeUndefined();
 const next=storage(),one=createDraftRepository(next);await one.prepareQueueReturn('A','queue',draft('queued')).ready;
 const two=createDraftRepository(next);await two.load();two.update('A',draft('edited elsewhere'));await two.flushStable();
 await expect(one.completeQueueReturn('A','queue')).rejects.toThrow('another tab');expect(next.rows.get('A').draft.text).toBe('edited elsewhere');expect(next.rows.get('A').queueReturns.queue.state).toBe('prepared');
});

test('legacy rows gain revisions and malformed revisions fail closed before any write',async()=>{
 const db=storage();db.rows.set('A',{sessionId:'A',draft:draft('old'),pending:[]});const repo=createDraftRepository(db);await repo.load();repo.update('A',draft('new'));await repo.flushStable();expect(db.rows.get('A').revision).toBe(1);
 for(const revision of [-1,1.5,Number.MAX_SAFE_INTEGER+1,'1']){
  const invalid=storage();invalid.rows.set('A',{sessionId:'A',revision,draft:draft('untrusted'),pending:[]});const bad=createDraftRepository(invalid);
  await expect(bad.load()).rejects.toThrow('revision');expect(bad.get('A').text).toBe('');await expect(bad.begin('A',draft('blocked')).ready).rejects.toThrow('revision');expect(invalid.rows.get('A').draft.text).toBe('untrusted');
 }
});

test('unknown recovery which commits first remains conservative and later accepted cleanup conflicts',async()=>{
 const db=storage(),a=createDraftRepository(db);const send=a.begin('A',draft('possibly accepted'));await send.ready;
 const unknown=createDraftRepository(db);await unknown.load();await expect(a.accepted('A',send.token)).rejects.toThrow('another tab');
 expect(db.rows.get('A').draft.text).toBe('possibly accepted');expect(db.rows.get('A').error).toContain('Delivery is unknown');
});

test('bootstrap retries cannot unfreeze a failed or conflicted repository; full reload can',async()=>{
 const db=storage(),a=createDraftRepository(db),b=createDraftRepository(db);await a.load();await b.load();a.update('A',draft('durable'));await a.flushStable();b.update('A',draft('local unsaved'));await expect(b.flushStable()).rejects.toThrow('another tab');
 await expect(b.load()).rejects.toThrow('another tab');expect(b.get('A').text).toBe('local unsaved');
 let fail=true;const bad=createDraftRepository({...db,load:async()=>{if(fail)throw Error('offline');return db.load()}});await expect(bad.load()).rejects.toThrow('offline');fail=false;
 await expect(bad.load()).rejects.toThrow('offline');await expect(bad.begin('A',draft('not sent')).ready).rejects.toThrow('offline');
 const fresh=createDraftRepository(db);await fresh.load();expect(fresh.get('A').text).toBe('durable');fresh.update('A',draft('after reload'));await fresh.flushStable();expect(db.rows.get('A').draft.text).toBe('after reload');
});
