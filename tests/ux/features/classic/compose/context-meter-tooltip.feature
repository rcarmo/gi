@classic @source-reviewed
Feature: Classic context meter tooltip
  Source: ContextPie in runtime/web/src/components/compose-box.ts.

  @ux-context-001
  Scenario: Show supplied usage in the context tooltip
    Given context usage contains token, context-window and percentage values
    When the context pie renders
    Then its title, tooltip data and accessible label include formatted token values and rounded percentage
    And the pie fill uses percentage clamped between zero and one hundred

  @ux-context-002
  Scenario: Display missing token counts without inventing them
    Given context usage does not contain a token count
    When the context pie renders
    Then its usage label contains the supplied percentage without fabricated token values
    And unavailable formatted numeric values use the formatter's unknown marker

  @ux-context-003
  Scenario: Offer compaction only when a callback exists
    Given no active compaction label is supplied
    When the context pie renders with a compaction callback
    Then its tooltip includes Compact context
    And clicking it invokes that callback
    When it renders without that callback
    Then the button is disabled and its tooltip says Context usage

  @ux-context-004
  Scenario: Show the supplied compaction title and elapsed label
    Given a non-empty compaction label is supplied
    When the context pie renders
    Then it has the is-compacting class and shows that label beside the pie
    And its tooltip includes the supplied compaction title or the Smart compaction fallback
    And this tooltip branch replaces the ordinary Compact context suffix

  @ux-context-005
  Scenario: Apply the coded usage warning colours
    Given the context pie has a supplied usage percentage
    When it renders
    Then values above ninety use the red context token
    And values above seventy-five through ninety use the amber context token
    And values up to seventy-five use the green context token
