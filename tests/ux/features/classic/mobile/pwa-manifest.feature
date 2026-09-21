@classic @pwa @source-reviewed
Feature: PWA manifest and home screen icon responses
  Source: runtime/src/channels/web/manifest.ts and http/dispatch-shell.ts.
  Home-screen installation itself is browser-controlled.

  @ux-pwa-001
  Scenario: Serve a manifest with declared application icons
    When I fetch the supported manifest endpoint
    Then the response describes the application name and icons
    And icon records contain source, sizes, type and purpose fields
    And the built-in icon configuration includes 192 and 512 pixel sizes

  @ux-pwa-002
  Scenario: Use configured agent-avatar URLs for manifest icons
    Given an agent avatar is configured for manifest use
    When the manifest is built
    Then avatar icon URLs request PNG format at the declared sizes
    And those URLs include an avatar version value

  @ux-pwa-003
  Scenario: Fall back to static icons without an avatar
    Given no agent avatar is configured for manifest use
    When the manifest is built
    Then static application icons are declared instead of a missing avatar URL

  @ux-pwa-004
  Scenario Outline: Request sized Apple touch icons
    Given an avatar handler can return a PNG
    When I request <path>
    Then the shell dispatcher requests PNG avatar output at <size>
    And it returns that response only when it is successful PNG content
    And otherwise it attempts the matching static fallback

    Examples:
      | path                           | size |
      | /apple-touch-icon-180x180.png    | 180  |
      | /apple-touch-icon-167x167.png    | 167  |
      | /apple-touch-icon-152x152.png    | 152  |
      | /apple-touch-icon.png           | 180  |

  @ux-pwa-005
  Scenario: Prefer PNG avatars for favicon compatibility
    When I request /favicon.ico
    Then the dispatcher first asks for the agent avatar in PNG format
    And it accepts only a successful PNG avatar response
    And otherwise it serves the static favicon fallback
    # The fallback is not specified as an unconditional decoded PNG guarantee.

  @ux-pwa-006
  Scenario: Vary avatar icon cache URLs with the avatar version
    Given the configured avatar version changes
    When the manifest is rebuilt
    Then its avatar icon URLs include the changed version value
