import {test,expect} from 'bun:test';
import {getProtectedRecoveryControlIntent as control} from '../../../web/src/gi-recovery-control.ts';
import {patchPostRecoveryControl as patch} from '../../../scripts/patch-post-recovery-control.mjs';
import {patchPostSpeech} from '../../../scripts/patch-post-speech.mjs';
import {patchPostOutcomes} from '../../../scripts/patch-post-outcomes.mjs';
import cases from '../fixtures/recovery-controls.json';
const base={type:'control_intent',intent:'protected_recovery_continuation',schema_version:1,source_message_id:'source',source_row_id:1,thread_id:1};
test('frozen recovery validator keeps malformed lookalikes visible and accepts legacy/typed controls',()=>{
 for(const c of cases){const blocks=c.fields===null?[]:[null,'bad',{...base,...c.fields}];expect(!!control(blocks)).toBe(c.hidden);}
 for(const bad of [null,undefined,{},'control_intent',42])expect(control(bad)).toBeNull();
 for(const key of ['source_row_id','thread_id','handoff_depth'])for(const n of [NaN,Infinity,-1,0,1.5,'1'])expect(control([{...base,[key]:n}])).toBeNull();
 const typed={...base,reason:'tools_required',compaction:'not_attempted',tools_required:true,retryable:false,recovery_attempts:0};
 for(const [reason,status] of [['post_compaction_tools_required','succeeded'],['compaction_failed','failed'],['recovery_budget_exhausted','not_attempted'],['unresolved_tool_execution','not_attempted'],['continuation_generation_exhausted','not_attempted'],['provider_retry_exhausted','not_attempted']])expect(control([{...typed,reason:reason,compaction:status}])).toBeTruthy();
 const timeout={...base,...cases.find(c=>c.id==='valid-timeout')!.fields};
 for(const [key,value] of [['primary_failure_detail','secret dump'],['primary_failure_category','error'],['primary_failure_elapsed_ms',2592000001],['primary_failure_tool_executions',1000001],['primary_failure_execution_tools','true']])expect(control([{...timeout,[key]:value}])).toBeNull();
 expect(control([{...base,intent:'bad'},base])).toEqual({label:'Recovery resumed with execution tools',sourceMessageId:'source',sourceRowId:1});
});
test('recovery adapter composes after speech/outcomes, preserves hook order and fails closed on drift',async()=>{
 const file=Bun.file('web/src/components/post.ts'),source=await file.text(),adapted=patch(patchPostOutcomes(patchPostSpeech(source)));
 expect(adapted).toContain('if (getProtectedRecoveryControlIntent(data.content_blocks)) return null;');
 expect(adapted.indexOf('if (getProtectedRecoveryControlIntent(data.content_blocks))')).toBeGreaterThan(adapted.lastIndexOf('useEffect('));
 expect(()=>patch(adapted)).toThrow();expect(()=>patch(source.replace('id=${`post-${post.id}`}','id="changed"'))).toThrow();expect(await file.text()).toBe(source);
});
