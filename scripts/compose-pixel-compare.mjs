import {readFile,writeFile} from 'node:fs/promises';
import {resolve,dirname} from 'node:path';
import {fileURLToPath} from 'node:url';
import {compareCaptureSet} from '../tests/ux/support/pixel-comparison.mjs';
const repo=resolve(dirname(fileURLToPath(import.meta.url)),'..');
const out=resolve(repo,process.env.PIXEL_RUN_DIR||'test-results/compose-pixels');
try{
 if(!process.env.PIXEL_RUN_DIR)throw Error('PIXEL_RUN_DIR must name one capture run directory');
 const result=await compareCaptureSet(out,JSON.parse(await readFile(resolve(out,'captures.json'),'utf8')));
 console.log(JSON.stringify(result,null,2));
 process.exitCode=!result.captureComplete?2:result.passed?0:1;
}catch(error){
 if(process.env.PIXEL_RUN_DIR)await writeFile(resolve(out,'comparison.json'),JSON.stringify({passed:false,captureComplete:false,reason:error.message},null,2)+'\n').catch(()=>{});
 console.error(error.message);process.exitCode=2;
}
