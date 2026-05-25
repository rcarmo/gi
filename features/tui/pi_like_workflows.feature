Feature: TUI Pi-like workflow affordances
  The gi TUI should expose the new Pi-like commands and editor hints in tmux-friendly text.

  Scenario: Discover commands, inspect models, and use shell shortcut help
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    When I type "/help" and press Enter
    Then the screen should contain "editor:"
    And the screen should contain "Alt+Enter queue"
    And the screen should contain "Ctrl+Z undo"
    When I type "/commands model" and press Enter
    Then the screen should contain "commands: palette"
    And the screen should contain "/model [name|index]"
    When I type "/model" and press Enter
    Then the screen should contain "model: current="
    And the screen should contain "enabled models:"
    When I type "!!printf local-ok" and press Enter
    Then the screen should contain "local$ printf local-ok"
    And the screen should contain "local-ok"

  Scenario: Inspect sessions and settings in a narrow terminal
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    When I resize the terminal to 60x18
    Then the tmux session should be alive
    When I type "/settings" and press Enter
    Then the screen should contain "settings: runtime"
    And the screen should contain "settings: model"
    And the screen should contain "settings: editor"
    When I type "/resume" and press Enter
    Then the screen should contain "resume: recent sessions"
    And the screen should contain "m="
    When I type "/tree" and press Enter
    Then the screen should contain "tree: sessions:"
