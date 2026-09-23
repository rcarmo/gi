import {test,expect} from 'bun:test';
import {writeClipboardTextBestEffort,writeClipboardDataViaExecCommand} from '../../../web/src/gi-clipboard-safety';

test('missing or denied native clipboard cannot claim a successful write',async()=>{
 for(const value of [undefined,null,{}, {writeText:undefined}, {get writeText(){throw new Error('blocked getter')}}])expect(await writeClipboardTextBestEffort(value,'source')).toBe(false);
 expect(await writeClipboardTextBestEffort({writeText:async()=>{throw new Error('denied')}},'source')).toBe(false);
 const target={calls:[] as string[],async writeText(value:string){this.calls.push(value)}};
 expect(await writeClipboardTextBestEffort(target,'**source**')).toBe(true);expect(target.calls).toEqual(['**source**']);
});
test('compatibility adapter retains supplied legacy helper exports',()=>{
 expect(writeClipboardDataViaExecCommand(undefined,{text:'source'})).toBe(false);
});
