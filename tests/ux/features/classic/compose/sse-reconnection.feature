@classic @source-reviewed
Feature: Classic SSE reconnection and refresh
  Source: runtime/web/src/ui/app-connection-lifecycle.ts.

  @ux-reconnect-001
  Scenario: Clear transient agent displays while disconnected
    Given the client connection changes to a state other than connected
    When the connection lifecycle handler runs
    Then agent status, agent draft, agent plan and thought previews are cleared
    And pending-request and agent-run state are reset
    # Agent previews here are distinct from the user's composer draft.

  @ux-reconnect-002
  Scenario: Refresh authoritative chat state after reconnect
    Given the client has previously connected
    When its connection state becomes connected again
    Then the client refreshes agent status, follow-up queue and context usage
    And it refreshes the main timeline when no hashtag or search view is active

  @ux-reconnect-003
  Scenario: Avoid replacing an active search with main-timeline refresh
    Given a hashtag or search view is active
    When the connection becomes connected
    Then the general reconnect path does not refresh the main timeline over that view
    And status, queue and context refresh still run

  @ux-reconnect-004
  Scenario: Show version drift without automatically reloading
    Given the server advertises a different UI asset version
    When the client receives that version for the first time
    Then it shows the New UI available warning with a manual reload instruction
    And it does not automatically reload even when editors and composer are clean
    And repeated notices for the same version are suppressed by the version guard

  @ux-reconnect-005
  Scenario: Avoid duplicate initial refresh after recent chat activation
    Given the first connected event follows a recent chat activation
    When initial-connection cleanup finishes
    Then the recent-activation guard may skip redundant refresh calls
    # This does not specify a fixed reconnect duration or exactly-once event delivery.
