# Full UX re-audit instructions

User request: audit ALL Gherkin scenarios for correct coverage of basic UX interactions, then cascade down to tests and code again. Do not trust historical mapped/pass counts.

Clean baseline repo: /workspace/tmp/gi-ux-reaudit-source (real /srv/piclaw-dev/workspace/tmp/gi-ux-reaudit-source), HEAD5dc6bf7. No repo changes. Uncommitted Shared40 in another worktree is PAUSED and excluded. Live8090 never mutate. No fresh tests requested in this audit stage. Source review is not execution or physical/device/visual acceptance.

Input: <group>-input.json in this directory, one object per scenario definition, including background, every step, tags, outline rows and expanded count. Audit EVERY input key; do not collapse or omit scenarios. Read original feature if any ambiguity. Trace tags/names/clauses to actual test files and code (search source narrow, no dist/vendor). Cite exact paths+line ranges. Treat a tag, helper, mock or screenshot alone as weaker than a behavioral assertion.

For each scenario assess:
- Is the requirement complete and testable for basic human interaction? Trigger (mouse/touch/keyboard), focus/caret, observable outcome, draft/data preservation, session ownership, busy/disabled/error/retry/cancel, navigation/reload/persistence, network/duplicate/reordered behavior, accessibility where relevant.
- Do expected steps conflict with other frozen contracts or actual product constraints? Do not silently choose a policy.
- Is an outline's EVERY example actually covered? Default assertion or test title substring may hide no-op/synthetic evidence.
- Does test exercise a native production path, mocked boundary, source-regex, API-only, screenshot-only, artificial DOM, or unsupported behavior? What exact assertions missing? Tests can be false positives even if current implementation happens to be good.
- Code: relevant handlers/ownership guards; code-backed defect vs missing evidence vs feature absent. No invented defect from omission of test alone.
- No claims of runtime bug without source/test evidence. Mark uncertain/open honestly.

Output JSON <group>-audit.json: {group,baseline,scenarios:[...] ,cross_cutting:[...]}.
Every scenarios row MUST include:
{key, verdict: 'adequate'|'requirement-gap'|'test-gap'|'code-gap'|'conflict'|'unverified', priority:'P0'|'P1'|'P2'|'P3', requirement_assessment, test_evidence:[{path,lines,kind,proves,missing}], code_evidence:[{path,lines,observation}], finding, correction:{scenario,test,code}, confidence:'high'|'medium'|'low'}.
Use 'none identified in bounded source review' rather than assert perfect coverage. Scenario rows need specific findings not copied generic boilerplate. Cite actual evidence, do not fabricate line numbers. For unmapped feature use no test evidence + relevant code absence/scaffold with clear distinction.

cross_cutting rows: {finding,priority,affected_keys,evidence,correction}. Keep concrete, bounded and actionable; list optional scope exclusions separately. Enough detail to turn findings into implementation tasks. Avoid generic aesthetic recommendations.

Final tool reply: row count, top3 high-confidence findings and outputpath, audit limitations. Do NOT edit production/tests/specs. Write only audit output. Do not commit or deploy. Prefer completeness for all keys over exhaustive testing; you have250s. Use bash/read efficiently; no delegates of your own.
