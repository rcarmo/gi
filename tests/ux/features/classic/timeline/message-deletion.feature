Feature: Message deletion from timeline
  As a user managing conversation history
  I want deletion prompts and cascade behavior to match the stored thread state
  So that I can remove messages without leaving orphaned replies

  Background:
    Given I am authenticated and on the main chat
    And the timeline contains messages

  Rule: Direct deletion applies when no replies are visible
    @ux-timeline-017
    Scenario: Delete a single message without visible replies
      Given a message with no visible thread replies exists
      When I click the delete button on that message
      Then the message should enter the removing state
      And the message should be removed from the DOM after the removal delay
      And refreshing the page should not show the deleted message

    @ux-timeline-018

    Scenario: Backend reply detection asks for a second confirmation before retrying cascade
      Given a message appears to have no replies in the current view
      And the backend rejects direct deletion with "Replies exist"
      When I confirm the follow-up cascade prompt
      Then deletion should retry with cascade enabled
      And the message and its replies should be removed

    @ux-timeline-019

    Scenario: Cancelling the backend follow-up prompt preserves the message
      Given a message appears to have no replies in the current view
      And the backend rejects direct deletion with "Replies exist"
      When I cancel the follow-up cascade prompt
      Then the message should remain visible

  Rule: Visible thread replies require explicit cascade confirmation
    @ux-timeline-020
    Scenario: Deleting a message with 3 visible replies asks for cascade confirmation
      Given a message that has 3 visible thread replies exists
      When I click the delete button on the parent message
      Then a confirmation prompt should ask "Delete this message and its 3 replies?"

    @ux-timeline-021

    Scenario: Confirming cascade deletes the parent and visible replies together
      Given a message that has visible thread replies exists
      When I click the delete button on the parent message
      And I confirm the cascade prompt
      Then the parent message should enter the removing state
      And its visible thread replies should enter the removing state
      And the parent and replies should be removed together

    @ux-timeline-022

    Scenario: Cancelling cascade preserves the parent and visible replies
      Given a message that has visible thread replies exists
      When I click the delete button on the parent message
      And I cancel the cascade prompt
      Then the parent message should remain visible
      And its visible thread replies should remain visible
