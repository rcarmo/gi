import {test,expect} from 'bun:test';
import {createSearchView} from '../../../web/src/gi-search-state';
test('search owns open/query/scope/close generations before rendering',()=>{
 const state=createSearchView(),timeline=state.capture();const open=state.enter();expect(state.isCurrent(timeline)).toBe(false);expect(open.active).toBe(true);
 const a=state.query(' needle ');expect(a.query).toBe('needle');const b=state.scope('all');expect(b.query).toBe('needle');expect(state.isCurrent(a)).toBe(false);
 state.close();state.enter();expect(state.isCurrent(b)).toBe(false);expect(state.capture()).toMatchObject({active:true,query:'',scope:'current'});
 expect(state.scope('bogus').scope).toBe('current');
});
