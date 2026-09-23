@gi-native @sessions @mobile
Feature: Gi swipe listener ownership during session activation
  This derived contract adds no frozen Classic or terminal credit.

  @gi-swipe-001
  Scenario: A fresh reverse swipe uses the committed selected session
    Given the native session catalogue contains adjacent sessions A and B
    And session A has a draft and an uploaded attachment
    When I swipe from A to B and reverse on the next animation frame
    Then the reverse contact returns to A without waiting for catalogue or timeline responses
    And repeating the sequence follows the same native carousel order
    And A's draft and attachment are preserved without submission

  @gi-swipe-002
  Scenario: A held target timeline cannot leak after reversing the selection
    Given a native timeline response for session B is held during activation
    When I type a draft in B and swipe back to A
    And the held response is delivered after the selection has changed
    Then A's timeline, draft and attachment remain selected
    And no B message is rendered into A's timeline
    And no additional turn is submitted to either session

  @gi-swipe-003
  Scenario: Streaming status passthrough stays inside the conversation surfaces
    Given native draft and thought streams render interactive links in their status panels
    When an eligible horizontal contact begins on either link
    Then the unchanged swipe resolver selects the adjacent native session
    And selected text and vertical contacts still prevent navigation
    And composer and Settings controls cannot initiate session navigation
    And excluded input targets retain their own touch and wheel handlers and defaults
    And the originating unsent draft returns unchanged
