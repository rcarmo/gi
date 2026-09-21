@classic @settings @layering
Feature: Classic settings dialog layering
  This audit scopes Classic overlay, backdrop, and body-portal behavior only.
  It does not claim that Visual or other skins share the same stacking values.

  Background:
    Given I am authenticated and on the main chat in the Classic shell

  @ux-settings-layering-001

  Scenario: Settings backdrop covers workspace pane
    Given the Classic workspace explorer is open
    When I open the settings dialog
    Then a semi-transparent backdrop should cover the viewport above the workspace pane
    And the backdrop should block clicks from reaching the workspace

  @ux-settings-layering-002

  Scenario: Settings dialog is above all other elements
    Given the Classic workspace explorer is open
    When I open the settings dialog
    Then the settings portal or backdrop should create a fixed overlay stacking context
    And the settings dialog should render above workspace content at its center point
    And the settings dialog should be fully visible within the viewport

  @ux-settings-layering-003

  Scenario: Backdrop is partially opaque (not fully transparent or opaque)
    When I open the settings dialog in the Classic shell
    Then the backdrop background should be rgba(0, 0, 0, 0.5)
    And the content behind should be dimmed beneath the overlay

  @ux-settings-layering-004

  Scenario: Only settings dialog is interactive above the backdrop
    Given the Classic workspace explorer is open
    And I open the settings dialog
    When I try to click a workspace file
    Then the click should not reach the workspace
    When I dismiss settings
    Then the backdrop should no longer be visible
    And the workspace should become interactive again
