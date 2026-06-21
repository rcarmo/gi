Feature: TUI Pi-like workflow affordances
  The gi TUI should expose the new Pi-like commands and editor hints in tmux-friendly text.

  Scenario: Discover commands, inspect models, and use shell shortcut help
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "(no messages yet)"
    When I type "/help" and press Enter
    Then the screen should contain "help"
    And the screen should contain "/commands all commands"
    And the screen should contain "!!cmd run locally"
    When I type "/commands model" and press Enter
    Then the screen should contain "commands: palette"
    And the screen should contain "/model [name|index]"
    When I type "/model" and press Enter
    Then the screen should contain "model test-model · low · test"
    And the screen should contain "1 test-model"
    When I type "!!printf local-ok" and press Enter
    Then the screen should contain "$ printf local-ok"
    And the screen should contain "local-ok"

  Scenario: Inspect sessions and settings in a narrow terminal
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "(no messages yet)"
    When I resize the terminal to 60x18
    Then the tmux session should be alive
    When I type "/settings" and press Enter
    Then the screen should contain "settings: session"
    And the screen should contain "settings: discovery"
    And the screen should contain "settings: compaction"
