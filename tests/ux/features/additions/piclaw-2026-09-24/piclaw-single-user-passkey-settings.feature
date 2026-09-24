@shared @implemented @browser-verified @single-user @passkeys @settings
Feature: Manage multiple single-user passkeys in Settings
  As the sole instance owner
  I want to add, recognise, rename and remove passkeys from Settings
  So that I can sign in from several authenticators without managing credentials in chat

  # Acceptance contract: Bun/Playwright mappings and manual-device limits are in
  # docs/reviews/single-user-passkey-settings.md. No generated Cucumber steps.
  # Run UI scenarios in Classic and Visual; the test matrix is in
  # docs/design/single-user-passkey-settings.md.
  # A passkey is a registered credential, not a count of physical devices.
  # Policies below are the conservative defaults used for the requested implementation.

  Background:
    Given a disposable instance uses single-user authentication
    And the instance has an enabled default owner account
    And the configured relying party ID is "piclaw.test"
    And the configured origin is "https://piclaw.test"

  Rule: The owner can recognise and manage each registered credential

    @ux-single-passkeys-001
    Scenario Outline: Open the same passkey controls in either Settings skin
      Given I am signed in as the owner
      And passkeys are enabled by the login policy
      And "Laptop" and "Backup key" are registered for "piclaw.test"
      When I open Settings then Authentication in <skin>
      Then the Passkeys section lists both registered credentials
      And each row shows its name, a distinguishing identifier, creation time and last-used time
      And a credential with no recorded use is labelled "Never used"
      And I can find Add passkey, Rename and Remove controls without a slash command
      And credential public keys, login cookies and enrolment tokens are not displayed

      Examples:
        | skin    |
        | Classic |
        | Visual  |

    @ux-single-passkeys-002
    Scenario: Add a first passkey from an authenticated TOTP session
      Given the login policy accepts TOTP and passkeys
      And no passkeys are registered
      And I have recently proved possession of the configured TOTP factor
      When I add a passkey named "Laptop" from Settings
      And I complete the browser's native passkey creation prompt
      Then the server verifies the registration before reporting success
      And Settings lists one passkey named "Laptop"
      And TOTP sign-in remains available
      And no enrolment token is put in a navigation URL or posted to chat

    @ux-single-passkeys-003
    Scenario: Add another passkey without a TOTP prerequisite
      Given the login policy is passkey-only
      And no TOTP factor is configured
      And "Laptop" is already registered for "piclaw.test"
      And I have recently proved possession of "Laptop"
      When I add a passkey named "Backup key" from Settings
      And I select a different authenticator in the browser's native prompt
      And the server verifies the new registration
      Then Settings lists both "Laptop" and "Backup key"
      And the existing "Laptop" credential is unchanged
      And I am not asked to configure TOTP or use a chat enrolment link

    @ux-single-passkeys-004
    Scenario Outline: Either passkey can independently sign in after a restart
      Given "Laptop" and "Backup key" are registered for "piclaw.test"
      And the instance restarts with its persistent store intact
      And a browser profile has no active login session
      When I sign in using only "<passkey>"
      Then I am authenticated as the default owner
      And both passkeys remain registered with their names intact
      And only the used credential's last-used time advances

      Examples:
        | passkey    |
        | Laptop     |
        | Backup key |

    @ux-single-passkeys-005 @manual-device
    Scenario: A synced credential is not displayed as several physical devices
      Given "Synced key" represents one credential available on two devices
      When each device successfully signs in with that credential
      Then Settings continues to show one row for "Synced key"
      And Settings does not claim to inventory the devices holding that credential

    @ux-single-passkeys-006
    Scenario: Rename an existing or previously unnamed credential
      Given two passkeys have no saved names
      And I have recently proved possession of an accepted factor
      When I rename one identified credential to "Tablet"
      And the server confirms the change
      Then only that credential is named "Tablet"
      And the name remains after I reload Settings
      And its credential ID, relying party ID, public key and sign counter are unchanged

    @ux-single-passkeys-007
    Scenario Outline: Validate names without changing authentication material
      Given I have recently proved possession of an accepted factor
      And I am editing the name of "Laptop"
      When I submit <name>
      Then Settings shows <result>
      And the credential remains usable for sign-in

      Examples:
        | name                                  | result                                      |
        | a blank or whitespace-only name       | a validation error with the old name intact |
        | a name longer than 80 Unicode characters | a validation error with the old name intact |
        | a name containing a control character | a validation error with the old name intact |
        | a name containing HTML markup         | the saved name as literal text              |

    @ux-single-passkeys-008
    Scenario: Confirm removal of one key without touching the other
      Given "Laptop" and "Backup key" are registered for "piclaw.test"
      And I have recently proved possession of an accepted factor
      When I choose Remove for "Laptop"
      Then a confirmation identifies "Laptop" and its distinguishing identifier
      And no removal has been sent yet
      When I confirm removal
      Then success is shown only after the server confirms removal
      And a fresh sign-in using "Laptop" is rejected
      And a fresh sign-in using "Backup key" succeeds
      And "Backup key" has not been renamed or replaced

    @ux-single-passkeys-009
    Scenario: Cancel a destructive confirmation
      Given "Laptop" and "Backup key" are registered
      When I choose Remove for "Laptop"
      And I cancel the confirmation
      Then no removal request is sent
      And both credentials remain registered
      And keyboard focus returns to the control that opened the confirmation

  Rule: Native prompts and failures do not leave misleading Settings state

    @ux-single-passkeys-010
    Scenario: Cancel native passkey creation and retry deliberately
      Given I have recently proved possession of an accepted factor
      And one passkey is registered
      When I start adding another passkey
      And I cancel the browser's native creation prompt
      Then Settings reports cancellation without claiming a server registration
      And the registered passkey list is unchanged
      And Add passkey is available for an explicit retry with a fresh ceremony
      And the application does not reopen the native prompt automatically

    @ux-single-passkeys-011 @native-focus-manual
    Scenario: The native prompt may temporarily take focus away from the page
      Given an authorised passkey creation ceremony is pending
      When the browser's native prompt takes focus away from Settings
      And I complete that prompt before the ceremony expires
      Then the registration can complete once
      And page blur alone has not cancelled the ceremony or opened another prompt

    @ux-single-passkeys-012
    Scenario: A duplicate credential never replaces the existing passkey
      Given "Laptop" is registered for "piclaw.test"
      When I start adding another passkey
      Then creation options exclude existing credentials for this owner and relying party
      When a duplicate credential is nevertheless submitted to the server
      Then the server rejects the duplicate without replacing the existing row
      And Settings does not report a second registered passkey

    @ux-single-passkeys-013
    Scenario Outline: Show failed reads and writes truthfully
      Given Settings previously loaded a passkey named "Laptop"
      When <operation> fails with a server or network error
      Then Settings shows an accessible error with an explicit retry action
      And it does not show a success message
      And <preserved_state>

      Examples:
        | operation                | preserved_state                                           |
        | refreshing the list      | the list is not represented as empty or freshly confirmed  |
        | saving a new name        | the last confirmed name remains visible                   |
        | sending a removal        | the row is not silently removed from the displayed list    |

    @ux-single-passkeys-014
    Scenario: Reconcile an uncertain registration result without blindly retrying
      Given the server has stored a newly created passkey
      But the registration response does not reach the browser
      When I return to the Passkeys section
      Then Settings refreshes the credential list from the server
      And the new credential appears once
      And the application does not automatically repeat registration with the consumed challenge

    @ux-single-passkeys-025
    Scenario: Explain an unregistered credential left on the authenticator
      Given the native authenticator has created a new local credential
      But the server rejects its registration finish request
      When Settings reports the failed enrolment
      Then it says the credential was not registered on the server
      And it explains that a local credential may remain in the authenticator or password manager
      And it does not claim to have removed that local credential
      And existing registered credentials are unchanged

    @ux-single-passkeys-015
    Scenario Outline: Explain unavailable passkey creation
      Given <condition>
      When I open Settings then Authentication
      Then Add passkey is unavailable with <explanation>
      And no native credential-creation call is attempted
      And no login policy or stored credential is changed

      Examples:
        | condition                                      | explanation                              |
        | the page is an insecure non-localhost origin    | passkey creation requires a secure origin |
        | the browser has no WebAuthn credential API      | this browser cannot create passkeys       |
        | the login policy is TOTP-only                  | passkeys are disabled by the login policy |
        | single-user authentication is not configured   | authentication must be configured first  |

    @ux-single-passkeys-016
    Scenario Outline: Settings remains operable at narrow widths and with a keyboard
      Given I am signed in as the owner in <skin>
      And the viewport is 390 CSS pixels wide
      And a passkey has an 80-character name
      When I use the Passkeys section with a keyboard or touch input
      Then names, dates, errors and actions remain readable without page-level horizontal scrolling
      And every input has a label and every action has an accessible name
      And I can reach Rename, Remove and cancellation controls without hover
      And progress and error messages are announced without stealing focus

      Examples:
        | skin    |
        | Classic |
        | Visual  |

  Rule: Registration and management are bound to the authenticated single user

    @ux-single-passkeys-017 @security
    Scenario Outline: Reject requests outside the management authority
      Given <request_context>
      When a caller directly requests <operation>
      Then the server refuses the operation
      And no credential or login policy is changed
      And the response reveals no other account's passkey inventory

      Examples:
        | request_context                                      | operation                            |
        | no valid owner login session                         | list passkeys                        |
        | only an internal automation token                    | register a passkey                   |
        | an expired or revoked owner login session            | rename a passkey                     |
        | an authenticated request from an unapproved origin   | remove a passkey                     |
        | a single-user management request selecting another account | list that account's passkeys   |
        | the instance is in family-shared mode                | use the single-user management route |

    @ux-single-passkeys-018 @security
    Scenario Outline: Fail closed when registration proof or its binding is invalid
      Given an authenticated owner has started a registration ceremony
      When the finish request contains <invalid_condition>
      Then no credential is added or replaced
      And existing passkeys continue to work
      And any displayed failure contains no credential secrets or challenge tokens

      Examples:
        | invalid_condition                          |
        | an expired challenge                       |
        | an already consumed challenge              |
        | a challenge issued to another login session |
        | a mismatched origin                        |
        | a mismatched relying party ID              |
        | an invalid attestation or credential proof |
        | a session revoked while the native prompt was open |

  Rule: Security changes require proof within the preceding five minutes
    # Policy: reuse the account flow's five-minute freshness model.
    # Listing is allowed for a valid owner session; writes require fresh proof.

    @ux-single-passkeys-019 @security
    Scenario Outline: Refresh authentication before changing a passkey
      Given I have a valid owner session with authentication older than five minutes
      When I attempt to <operation>
      Then I must prove possession of a factor accepted by the current login policy
      And no write is authorised until that proof succeeds for this session
      And opening Settings or refreshing the page does not renew authentication freshness
      When I cancel re-authentication
      Then no credential change occurs

      Examples:
        | operation       |
        | add a passkey   |
        | rename a passkey |
        | remove a passkey |

    @ux-single-passkeys-026 @security
    Scenario: Fresh proof in one browser does not authorise another browser
      Given two browser sessions belong to the same owner
      And neither session has recent authentication
      When I successfully re-authenticate in the first session
      Then the first session can request authorised passkey changes
      But add, rename and remove requests from the second session still require its own fresh proof
      And a direct write request from the second session is refused without changing credentials

  Rule: Removal must preserve a usable future sign-in method
    # Policy: usability is determined by current server policy and current RP ID.
    # Active sessions, internal tokens and credentials for another RP are not factors.

    @ux-single-passkeys-020 @security
    Scenario Outline: Decide last-key removal from effective login methods
      Given I have recently proved possession of an accepted factor
      And "Laptop" is the selected registered passkey for "piclaw.test"
      And the login policy is "<policy>"
      And <remaining_method>
      When I request removal of "Laptop"
      Then the server <decision> removal
      And Settings explains <explanation>

      Examples:
        | policy       | remaining_method                                         | decision | explanation                                 |
        | passkey-only | another passkey is registered for piclaw.test             | allows   | the selected key was removed                |
        | passkey-only | a TOTP secret exists but there is no other passkey         | refuses  | no other sign-in method is accepted          |
        | passkey-only | only an active browser session remains                    | refuses  | a session is not a future sign-in method     |
        | passkey-only | the only other passkey is registered for old.piclaw.test   | refuses  | the other passkey cannot sign in here        |
        | either       | a TOTP factor is configured and accepted         | allows   | TOTP remains available for sign-in           |
        | either       | no other accepted factor is configured                    | refuses  | add another sign-in method before removing this key |
        | either       | a pending TOTP enrolment exists but has not been verified | refuses  | unverified TOTP setup is not a sign-in method |

    @ux-single-passkeys-021 @security @concurrency
    Scenario: Two concurrent removals cannot delete both remaining usable keys
      Given "Laptop" and "Backup key" are the only accepted sign-in methods
      And two recently authenticated browser sessions display both keys
      When one session requests removal of "Laptop" as the other requests removal of "Backup key"
      Then exactly one removal succeeds
      And the other is refused by the server's write-time lockout check
      And both lists can refresh to show the one remaining key

    @ux-single-passkeys-022 @security
    Scenario: Recheck policy and recovery methods at commit time
      Given a removal confirmation was opened while TOTP was an accepted fallback
      And the effective policy changes to passkey-only before removal is committed
      When I confirm removal of the last passkey usable for "piclaw.test"
      Then the server refuses removal
      And Settings refreshes the reason instead of trusting the earlier enabled button

    @ux-single-passkeys-023 @security @legacy
    Scenario: Slash commands cannot bypass the last-factor guard
      Given "Laptop" is the only sign-in method accepted by the current policy
      When an authenticated owner submits a legacy passkey delete command for "Laptop"
      Then the command refuses removal and directs me to authenticated Settings
      And the command does not report success

  Rule: Removing a passkey does not silently revoke existing login sessions

    @ux-single-passkeys-024
    Scenario: Explain the difference between removing a passkey and signing out a device
      Given I have recently proved possession of an accepted factor
      And another usable sign-in method will remain
      When I open the confirmation to remove "Laptop"
      Then the confirmation says removal blocks future sign-ins with that credential
      And it says existing login sessions are not signed out by this action
      When removal succeeds
      Then an already authenticated session remains subject to its normal expiry and explicit logout
