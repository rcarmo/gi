import {test,expect} from 'bun:test';
import {compactionNotice,compactionElapsed,createActivityRevision,compactionUnavailableReason} from '../../../web/src/gi-compaction-state';
test('disabled compaction reasons prefer connection and freshness over stale capability',()=>{
 const state={fresh:true,disconnected:false,pending:false,status:'idle',capability:{available:false,reason:'  Context is too short to compact  '}};
 expect(compactionUnavailableReason(state)).toBe('Context is too short to compact');
 expect(compactionUnavailableReason({...state,capability:{available:true}})).toBe('');
 expect(compactionUnavailableReason({...state,capability:{reason:23}})).toBe('Compaction is unavailable');
 expect(compactionUnavailableReason({...state,capability:{reason:'  '}})).toBe('Compaction is unavailable');
 expect(compactionUnavailableReason({...state,status:'running'})).toBe('Session has active or queued work');
 expect(compactionUnavailableReason({...state,fresh:false})).toBe('Refreshing compaction availability');
 expect(compactionUnavailableReason({...state,pending:true,fresh:false})).toBe('Compaction request pending');
 expect(compactionUnavailableReason({...state,disconnected:true,pending:true})).toBe('Reconnect to check compaction availability');
});
test('native compaction snapshot owns notice; stale revisions and foreign runs cannot paint state',()=>{
 const guard=createActivityRevision();const captured=guard.capture();expect(guard.accepts(captured)).toBe(true);guard.invalidate();expect(guard.accepts(captured)).toBe(false);
 const now=Date.parse('2026-09-22T12:00:04Z');const c={active:true,turn_id:'one',seq:3,timestamp:'2026-09-22T12:00:00Z'};
 expect(compactionNotice({turn_id:'two',status:'running',compaction:c},now)).toBeNull();
 const active=compactionNotice({turn_id:'one',status:'running',compaction:c},now);expect(active.title).toBe('Compacting context');expect(compactionElapsed(active,now)).toBe('0:04');
 expect(compactionNotice({turn_id:'one',status:'cancelling',compaction:c},now).title).toBe('Cancelling compaction');
 expect(compactionNotice({turn_id:'one',status:'idle',compaction:c},now)).toBeNull();
 const suppressed={turn_id:'one',status:'idle',compaction:{...c,active:false,event_type:'compaction.suppressed',detail:'hook policy'}};
 expect(compactionNotice(suppressed,now).detail).toBe('hook policy');expect(compactionNotice(suppressed,now+11000)).toBeNull();
});
