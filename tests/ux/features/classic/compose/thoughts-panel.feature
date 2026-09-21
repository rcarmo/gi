@classic @source-reviewed
Feature: Classic thought and draft panel disclosure
  Source: runtime/web/src/components/status.ts and Classic agent-status CSS.

  @ux-thoughts-001
  Scenario: Render collapsed thought content with disclosure state
    Given thought content is available and its panel is collapsed
    When the status panel renders
    Then it exposes the collapsed data-expanded state
    And the collapsed height and overflow follow the panel's CSS

  @ux-thoughts-002
  Scenario: Continue updating content independently of disclosure
    Given a thought panel is collapsed
    When accepted thought updates change its content
    Then the rendered thought content updates without requiring the panel to be expanded

  @ux-thoughts-003
  Scenario: Toggle thought panel expansion
    Given thought content has a disclosure control
    When I activate that control
    Then the panel's expansion state toggles
    And the supplied panel-toggle callback is used when provided
    And otherwise the component manages its own expansion set

  @ux-thoughts-004
  Scenario: Collapse an expanded status panel with Escape
    Given an expanded status panel can be resolved
    And the Escape event is unmodified and does not target an editable field
    When I press Escape
    Then that panel is collapsed through the panel-toggle path

  @ux-thoughts-005
  Scenario: Preserve text when changing disclosure state
    Given a status panel contains streamed text
    When I expand and collapse the panel
    Then changing disclosure does not itself replace the stored thought or draft text
    And scroll behavior follows the component's content and expansion effects
