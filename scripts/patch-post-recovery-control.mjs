/** Port frozen extra012 display suppression without editing the supplied Post. */
export function patchPostRecoveryControl(source) {
 const importAnchor="import { ImageModal } from './image-modal.js';";
 const renderAnchor='    return html`\n        <div id=${`post-${post.id}`}';
 for(const anchor of [importAnchor,renderAnchor])if(source.split(anchor).length!==2)throw Error(`Post recovery control anchor changed: ${anchor}`);
 if(source.includes('getProtectedRecoveryControlIntent'))throw Error('Post recovery control already adapted');
 return source.replace(importAnchor,importAnchor+"\nimport { getProtectedRecoveryControlIntent } from '../gi-recovery-control.js';")
  .replace(renderAnchor,'    // Keep hook order stable; suppress only after all hooks have run.\n    if (getProtectedRecoveryControlIntent(data.content_blocks)) return null;\n\n'+renderAnchor);
}
