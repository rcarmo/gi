import {test,expect} from 'bun:test';
import {createTimelineRevision,createAssetVersionGuard,loadedAssetVersion} from '../../../web/src/gi-refresh-guards';
test('HTTP timeline ownership rejects older reloads, pagination and disconnect generations',()=>{
 const g=createTimelineRevision();const old=g.begin();const fresh=g.begin();expect(g.accepts(old)).toBe(false);expect(g.accepts(fresh)).toBe(true);
 g.invalidate();expect(g.accepts(fresh)).toBe(false);const reconnect=g.begin();g.invalidate();expect(g.accepts(reconnect)).toBe(false);
});
test('asset drift compares loaded page version and deduplicates without reload side effects',()=>{
 const g=createAssetVersionGuard('loaded');expect(g.observe('loaded')).toBe(false);expect(g.observe('new')).toBe(true);expect(g.observe('new')).toBe(false);expect(g.observe('next')).toBe(true);expect(g.observe('new')).toBe(false);expect(g.observe(null)).toBe(false);
 const initial=createAssetVersionGuard(null);expect(initial.observe('first')).toBe(false);expect(initial.observe('second')).toBe(true);
 expect(loadedAssetVersion({baseURI:'https://example.test',querySelector:()=>({getAttribute:()=>'/dist/app.bundle.js?v=123'})} as any)).toBe('123');
});
