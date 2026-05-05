Feature: TUI active tool controls
  The gi TUI should expose active tool visibility and activation controls.

  Scenario: Activate and reset a focused tool set
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    When I type "/tools active" and press Enter
    Then the screen should contain "tools: active:"
    When I type "/tools activate read shell" and press Enter
    Then the screen should contain "Activated tools:"
    And the screen should contain "read"
    And the screen should contain "shell"
    When I type "/tools active" and press Enter
    Then the screen should contain "tools: active: tools, read, shell"
    When I type "/tools reset" and press Enter
    Then the screen should contain "Reset active tools"
