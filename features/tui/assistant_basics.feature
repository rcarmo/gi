Feature: TUI assistant basics
  The gi TUI should be usable like a compact Pi-style assistant from a terminal.

  Scenario: Boot, discover help and tools, and submit a prompt
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    And the screen should contain "Hints: /help"
    When I type "/help" and press Enter
    Then the screen should contain "gi TUI help"
    And the screen should contain "commands: /help, /tools"
    When I type "/compact" and press Enter
    Then the screen should contain "compact:"
    And the screen should contain "threshold_tokens"
    When I type "/model test-alt" and press Enter
    Then the screen should contain "Model: test-alt"
    When I type "/model test-model" and press Enter
    Then the screen should contain "Model: test-model"
    When I type "/thinking high" and press Enter
    Then the screen should contain "Thinking: high"
    When I type "/cancel" and press Enter
    Then the screen should contain "no running or queued turn to cancel"
    When I type "/tools rtk" and press Enter
    Then the screen should contain "tools:"
    And the screen should contain "rtk"
    When I type "hello from gherkin" and press Enter
    Then the database should contain a user message "hello from gherkin"
    And the database should contain an assistant message "Gi received: hello from gherkin"
