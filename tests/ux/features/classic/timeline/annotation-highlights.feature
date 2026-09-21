Feature: Timeline annotations and highlights
  As a reviewer working in the timeline
  I want inline image markup and persistent text annotations
  So that I can inspect visual and textual output without leaving the chat

  Rule: Inline image annotation is limited to iPad-class devices
    Background:
      Given I am authenticated and on the main chat
      And the timeline contains a message with an image

    @ux-timeline-001

    Scenario: iPad image tap opens the inline annotator
      Given annotation actions are enabled
      And I am on an iPad
      When I tap an image in the timeline
      Then the inline image annotator dialog should open
      And the toolbar should contain pen, highlighter, arrow, rectangle, text, crop, eraser, and undo controls

    @ux-timeline-002

    Scenario: Two-finger gestures pinch instead of drawing
      Given the inline image annotator is open
      When I place two fingers on the canvas and pinch
      Then the annotator should enter pinch mode
      And no drawing stroke should be committed from that gesture

    @ux-timeline-003

    Scenario: Applying crop reduces the working image and resets crop state
      Given the inline image annotator is open
      And I have drawn a crop region
      When I apply the crop
      Then the raster source and existing annotations should be reduced to the crop bounds
      And the crop selection should clear
      And the active tool should return to pen

    @ux-timeline-004

    Scenario: Done uploads a flattened PNG and queues a preview
      Given the inline image annotator is open
      And I have added annotations
      When I tap Done
      Then a flattened PNG should be uploaded
      And an annotation preview with Send and Discard actions should appear

    @ux-timeline-005

    Scenario: Cancel closes the annotator without queuing a preview
      Given the inline image annotator is open
      When I tap Cancel
      Then the annotator should close
      And no annotation preview should appear

    @ux-timeline-006

    Scenario: Non-iPad image activation opens the lightbox instead
      Given annotation actions are enabled
      And I am not on an iPad
      When I activate an image in the timeline
      Then the inline image annotator should not open
      And the image lightbox should open

    @ux-timeline-007

    Scenario: SVG sources are rasterized before PNG export
      Given the inline image annotator is open for an SVG image
      When I tap Done
      Then the source image should be rasterized into a canvas before export
      And the uploaded result should be a PNG

  Rule: Text annotations persist on posts
    Background:
      Given I am authenticated and on the main chat
      And a post with text content is visible

    @ux-timeline-008

    Scenario: Selecting text shows highlight colors
      When I select text in a post
      Then a highlight toolbar should appear
      And it should offer yellow, green, blue, pink, and orange colors

    @ux-timeline-009

    Scenario: Clicking a highlight color persists the saved selection snapshot
      Given I have selected text and the highlight toolbar is visible
      When the live browser selection clears before I click a highlight color
      And I click the yellow highlight color
      Then the original selected text should still be highlighted
      And the saved annotation should use the original text and textOffset

    @ux-timeline-010

    Scenario: Highlights persist via post annotations
      Given I have highlighted text in a post
      Then the annotation should be saved via PATCH /post/:id/annotations
      And the annotation should include type, text, textOffset, and color

    @ux-timeline-011

    Scenario: Desktop highlight toolbar stays near the selection
      Given I am using a fine-pointer desktop browser
      When I select text in a post
      Then the highlight toolbar should be positioned near the selected range
      And it should remain inside the viewport

    @ux-timeline-012

    Scenario: Coarse-pointer highlight toolbar docks away from the selection
      Given I am using iOS or another coarse-pointer device
      When I select text in a post
      Then the highlight toolbar should render in its docked placement
