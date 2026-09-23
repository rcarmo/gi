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
