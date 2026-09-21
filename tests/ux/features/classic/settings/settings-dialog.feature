@classic @settings
Feature: Classic settings dialog
  This audit scopes the shipped Classic Settings dialog only.
  It keeps only source-backed expectations from runtime/web and the current tests.

  Background:
    Given I am authenticated and on the main chat in the Classic shell

  @ux-settings-dialog-001

  Scenario: Rapid shortcut presses open exactly one settings dialog
    When I trigger the current Classic Settings keyboard shortcut three times quickly
    Then exactly one settings dialog should be visible
    And no duplicate settings dialog should remain visible in the UI

  @ux-settings-dialog-002

  Scenario: Second settings open is instant
    When I open settings
    And I close settings
    And I open settings again on the cached path
    Then the settings dialog should appear within 1 second
    And the reopen path should reuse cached settings data instead of a fresh cold load

  @ux-settings-dialog-003

  Scenario: Settings shows loading shell then content
    When I open settings for the first time in the Classic session
    Then a loading shell should appear immediately instead of a blank frame
    And the General pane should become visible within 2 seconds
    And General should be the first resolved built-in section

  @ux-settings-dialog-004

  Scenario: User can type a number in stepper fields
    Given I am on a Classic settings pane with a numeric stepper field
    When I focus the field
    And I replace its contents with "128000"
    Then the field should display "128000"

  @ux-settings-dialog-005

  Scenario: Non-General panes load only on click
    When I open settings
    Then only the General pane content should be available without a lazy section import
    And unopened built-in panes should not be loaded yet
    When I click on the Models nav item
    Then the Models pane should load and render
