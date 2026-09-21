Feature: Terminal pane — standalone mode
  As a user
  I want to open a terminal pane without garbled display
  And manage it via tab controls and keyboard shortcuts
  So that I can run commands alongside the chat

  Background:
    Given I am authenticated and on the main chat

  @ux-terminal-001

  Scenario: Open terminal standalone without garbled output
    When I open a terminal pane
    Then a terminal surface should be visible
    And the terminal should render a canvas or text layer
    And the terminal background should match the active theme

  @ux-terminal-002

  Scenario: Execute ls -al in terminal
    Given a terminal pane is open
    When I type "ls -al" and press Enter in the terminal
    Then the terminal should display file listing output
    And the output should contain recognizable shell text

  @ux-terminal-003

  Scenario: Terminal opens clean without IME active
    Given a terminal pane is freshly opened
    When I type plain ASCII "echo test123" in the terminal
    Then the terminal should echo "echo test123" as terminal text

  @ux-terminal-004

  Scenario: Close terminal via tab close button (click)
    Given a terminal pane is open as a tab
    When I click the close button on the terminal tab
    Then the terminal tab should disappear
    And no terminal content should remain visible

  @ux-terminal-005

  Scenario: Close terminal via tab close button (tap)
    Given a terminal pane is open as a tab on a touch device
    When I tap the close button on the terminal tab
    Then the terminal tab should disappear

  @ux-terminal-006

  Scenario: Pop out terminal to new window (desktop)
    Given a terminal pane is open
    When I use the terminal pop-out control
    Then the original shell should expose a reattach affordance for the terminal

  @ux-terminal-007

  Scenario: Terminal theme matches UI theme
    Given a terminal pane is open
    When I change the UI theme
    Then the terminal foreground and background colors should update
    And the terminal should remain interactive

  Rule: Terminal dock — beneath editor

    Background:
      Given a file is open in the editor

    @ux-terminal-008

    Scenario: Toggle terminal dock via keyboard shortcut
      When I press Ctrl+Backtick
      Then the terminal dock should appear below the editor
      And the dock should have a visible splitter handle
      When I press Ctrl+Backtick again
      Then the terminal dock should hide

    @ux-terminal-009

    Scenario: Toggle terminal dock via tab strip button
      When I click the terminal dock toggle button in the tab strip
      Then the terminal dock should appear below the editor
      When I click the toggle button again
      Then the terminal dock should hide

    @ux-terminal-010

    Scenario: Dock splitter resizes terminal height
      Given the terminal dock is visible below the editor
      When I drag the dock splitter upward
      Then the terminal dock should remain visible
      And the dock height should increase
      When I drag the dock splitter downward
      Then the terminal dock should remain visible
      And the dock height should decrease

    @ux-terminal-011

    Scenario: Terminal dock is interactive alongside editor
      Given the terminal dock is visible below the editor
      When I type "echo hello" in the terminal dock
      Then the terminal should show terminal text output
      When I focus the editor and type "test"
      Then the editor should receive the input

    @ux-terminal-012

    Scenario: Dock hidden in zen mode
      Given the terminal dock is visible below the editor
      When I enter zen mode
      Then the terminal dock should be hidden

  Rule: Terminal zen mode

    Background:
      Given a file is open in the editor

    @ux-terminal-013

    Scenario: Zen mode hides all chrome except the terminal/editor
      When I enter zen mode
      Then the workspace sidebar should be hidden
      And the chat container should be hidden
      And the editor or terminal pane should remain visible

    @ux-terminal-014

    Scenario: Zen mode has a hover-discoverable exit control
      Given I am in zen mode
      When I hover at the top edge of the viewport
      Then the zen toggle control should become visible
      And the control should be clickable

    @ux-terminal-015

    Scenario: Clicking zen exit indicator reverts to normal layout
      Given I am in zen mode
      When I click the zen toggle control
      Then zen mode should deactivate
      And the workspace sidebar should reappear
      And the chat container should reappear

    @ux-terminal-016

    Scenario: Escape key exits zen mode
      Given I am in zen mode
      When I press Escape
      Then zen mode should deactivate
      And the normal layout should restore

    @ux-terminal-017

    Scenario: Hover-reveal tab strip in zen mode
      Given I am in zen mode
      Then the tab strip should be hidden by default
      When I hover near the top edge of the viewport
      Then the tab strip should become visible
