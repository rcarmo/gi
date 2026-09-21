@classic @source-reviewed
Feature: Classic composer draft and queue behavior
  Source: runtime/web/src/components/compose-box.ts.

  @ux-compose-001
  Scenario: Clear captured content while allowing a new draft
    Given the composer contains text and references
    When I submit the draft
    Then the client captures text, references, media and chat identifier for that submission
    And it clears the displayed draft before awaiting the background send
    And further typing belongs to the new displayed draft

  @ux-compose-002
  Scenario: Restore a failed submission alongside newer text
    Given a submitted draft is being sent in the background
    And I have typed a different new draft
    When that submission fails and the restore path runs
    Then captured text is restored ahead of the newer text unless it is already present
    And captured references are merged with current references
    And the failure is reported without claiming delivery succeeded

  @ux-compose-003
  Scenario: Reject an entirely empty submission
    Given there is no non-whitespace text, media or file, folder or message reference
    When the submission handler runs
    Then it returns without beginning a message submission

  @ux-compose-004
  Scenario: Return a queued message replaces the current editor draft
    Given a queued follow-up contains text and serialised references
    When I return it to the editor
    Then the client restores its text and references and clears the editor media list
    And it schedules queued-item removal after updating the editor
    And the text area receives focus and its cursor moves to the restored text's end

  @ux-compose-005
  Scenario: Keep upload progress separate from sending state
    Given a draft includes files
    When the submission uploads those files
    Then the transfer indicator describes attachment upload progress
    And that indicator is cleared before the message request is sent
    And message submission has its own button state

  @ux-compose-006
  Scenario: Submit captures the destination chat
    Given a draft is submitted in session "main"
    When the asynchronous upload finishes
    Then the message request uses the chat identifier captured at submission
    And it does not derive its destination from the current picker selection at completion time
