Feature: TUI keyboard behavior
  The gi TUI should support predictable keyboard-only operation in tmux.

  Scenario: Blur, focus, history, scroll, resize, and quit
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    When I press Escape
    And I type "ignored while blurred" and press Enter
    Then the database message count should be 0
    When I press Tab
    And I type "focus restored" and press Enter
    Then the database should contain a user message "focus restored"
    And the database should contain an assistant message "Gi received: focus restored"
    Then the screen should contain "F2/F3 history"
    When I press PageUp
    Then the screen should contain "focus restored"
    When I press End
    Then the screen should contain "Gi received: focus restored"
    When I resize the terminal to 100x22
    Then the tmux session should be alive
    And the screen should contain "Messages:"
    When I press Ctrl-D
    Then the tmux session should exit
