import {test,expect} from 'bun:test';
// Rendering comes from the production browser bundle. Exercise pure timing
// without loading the browser-only vendor module in Bun.
import {readFileSync} from 'node:fs';
const source=readFileSync('web/src/gi-tool-activity.ts','utf8');
const body=source.slice(source.indexOf('export function toolElapsed'),source.indexOf('export function ToolActivity')).replace('export ','');
const elapsed=new Function(`${body}; return toolElapsed;`)();
test('running tool uses authoritative start and completed duration freezes without guessing',()=>{
 expect(elapsed({state:'running',started_at:'2026-01-01T00:00:00Z'},Date.parse('2026-01-01T00:00:03Z'))).toBe('3s');
 for(const state of ['completed','failed','cancelled','aborted','interrupted'])expect(elapsed({state,duration_ms:2450},99999)).toBe('2s');
 for(const tool of [{state:'running'},{state:'completed',duration_ms:null},{state:'running',started_at:'2099-01-01'}])expect(elapsed(tool,0)).toBe('?');
 expect(source).toContain('clearInterval(timer)');expect(source).toContain('1000');
});
