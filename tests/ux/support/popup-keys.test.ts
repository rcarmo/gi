import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {blocksQuickActions,settingsOwnsKeyboard} from '../../../web/src/gi-quick-actions';
import {patchQuickActionKeys,patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';

test('palette declines disallowed keys without consuming target events',()=>{
 for(const entry of [{ready:false},{defaultPrevented:true},{repeat:true},{interactive:true},{}]){
  const calls:string[]=[];const event={target:{closest:(selector:string)=>{expect(selector).toContain('button, a');return entry.interactive?{}:null;}},defaultPrevented:Boolean(entry.defaultPrevented),repeat:Boolean(entry.repeat),preventDefault:()=>calls.push('prevent'),stopPropagation:()=>calls.push('stop'),stopImmediatePropagation:()=>calls.push('stopImmediate')};
  expect(blocksQuickActions(event as any,entry.ready!==false)).toBe(Object.keys(entry).length>0);expect(calls).toEqual([]);
 }
 expect(settingsOwnsKeyboard({querySelector:selector=>{expect(selector).toBe('.settings-dialog[aria-modal="true"]');return {} as any;}})).toBe(true);
 expect(settingsOwnsKeyboard({querySelector:()=>null})).toBe(false);
});
test('guarded popup adaptations preserve render content and reject source drift',()=>{
 for(const [path,patch,anchor] of [
  ['web/src/components/timeline-quick-actions.ts',patchQuickActionKeys,'const onKeyDown ='],
  ['web/src/components/compose-box.ts',patchComposePopupKeys,'const handlePopupKeyboardEvent ='],
 ] as const){
  const source=readFileSync(path,'utf8'),out=patch(source);expect(out).toContain('settingsOwnsKeyboard()');
  const render=source.lastIndexOf('    return html`');expect(render).toBeGreaterThan(0);expect(out.slice(out.lastIndexOf('    return html`'))).toBe(source.slice(render));expect(readFileSync(path,'utf8')).toBe(source);
  for(const invalid of ['',source+source,source.replace(anchor,'const changed ='),out])expect(()=>patch(invalid)).toThrow('anchor changed');
 }
 const source=readFileSync('web/src/app.ts','utf8');expect(source).not.toContain('guardQuickActionsTyping');
});
