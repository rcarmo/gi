import {readFile,writeFile} from 'node:fs/promises';
import {resolve,basename} from 'node:path';
import {createHash} from 'node:crypto';
import {PNG} from 'pngjs';
import {comparePixels,unionRegion,extractRegion} from './pixel-images.mjs';

export async function compareCaptureSet(out,captures){
 const {records,viewports,themes,scenarios}=captures;
 const comparisons=[],failures=[];
 for(const [values,allowed]of [[viewports,['phone','tablet','desktop']],[themes,['light','dark']],[scenarios,['compose','models','sessions']]]){
  if(!Array.isArray(values)||!values.length||new Set(values).size!==values.length||values.some(v=>!allowed.includes(v)))throw Error('Invalid capture matrix');
 }
 const expected=viewports.length*themes.length*scenarios.length;
 if(!captures.captureFinished)failures.push('Capture finalisation missing (interrupted run)');
 if(captures.candidateUnchanged!==true)failures.push('Candidate tree final verification missing or failed');
 if(records.length!==expected*4)failures.push('Capture count mismatch');
 const keys=new Set();
 for(const viewport of viewports)for(const theme of themes)for(const scenario of scenarios)keys.add(`${viewport}-${theme}-${scenario}`);
 for(const record of records)if(!keys.has(record.key)||record.status!=='captured'||record.failures?.length)failures.push(`Invalid capture: ${record.key}/${record.host}/${record.repeat}`);
 for(const key of keys){
  const group=records.filter(r=>r.key===key);
  if(group.length!==4||group.some(r=>r.status!=='captured')){failures.push(`Incomplete group: ${key}`);continue;}
  const find=(host,repeat)=>{const matches=group.filter(r=>r.host===host&&r.repeat===repeat);if(matches.length!==1)throw Error(`Duplicate/missing capture: ${key}/${host}/${repeat}`);return matches[0];};
  try{
   const images=new Map();
   for(const record of group){
    if(basename(record.file)!==record.file)throw Error('Invalid capture filename');
    const bytes=await readFile(resolve(out,record.file));
    if(record.sha256!==createHash('sha256').update(bytes).digest('hex'))failures.push(`Capture digest missing or changed: ${record.file}`);
    images.set(record,PNG.sync.read(bytes));
   }
   for(const [label,left,right]of [['piclaw-repeat',find('piclaw',1),find('piclaw',2)],['gi-repeat',find('gi',1),find('gi',2)],['cross',find('piclaw',1),find('gi',1)]]){
    if(label!=='cross'&&JSON.stringify(Object.entries(left.assets).sort())!==JSON.stringify(Object.entries(right.assets).sort()))failures.push(`${key}/${label}: asset set changed`);
    const a=images.get(left),b=images.get(right);
    for(const regionName of ['full','compose',...(left.scenario==='compose'?[]:['panel'])]){
     const region=regionName==='full'?null:unionRegion([left.dom[regionName],right.dom[regionName]],a.width,a.height);
     const {diff,overlay,...metrics}=comparePixels(region?extractRegion(a,region):a,region?extractRegion(b,region):b);
     const prefix=`${key}-${label}-${regionName}`;
     // Low compression changes only PNG encoding, never decoded RGBA values.
     if(diff)await writeFile(resolve(out,`${prefix}-diff.png`),PNG.sync.write(diff,{deflateLevel:1}));
     if(overlay&&label==='cross')await writeFile(resolve(out,`${prefix}-overlay.png`),PNG.sync.write(overlay,{deflateLevel:1}));
     comparisons.push({key,label,regionName,region,...metrics});
    }
   }
  }catch(error){failures.push(`${key}: ${error.message}`);}
 }
 const full=comparisons.filter(c=>c.regionName==='full');
 const repeatStable=full.filter(c=>c.label!=='cross').length===expected*2&&full.filter(c=>c.label!=='cross').every(c=>c.exactMatch);
 const crossEqual=full.filter(c=>c.label==='cross').length===expected&&full.filter(c=>c.label==='cross').every(c=>c.exactMatch);
 const result={captureComplete:failures.length===0,repeatStable,crossEqual,selectedMatrixOnly:viewports.length!==3||themes.length!==2||scenarios.length!==3,failures,comparisons};
 result.passed=result.captureComplete&&repeatStable&&crossEqual;
 await writeFile(resolve(out,'comparison.json'),JSON.stringify(result,null,2)+'\n');
 return result;
}
