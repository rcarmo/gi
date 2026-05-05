Feature: TUI settings and approval visibility
  The gi TUI should make runtime settings and approval-gate state discoverable.

  Scenario: Inspect runtime settings and approval state
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    When I type "/settings" and press Enter
    Then the screen should contain "settings: runtime:"
    And the screen should contain "provider=test"
    And the screen should contain "compaction enabled"
    When I type "/approvals" and press Enter
    Then the screen should contain "approvals: no approval gates"
