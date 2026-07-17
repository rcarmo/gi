Feature: TUI settings and approval visibility
  The gi TUI should make runtime settings and approval-gate state discoverable.

  Scenario: Inspect runtime settings and approval state
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "(no messages yet)"
    When I type "/settings" and press Enter
    Then the screen should contain "model: test-model"
    And the screen should contain "settings: editor"
    And the screen should contain "scrollback_limit:"
    And the screen should contain "settings: compaction"
    When I type "/approvals" and press Enter
    Then the screen should contain "approvals: no approval gates are configured in gi yet"
