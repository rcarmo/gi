import {test,expect} from 'bun:test';
import {createActivationRefreshGate,createTimelineRevision,createAssetVersionGuard,loadedAssetVersion} from '../../../web/src/gi-refresh-guards';
test('HTTP timeline ownership rejects older reloads, pagination and disconnect generations',()=>{
 const g=createTimelineRevision();const old=g.begin();const fresh=g.begin();expect(g.accepts(old)).toBe(false);expect(g.accepts(fresh)).toBe(true);
 g.invalidate();expect(g.accepts(fresh)).toBe(false);const reconnect=g.begin();g.invalidate();expect(g.accepts(reconnect)).toBe(false);
});
test('asset drift compares loaded page version and deduplicates without reload side effects',()=>{
 const g=createAssetVersionGuard('loaded');expect(g.observe('loaded')).toBe(false);expect(g.observe('new')).toBe(true);expect(g.observe('new')).toBe(false);expect(g.observe('next')).toBe(true);expect(g.observe('new')).toBe(false);expect(g.observe(null)).toBe(false);
 const initial=createAssetVersionGuard(null);expect(initial.observe('first')).toBe(false);expect(initial.observe('second')).toBe(true);
 expect(loadedAssetVersion({baseURI:'https://example.test',querySelector:()=>({getAttribute:()=>'/dist/app.bundle.js?v=123'})} as any)).toBe('123');
});
test('activation/readiness share one initial refresh and real reconnect opens a new epoch',()=>{
 const g=createActivationRefreshGate();
 expect(g.activate(1)).toBe(false);expect(g.ready(1)).toBe(false);expect(g.status(1,'connected')).toBe(true);expect(g.activate(1)).toBe(false);expect(g.status(1,'connected')).toBe(false);
 expect(g.status(1,'disconnected')).toBe(false);expect(g.ready(1)).toBe(false);expect(g.status(1,'connected')).toBe(true);
 // Opposite effect ordering and A -> B -> A use distinct selection generations.
 expect(g.status(2,'connected')).toBe(true);expect(g.activate(2)).toBe(false);expect(g.ready(1)).toBe(false);
 g.select(3);expect(g.ready(3)).toBe(false);expect(g.status(3,'connected')).toBe(true);
});
