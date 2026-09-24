// @ts-nocheck
// Display-only validator ported from Piclaw 70d33bc93 runtime/web/src/components/post.ts.
// This does not grant control authority or change native stored messages.
// Gi also validates primary-failure fields on legacy-shaped blocks: malformed
// typed recovery data must remain visible even when no handoff fields exist.
const PROTECTED_RECOVERY_CONTROL_INTENT = 'protected_recovery_continuation';
const PROTECTED_RECOVERY_CONTROL_LABEL = 'Recovery resumed with execution tools';
const PROTECTED_RECOVERY_REASONS = new Set([
    'post_compaction_tools_required',
    'tools_required',
    'compaction_failed',
    'recovery_budget_exhausted',
    'unresolved_tool_execution',
    'continuation_generation_exhausted',
    'provider_retry_exhausted',
]);
const PROTECTED_RECOVERY_TYPED_KEYS = ['reason', 'compaction', 'tools_required', 'retryable', 'recovery_attempts'];
const PRIMARY_FAILURE_KEYS = [
    'primary_failure_category',
    'primary_failure_detail',
    'primary_failure_elapsed_ms',
    'primary_failure_execution_tools',
    'primary_failure_had_partial_output',
    'primary_failure_had_tool_activity',
    'primary_failure_tool_executions',
];

function hasValidPrimaryFailureFields(block) {
    const count = PRIMARY_FAILURE_KEYS.filter(key => Object.prototype.hasOwnProperty.call(block, key)).length;
    if (count === 0) return true;
    return count === PRIMARY_FAILURE_KEYS.length
        && block.primary_failure_category === 'timeout'
        && typeof block.primary_failure_detail === 'string'
        && block.primary_failure_detail.trim().length <= 500
        && /^Timed out after \d+s\.$/.test(block.primary_failure_detail.trim())
        && Number.isInteger(block.primary_failure_elapsed_ms)
        && block.primary_failure_elapsed_ms >= 0
        && block.primary_failure_elapsed_ms <= 30 * 24 * 60 * 60 * 1000
        && typeof block.primary_failure_execution_tools === 'boolean'
        && typeof block.primary_failure_had_partial_output === 'boolean'
        && typeof block.primary_failure_had_tool_activity === 'boolean'
        && Number.isInteger(block.primary_failure_tool_executions)
        && block.primary_failure_tool_executions >= 0
        && block.primary_failure_tool_executions <= 1000000;
}

function hasValidProtectedRecoveryHandoffFields(block) {
    const hasTypedFields = PROTECTED_RECOVERY_TYPED_KEYS.some(key => Object.prototype.hasOwnProperty.call(block, key));
    if (!hasTypedFields) return hasValidPrimaryFailureFields(block);
    const valid = PROTECTED_RECOVERY_REASONS.has(block.reason)
        && ['not_attempted', 'succeeded', 'failed'].includes(block.compaction)
        && typeof block.tools_required === 'boolean'
        && typeof block.retryable === 'boolean'
        && Number.isInteger(block.recovery_attempts)
        && block.recovery_attempts >= 0;
    if (!valid || !hasValidPrimaryFailureFields(block)) return false;
    if (block.reason === 'post_compaction_tools_required') return block.compaction === 'succeeded' && block.tools_required;
    if (block.reason === 'compaction_failed') return block.compaction === 'failed';
    if (block.reason === 'tools_required' || block.reason === 'unresolved_tool_execution') return block.tools_required;
    return true;
}

export function getProtectedRecoveryControlIntent(contentBlocks) {
    const block = Array.isArray(contentBlocks)
        ? contentBlocks.find((candidate) => (
            candidate
            && typeof candidate === 'object'
            && candidate.type === 'control_intent'
            && candidate.intent === PROTECTED_RECOVERY_CONTROL_INTENT
            && candidate.schema_version === 1
            && typeof candidate.source_message_id === 'string'
            && candidate.source_message_id.trim().length > 0
            && Number.isInteger(candidate.source_row_id)
            && Number(candidate.source_row_id) > 0
            && Number.isInteger(candidate.thread_id)
            && Number(candidate.thread_id) > 0
            && Number.isInteger(candidate.handoff_depth ?? 1)
            && Number(candidate.handoff_depth ?? 1) > 0
            && hasValidProtectedRecoveryHandoffFields(candidate)
        ))
        : null;
    if (block) {
        return {
            label: PROTECTED_RECOVERY_CONTROL_LABEL,
            sourceMessageId: typeof block.source_message_id === 'string' ? block.source_message_id : '',
            sourceRowId: Number.isFinite(Number(block.source_row_id)) ? Number(block.source_row_id) : null,
        };
    }
    return null;
}

