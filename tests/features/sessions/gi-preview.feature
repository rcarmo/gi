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

  @gi-preview-003
  Scenario: Wrapped paragraph disclosure follows rendered overflow without false line counts
    Given a native streamed paragraph has no internal newline
    When its collapsed body clips at the current width
    Then a generic Show more control appears without changing the source or line metadata
    And widening the collapsed body removes the control if the full content now fits
    And stream updates recheck clipping even when the collapsed box does not grow
    And expansion retains a collapse action through width changes
    And window resize and stream updates still work when ResizeObserver is absent
