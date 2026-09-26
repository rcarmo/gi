// Native events invalidate a snapshot; no unguarded payload can invent active
// compaction. Selection/connection generations stay in the host app.
export function createActivityRevision() {
 let revision=0;
 return {capture:()=>revision, invalidate:()=>++revision, accepts:(value:number)=>value===revision};
}
export function compactionNotice(activity:any, now=Date.now()) {
 const c=activity?.compaction;
 if (!c || c.turn_id!==activity.turn_id) return null;
 if (c.active && ['running','cancelling'].includes(activity.status)) return {
  type:'intent',intent_key:'compaction',title:activity.status==='cancelling'?'Cancelling compaction':'Compacting context',
  started_at:c.timestamp,turn_id:activity.turn_id,started_seq:c.seq,
  tokens_before:c.tokens_before,tokens_source:c.tokens_source,
 };
 const age=now-Date.parse(c.timestamp);
 if (!Number.isFinite(age)||age<0||age>10000) return null;
 if (c.event_type==='compaction.suppressed') return {type:'notice',title:'Compaction temporarily suppressed',detail:c.detail||'Before-compact hook suppressed this attempt',turn_id:activity.turn_id};
 if (c.event_type==='compaction.failed') return {type:'notice',title:'Compaction failed',detail:c.detail||'Context retained',turn_id:activity.turn_id};
 return null;
}
// Explain the disabled action without advertising stale capability. Native
// admission tokens and server-side context checks still own mutation.
export function compactionUnavailableReason({fresh, disconnected, pending, status, capability}: {
 fresh:boolean; disconnected:boolean; pending:boolean; status?:string; capability?:any;
}):string {
 if(disconnected)return 'Reconnect to check compaction availability';
 if(pending)return 'Compaction request pending';
 if(!fresh)return 'Refreshing compaction availability';
 if(status!=='idle')return 'Session has active or queued work';
 if(capability?.available)return '';
 return typeof capability?.reason==='string'&&capability.reason.trim()?capability.reason.trim():'Compaction is unavailable';
}
// Locally estimated history is separate from measured provider request usage.
export function compactionEstimateLabel(notice:any):string {
 const n=notice?.tokens_before;
 if(notice?.intent_key!=='compaction'||notice?.tokens_source!=='estimate'||typeof n!=='number'||!Number.isSafeInteger(n)||n<0)return '';
 return `estimated history: ${n} tokens`;
}
export function compactionElapsed(notice:any, now=Date.now()) {
 const elapsed=Math.max(0,Math.floor((now-Date.parse(notice?.started_at))/1000));
 return Number.isFinite(elapsed)?`${Math.floor(elapsed/60)}:${String(elapsed%60).padStart(2,'0')}`:'0:00';
}
