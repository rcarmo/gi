import {test,expect} from 'bun:test';
import {mkdtempSync,readFileSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
import {execFileSync} from 'node:child_process';

const script=resolve(import.meta.dir,'../../../scripts/ux-parity-report.mjs');
const projects=['chromium-phone','chromium-tablet','chromium-desktop','webkit-phone','webkit-tablet','webkit-desktop'];
test('parity report keeps shared evidence separate and requires every project',()=>{
 const dir=mkdtempSync(join(tmpdir(),'gi-parity-report-'));
 try{
  const input=join(dir,'results.json');
  const spec=(id:string,count:number)=>({title:`${id} mapped test`,tests:projects.slice(0,count).map(projectName=>({projectName,results:[{status:'passed'}]}))});
  writeFileSync(input,JSON.stringify({suites:[{specs:[spec('@shared-28',6),spec('@ux-original-001',5)]}]}));
  execFileSync(process.execPath,[script,input],{cwd:dir});
  const report=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(report.scenarios).toBe(236);expect(report.sharedContractCases).toBe(42);
  expect(report.sharedCounts).toEqual({unmapped:40,pass:1,'not-run':1});expect(report.counts.pass).toBeUndefined();
  expect(report.rows.find((row:any)=>row.id==='@ux-original-001').status).toBe('partial-matrix');
  expect(report.rows.some((row:any)=>row.id==='@shared-28')).toBe(false);
  expect(report.sharedRows.find((row:any)=>row.id==='@shared-28').status).toBe('pass');
  expect(readFileSync(join(dir,'test-results/ux-parity/matrix.md'),'utf8')).toContain('Shared contract (separate evidence)');
  const steer=join(dir,'steer.json');writeFileSync(steer,JSON.stringify({suites:[{specs:[spec('@shared-30',6)]}]}));
  execFileSync(process.execPath,[script,input,steer],{cwd:dir});
  const combined=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(combined.sharedCounts).toEqual({unmapped:40,pass:2});
  execFileSync(process.execPath,[script,steer,steer],{cwd:dir});
  const repeated=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(repeated.sharedRows.find((row:any)=>row.id==='@shared-30').status).toBe('partial-matrix');
 }finally{rmSync(dir,{recursive:true,force:true});}
});
