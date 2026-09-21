@canonical @auth @classic
Feature: Classic core auth UX
  Classic login, invitation, and provider OOBE expose the currently shipped browser behavior.
  These scenarios describe observed code paths in runtime/web and existing runtime/test/web coverage.

  Background:
    Given Classic auth is served from the shipped web bundles
    And all actions use visible enabled controls only

  @ux-auth-001 @family @totp
  Scenario: Family-shared code sign-in requires and normalizes the account username
    Given the login policy enables TOTP in family-shared mode
    When the login page finishes loading
    Then the username field is visible and required
    And the code field and Verify button are enabled
    And the description says to enter the account username and authenticator code
    When I submit username " Alice " with code "123456"
    Then the verification request sends username "alice" and code "123456"
    And a network failure reports that the server could not be reached

  @ux-auth-002 @single-user @totp
  Scenario: Single-user code sign-in omits the username field and submits only the code
    Given the login policy enables TOTP in single-user mode
    When the login page finishes loading
    Then the username field is hidden and not required
    And the description says to use the six-digit code from the authenticator app
    When I submit code "123456"
    Then the verification request sends only code "123456"

  @ux-auth-003 @single-user @passkey
  Scenario: Single-user passkey-only mode hides the TOTP form
    Given the login policy enables passkey and disables TOTP in single-user mode
    When the login page finishes loading
    Then the login form is hidden
    And the passkey button is visible
    And the description says to use a passkey registered for this site

  @ux-auth-004 @login @failure
  Scenario: Failed sign-in policy loading never exposes stale credential controls
    Given the first sign-in options request fails
    When the login page finishes loading
    Then the login form is hidden
    And the passkey button is hidden
    And the page says sign-in options could not be loaded
    And a Retry control is visible
    When Retry reloads a valid single-user TOTP policy
    Then the code field becomes visible

  @ux-auth-005 @login @passkey @race
  Scenario: Explicit passkey sign-in supersedes ambient passkey work and code submission aborts passkey work
    Given the login policy enables both TOTP and passkey
    And conditional passkey mediation is available
    When the page starts an ambient passkey request
    And I trigger explicit passkey sign-in
    Then the ambient passkey request is aborted before the explicit request proceeds
    When I submit a TOTP code while passkey work is pending
    Then the pending passkey request is aborted before code verification is sent

  @ux-auth-006 @oobe
  Scenario: Classic OOBE shows provider-missing only for an unconfigured current instance
    Given model readiness is known for the current Classic instance
    And the instance has no available models and no configured current model
    And the provider-missing panel was not dismissed
    Then the OOBE panel kind is provider-missing
    And it shows Getting started copy
    And it offers Open settings and Dismiss actions

  @ux-auth-007 @oobe
  Scenario Outline: Classic OOBE stays hidden when the instance is already configured or still unresolved
    Given <state>
    Then the OOBE panel kind is hidden

    Examples:
      | state                                             |
      | available model options exist                     |
      | a current model hint already exists               |
      | the provider-missing panel was dismissed          |
      | model readiness has not finished loading          |
      | the pane is running in popout mode                |

  @ux-auth-008 @invitation @totp
  Scenario: TOTP invitation strips the link token from the URL, claims only on click, and erases setup secrets after confirmation
    Given I open a valid TOTP invitation link with a fragment token
    When the invitation page initializes
    Then the visible URL is reduced to /auth/invitation before any fetch
    And no claim request is sent automatically
    When I click Claim invitation
    Then the claim request body contains the token without exposing it in the request URL
    And the page shows the bound account, shared secret, and QR code
    When I confirm with code "123456"
    Then the confirmation request sends the token, enrolment token, and code
    And the shared secret and QR code are erased
    And the page remains on /auth/invitation with a Sign in link

  @ux-auth-009 @invitation @recovery
  Scenario: Recovery-only invitation completion hides the normal sign-in action
    Given a valid invitation completes with recovery_only true
    Then the status says recovery is complete
    And the Sign in link is hidden
    And all enrolment secrets are cleared

  @ux-auth-010 @invitation @failure
  Scenario: Invalid or rejected invitations never reveal enrolment or auto-retry a consumed claim
    Given the invitation page has no valid fragment token or the claim is rejected
    Then no enrolment form is revealed
    And no automatic claim retry is made
    And the page tells me to ask the administrator for a fresh link

  @ux-auth-011 @invitation @passkey
  Scenario: Passkey invitations are account-bound, proof-checked, and never sign the browser in automatically
    Given I open a valid passkey invitation link
    When the page initializes
    Then the visible URL is reduced to /auth/invitation before any fetch
    And no claim request is sent automatically
    When I click Begin passkey setup
    Then the account label appears only after a successful claim
    When I create the passkey successfully
    Then the browser rechecks the restricted invitation before confirmation
    And confirmation succeeds only for the claimed account binding
    And the page stays on /auth/invitation with no session cookies

  @ux-auth-012 @invitation @passkey @cancellation
  Scenario Outline: Leaving a passkey invitation flow discards the one-use ceremony
    Given a passkey invitation is waiting for native credential creation
    When I <action>
    Then no passkey confirmation request is sent
    And the passkey enrolment panel is hidden
    And the page tells me to request a fresh invitation

    Examples:
      | action                        |
      | cancel from the page control  |
      | blur the tab outside the native prompt |
      | leave the page                |

  @ux-auth-013 @family @privacy
  Scenario: Mask the family client when its page loses visibility
    Given the family client has loaded an authenticated identity
    When the page blurs or becomes hidden
    Then the client masks its content
    When focus or visibility returns
    Then the family resume path runs
    And busy results or memory operations recheck identity before they continue

  @ux-auth-014 @family @logout
  Scenario: Sign out of the family client
    Given the family client is loaded and not stopped
    When I activate Sign out
    Then the sign-out control is disabled while logout runs
    And notification cleanup is attempted before the logout request
    And successful logout invalidates client state and navigates to /login
    And failed logout reports an error and re-enables the control while the client remains active
