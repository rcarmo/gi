import {test, expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {filterModelOptions, modelPickerKey} from '../../../web/src/gi-model-picker';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';
const event=(key:string, extra:any={})=>({key, ...extra}) as KeyboardEvent;
const entries=[{label:'substring pine',disabled:false},{label:'pine blocked',disabled:true},{label:'pine',disabled:false},{label:'piper',disabled:false}];
const empty={value:'',updatedAt:0};
test('filter retains order and identity across case-insensitive metadata terms',()=>{
 const options=[{label:'Provider/one Name One 32K ctx'}, {label:'Provider/two Name Two 200 ctx'}];
 expect(filterModelOptions(options,' provider  ONE 32k ',x=>x.label)).toEqual([options[0]]);
 expect(filterModelOptions(options,'',x=>x.label)[0]).toBe(options[0]);
 expect(filterModelOptions(options,'missing',x=>x.label)).toEqual([]);
});
test('enabled navigation preserves full indices; prefix wins and native editing/buttons retain ownership',()=>{
 for(const [key,index] of [['ArrowDown',2],['ArrowUp',3],['Home',0],['End',3],['PageDown',3],['PageUp',0]] as const)
  expect(modelPickerKey(event(key),entries,0,empty)?.index).toBe(index);
 const p=modelPickerKey(event('p'),entries,0,empty)!;expect(p.index).toBe(2);
 expect(modelPickerKey(event('i'),entries,p.index,p.buffer)?.index).toBe(2);
 expect(modelPickerKey(event('x'),entries,0,empty)?.index).toBe(-1);
 const target={closest:(selector:string)=>selector.includes('input')?{}:null};
 for(const key of ['Home','End','p',' ','Backspace','Tab']) expect(modelPickerKey(event(key,{target}),entries,0,empty)).toBeNull();
 expect(modelPickerKey(event('ArrowDown',{target}),entries,0,empty)).toMatchObject({index:2,focus:false});
 expect(modelPickerKey(event('Enter'),entries,1,empty)?.activate).toBe(false);
 expect(modelPickerKey(event('Enter'),[],0,empty)?.activate).toBe(false);
 expect(modelPickerKey(event('Enter',{repeat:true}),entries,2,empty)?.activate).toBe(false);
 const button={closest:(s:string)=>s==='button'?{}:s.includes('data-model')?{getAttribute:()=> '3'}:null};
 expect(modelPickerKey(event('Enter',{target:button}),entries,0,empty)).toBeNull();
 expect(modelPickerKey(event(' ' ,{target:button}),entries,0,empty)).toBeNull();
 expect(modelPickerKey(event('ArrowDown',{target:button}),entries,0,empty)?.index).toBe(0);
 for(const flag of ['defaultPrevented','isComposing','ctrlKey','metaKey','altKey'])
  expect(modelPickerKey(event('ArrowDown',{[flag]:true}),entries,0,empty)).toBeNull();
 expect(modelPickerKey(event('End'),entries.map(x=>({...x,disabled:true})),0,empty)?.index).toBe(-1);
});
test('guarded filter adapter composes without editing supplied bytes and fails closed on drift',()=>{
 const path='web/src/components/compose-box.ts',source=readFileSync(path,'utf8');
 const adapted=patchModelPicker(patchComposePopupKeys(source));
 expect(adapted).toContain('aria-label="Search models"');
 expect(adapted).toContain('visibleModels.map((modelOption, index)');
 expect(adapted).toContain('modelPickerKey(e, modelEntries');
 expect(readFileSync(path,'utf8')).toBe(source);
 for(const bad of ['',source+source,source.replace('Select model</div>','Changed</div>'),adapted])
  expect(()=>patchModelPicker(bad)).toThrow('anchor changed');
});
