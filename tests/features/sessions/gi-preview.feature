@gi-native @sessions @streaming
Feature: Gi streamed preview ownership and bounds
  Frozen disclosure criteria are verified separately. This contract covers Gi host ownership.

  @gi-preview-001
  Scenario: Preview expansion belongs to one session and native turn
    Given the native provider is streaming thought and draft text for A
    When I expand a preview and switch to B
    Then A's later stream does not appear in B
    And B's composer draft remains isolated
    When I return to A or reload
    Then only newly received preview bytes appear and expansion starts collapsed
    When a new native turn begins
    Then its preview buffers and expansion do not inherit the earlier turn

  @gi-preview-002
  Scenario: Preview metadata and scrolling preserve the composer
    Given multi-line native draft and thought streams exceed eight lines
    Then the supplied status renderer receives text without a false zero-line override
    And it owns disclosure state and Escape handling
    When I expand a long preview or resize the viewport
    Then the status region stays within forty percent of the conversation height and scrolls
    And the full preview text and unsent composer draft remain accessible
    And no terminal panel or idle row is added
