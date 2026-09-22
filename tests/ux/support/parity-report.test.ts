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
  const fit=join(dir,'fit.json');writeFileSync(fit,JSON.stringify({suites:[{specs:[spec('@ux-compaction-006',6),spec('@ux-compaction-007',6)]}]}));
  execFileSync(process.execPath,[script,input,steer,fit],{cwd:dir});
  const all=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(all.counts.pass).toBe(2);expect(all.counts.unmapped).toBe(206);expect(all.sharedCounts.pass).toBe(2);
  const meter=join(dir,'meter.json');writeFileSync(meter,JSON.stringify({suites:[{specs:[spec('@ux-context-001',6),spec('@ux-context-005',6)]}]}));
  execFileSync(process.execPath,[script,input,steer,fit,meter],{cwd:dir});
  const withMeter=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(withMeter.counts.pass).toBe(4);expect(withMeter.counts.unmapped).toBe(206);expect(withMeter.sharedCounts.pass).toBe(2);
  const compact=join(dir,'compact.json');writeFileSync(compact,JSON.stringify({suites:[{specs:['@ux-compaction-001','@ux-compaction-002','@ux-compaction-003','@ux-compaction-004','@ux-compaction-005','@ux-context-004'].map(id=>spec(id,6))}]}));
  execFileSync(process.execPath,[script,input,steer,fit,meter,compact],{cwd:dir});
  const withCompact=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(withCompact.counts.pass).toBe(10);expect(withCompact.counts.unmapped).toBe(206);
  const reconnect=join(dir,'reconnect.json');writeFileSync(reconnect,JSON.stringify({suites:[{specs:[spec('@ux-reconnect-002',6),spec('@ux-reconnect-003',6),spec('@ux-reconnect-004',6),spec('@ux-reconnect-005',6)]}]}));
  execFileSync(process.execPath,[script,reconnect],{cwd:dir});
  const reconnected=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(reconnected.counts.pass).toBe(4);expect(reconnected.counts.unmapped).toBe(206);
  execFileSync(process.execPath,[script,steer,steer],{cwd:dir});
  const repeated=JSON.parse(readFileSync(join(dir,'test-results/ux-parity/matrix.json'),'utf8'));
  expect(repeated.sharedRows.find((row:any)=>row.id==='@shared-30').status).toBe('partial-matrix');
 }finally{rmSync(dir,{recursive:true,force:true});}
});
