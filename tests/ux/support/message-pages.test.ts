import {test,expect} from 'bun:test';
import {mergeMessagePages,newMessageWindow} from '../../../web/src/gi-message-pages';
test('page merge keeps chronological timestamp/id order and dedupes live replies',()=>{
 const rows=mergeMessagePages([{id:'b',timestamp:'1',content:'old'},{id:'d',timestamp:'2'}],[{id:'a',timestamp:'1'},{id:'b',timestamp:'1',content:'fresh'},{id:'c',timestamp:'1'}]);
 expect(rows.map(x=>x.id)).toEqual(['a','b','c','d']);expect(rows[1].content).toBe('fresh');expect(newMessageWindow()).toEqual({loaded:false,before:null,after:null,hasMore:false});
});
