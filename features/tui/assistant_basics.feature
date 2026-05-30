Feature: TUI assistant basics
  The gi TUI should be usable like a compact Pi-style assistant from a terminal.

  Scenario: Invoking gi without arguments starts the TUI
    Given a fresh gi TUI workspace
    When I start gi without arguments in tmux
    Then the screen should contain "(no messages yet)"
    And the screen should contain "m0/t0"

  Scenario: Boot, discover help and tools, and submit a prompt
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "(no messages yet)"
    And the screen should contain "m0/t0"
    When I type "/help" and press Enter
    Then the screen should contain "help"
    And the screen should contain "/commands all commands"
    When I type "/compact" and press Enter
    Then the screen should contain "compact:"
    And the screen should contain "threshold_tokens"
    When I type "/model test-alt" and press Enter
    Then the screen should contain "model: test-alt"
    When I type "/model test-model" and press Enter
    Then the screen should contain "model: test-model"
    When I type "/thinking high" and press Enter
    Then the screen should contain "thinking set to high"
    When I type "/cancel" and press Enter
    Then the screen should contain "no running or queued turn to cancel"
    When I type "/tools rtk" and press Enter
    Then the screen should contain "tools:"
    And the screen should contain "rtk"
    When I type "hello from gherkin" and press Enter
    Then the database should contain an assistant message "Gi received: hello from gherkin"
    And the screen should contain "you: hello from gherkin"
