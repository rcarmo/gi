@gi @settings
Feature: Gi settings backed by native capabilities
  Gi reuses the Piclaw dialog structure, not its unsupported service contracts.
  These Gi cases are separate from frozen Classic and shared parity counts.

  @gi-settings-001
  Scenario: Open one modal and return to the unchanged draft
    Given a native Gi session with an unsent draft and an open workspace
    When I open Settings from the timeline menu or press Control+Comma repeatedly
    Then exactly one body-portal dialog named Gi Settings is visible above the workspace
    And its half-opaque backdrop and focus trap prevent interaction with the workspace
    And the dialog fits phone, tablet and desktop viewports
    When I dismiss with Escape, Close or a backdrop click
    Then focus returns to the originating control or composer
    And the selected session, workspace and unsent draft are unchanged

  @gi-settings-002
  Scenario: Show only implemented settings sections and explicit scopes
    When I open Gi Settings
    Then General is selected and separates active instance values from saved display-name fields
    And Models, Appearance, Compaction and Providers are the other enabled sections
    And General explains that startup settings are loaded from files and require restart
    And General edits only assistant and user display names
    And credential controls are confined to supported Providers entries
    And no environment, budget, recording or add-on write controls are present
    And Models identifies its destination session and does not edit global defaults

  @gi-settings-003
  Scenario: Load General without a blank frame and recover from failure
    Given the runtime config response is held
    When I open Gi Settings
    Then the shell immediately shows a loading status
    When the request fails
    Then an error and Retry action replace the loading status
    When I retry successfully
    Then General shows the native instance snapshot
    And reopening can show the cached snapshot while refreshing it

  @gi-settings-004
  Scenario: Fetch model choices only when Models is opened
    Given Gi Settings is on General
    When I select Models
    Then a loading state is shown until the current session model catalogue arrives
    And I can filter the native model labels
    And at most 50 matching choices are rendered with a refine-filter hint when truncated
    And unknown context capacity is labelled unknown
    And the current thinking level is read-only

  @gi-settings-005
  Scenario: Apply a model only after server confirmation to its captured session
    Given two native sessions and an unsent draft with an attachment
    When I choose another model in Gi Settings
    Then no mutation occurs before I press Apply model
    When I apply and the accepted response is held
    Then Apply is disabled and the previous confirmed model remains visible
    When the response arrives
    Then the accepted model and context state appear in Settings and the composer
    And the selection survives reload while the other session and global defaults remain unchanged
    And draft text and attachment survive without submitting a turn

  @gi-settings-006
  Scenario: Keep failures and incompatible models truthful
    Given Models has loaded native catalogue and context data
    When the native server rejects an unavailable model
    Then Settings shows the error and re-enables Apply without changing the confirmed model
    And the current draft is untouched
    And a model whose known capacity is below measured usage cannot be applied
    And unknown capacity is not fabricated or treated as a known fit failure

  @gi-settings-007
  Scenario: Ignore responses from a closed or superseded settings view
    Given a model read or write response for session A is held
    When I close Settings, switch to session B and reopen Models
    And the response for A arrives
    Then B's model, error feedback and draft are unchanged
    And any accepted write to A remains persisted only in A

  @gi-settings-008
  Scenario: Keep existing authentication and API contracts
    Given native model and runtime config routes require authentication
    When an unauthenticated caller requests those settings or attempts a model write
    Then the existing authentication boundary denies the request
    And the settings UX introduces no new secret disclosure or unauthenticated mutation route

  @gi-settings-012
  Scenario: Save display names explicitly and require restart for activation
    Given General shows active instance names and saved names with a revision
    When I edit assistant and user display names
    Then nothing is written until I activate Save names
    When the native server confirms the save
    Then only those names are changed in the workspace Piclaw config
    And unrelated keys, avatars and runtime configuration are preserved
    And General shows a restart-required notice while active names remain unchanged
    And reload retains the saved names without automatically restarting or changing sessions
    When a fresh instance loads that config
    Then the saved names become its active names

  @gi-settings-013
  Scenario: Protect identity saves against conflict and invalid or unsafe configuration
    Given General holds a revision of the saved config
    When another writer changes that config before I save
    Then my save is rejected as a conflict and my draft remains visible
    And Reload saved names explicitly discards my draft and obtains the new revision
    When a name is empty, exceeds 128 characters or contains control characters
    Then no write occurs and a validation error appears
    And corrupt, oversized, nonregular or symlinked configuration is not silently replaced
    And write failure leaves the original file intact and reports failure without success
    And unauthenticated writes and cross-origin browser writes are denied before parsing

  @gi-settings-009
  Scenario: Explicitly save browser-local appearance without changing server settings
    Given Appearance is open and labelled as applying to this browser across sessions
    When I select a supported preset or enter a default-theme hex tint
    Then only the unsaved fields change until I press Save appearance
    When I save successfully
    Then a single versioned browser preference is persisted before the theme is rendered
    And the page palette and theme metadata reflect the selection
    And changing session or reloading retains the preference
    And no server mutation, legacy per-chat theme write or TUI change occurs

  @gi-settings-010
  Scenario: Reject invalid appearance input and report storage failure honestly
    Given Appearance is open
    When I enter a tint other than an empty value or a three- or six-digit hex colour
    Then Save reports a validation error without changing storage or the rendered theme
    When storage denies a valid save
    Then Settings retains the fields and reports failure without changing the rendered theme
    And retry can persist and apply the same fields

  @gi-settings-011
  Scenario: Reset local appearance and synchronise other tabs
    Given a saved browser appearance and another tab on the same origin
    When I save another preset or use Reset appearance
    Then the other tab renders the new preference without a reload
    And Reset persists the default theme with no tint and follows the system colour mode
    And a dirty Appearance form reports an external change without discarding its fields
    And malformed saved preference data is ignored without crashing startup

  @gi-settings-014
  Scenario: Inspect effective compaction policy and session capability without changing it
    When I open Compaction in Gi Settings
    Then the effective engine startup policy is displayed separately from saved policy fields with restart guidance
    And the destination session and authoritative manual capability or disabled reason are visible
    And no provider-model, watchdog or backoff controls are advertised
    And a failed read disables actions and offers Refresh without using stale capability

  @gi-settings-015
  Scenario: Compact with an explicit snapshot and report authoritative completion
    Given the destination session is idle with eligible native context
    When I press Compact now
    Then the captured context token is posted once and controls wait for admission
    And admission is labelled as accepted rather than completed
    When authoritative progress and completion arrive
    Then the matching compaction state is displayed
    And the durable context boundary is persisted without deleting timeline or submitting my draft
    And draft text, attachments and the selected model remain unchanged

  @gi-settings-016
  Scenario: Stop the displayed compaction turn and recover from stale capability
    Given a matching active compaction is displayed in Settings
    When I press Stop turn
    Then cancellation targets that exact turn identifier and waits for authoritative state
    And a cancellation failure reports an error without clearing draft state
    When context or active work changes after a manual capability read
    Then native admission rejects the stale request without a duplicate or queued compaction
    And Settings refreshes capability and retains explicit failure feedback

  @gi-settings-017
  Scenario: Fence Compaction reads and actions by dialog and session
    Given a Compaction read or action response for session A is held
    When I close Settings, switch to B and reopen Compaction
    And the response for A arrives
    Then B's capability, progress, error feedback and draft are unchanged
    And accepted work on A remains bound to A

  @gi-settings-018
  Scenario: Explicitly save automatic compaction policy for the next restart
    Given Compaction shows active engine policy and a saved file revision
    When I change automatic enablement or token budgets
    Then the engine policy is unchanged and no write occurs until Save policy
    When the server confirms Save policy
    Then the saved snapshot and restart-required notice are shown
    And only supported compaction fields change in Pi settings while unknown fields and strategy remain intact
    And model and TUI preference writers share the atomic settings-file lock
    When a fresh process starts
    Then it loads the saved policy without an automatic restart or policy change in the old process

  @gi-settings-019
  Scenario: Reject conflicting or invalid automatic compaction policy saves
    Given another writer changed Pi settings after my policy read
    When I save my policy
    Then the stale revision is rejected and the unsaved fields are preserved
    And Reload saved policy explicitly discards those fields and reads the new revision
    When values are non-integers or budgets are out of bounds or inconsistent
    Then validation reports failure and the file remains unchanged
    And file failures preserve the original settings with no success notice
    And reads and writes require existing authentication and writes reject cross-origin requests
    And closing a pending save cannot report success in a reopened dialog

  @gi-settings-020
  Scenario: Inspect provider credential metadata without secrets or network probing
    When I open Providers
    Then OpenAI and Anthropic show explicit API-key setup capability
    And existing OAuth or other credential entries are read-only with native setup guidance
    And stored means present locally rather than verified by the provider
    And no key, access token, refresh token or credential-file content is returned
    And credential writes require existing authentication plus HTTPS or a loopback connection

  @gi-settings-021
  Scenario: Save a supported provider key explicitly and use it in native inference
    Given an isolated native provider fixture and no key for its allowlisted provider
    When I enter its API key in the password field
    Then no mutation occurs before Save key
    When the server confirms Save key
    Then the input is cleared and only metadata confirms it was stored, not verified
    And only that provider entry changes atomically with private file permissions
    And a subsequent native inference request consumes the saved key
    And session model and draft are unchanged by the save
    And the secret is never written to browser storage or returned in API responses

  @gi-settings-022
  Scenario: Reject unsafe, stale and failed credential changes without overwriting other entries
    Given Providers holds a revision of the credential store
    When another writer updates it before my save or remove
    Then the stale revision is rejected and no entry is overwritten
    And errors contain no secrets and permit metadata refresh and explicit retry
    And empty, oversized, whitespace-containing keys and unsupported providers are rejected
    And an existing OAuth or token entry cannot be overwritten by an API-key control
    And corrupt, nonregular or symlinked credential files are not replaced
    And unauthenticated or cross-origin writes are rejected before parsing

  @gi-settings-023
  Scenario: Remove only the confirmed API-key entry and ignore closed-view responses
    Given a supported API-key entry is stored
    When I cancel Remove key confirmation
    Then no delete request is sent
    When I confirm removal with its current revision
    Then only that entry is removed and its configured state refreshes
    And unrelated credentials and session state remain intact
    When a save or read response arrives after closing Providers
    Then it cannot update a reopened view or restore a cleared secret field

  @gi-settings-024
  Scenario: Load non-General pane code on demand without caching native state
    Given General is available in the app module
    When I select an unopened pane
    Then its content-hashed module is requested and a pane loading status appears
    And unopened pane modules are not requested
    When I return to a visited pane
    Then the browser reuses its module but native data follows its existing refresh contract
    And late imports cannot replace a newer section or reopen a closed dialog
    And a failed module load shows a bounded error without a blank page or automatic reload

  @proposal @gi-settings-next-004
  Scenario: Add browser OAuth only with provider-specific native login and refresh lifecycle
    Given a provider has a supported browser login contract with bounded lifetime and cancellation
    When its OAuth controls are enabled
    Then callback ownership, refresh persistence and sign-out are tested separately
    And unsupported custom authentication fields remain absent
