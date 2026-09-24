import {expect,test} from 'bun:test';
import {ownsHorizontalGesture} from '../../../web/src/gi-scroll-gesture';
test('nested horizontal scrollers own gestures at either edge, ordinary content and boundary do not',()=>{
 const previous=globalThis.getComputedStyle;globalThis.getComputedStyle=((el:any)=>({overflowX:el.overflow})) as any;
 try{
  const boundary:any={clientWidth:100,scrollWidth:500,overflow:'auto'},outer:any={parentElement:boundary,clientWidth:200,scrollWidth:200,overflow:'auto'};
  const scroller:any={parentElement:outer,clientWidth:100,scrollWidth:400,overflow:'auto',scrollLeft:0},cell:any={parentElement:scroller,clientWidth:40,scrollWidth:40,overflow:'visible'};
  expect(ownsHorizontalGesture(cell,boundary)).toBe(true);scroller.scrollLeft=300;expect(ownsHorizontalGesture(cell,boundary)).toBe(true);
  scroller.overflow='hidden';expect(ownsHorizontalGesture(cell,boundary)).toBe(false);scroller.overflow='scroll';expect(ownsHorizontalGesture(cell,boundary)).toBe(true);
  scroller.scrollWidth=101;expect(ownsHorizontalGesture(cell,boundary)).toBe(false);expect(ownsHorizontalGesture(outer,boundary)).toBe(false);expect(ownsHorizontalGesture(boundary,boundary)).toBe(false);expect(ownsHorizontalGesture(null,boundary)).toBe(false);
 }finally{globalThis.getComputedStyle=previous;}
});
