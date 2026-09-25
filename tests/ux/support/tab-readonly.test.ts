import {expect,test} from 'bun:test';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {tabStore, type TabStore} from '../../../web/src/panes/tab-store';
import {patchTabReadonly} from '../../../scripts/patch-tab-readonly.mjs';
import {readonlyTabDestination} from '../../../web/src/gi-readonly-tab-focus';
test('read-only tab navigation wraps and ignores unrelated keys',()=>{
 expect(readonlyTabDestination('Home',2,3)).toBe(0);expect(readonlyTabDestination('End',0,3)).toBe(2);
 expect(readonlyTabDestination('ArrowRight',2,3)).toBe(0);expect(readonlyTabDestination('ArrowLeft',0,3)).toBe(2);
 expect(readonlyTabDestination('Enter',1,3)).toBeNull();expect(readonlyTabDestination('Home',0,0)).toBeNull();expect(readonlyTabDestination('ArrowRight',-1,3)).toBeNull();
});

test('supplied tab store preserves pinned MRU, atomic snapshots and individual close',()=>{
 const store = new (tabStore.constructor as {new():TabStore})();
 const changes:any[]=[];const stop=store.onChange((tabs,active)=>changes.push({ids:tabs.map(t=>t.id),active}));
 store.open('a');store.open('b');store.open('c');store.togglePin('a');
 store.activate('a');store.activate('b');store.close('b');expect(store.getActiveId()).toBe('a');
 expect(changes.at(-1)).toEqual({ids:['a','c'],active:'a'});
 store.open('b');store.closeOthers('c');expect(store.getTabs().map(t=>t.id)).toEqual(['a','c']);expect(store.getActiveId()).toBe('c');
 store.closeAll();expect(store.getTabs().map(t=>t.id)).toEqual(['a']);expect(store.getActiveId()).toBe('a');
 store.close('a');expect(store.size).toBe(0);expect(store.getActiveId()).toBeNull();
 const n=changes.length;stop();store.open('later');expect(changes).toHaveLength(n);
 for(const change of changes)expect(change.active===null||change.ids.includes(change.active)).toBe(true);
});

test('read-only tab adapter is opt-in, fail-closed and preserves supplied actions',()=>{
 const source=readFileSync('web/src/components/tab-strip.ts','utf8');const patched=patchTabReadonly(source);
 expect(patched).toContain('readOnlyHost = false');expect(patched).toContain("if (readOnlyHost) return null;");
 expect(patched).toContain("document.querySelector('.settings-dialog[aria-modal=\"true\"]')");
 expect(patched).toContain('window.innerWidth - 148');expect(patched).toContain('window.innerHeight - 208');
 expect(patched).toContain('Close context menu on outside click or Escape\n    useLayoutEffect');
 for(const action of ['onClose?.(contextMenu.id)','onCloseOthers?.(contextMenu.id)','onCloseAll?.()','onTogglePin?.(contextMenu.id)'])expect(patched).toContain(action);
 for(const changed of [source.replace('const onKeyDown = (e) => {','changed'),source+source,patched])expect(()=>patchTabReadonly(changed)).toThrow('anchor changed');
 // The supplied store is part of the same named Piclaw subset, not a local fork.
 expect(createHash('sha256').update(readFileSync('web/src/panes/tab-store.ts')).digest('hex')).toBe('858522603c715ccbdb6ea52b3504b41fe488a4ad119303a2e193a9f91833fb36');
});
