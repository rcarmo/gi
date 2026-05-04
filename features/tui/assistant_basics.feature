Feature: TUI assistant basics
  The gi TUI should be usable like a compact Pi-style assistant from a terminal.

  Scenario: Boot, discover help and tools, and submit a prompt
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    And the screen should contain "Input (click to focus"
    When I type "/help" and press Enter
    Then the screen should contain "gi TUI help"
    And the screen should contain "commands: /help, /tools"
    When I type "/tools rtk" and press Enter
    Then the screen should contain "tools:"
    And the screen should contain "rtk"
    When I type "hello from gherkin" and press Enter
    Then the database should contain a user message "hello from gherkin"
    And the database should contain an assistant message "Gi received: hello from gherkin"
