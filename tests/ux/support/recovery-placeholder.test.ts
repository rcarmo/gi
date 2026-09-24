import {test,expect} from 'bun:test';
import {isSilentRecoveryPlaceholder as silent} from '../../../web/src/gi-recovery-placeholder.ts';
import {patchPostRecoveryControl} from '../../../scripts/patch-post-recovery-control.mjs';
import cases from '../fixtures/recovery-placeholders.json';
test('empty informational recovery requires exact metadata, agent origin and no visible content',()=>{
 for(const c of cases)expect(silent({isAgent:c.role==='assistant',contentBlocks:c.blocks,hasRenderableContent:Boolean(c.content),hasVisibleExtras:Boolean(c.media||c.blocks.some(b=>['adaptive_card','adaptive_card_submission','resource'].includes(b.type)||b.annotations))})).toBe(c.hidden);
 const base={isAgent:true,contentBlocks:[{type:'turn_outcome_marker',kind:'recovery',severity:'info'}],hasRenderableContent:false,hasVisibleExtras:false};
 expect(silent(base)).toBe(true);expect(silent({...base,hasVisibleExtras:true})).toBe(false);expect(silent({...base,hasRenderableContent:true})).toBe(false);
 for(const blocks of [null,{},'recovery',[],[{type:'turn_outcome_marker',kind:'recovery',severity:true}],[{type:'turn_outcome_marker',kind:'Recovery',severity:'info'}]])expect(silent({...base,contentBlocks:blocks})).toBe(false);
});
test('placeholder guard uses the supplied renderability and keeps every visible extra after hooks',async()=>{
 const source=await Bun.file('web/src/components/post.ts').text(),out=patchPostRecoveryControl(source);
 expect(out.indexOf('if (isSilentRecoveryPlaceholder({')).toBeGreaterThan(out.lastIndexOf('useEffect('));
 expect(out).toContain('hasRenderableContent: shouldRenderContent');
 for(const key of ['mediaIds.length','cardBlocks.length','submissionBlocks.length','fileRefs.length','messageRefs.length','attachments.length','data.link_previews?.length','generatedWidgets.length','resources.length','resourceLinks.length','textAnnotations.length'])expect(out).toContain(key);
 expect(out).toContain('if (getProtectedRecoveryControlIntent(data.content_blocks)) return null;');
 expect(()=>patchPostRecoveryControl(source+'\nconst isSilentRecoveryPlaceholder = true;')).toThrow();
 expect(await Bun.file('web/src/components/post.ts').text()).toBe(source);
});
