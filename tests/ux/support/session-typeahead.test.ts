import {test,expect} from 'bun:test';
import {sessionTypeahead} from '../../../web/src/gi-session-typeahead';
const event=(key:string,props:any={})=>({key,target:{closest:()=>null},...props} as any);
test('session typeahead prefers prefixes and returns original enabled index',()=>{
 const entries=[{label:'@z-alpha'}, {label:'@alpha',disabled:true},{label:'@alpha-one'},{label:'@alpine'}];
 const a=sessionTypeahead(event('a'),entries,null)!;expect(a.index).toBe(2);expect(a.buffer.value).toBe('a');
 const l=sessionTypeahead(event('l'),entries,a.buffer)!;const p=sessionTypeahead(event('p'),entries,l.buffer)!;const i=sessionTypeahead(event('i'),entries,p.buffer)!;expect(i.index).toBe(3);expect(i.buffer.value).toBe('alpi');
 expect(sessionTypeahead(event('z'),entries,{value:'old',updatedAt:0})?.index).toBe(0);
 expect(sessionTypeahead(event('x'),entries,null)?.index).toBe(-1);
 expect(sessionTypeahead(event('a'),[{label:'alpha',disabled:true}],null)?.index).toBe(-1);
});
test('session typeahead leaves native editing, modifiers and composing keys alone',()=>{
 for(const props of [{repeat:true},{defaultPrevented:true},{isComposing:true},{ctrlKey:true},{altKey:true},{metaKey:true},{target:{closest:()=>({})}}])expect(sessionTypeahead(event('a',props),[{label:'alpha'}],null)).toBeNull();
 for(const key of [' ','Tab','Enter','ArrowDown'])expect(sessionTypeahead(event(key),[{label:'alpha'}],null)).toBeNull();
});
