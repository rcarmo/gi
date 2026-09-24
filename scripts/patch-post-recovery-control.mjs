/** Port frozen extra012/013 display suppression without editing the supplied Post. */
export function patchPostRecoveryControl(source) {
 const importAnchor="import { ImageModal } from './image-modal.js';";
 const renderAnchor='    return html`\n        <div id=${`post-${post.id}`}';
 for(const anchor of [importAnchor,renderAnchor])if(source.split(anchor).length!==2)throw Error(`Post recovery control anchor changed: ${anchor}`);
 if(source.includes('getProtectedRecoveryControlIntent')||source.includes('isSilentRecoveryPlaceholder'))throw Error('Post recovery control already adapted');
 return source.replace(importAnchor,importAnchor+"\nimport { getProtectedRecoveryControlIntent } from '../gi-recovery-control.js';\nimport { isSilentRecoveryPlaceholder } from '../gi-recovery-placeholder.js';")
  .replace(renderAnchor,`    // Keep hook order stable; suppress only after all hooks have run.
    if (getProtectedRecoveryControlIntent(data.content_blocks)) return null;
    if (isSilentRecoveryPlaceholder({
        isAgent, contentBlocks: data.content_blocks,
        hasRenderableContent: shouldRenderContent,
        hasVisibleExtras: Boolean(mediaIds.length || cardBlocks.length || submissionBlocks.length
            || fileRefs.length || messageRefs.length || attachments.length || data.link_previews?.length
            || generatedWidgets.length || resources.length || resourceLinks.length || textAnnotations.length),
    })) return null;

`+renderAnchor);
}
