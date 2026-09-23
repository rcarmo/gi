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
    Then General is selected and contains read-only instance identity and startup defaults
    And Models is the only other enabled section in this slice
    And General explains that startup settings are loaded from files and require restart
    And no credential, environment, budget, recording or add-on write controls are present
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

  @proposal @gi-settings-next-001
  Scenario: Edit instance identity with explicit validated persistence
    Given a dedicated native settings write API with revision checks and atomic file preservation exists
    When I save instance identity
    Then success reports the applied snapshot and any restart requirement
    And failure retains the draft without claiming live application

  @proposal @gi-settings-next-002
  Scenario: Separate browser appearance from session and instance policies
    Given a Gi appearance storage contract has been chosen
    When I edit a supported browser preference
    Then its browser-local scope is explicit and no global or TUI policy is changed

  @proposal @gi-settings-next-003
  Scenario: Expose provider and compaction controls only with native write contracts
    Given Gi supports a validated provider credential or compaction policy operation
    When its settings section is enabled
    Then it exposes only implemented controls with secret-safe failure and persistence tests
