Feature: Timeline rendering and post actions
  As a user
  I want rendered posts and post-level actions to match the implemented classic timeline behavior
  So that visual affordances stay trustworthy

  Background:
    Given I am authenticated and on the main chat

  @ux-timeline-023

  Scenario: Markdown tables render as full-width tables with automatic layout
    Given the agent has posted a message containing a markdown table
    Then the rendered table should use table display
    And the table should span the post width with automatic column layout

  @ux-timeline-024

  Scenario: Code blocks expose a copy button in the top-right corner
    Given the agent has posted a message with a code block
    Then the code block should render a copy button in its top-right corner
    When I click the code copy button
    Then the code block text should be copied to my clipboard

  @ux-timeline-025

  Scenario: Resource links and link previews open in a new tab
    Given the agent has posted a resource link or link preview with a remote URL
    When I activate that rendered link
    Then it should open in a new browser tab
    And it should use noopener noreferrer isolation

  @ux-timeline-026

  Scenario: Outcome chips render after the timestamp in post metadata
    Given the agent has completed a turn with an outcome marker
    Then the post metadata should show the timestamp first
    And the outcome chip should render after the timestamp on the same metadata row

  @ux-timeline-027

  Scenario: Read aloud appears only when browser speech support and speakable text both exist
    Given an agent post has speakable text
    When the browser supports speech synthesis
    Then the post should show a Read aloud action
    But without speech synthesis support the Read aloud action should not be shown

  @ux-timeline-028

  Scenario: Starting read aloud on another post transfers playback ownership
    Given one agent post is already being read aloud
    When I start Read aloud on a different agent post
    Then the earlier speech playback should be cancelled
    And the new post should become the active speaking post
