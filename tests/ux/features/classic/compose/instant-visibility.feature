@classic @source-reviewed
Feature: Classic accepted message visibility
  Source: runtime/web/src/components/compose-box.ts and UI timeline refresh paths.
  Network latency is not a product-level one-second delivery guarantee.

  @ux-compose-007
  Scenario: Display an accepted text submission
    Given the composer contains a non-empty text draft
    When I submit it and the message API accepts it
    Then the post-response path is notified
    And the selected chat's timeline refresh can display the stored message

  @ux-compose-008
  Scenario: Serialize text and references into one submission
    Given the draft contains multiline text and file, folder and message references
    When I submit it
    Then the outgoing content contains the trimmed text and the corresponding reference blocks
    And selecting references alone is sufficient to create a non-empty submission

  @ux-compose-009
  Scenario: Preserve the association between uploaded files and media identifiers
    Given a submitted draft has multiple attachments
    When the upload batch succeeds
    Then each returned media identifier stays paired with its source filename
    And the message request includes those identifiers and attachment references

  @ux-compose-010
  Scenario: Do not erase newer typing after send completes
    Given a captured draft has already cleared from the composer
    And I have started another draft
    When the earlier background send completes
    Then completion does not clear the newly entered text as a second submission reset

  @ux-compose-011
  Scenario: Reconcile visible messages through timeline state
    Given a submitted message is returned by the server or a later timeline refresh
    When the timeline updates its post collection
    Then the current chat's message records drive visible posts
    And scrolling uses the current near-bottom and user-scroll policy
    # Always scrolling to bottom would discard the user's history-reading position.
