Feature: Editor pane
  As a user
  I want the code editor to be stable and responsive
  So that I can edit files without fighting the UI

  Background:
    Given I am authenticated and on the main chat
    And the workspace explorer is visible

  Rule: Editor tabs stay stable while switching and closing

    @ux-editor-001

    Scenario: Switching files does not cause visible flicker
      Given I have two files open in editor tabs
      When I switch from one editor tab to the other
      Then the editor pane should remain visible
      And the editor should not show a loading placeholder

    @ux-editor-002

    Scenario: Closing an unsaved tab shows confirmation
      Given I have a dirty editor tab
      When I attempt to close the tab
      Then the browser should show a confirmation dialog
      And dismissing the dialog should keep the tab open

    @ux-editor-003

    Scenario: Clicking a tab activates it immediately
      Given I have two files open in editor tabs
      When I press the primary mouse button on an inactive tab
      Then that tab should become active before mouse-up
      And the editor content should switch to that file

  Rule: Markdown preview and zen mode stay stable

    @ux-editor-004

    Scenario: Markdown preview is stable during splitter resize
      Given I have a markdown file open with preview enabled
      When I drag the preview splitter
      Then the preview pane should remain visible
      And the preview pane should keep rendered content
      And the preview height should persist after release

    @ux-editor-005

    Scenario: Zen mode keeps editor content visible while other shell panes are hidden
      Given I have a file open in the editor
      When I enter zen mode
      Then the workspace sidebar should be hidden
      And the chat container should be hidden
      And the editor pane should remain visible
