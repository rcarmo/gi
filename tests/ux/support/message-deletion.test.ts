import {test,expect} from 'bun:test';
import {createMessageDeletionState} from '../../../web/src/gi-message-deletion';
test('only successful removal fences late pages and repeated delete admission',()=>{
 const s=createMessageDeletionState(),rows=[{id:'a'},{id:'b'}];expect(s.begin('a')).toBe(true);expect(s.begin('a')).toBe(false);
 expect(s.filter(rows)).toEqual(rows);s.finish('a',false);expect(s.begin('a')).toBe(true);s.finish('a',true);
 expect(s.filter(rows)).toEqual([{id:'b'}]);expect(s.filter(rows,new Set(['a']))).toEqual(rows);expect(s.begin('a')).toBe(false);
});
