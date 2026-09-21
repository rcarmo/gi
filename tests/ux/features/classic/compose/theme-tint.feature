@classic @compose @theme
Feature: Classic theme and tint commands
  This audit scopes /theme and /tint to the shipped Classic web shell.
  It keeps only behavior backed by the current Classic source and existing tests.

  Background:
    Given I am authenticated and on the main chat in the Classic shell

  @ux-theme-001

  Scenario: /theme with no arguments shows available themes
    When I type "/theme" and press Enter in the Classic compose box
    Then the timeline should show a message containing "Available themes"

  @ux-theme-002

  Scenario: /theme ristretto applies dark theme visually
    When I type "/theme ristretto" and press Enter in the Classic compose box
    Then the page background color should change
    And the root data-theme should be "dark"
    And the root data-color-theme should be "ristretto"
    And the root data-tint should be empty
    And the root CSS variable --bg-primary should be set
    And the root CSS variable --accent-color should be set
    And localStorage piclaw_theme should be "ristretto"
    And the timeline should show "Theme set to"

  @ux-theme-003

  Scenario: /theme default restores from ristretto visually
    Given the Classic shell theme is set to "ristretto"
    When I type "/theme default" and press Enter in the Classic compose box
    Then the page background color should change back from the ristretto state
    And the root data-color-theme should be "default"
    And the root data-tint should be empty
    And the untinted default theme should clear the root CSS variables it had applied
    And localStorage piclaw_theme should be "default"

  @ux-theme-004

  Scenario: /theme dark returns error — not a valid theme name
    When I type "/theme dark" and press Enter in the Classic compose box
    Then the Classic shell theme should not change
    And the timeline should show "Unknown theme"

  @ux-theme-005

  Scenario: /theme survives page refresh
    Given the Classic shell theme is set to "ristretto"
    When I refresh the Classic page
    Then the root data-theme should still be "dark"
    And localStorage piclaw_theme should be "ristretto"

  @ux-theme-006

  Scenario: /tint hex changes accent and background on default theme
    Given the Classic shell theme is "default" with no tint
    When I type "/tint #e11d48" and press Enter in the Classic compose box
    Then the root data-color-theme should be "default"
    And the root data-tint should be "#e11d48"
    And the root CSS variable --bg-primary should be set to a tinted value
    And the root CSS variable --accent-color should be set
    And localStorage piclaw_tint should contain "e11d48"
    And the timeline should show "Tint set to"

  @ux-theme-007

  Scenario: /tint named color works on default theme
    Given the Classic shell theme is "default" with no tint
    When I type "/tint orange" and press Enter in the Classic compose box
    Then the root data-color-theme should be "default"
    And the root data-tint should be "orange"
    And the root CSS variables for the tinted default theme should be applied
    And localStorage piclaw_tint should be "orange"

  @ux-theme-008

  Scenario: Switching tints visibly changes accent color
    Given the Classic shell theme is "default" with tint "#e11d48"
    When I type "/tint #3b82f6" and press Enter in the Classic compose box
    Then the root --accent-color should differ from the previous tint
    And the root --bg-primary should differ from the previous tint

  @ux-theme-009

  Scenario: /tint off clears tint and restores vanilla default
    Given the Classic shell theme is "default" with tint "#3b82f6"
    When I type "/tint off" and press Enter in the Classic compose box
    Then the root data-color-theme should be "default"
    And the root data-tint should be empty
    And the untinted default theme should clear the root CSS variables it had applied
    And the timeline should show "Tint cleared"

  @ux-theme-010

  Scenario: /tint with no args shows usage
    When I type "/tint" and press Enter in the Classic compose box
    Then the timeline should show "Usage"

  @ux-theme-011

  Scenario: /tint invalid value returns error
    When I type "/tint $$notacolor" and press Enter in the Classic compose box
    Then the Classic shell tint should not change
    And the timeline should show "Invalid tint"

  @ux-theme-012

  Scenario: /tint survives page refresh
    Given the Classic shell theme is "default" with tint "#e11d48"
    When I refresh the Classic page
    Then the root data-tint should still be set
    And the root CSS variable --bg-primary should be reapplied from the stored tint

  @ux-theme-013

  Scenario: Tint on default, switch to ristretto, switch back
    Given the Classic shell theme is "default" with tint "#e11d48"
    When I type "/theme ristretto" and press Enter in the Classic compose box
    Then the root data-color-theme should be "ristretto"
    When I type "/theme default" and press Enter in the Classic compose box
    Then the root data-color-theme should be "default"
    And the root data-tint should be empty because /theme clears tint

  @ux-theme-014

  Scenario: /tint on ristretto switches to default+tint
    Given the Classic shell theme is set to "ristretto"
    When I type "/tint #3b82f6" and press Enter in the Classic compose box
    Then the root data-color-theme should be "default"
    And the root data-tint should be "#3b82f6"

  @ux-theme-015

  Scenario: Round-trip visual consistency
    Given the Classic shell theme is "default" with no tint
    And I note the current page background color
    When I apply tint "#e11d48"
    Then the background should differ from the noted color
    When I switch to "/theme ristretto"
    Then the background should differ again
    When I switch back to "/theme default"
    Then the background should match the original noted color
