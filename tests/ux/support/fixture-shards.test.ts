import { test, expect } from 'bun:test';
import { fixtureProjects, mergeFixtureResults } from '../../../scripts/merge-fixture-results';
const reports = () => fixtureProjects.map(projectName => ({ config:{projects:[{name:projectName}]}, errors:[], suites:[{specs:[{title:'scenario',tests:[{projectName,status:'expected'}]}]}],stats:{expected:1} }));
test('six isolated projects merge without dropping scenario evidence',()=>{const merged=mergeFixtureResults(reports());expect(merged.suites).toHaveLength(6);expect(merged.stats.expected).toBe(6)});
test('missing duplicate errored empty failed and flaky shards fail closed',()=>{
 expect(()=>mergeFixtureResults(reports().slice(1))).toThrow();
 const duplicate=reports();duplicate[5]=duplicate[0];expect(()=>mergeFixtureResults(duplicate)).toThrow();
 for(const status of ['unexpected','flaky']){const rows=reports();rows[0].suites[0].specs[0].tests[0].status=status;expect(()=>mergeFixtureResults(rows)).toThrow()}
 const empty=reports();empty[0].suites=[];expect(()=>mergeFixtureResults(empty)).toThrow();
 const errors=reports();(errors[0].errors as any[]).push({message:'crash'});expect(()=>mergeFixtureResults(errors)).toThrow();
});
