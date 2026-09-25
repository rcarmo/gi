import {test,expect} from 'bun:test';
import {PNG} from 'pngjs';
import {comparePixels,unionRegion,extractRegion} from './pixel-images.mjs';
const image=()=>{const p=new PNG({width:4,height:3});p.data.fill(255);return p;};
test('exact RGBA rejects a one-channel one-bit difference including alpha',()=>{
 for(const channel of [0,1,2,3]){const a=image(),b=image();b.data[channel]=254;const result=comparePixels(a,b);expect(result.exactMatch).toBe(false);expect(result.changedPixels).toBe(1);expect(result.totalPixels).toBe(12);}
 expect(comparePixels(image(),image()).exactMatch).toBe(true);
});
test('different dimensions fail without scaling',()=>{
 const result=comparePixels(image(),new PNG({width:3,height:4}));expect(result.dimensionsMatch).toBe(false);expect(result.exactMatch).toBe(false);
});
test('region union preserves absolute positions, floors/ceils and clips',()=>{
 const region=unionRegion([{x:-0.2,y:0.8,width:2,height:2},{x:2,y:1,width:3,height:4}],4,3);
 expect(region).toEqual({x:0,y:0,width:4,height:3});expect(extractRegion(image(),region).data).toEqual(image().data);
 expect(()=>unionRegion([null],4,3)).toThrow();expect(()=>unionRegion([{x:9,y:9,width:1,height:1}],4,3)).toThrow();
});
test('diagnostic region cannot make full-frame displacement equal',()=>{
 const a=image(),b=image();a.data[0]=0;b.data[4]=0;
 expect(comparePixels(a,b).exactMatch).toBe(false);
 const region=unionRegion([{x:0,y:0,width:1,height:1},{x:1,y:0,width:1,height:1}],4,3);
 expect(comparePixels(extractRegion(a,region),extractRegion(b,region)).changedPixels).toBe(2);
});
