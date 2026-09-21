@classic @source-reviewed
Feature: Classic session selection
  Source: runtime/web/src/ui/compose-session-switcher.ts and chat-scoped refresh orchestration.

  @ux-session-001
  Scenario: Show the selected chat's timeline
    Given sessions "main" and "research" contain different posts
    When I choose "research" through the session switcher
    Then the client changes the current chat identifier
    And the selected-chat timeline refresh uses "research"
    And responses guarded by a superseded chat generation are ignored

  @ux-session-002
  Scenario: Group picker entries using the current session metadata
    Given the session catalogue has current, pinned, active, tree, other and archived entries
    When the Classic picker groups them
    Then it uses those native groups rather than a single flat alphabetical list
    And the current session remains distinguishable in the picker

  @ux-session-003
  Scenario: Filter session entries using their search metadata
    Given the Classic session picker is open
    When I enter a search query
    Then the session matchers filter its entries
    And keyboard navigation operates on that filtered result list

  @ux-session-004
  Scenario: Use archive and restore actions supplied for session entries
    Given a session entry offers its archive or restore action
    When I activate the action
    Then the corresponding branch action callback receives that session identifier
    And accepted changes are followed by a catalogue refresh
    And failures report the branch action error

  @ux-session-005
  Scenario: Keep touch swipe eligibility independent of picker grouping
    Given chat swipe navigation is enabled
    When candidates for a swipe are resolved
    Then archived candidates are excluded
    And the swipe carousel uses active-first and chat-identifier order
    And interactive target and text-selection exclusions still apply

  @ux-session-006
  Scenario: Dismiss the session picker without choosing an entry
    Given the Classic session picker is open
    When I press Escape
    Then the picker closes and clears its query/typeahead state
    And focus is restored to its trigger
    And no new session entry is activated
