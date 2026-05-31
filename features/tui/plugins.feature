Feature: TUI plugin and hook visibility
  The gi TUI should expose loaded extensions and registered hooks for debugging.

  Scenario: Inspect plugin and hook state
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "m0/t0"
    When I type "/plugins" and press Enter
    Then the screen should contain "plugins: extensions:"
    And the screen should contain "plugins: hooks:"
