@classic @source-reviewed
Feature: Classic compaction and model controls
  Source: compose-box ContextPie, model-picker, app-agent-turn-events and model-state helpers.

  @ux-compaction-001
  Scenario: Render compaction using supplied status state
    Given a non-empty compaction elapsed label and title are supplied to the composer
    When the context pie renders
    Then its compacting style and elapsed label are visible
    And its accessible label includes the compaction title

  @ux-compaction-002
  Scenario: Reconcile compaction events with client status
    Given a compaction event is accepted by the current turn's event guards
    When the event updates agent state
    Then the composer renders the resulting compaction status
    And completion updates follow the accepted terminal event and usage refresh
    # No fixed SSE delivery deadline or immediate atomic refresh is promised.

  @ux-compaction-003
  Scenario: Request stop through the visible compaction control
    Given the composer renders a stop control for active work
    When I activate stop while compaction is active
    Then the client sends the supported cancellation request
    And the displayed outcome follows later status updates

  @ux-compaction-004
  Scenario: Use refreshed usage rather than assume compaction always shrinks context
    Given compaction completes
    When the context usage request succeeds
    Then the meter renders the returned usage values
    And no assertion requires the new percentage to be smaller than its previous value

  @ux-compaction-005
  Scenario: Display temporary compaction suppression
    Given the client receives a supported compaction-suppressed notification
    When that notification is handled
    Then the status notice identifies temporary suppression
    And available retry or failure detail is displayed

  @ux-compaction-006
  Scenario: Check model context compatibility before switching
    Given the picker has model metadata and current context usage
    When I select an entry whose context-fit state is blocked
    Then the model picker's selection handler does not switch to it
    And compatible entries still use the normal model mutation path

  @ux-compaction-007
  Scenario: Refresh model information after an accepted switch
    Given a selectable model exists in the configured catalogue
    When the server accepts selecting that model for the current chat
    Then displayed model state updates from the accepted result
    And context information follows the refreshed model and usage data

  @ux-compaction-008
  Scenario: Handle a model command using the configured provider catalogue
    Given the configured catalogue contains a usable model
    When I submit its supported model command while idle
    Then the command is resolved by the native model command handler
    And an unknown or unusable model produces a command error instead of a fabricated selection
