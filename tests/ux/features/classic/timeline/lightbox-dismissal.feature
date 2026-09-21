Feature: Image lightbox dismissal
  As a user viewing an image in the lightbox
  I want the modal to close through the supported dismissal gestures
  So that I can return to the timeline quickly

  Background:
    Given I am authenticated and on the main chat
    And a message with an image attachment is visible in the timeline

  @ux-timeline-013

  Scenario: Escape key dismisses the lightbox
    Given the lightbox is open showing an image
    When I press Escape
    Then the lightbox should close
    And the timeline should be visible again

  @ux-timeline-014

  Scenario: Non-Escape keys do not dismiss the lightbox
    Given the lightbox is open showing an image
    When I press Space, Enter, a letter, or an arrow key
    Then the lightbox should remain open

  @ux-timeline-015

  Scenario: Clicking anywhere inside the modal dismisses the lightbox
    Given the lightbox is open showing an image
    When I click the backdrop or the image
    Then the lightbox should close

  @ux-timeline-016

  Scenario: Tapping the modal surface on a touch device dismisses the lightbox
    Given the lightbox is open showing an image on a touch device
    When I tap the modal surface
    Then the lightbox should close
