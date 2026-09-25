import {test,expect} from 'bun:test';
import {mkdtemp,writeFile,rm,readFile} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {createHash} from 'node:crypto';
import {PNG} from 'pngjs';
import {compareCaptureSet} from './pixel-comparison.mjs';
async function fixture(run){
 const out=await mkdtemp(join(tmpdir(),'gi-pixels-'));
 const png=new PNG({width:4,height:3});png.data.fill(255);const bytes=PNG.sync.write(png),sha256=createHash('sha256').update(bytes).digest('hex');
 const records=[];for(const host of ['piclaw','gi'])for(const repeat of [1,2]){const file=`${host}-${repeat}.png`;await writeFile(join(out,file),bytes);records.push({key:'phone-light-compose',host,repeat,scenario:'compose',status:'captured',file,sha256,assets:{},failures:[],dom:{compose:{x:0,y:0,width:4,height:3}}});}
 const captures={viewports:['phone'],themes:['light'],scenarios:['compose'],captureFinished:true,candidateUnchanged:true,records};
 try{await run(out,captures);}finally{await rm(out,{recursive:true,force:true});}
}
test('bounded exact capture set passes with selected-matrix disclosure',()=>fixture(async(out,captures)=>{
 const result=await compareCaptureSet(out,captures);expect(result.passed).toBe(true);expect(result.selectedMatrixOnly).toBe(true);expect(result.comparisons).toHaveLength(6);
 expect(JSON.parse(await readFile(join(out,'comparison.json'),'utf8')).passed).toBe(true);
}));
test('empty matrix cannot pass vacuously',()=>fixture(async(out,captures)=>{
 captures.viewports=[];captures.records=[];await expect(compareCaptureSet(out,captures)).rejects.toThrow('Invalid capture matrix');
}));
test('missing finalisation or changed/missing digest fails closed',()=>fixture(async(out,captures)=>{
 captures.captureFinished=false;delete captures.records[0].sha256;
 const result=await compareCaptureSet(out,captures);expect(result.captureComplete).toBe(false);expect(result.passed).toBe(false);expect(result.failures).toHaveLength(2);
}));
test('missing or duplicate repeats fail closed',()=>fixture(async(out,captures)=>{
 captures.records[1].repeat=1;
 expect((await compareCaptureSet(out,captures)).passed).toBe(false);
 captures.records.pop();expect((await compareCaptureSet(out,captures)).captureComplete).toBe(false);
}));
test('a single altered pixel fails exact repeat and cross comparison',()=>fixture(async(out,captures)=>{
 const record=captures.records[2],png=PNG.sync.read(await readFile(join(out,record.file)));png.data[0]=254;const bytes=PNG.sync.write(png);await writeFile(join(out,record.file),bytes);record.sha256=createHash('sha256').update(bytes).digest('hex');
 const result=await compareCaptureSet(out,captures);expect(result.captureComplete).toBe(true);expect(result.repeatStable).toBe(false);expect(result.crossEqual).toBe(false);expect(result.passed).toBe(false);
}));
