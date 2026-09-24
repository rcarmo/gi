import {test,expect} from 'bun:test';
test('unsupported API action rejects asynchronously for every payload without success acknowledgement',async()=>{
 const source=await Bun.file('web/src/api.ts').text();
 const fn=source.match(/export async function submitAdaptiveCardAction\(_payload: unknown\) \{[\s\S]*?\n\}/)?.[0];expect(fn).toBeTruthy();
 const js=new Bun.Transpiler({loader:'ts'}).transformSync(fn!.replace('export ',''));
 const submit=new Function(js+';return submitAdaptiveCardAction;')();
 for(const payload of [null,{}, {card_id:'native',data:{answer:'private value'}}]){
  const result=submit(payload);expect(result).toBeInstanceOf(Promise);
  await expect(result).rejects.toThrow('Card submissions are not supported by Gi yet. Your inputs have not been submitted.');
 }
 expect(fn).not.toContain('request(');expect(fn).not.toContain('JSON.stringify');
});
