@oracle-only @piclaw-3.2.4 @not-gi-parity
Feature: Observed basic Classic interactions in the shipped Piclaw 3.2.4 UI
  The oracle is the installed 3.2.4 release, not the older frozen 70d33bc
  Classic contract. These cases describe the shipped UI with an isolated
  backend fixture. They do not certify Gi, physical devices or authentication.

  Background:
    Given the shipped Classic UI assets with the server's default SVG sanitization flag
    And an isolated session "web:default" with no real chat or auth writes

  @oracle-quick-actions @conflicts-classic-007 @conflicts-shared-16
  Scenario: A slash Quick Action replaces an existing unsent composer draft
    Given the composer contains "ORACLE EXISTING DRAFT"
    When I open Quick Actions from noninteractive timeline content
    And I activate the "/model" slash command
    Then the composer contains "/model"
    And the composer receives focus
    And no prompt submission request is made
    But the former unsent text is not preserved in the composer

  @oracle-queue-return @conflicts-shared-28
  Scenario: Return to editor replaces the existing draft before removing the queued row
    Given the composer contains "NEWER UNSENT TEXT"
    And the queue contains one row with "QUEUED ORACLE TEXT"
    When I activate Return queued message to editor
    Then the composer contains "QUEUED ORACLE TEXT" and receives focus
    And the removal request uses that row ID and the selected chat identifier
    But the former unsent text is not preserved in the composer
    # Failure/retry ownership and attached media are not proven by this oracle probe.

  @oracle-svg @conflicts-classic-029 @not-inline-svg-dom
  Scenario: A safe fenced SVG is rendered as an isolated image by default
    Given an assistant post contains a fenced SVG with a title and safe vector geometry
    When the shipped Classic Markdown renderer processes it with sanitization enabled
    Then the post contains an image with a data:image/svg+xml source and accessible title
    And the original SVG source remains available for the code-copy path
    And raw model SVG geometry is not injected as privileged descendants of the post
    When a second fenced SVG contains a script, an event handler and an external image reference
    Then that unsafe fence remains visible as escaped source code without a preview image
    And it does not execute script or fetch its external reference
    # Malformed/oversized input, other unsafe categories and theme controls need
    # separate adversarial tests; this observation alone earns no Shared41 credit.
