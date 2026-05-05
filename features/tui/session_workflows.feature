Feature: TUI session and agent workflows
  The gi TUI should support Pi-like session, fork, switch, and peer-send flows.

  Scenario: List agents, fork, switch, and send to a peer agent
    Given a fresh gi TUI workspace
    When I start the gi TUI in tmux
    Then the screen should contain "Session:"
    When I type "/agents" and press Enter
    Then the screen should contain "sys: agents:"
    And the screen should contain "@agent"
    When I type "/where" and press Enter
    Then the screen should contain "Session:"
    And the screen should contain "Agent: @agent"
    When I type "/fork @agent1" and press Enter
    Then the screen should contain "switched to @agent1"
    And the screen should contain "Agent: @agent1"
    When I type "/agents" and press Enter
    Then the screen should contain "@agent1"
    When I type "/switch @agent" and press Enter
    Then the screen should contain "switched to @agent"
    And the screen should contain "Agent: @agent"
    When I type "/send @agent1 hello peer" and press Enter
    Then the screen should contain "delivered to @agent1"
