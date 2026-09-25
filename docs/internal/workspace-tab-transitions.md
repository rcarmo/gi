# Read-only workspace tab transitions

Gi can return from a file preview to the conversation without closing its tabs.
The tab strip also has roving keyboard focus, Arrow/Home/End activation and
keyboard context-menu access. These are read-only host adaptations; document
editing, dirty-buffer saves, popouts and docking are still unsupported.

## Conversation and preview

Opening a file shows its preview and collapses the workspace drawer on narrow
layouts. **Return to conversation** hides the preview column and splitter, keeps
the active pane mounted and focuses the composer if another control or Settings
has not taken focus. **Show read-only tabs** restores the same preview and focuses
the active tab. Hiding and showing does not fetch the file again or change tab
order, pins, chat selection, draft text, attachment bytes or file references.
The controls exist only while tabs are open; no idle row is added.

Hidden tabs do not handle global Ctrl+Tab/Ctrl+W. Hiding also clears any open tab
context menu, so it cannot reappear on restore. A pending file read can finish
while hidden, but cannot make the pane visible or change focus. Preview lifecycle
cleanup still fences reads from closed/replaced panes. Closing the final tab uses
the existing guarded composer-focus path. Pins protect bulk closes, not explicit
individual closes.

## Keyboard and touch

Only the active tab is in the tab strip's normal focus order. ArrowLeft/ArrowRight
wrap through tabs; Home/End select the first/last. Activation causes one native
preview read. Close buttons retain their own Enter/Space handling. Shift+F10 or
the ContextMenu key opens the focused tab's menu and focuses its first action.
Dismissal returns to that tab, or the surviving active tab after close, unless
focus moved elsewhere. Modal Settings and IME/default-prevented events retain
keyboard ownership.

The read-only build adapter now allows the touch pointer default on Close while
stopping propagation. WebKit otherwise suppressed the compatibility click and
left background tabs open. Mouse close still prevents activation; click handling
continues to own removal. Coarse-pointer controls have44px minimum touch targets.
The supplied tab-strip source and tab store are unchanged.

## Verification

`make test-ux-workspace-tabs` runs the workspace-tab suite and saves a distinct
`workspace-tabs-results.json`. The required browser CI job runs it after startup,
picker geometry and slash ownership. It is one bounded suite, not the full
browser acceptance matrix.

The expanded suite has72 executions across Chromium/WebKit at three configured
sizes. Six touch executions explicitly use390px contexts; those repeat mobile
coverage rather than adding tablet/desktop touch acceptance. Cases cover:

- Roving tab focus and one native read per Arrow/Home/End activation.
- Context-menu focus, surviving-tab close and clearing menus when hidden.
- Return/reopen without another read, with exact saved drafts/media preserved.
- Pending hidden reads completing while Settings owns focus and newer chat text.
- Trusted touch opening, background close without activation/read, conversation
  return and reopening retained tabs.
- Existing late-read, MRU, pin, bulk-close, code-font and Settings isolation tests.

The initial12 keyboard/return cases failed before implementation. The first full
run also found three WebKit background-touch-close failures; all passed after the
pointer change. Focused review found hidden menu state surviving hide, which was
fixed and re-reviewed without another blocker.

Local verification passed72tab,54preview/shell,42startup,24geometry,72slash and113
functional cases (five existing skips), Go/vet/hooks and113helper tests with3,020
assertions. The first functional run had one unrelated session-typeahead
active-class failure after focus moved correctly; an unchanged rerun passed.
That intermittent failure is retained in evidence and is not considered fixed.

No new frozen scenario mapping is assigned. Reference Piclaw editor transitions,
full visual equivalence and physical mobile-keyboard behaviour remain outside
this read-only adaptation. Terminal behaviour is unchanged; these browser
controls do not add terminal chrome.
