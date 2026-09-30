SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

# ── Tool commands ───────────────────────────────────────────────────────

GO ?= go
BUN ?= bun
PLAYWRIGHT ?= scripts/run-playwright.sh

# ── Runtime defaults ────────────────────────────────────────────────────

PORT ?= 8090
BIND ?= 0.0.0.0
LISTEN ?=
MODEL ?= github-copilot/gpt-5-mini
WORKSPACE ?= /workspace

# ── Local paths ─────────────────────────────────────────────────────────

RUN_DIR ?= .gi-run
BIN_DIR ?= bin
BIN ?= $(BIN_DIR)/gi
DB ?= $(RUN_DIR)/gi.db
LOG ?= $(RUN_DIR)/gi.log
PID ?= $(RUN_DIR)/gi.pid

TEST_PORT ?= 19090
TEST_DIR ?= .gi-test
TEST_DB ?= $(TEST_DIR)/gi.db
TEST_LOG ?= $(TEST_DIR)/gi.log
TEST_PID ?= $(TEST_DIR)/gi.pid
TEST_WORKSPACE ?= $(TEST_DIR)/workspace
TEST_RESULTS ?= test-results
TUI_TEST_DIR ?= .gi-tui-test

# ── Derived arguments and data ──────────────────────────────────────────

SERVER_LISTEN_ARGS = $(if $(LISTEN),-listen $(LISTEN),-bind $(BIND) -port $(PORT))
SERVER_RUN_ARGS = $(SERVER_LISTEN_ARGS) -model $(MODEL) -db $(DB) -workspace $(WORKSPACE)
SERVER_DAEMON_ARGS = $(SERVER_LISTEN_ARGS) -model $(MODEL) -db $(abspath $(DB)) -workspace $(WORKSPACE) -log-file $(abspath $(LOG)) -pid-file $(abspath $(PID))
SERVER_STATUS_ADDR = $(if $(LISTEN),$(LISTEN),$(BIND):$(PORT))
TEST_SERVER_ARGS = -bind 127.0.0.1 -port $(TEST_PORT) -model test-model -db $(abspath $(TEST_DB)) -workspace $(abspath $(TEST_WORKSPACE)) -log-file $(abspath $(TEST_LOG)) -pid-file $(abspath $(TEST_PID))
TEST_PICLAW_CONFIG_JSON = {"assistant":{"assistantName":"Gi Test"},"user":{"userName":"Test User"}}
TEST_ENABLED_MODELS ?= ["test-model"]
TEST_PI_SETTINGS_JSON = {"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":$(TEST_ENABLED_MODELS),"agents":{"list":[{"id":"web","name":"Gi Test","default":true,"model":"test-model"}]}}

# ── Helper macros ───────────────────────────────────────────────────────

define require-command
	@command -v $(1) >/dev/null || { echo "$(2)"; exit 1; }
endef

# ── Public targets ──────────────────────────────────────────────────────

.PHONY: \
	help bootstrap deps \
	build-web build \
	run start stop restart status logs \
	test vet bun-checks check \
	test-instance-start test-instance-stop test-ux test-ux-parity test-ux-index-config ux-parity-inventory test-tui-smoke test-tui-gherkin test-tui-input-history test-tui-sessions test-tui-source-copy test-tui-markdown \
	clean

# ── Help and bootstrap ──────────────────────────────────────────────────

help:
	@printf "%s\n" \
		"gi Make targets" \
		"" \
		"Bootstrap" \
		"  make bootstrap        Install Go/Bun deps, install Playwright Chromium, build gi" \
		"  make deps             Download Go modules and install Bun packages" \
		"" \
		"Build" \
		"  make build-web        Bundle web assets with Bun" \
		"  make build            Build web assets and the gi binary" \
		"" \
		"Run" \
		"  make run              Build and run gi in the foreground" \
		"  make start            Build and start gi detached" \
		"  make stop             Stop the detached gi process" \
		"  make restart          Restart the detached gi process" \
		"  make status           Show detached process status" \
		"  make logs             Tail the detached process log" \
		"" \
		"Checks and tests" \
		"  make test             Run Go unit tests" \
		"  make vet              Run go vet" \
		"  make bun-checks       Run Bun/TS hook checks" \
		"  make check            Run the standard verification suite" \
		"  make test-ux          Run Playwright tests against an isolated instance" \
		"  make test-tui-smoke   Run the tmux-based TUI smoke harness" \
		"  make test-tui-gherkin Run the TUI gherkin harness" \
		"  make test-tui-markdown Run Markdown/ANSI Gherkin scenarios in tmux" \
		"  make test-tui-sessions Verify draft isolation and compact picker sizes" \
		"  make test-ux-parity   Run mapped frozen Piclaw scenarios in Chromium/WebKit" \
		"  make ux-parity-inventory Verify all frozen feature hashes and list coverage" \
		"" \
		"Isolated test instance" \
		"  make test-instance-start  Start the isolated test server on 127.0.0.1:$(TEST_PORT)" \
		"  make test-instance-stop   Stop and clean up the isolated test server" \
		"" \
		"Cleanup" \
		"  make clean            Remove build, run, and test artifacts" \
		"" \
		"Common overrides" \
		"  PORT=$(PORT) BIND=$(BIND) MODEL=$(MODEL) WORKSPACE=$(WORKSPACE) LISTEN=$(LISTEN)"

bootstrap: deps
	$(PLAYWRIGHT) install chromium
	$(MAKE) --no-print-directory build
	@echo "Bootstrap complete. Run 'make start' or 'make run'."

deps:
	$(call require-command,$(GO),Go is required but not installed or not on PATH)
	$(call require-command,$(BUN),Bun is required but not installed or not on PATH)
	$(GO) mod download
	$(BUN) install --frozen-lockfile

# ── Build ───────────────────────────────────────────────────────────────

build-web:
	$(BUN) run build:web

build: build-web
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) ./cmd/gi

# ── Run lifecycle ───────────────────────────────────────────────────────

run: build
	mkdir -p $(RUN_DIR)
	$(BIN) $(SERVER_RUN_ARGS)

start: build
	mkdir -p $(RUN_DIR)
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		echo "Gi already running with PID $$(cat $(PID))"; \
		exit 0; \
	fi
	$(abspath $(BIN)) $(SERVER_DAEMON_ARGS) >/dev/null 2>&1 </dev/null &
	@sleep 2
	@$(MAKE) --no-print-directory status BIND=$(BIND) PORT=$(PORT) LISTEN=$(LISTEN)

stop:
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		kill $$(cat $(PID)) && echo "Stopped Gi ($$(cat $(PID)))"; \
		rm -f $(PID); \
	else \
		echo "Gi is not running"; \
	fi

restart: stop start

status:
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		addr='$(SERVER_STATUS_ADDR)'; \
		port='$(PORT)'; \
		if [ -n '$(LISTEN)' ]; then \
			case "$$addr" in \
				*:* ) port="$${addr##*:}"; port="$${port#\[}"; port="$${port#\]}" ;; \
			esac; \
		fi; \
		echo "Gi running on $$addr with PID $$(cat $(PID))"; \
		if command -v ss >/dev/null 2>&1 && [ -n "$$port" ]; then \
			ss -ltnp | grep -F ":$$port" || true; \
		else \
			echo "Listener probe unavailable (missing 'ss' or unresolved port)"; \
		fi; \
	else \
		echo "Gi is not running"; \
		exit 1; \
	fi

logs:
	@mkdir -p $(RUN_DIR)
	@touch $(LOG)
	tail -f $(LOG)

# ── Checks and tests ────────────────────────────────────────────────────

.PHONY: test-session-thinking test-ux-thinking test-shared-capability-evidence
test-shared-capability-evidence: test-ux-thinking
	$(BUN) test tests/ux/support/parity-report.test.ts tests/ux/support/passkey-criteria.test.ts
	$(BUN) scripts/ux-parity-report.mjs test-results/ux-parity/results.json

.PHONY: test-message-retrieval
test-message-retrieval:
	$(GO) test -race ./internal/store ./internal/tools ./internal/turn -run 'TestMessageRows|TestMessageRetrieval' -count=3

.PHONY: test-ux-message-retrieval
test-ux-message-retrieval: build-web test-message-retrieval
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_MESSAGE_RETRIEVAL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test -c playwright.ux.config.mjs tests/ux/message-retrieval.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-shared-message-evidence
test-shared-message-evidence: test-ux-message-retrieval
	$(BUN) test tests/ux/support/parity-report.test.ts tests/ux/support/passkey-criteria.test.ts
	$(BUN) scripts/ux-parity-report.mjs test-results/ux-parity/message-retrieval-results.json

test-session-thinking:
	$(GO) test -race -count=3 ./internal/store ./internal/inference ./internal/turn ./internal/web -run SessionThinking

test-ux-thinking: build-web test-session-thinking
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_THINKING=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config playwright.ux.config.mjs tests/ux/session-thinking.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-web-queue-hold
test-web-queue-hold:
	$(GO) test -race -count=3 ./internal/store ./internal/turn ./internal/web -run WebQueueHold

.PHONY: test-web-send-receipts
test-web-send-receipts:
	$(GO) test -race -count=3 ./internal/store ./internal/web -run WebSendReceipt

.PHONY: test-web-http-helpers test-web-basic-send test-web-basic-controls

test-web-basic-controls:
	$(MAKE) test-ux-steer UX_LOCAL_ENV="GI_UX_BASIC_HTTP=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN))" UX_LOCAL_SPEC=tests/ux/basic-http-control.spec.mjs
test-web-http-helpers:
	$(BUN) test tests/ux/support/random-id.test.ts tests/ux/support/drafts.test.ts tests/ux/support/send-recovery.test.ts

test-web-basic-send: test-instance-start
	@GI_TEST_URL=http://127.0.0.1:$(TEST_PORT) $(PLAYWRIGHT) test tests/functional/18-basic-http-send.spec.ts tests/functional/19-http-delivery.spec.ts --reporter=line --output=$(TEST_RESULTS)/basic-http-send; \
	status=$$?; $(MAKE) test-instance-stop; exit $$status

test:
	$(GO) test ./...

.PHONY: test-shell-runtime check-cross-build test-active-steering

test-active-steering:
	$(GO) test -race -count=50 ./internal/turn -run '^TestSubmitPromptSteersSecondPromptToActiveTurn$$'

test-shell-runtime:
	$(GO) test -race -count=3 ./internal/tools -run 'RunShellPrompt|KillShellProcess'
	$(GO) test -race -count=3 ./internal/turn -run 'CancelTurn|Cancelled|CancelQueuedTurn'

# Keep portable compilation reproducible outside GitHub Actions too.
check-cross-build:
	@set -eu; out=$$(mktemp -d); trap 'rm -rf "$$out"' EXIT; \
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		os=$${target%/*}; arch=$${target#*/}; echo "Building $$target"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build -o "$$out/gi-$$os-$$arch" ./cmd/gi; \
	done

.PHONY: test-tui-tables
test-tui-tables: test-pi-table-oracle test-tui-tables-unit
	mkdir -p bin
	$(GO) test -c -o bin/gi-tui-table-test ./internal/tui
	$(BUN) scripts/test-tui-tables.mjs

.PHONY: test-pi-table-oracle
test-pi-table-oracle:
	$(BUN) tests/ux/oracle/pi-table-probe.mjs

.PHONY: test-tui-tables-unit
test-tui-tables-unit: test-pi-table-oracle
	PI_TABLE_ORACLE=$(abspath test-results/tui-tables/pi-oracle.json) $(GO) test -race -count=3 ./internal/tui -run 'TestMarkdownTable'

.PHONY: test-tool-input-contract
test-tool-input-contract:
	$(BUN) tests/ux/oracle/pi-tool-input-probe.mjs
	$(GO) test -race -count=3 ./internal/inference -run 'ToolInput'

.PHONY: test-piclaw-stop-queue
test-piclaw-stop-queue:
	$(BUN) tests/ux/oracle/piclaw-stop-queue-probe.mjs
	$(BUN) tests/ux/oracle/piclaw-stop-ui-probe.mjs

.PHONY: test-piclaw-idle-steer
test-piclaw-idle-steer:
	$(BUN) tests/ux/oracle/piclaw-idle-steer-probe.mjs

.PHONY: test-ux-idle-steer
test-ux-idle-steer: test-idle-queue-steer
	$(MAKE) test-ux-parity UX_PARITY_ARGS='tests/ux/queue-idle-steer.spec.mjs'

.PHONY: test-ux-ended-steer
test-ux-ended-steer: build-web
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_ENDED_STEER=1 GI_UX_STEER=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/queue-ended-steer.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-idle-queue-steer
test-idle-queue-steer:
	$(GO) test -race -count=3 ./internal/store ./internal/turn ./internal/web -run 'IdleQueue|QueueSteer|WebQueueHold'

.PHONY: test-ux-message-reference-labels
test-ux-message-reference-labels: test-message-reference-labels
	$(MAKE) test-ux-parity UX_PARITY_ARGS='tests/ux/message-reference-labels.spec.mjs'

.PHONY: test-message-reference-labels
test-message-reference-labels:
	$(BUN) tests/ux/oracle/piclaw-reference-label-probe.mjs
	$(GO) test -race -count=3 ./internal/store -run TestMessageDisplayRows
	$(BUN) test tests/ux/support/message-reference-label.test.ts

.PHONY: test-piclaw-settings-title test-ux-settings-title
test-piclaw-settings-title:
	$(BUN) tests/ux/oracle/piclaw-settings-title-probe.mjs

.PHONY: test-piclaw-settings-shell test-piclaw-settings-shortcut test-piclaw-settings-reopen test-piclaw-settings-stepper test-piclaw-settings-pane-load test-piclaw-settings-layering test-piclaw-non-ipad-lightbox test-piclaw-direct-delete test-piclaw-markdown-table test-piclaw-code-copy test-piclaw-remote-links test-piclaw-outcome-chip
test-piclaw-settings-shell:
	$(BUN) tests/ux/oracle/piclaw-settings-shell-probe.mjs

test-piclaw-settings-shortcut:
	$(BUN) tests/ux/oracle/piclaw-settings-shortcut-probe.mjs

test-piclaw-settings-reopen:
	$(BUN) tests/ux/oracle/piclaw-settings-reopen-probe.mjs

test-piclaw-settings-stepper:
	$(BUN) tests/ux/oracle/piclaw-settings-stepper-probe.mjs

test-piclaw-settings-pane-load:
	$(BUN) tests/ux/oracle/piclaw-settings-pane-load-probe.mjs

test-piclaw-settings-layering:
	$(BUN) tests/ux/oracle/piclaw-settings-layering-probe.mjs

test-piclaw-non-ipad-lightbox:
	$(BUN) tests/ux/oracle/piclaw-non-ipad-lightbox-probe.mjs

test-piclaw-direct-delete:
	$(BUN) tests/ux/oracle/piclaw-direct-delete-probe.mjs

test-piclaw-markdown-table:
	$(BUN) tests/ux/oracle/piclaw-markdown-table-probe.mjs

test-piclaw-code-copy:
	$(BUN) tests/ux/oracle/piclaw-code-copy-probe.mjs

test-piclaw-remote-links:
	$(BUN) tests/ux/oracle/piclaw-remote-links-probe.mjs

test-piclaw-outcome-chip:
	$(BUN) tests/ux/oracle/piclaw-outcome-chip-probe.mjs

.PHONY: test-piclaw-editor-tabs
test-piclaw-editor-tabs:
	$(BUN) tests/ux/oracle/piclaw-editor-tabs-probe.mjs

.PHONY: test-piclaw-login-policy test-piclaw-login-preemption test-piclaw-oobe-panel test-piclaw-invitation-totp test-piclaw-invitation-boundaries test-piclaw-invitation-passkey test-piclaw-family-privacy test-piclaw-pwa-icons test-piclaw-swipe-selection test-piclaw-session-picker-dismiss test-piclaw-session-groups test-piclaw-session-restore
test-piclaw-login-policy:
	$(BUN) tests/ux/oracle/piclaw-login-policy-probe.mjs

test-piclaw-login-preemption:
	$(BUN) tests/ux/oracle/piclaw-login-preemption-probe.mjs

test-piclaw-oobe-panel:
	$(BUN) tests/ux/oracle/piclaw-oobe-panel-probe.mjs

test-piclaw-invitation-totp:
	$(BUN) tests/ux/oracle/piclaw-invitation-totp-probe.mjs

test-piclaw-invitation-boundaries:
	$(BUN) tests/ux/oracle/piclaw-invitation-boundaries-probe.mjs

test-piclaw-invitation-passkey:
	$(BUN) tests/ux/oracle/piclaw-invitation-passkey-probe.mjs

test-piclaw-family-privacy:
	$(BUN) tests/ux/oracle/piclaw-family-privacy-probe.mjs

test-piclaw-pwa-icons:
	$(BUN) tests/ux/oracle/piclaw-pwa-icons-probe.mjs

test-piclaw-swipe-selection:
	$(BUN) tests/ux/oracle/piclaw-swipe-selection-probe.mjs

test-piclaw-session-picker-dismiss:
	$(BUN) tests/ux/oracle/piclaw-session-picker-dismiss-probe.mjs

test-piclaw-session-groups:
	$(BUN) tests/ux/oracle/piclaw-session-groups-probe.mjs

test-piclaw-session-restore:
	$(BUN) tests/ux/oracle/piclaw-session-restore-probe.mjs

.PHONY: test-piclaw-notification-presence test-piclaw-notification-leader
test-piclaw-notification-presence:
	$(BUN) tests/ux/oracle/piclaw-notification-presence-probe.mjs

test-piclaw-notification-leader:
	$(BUN) tests/ux/oracle/piclaw-notification-leader-probe.mjs

.PHONY: test-piclaw-recovery-classic-shapes
test-piclaw-recovery-classic-shapes:
	$(BUN) tests/ux/oracle/piclaw-recovery-classic-shapes-probe.mjs

.PHONY: test-piclaw-recovery-display
test-piclaw-recovery-display:
	$(BUN) tests/ux/oracle/piclaw-recovery-display-probe.mjs

.PHONY: test-piclaw-card-rejection
test-piclaw-card-rejection:
	$(BUN) tests/ux/oracle/piclaw-card-rejection-probe.mjs

.PHONY: test-piclaw-card-identity
test-piclaw-card-identity:
	$(BUN) tests/ux/oracle/piclaw-card-submission-identity-probe.mjs

.PHONY: test-piclaw-btw-panel
test-piclaw-btw-panel:
	$(BUN) tests/ux/oracle/piclaw-btw-panel-probe.mjs

.PHONY: test-piclaw-widget-events test-piclaw-widget-persisted
test-piclaw-widget-events:
	$(BUN) tests/ux/oracle/piclaw-widget-lifecycle-probe.mjs

test-piclaw-widget-persisted:
	$(BUN) tests/ux/oracle/piclaw-widget-persisted-probe.mjs

UX_SETTINGS_PROJECTS ?=
test-ux-settings-title: test-piclaw-settings-title
	$(MAKE) test-ux-parity UX_PARITY_ARGS='tests/ux/settings-shell.spec.mjs tests/ux/gi-settings.spec.mjs $(UX_SETTINGS_PROJECTS)'

.PHONY: test-piclaw-output-oracle test-piclaw-failed-tool-wait
test-piclaw-output-oracle: build-web
	$(BUN) tests/ux/oracle/piclaw-output-contract.mjs

test-piclaw-failed-tool-wait:
	$(BUN) tests/ux/oracle/piclaw-failed-tool-wait-probe.mjs

.PHONY: test-piclaw-tool-output test-piclaw-tool-output-window
test-piclaw-tool-output-window:
	$(BUN) tests/ux/oracle/piclaw-tool-output-window-probe.mjs

test-piclaw-tool-output:
	$(GO) test -race -count=3 ./internal/store ./internal/tools ./internal/turn -run 'TestToolOutput|TestShellToolOutput'
	$(BUN) test tests/ux/support/conversation.test.ts tests/ux/support/piclaw-status-adapter.test.ts

.PHONY: test-conversation-projection
test-conversation-projection:
	$(GO) test -race -count=3 ./internal/store ./internal/web -run 'TestConversation|TestMessagePage'
	$(BUN) test tests/ux/support/conversation.test.ts

.PHONY: test-provider-retry-oracle test-provider-retry
test-provider-retry-oracle:
	$(BUN) tests/ux/oracle/provider-retry-probe.mjs

test-provider-retry:
	$(GO) test -race -count=3 ./internal/config ./internal/turn ./internal/tui -run 'TestProviderRetry|TestProviderHeaderTimeout|TestTransientProviderFailure'

.PHONY: test-codex-tui-regression
test-codex-tui-regression:
	$(GO) test -race -count=3 ./internal/inference -run 'TestCodex'
	$(GO) test -race -count=3 ./internal/tui -run 'TestStreamedErrorAndDurableSystemPost|TestSystemPostDoesNotBecomeAssistant'

vet:
	$(GO) vet ./...

bun-checks:
	$(BUN) run check:hook-tdz

check: test vet build-web bun-checks test-ux

# ── Isolated UX test instance ───────────────────────────────────────────

test-instance-start: build
	@mkdir -p $(TEST_DIR)
	@if [ -f $(TEST_PID) ] && kill -0 $$(cat $(TEST_PID)) 2>/dev/null; then \
		kill $$(cat $(TEST_PID)) 2>/dev/null || true; \
		sleep 1; \
	fi
	@rm -rf $(TEST_WORKSPACE) $(TEST_DB) $(TEST_LOG) $(TEST_PID)
	@mkdir -p $(TEST_WORKSPACE)/.piclaw $(TEST_WORKSPACE)/.pi
	@printf '%s\n' '$(TEST_PICLAW_CONFIG_JSON)' > $(TEST_WORKSPACE)/.piclaw/config.json
	@printf '%s\n' '$(TEST_PI_SETTINGS_JSON)' > $(TEST_WORKSPACE)/.pi/settings.json
	@if [ -n "$(TEST_FIXTURES_DIR)" ]; then cp -R "$(TEST_FIXTURES_DIR)/." "$(TEST_WORKSPACE)/"; fi
	$(abspath $(BIN)) $(TEST_SERVER_ARGS) >$(TEST_DIR)/process.log 2>&1 </dev/null &
	@for _ in 1 2 3 4 5 6 7 8 9 10; do \
		if [ -f $(TEST_PID) ] && kill -0 $$(cat $(TEST_PID)) 2>/dev/null && curl -fsS http://127.0.0.1:$(TEST_PORT)/api/sessions >/dev/null; then \
			echo "Test instance running on 127.0.0.1:$(TEST_PORT) with PID $$(cat $(TEST_PID))"; exit 0; \
		fi; \
		sleep 0.5; \
	done; \
	echo "Test instance failed to start"; cat $(TEST_LOG) 2>/dev/null; exit 1

test-instance-stop:
	@if [ -f $(TEST_PID) ] && kill -0 $$(cat $(TEST_PID)) 2>/dev/null; then \
		kill $$(cat $(TEST_PID)) && echo "Stopped test instance"; \
	fi
	@rm -rf $(TEST_DIR)

test-ux: test-instance-start
	mkdir -p $(TEST_RESULTS)
	GI_TEST_URL=http://127.0.0.1:$(TEST_PORT) $(PLAYWRIGHT) test tests/functional/ --reporter=line --output=$(TEST_RESULTS)/playwright; \
	rc=$$?; \
	$(MAKE) --no-print-directory test-instance-stop; \
	exit $$rc

# Frozen Vibes/Tau Piclaw Classic corpus, with a separate disposable process.
UX_PARITY_PORT ?= 19091
UX_PARITY_ARGS ?=

# Real local inference checkpoints (no paid provider).
UX_LOCAL_ENV ?= GI_UX_STEER=1
UX_LOCAL_SPEC ?= tests/ux/queue-steer.spec.mjs
UX_LOCAL_BIN ?= bin/gi-ux-steer
UX_LOCAL_PORT ?= 19092
.PHONY: test-shared-stop-evidence
test-shared-stop-evidence: test-ux-reconnect
	$(BUN) test tests/ux/support/parity-report.test.ts tests/ux/support/passkey-criteria.test.ts
	$(BUN) scripts/ux-parity-report.mjs test-results/ux-parity/reconnect-results.json

test-ux-reconnect: build-web
	@mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_RECONNECT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/reconnect.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-status-swipes test-ux-mobile-exclusions test-ux-thoughts test-ux-direct-reply-steer
test-ux-direct-reply-steer:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_DIRECT_REPLY_STEER=1 UX_LOCAL_SPEC=tests/ux/direct-reply-steer.spec.mjs

test-ux-thoughts:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_THOUGHTS=1 UX_LOCAL_SPEC=tests/ux/thoughts.spec.mjs

test-ux-status-swipes:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_STATUS_SWIPES=1 UX_LOCAL_SPEC=tests/ux/status-swipes.spec.mjs

test-ux-mobile-exclusions:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV='GI_UX_CARD_REJECTION=1 GI_UX_MOBILE_EXCLUSIONS=1' UX_LOCAL_SPEC=tests/ux/mobile-exclusions.spec.mjs

test-ux-compaction:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_COMPACTION=1 UX_LOCAL_SPEC=tests/ux/compaction.spec.mjs

test-ux-context-meter:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_METER=1 UX_LOCAL_SPEC=tests/ux/context-meter.spec.mjs

test-ux-context-fit:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_CONTEXT=1 UX_LOCAL_SPEC=tests/ux/context-fit.spec.mjs

test-ux-index-config:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_INDEX_CONFIG=1 UX_LOCAL_SPEC=tests/ux/workspace-index-config.spec.mjs

test-ux-shared-copy-delete:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_SHARED_COPY_DELETE=1 UX_LOCAL_SPEC=tests/ux/shared-copy-delete.spec.mjs

test-ux-steer: build-web
	@mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity/queue-gates
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	@set -e; \
	PATH=$(abspath tests/ux/shell):$$PATH $(UX_LOCAL_ENV) GI_UX_LISTEN=127.0.0.1:$(UX_LOCAL_PORT) GI_UX_QUEUE_GATES=$(abspath test-results/ux-parity/queue-gates) $(UX_LOCAL_BIN) >test-results/ux-parity/steer-server.log 2>&1 & pid=$$!; \
	trap 'kill $$pid 2>/dev/null || true; wait $$pid 2>/dev/null || true' EXIT; \
	ready=0; for i in $$(seq 1 100); do kill -0 $$pid || exit 1; if curl -fsS http://127.0.0.1:$(UX_LOCAL_PORT)/api/runtime/config >/dev/null 2>&1; then ready=1; break; fi; sleep .1; done; \
	test $$ready -eq 1; \
	$(UX_LOCAL_ENV) GI_TEST_URL=http://127.0.0.1:$(UX_LOCAL_PORT) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs $(UX_LOCAL_SPEC) $(UX_PARITY_ARGS); \
	if [ -n "$(UX_LOCAL_FUNCTIONAL)" ]; then $(UX_LOCAL_ENV) GI_TEST_URL=http://127.0.0.1:$(UX_LOCAL_PORT) $(PLAYWRIGHT) test --config=playwright.config.ts --output=test-results/local-functional-artifacts $(UX_LOCAL_FUNCTIONAL); fi
ux-parity-inventory:
	$(BUN) test tests/ux/support/
	$(BUN) scripts/ux-parity-report.mjs

.PHONY: ux-parity-report
ux-parity-report:
	$(BUN) scripts/ux-parity-report.mjs $(UX_PARITY_REPORT_ARGS)

.PHONY: build-pane-host-fixture
build-pane-host-fixture:
	@mkdir -p test-results
	$(BUN) scripts/build-pane-host-fixture.mjs

.PHONY: test-web-skills test-ux-skills
.PHONY: test-tool-activity
test-tool-activity:
	$(GO) test -race -count=3 ./internal/store ./internal/web ./internal/turn -run 'ToolActivity|ToolPreview|SessionActivity|ToolTerminal'
	$(BUN) test tests/ux/support/tool-activity.test.ts

.PHONY: test-ux-tool-terminal
test-ux-tool-terminal: test-tool-activity
	$(MAKE) test-ux-parity UX_PARITY_ARGS='tests/ux/tool-activity.spec.mjs'

.PHONY: test-terminal-tool-identity test-tui-tool-timing
.PHONY: test-terminal-links
.PHONY: test-tui-session-actions
test-tui-session-actions: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-session-actions.mjs

.PHONY: test-session-actions
test-session-actions:
	$(GO) test -race -count=3 ./internal/tui ./internal/store -run 'SessionActions|SessionRename|SessionDisplayCapabilities|SessionPicker'

test-terminal-links:
	$(GO) test -race -count=3 ./internal/tui -run 'TranscriptLink|TranscriptSelection|TranscriptSearch'

.PHONY: test-tui-links
test-tui-links: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-links.mjs

.PHONY: test-tui-selection-edge
test-tui-selection-edge: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-selection-edge.mjs

test-terminal-tool-identity:
	$(GO) test -race -count=3 ./internal/tui -run 'ToolRuntime|ToolEndWithout|RenderToolEvent|BuildTranscriptRenderable'

test-tui-tool-timing: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-tool-timing.mjs

test-web-skills:
	$(GO) test -race -count=3 ./internal/web -run 'LoadedWebSkill|WebSkillOpen|QuickActions'

test-ux-skills:
	GI_UX_SKILLS=1 $(MAKE) test-ux-parity TEST_FIXTURES_DIR=tests/ux/fixtures/skills UX_PARITY_ARGS='tests/ux/skills.spec.mjs'

.PHONY: test-recovery-marker test-ux-outcomes
test-recovery-marker:
	$(GO) test -race -count=3 ./internal/store ./internal/turn -run 'RecoveryMarker|StartupRecoveryRequeuesCompactingTurn'
.PHONY: test-ux-card-rejection
test-ux-card-rejection:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_CARD_REJECTION=1 GI_UX_RECOVERY_PLACEHOLDERS=1' UX_LOCAL_SPEC='tests/ux/card-rejection.spec.mjs tests/ux/recovery-placeholders.spec.mjs tests/ux/speech.spec.mjs' UX_LOCAL_FUNCTIONAL='tests/functional/16-card-rejection.spec.ts tests/functional/15-recovery-placeholders.spec.ts'
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

.PHONY: test-ux-recovery-placeholders
test-ux-recovery-placeholders:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_RECOVERY_CONTROLS=1 GI_UX_RECOVERY_PLACEHOLDERS=1' UX_LOCAL_SPEC='tests/ux/recovery-placeholders.spec.mjs tests/ux/recovery-controls.spec.mjs tests/ux/speech.spec.mjs' UX_LOCAL_FUNCTIONAL='tests/functional/14-recovery-controls.spec.ts tests/functional/15-recovery-placeholders.spec.ts'
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

.PHONY: test-ux-recovery-controls
test-ux-recovery-controls:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_RECOVERY_CONTROLS=1' UX_LOCAL_SPEC='tests/ux/recovery-controls.spec.mjs tests/ux/message-copy.spec.mjs tests/ux/message-delete.spec.mjs' UX_LOCAL_FUNCTIONAL=tests/functional/14-recovery-controls.spec.ts
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

test-ux-outcomes:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_OUTCOMES=1' UX_LOCAL_SPEC='tests/ux/outcomes.spec.mjs tests/ux/message-copy.spec.mjs tests/ux/speech.spec.mjs' UX_LOCAL_FUNCTIONAL=tests/functional/13-outcomes.spec.ts
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

.PHONY: test-ux-card-identity
test-ux-card-identity:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_CARD_IDENTITY=1' UX_LOCAL_SPEC='tests/ux/card-submission-identity.spec.mjs'

.PHONY: test-ux-btw-mount
test-ux-btw-mount: build-web
	@mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_BTW_MOUNT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs

.PHONY: test-ux-widget-persisted
test-ux-widget-persisted:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_WIDGETS=1' UX_LOCAL_SPEC='tests/ux/widget-persisted.spec.mjs'

.PHONY: test-ux-links
test-ux-links:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_LINKS=1' UX_LOCAL_SPEC='tests/ux/remote-links.spec.mjs tests/ux/rendering.spec.mjs tests/ux/lightbox.spec.mjs' UX_LOCAL_FUNCTIONAL=tests/functional/12-remote-links.spec.ts
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

.PHONY: test-ux-speech-contract
test-ux-speech-contract:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_SPEECH=1' UX_LOCAL_SPEC='tests/ux/speech-contract.spec.mjs tests/ux/speech.spec.mjs tests/ux/message-copy.spec.mjs tests/ux/rendering.spec.mjs'
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

.PHONY: capture-gi-chat-baseline
capture-gi-chat-baseline:
	$(BUN) tests/ux/oracle/gi-chat-baseline-probe.mjs

.PHONY: test-pi-piclaw-queue-oracle test-pi-idle-escape-oracle test-pi-tool-boundary-steer test-pi-direct-reply-steer test-pi-steering-modes test-pi-abort-steering test-pi-steering-mode-switch test-pi-preparation-steer
test-pi-piclaw-queue-oracle:
	$(BUN) tests/ux/oracle/pi-piclaw-queue-probe.mjs

test-pi-idle-escape-oracle:
	$(BUN) tests/ux/oracle/pi-idle-escape-pty.mjs

test-pi-tool-boundary-steer:
	$(BUN) tests/ux/oracle/pi-tool-boundary-steer-probe.mjs

test-pi-direct-reply-steer:
	$(BUN) tests/ux/oracle/pi-direct-reply-steer-probe.mjs

test-pi-steering-modes:
	$(BUN) tests/ux/oracle/pi-steering-mode-order-probe.mjs

test-pi-abort-steering:
	$(BUN) tests/ux/oracle/pi-abort-steering-probe.mjs

test-pi-steering-mode-switch:
	$(BUN) tests/ux/oracle/pi-steering-mode-switch-probe.mjs

test-pi-preparation-steer:
	$(BUN) tests/ux/oracle/pi-preparation-steer-probe.mjs

.PHONY: test-piclaw-queue-reorder test-piclaw-queue-return test-ux-queue-return
test-piclaw-queue-reorder:
	$(BUN) tests/ux/oracle/piclaw-queue-reorder-probe.mjs

test-piclaw-queue-return:
	$(BUN) tests/ux/oracle/piclaw-return-probe.mjs

test-ux-queue-return:
	$(BUN) test tests/ux/support/drafts.test.ts tests/ux/support/queue-return.test.ts
	$(MAKE) test-ux-parity UX_PARITY_ARGS='tests/ux/queue-return.spec.mjs tests/ux/message-reference-labels.spec.mjs'

.PHONY: diagnose-webkit-unload
diagnose-webkit-unload:
	$(BUN) tests/ux/oracle/webkit-unload-diagnostic.mjs

.PHONY: test-piclaw-chat-lifecycle
test-piclaw-chat-lifecycle:
	$(BUN) tests/ux/oracle/piclaw-chat-lifecycle-probe.mjs

.PHONY: test-piclaw-oracle-basic
test-piclaw-oracle-basic:
	$(BUN) test tests/ux/oracle/piclaw-basic-contract.test.mjs
	$(BUN) tests/ux/oracle/piclaw-basic-probe.mjs

.PHONY: test-piclaw-oracle-matrix
test-piclaw-oracle-matrix:
	$(BUN) test tests/ux/oracle/piclaw-basic-contract.test.mjs
	@set -e; for browser in chromium webkit; do for viewport in phone tablet desktop; do ORACLE_BROWSER=$$browser ORACLE_VIEWPORT=$$viewport $(BUN) tests/ux/oracle/piclaw-basic-probe.mjs > test-results/ux-oracle-$$browser-$$viewport.log; done; done

test-ux-parity:
	@mkdir -p test-results/ux-parity/queue-gates
	PATH="$(abspath tests/ux/shell):$$PATH" GI_UX_QUEUE_GATES="$(abspath test-results/ux-parity/queue-gates)" $(MAKE) --no-print-directory test-instance-start TEST_PORT=$(UX_PARITY_PORT) TEST_DIR=.gi-ux-parity TEST_ENABLED_MODELS='["test-model","bootstrap","test/unavailable-model"]'
	@trap '$(MAKE) --no-print-directory test-instance-stop TEST_DIR=.gi-ux-parity' EXIT; \
		$(MAKE) --no-print-directory build-pane-host-fixture || exit 1; \
		rm -f test-results/ux-parity/results.json; \
		GI_TEST_URL=http://127.0.0.1:$(UX_PARITY_PORT) $(PLAYWRIGHT) test -c playwright.ux.config.mjs $(UX_PARITY_ARGS); \
		rc=$$?; \
		$(BUN) scripts/ux-parity-report.mjs test-results/ux-parity/results.json || exit 1; \
		exit $$rc

test-tui-compaction:
	@mkdir -p bin
	$(GO) test -c -o bin/gi-tui-compaction-test ./internal/tui
	$(BUN) scripts/test-tui-compaction.mjs


.PHONY: test-tui-queue-input test-tui-queue-input-core test-tui-queue-input-binary test-tui-queue-restore test-tui-queue-escape-restore test-tui-queue-display
test-tui-queue-input: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-queue-input.mjs

test-tui-queue-input-binary:
	$(BUN) scripts/test-tui-queue-input.mjs

test-tui-queue-restore: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-queue-restore.mjs

test-tui-queue-escape-restore: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-queue-escape-restore.mjs

test-tui-queue-display: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-queue-display.mjs

test-tui-queue-input-core:
	$(GO) test -race -count=3 ./internal/tui -run 'TestPiEnter|TestPiFollowUp'
	$(GO) test -race -count=3 ./internal/turn -run 'TestTUIComposerFollowUp|TestTUIComposerRejectsUnknownDelivery|TestTUIComposerSubmit'

.PHONY: test-tui-inline-prose test-tui-inline-prose-binary
test-tui-inline-prose: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-inline-prose.mjs

# Probe an already-built binary (for retained failing-baseline evidence).
test-tui-inline-prose-binary:
	$(BUN) scripts/test-tui-inline-prose.mjs

test-tui-markdown: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-markdown.mjs

test-tui-source-copy: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-source-copy.mjs

.PHONY: test-tui-model-picker
test-tui-model-picker:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/gi-tui-model-picker ./tests/tui-model-picker
	GI_TUI_MODEL_BIN=$(abspath $(BIN_DIR)/gi-tui-model-picker) $(BUN) scripts/test-tui-model-picker.mjs

.PHONY: test-tui-reading test-tui-outcomes test-tui-regular test-tui-search test-tui-selection test-tui-index

test-tui-index: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-index.mjs
test-tui-selection: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-selection.mjs

.PHONY: bench-tui-editor-layout
bench-tui-editor-layout:
	$(GO) test ./internal/tui -run '^$$' -bench '^BenchmarkEditorLayout$$' -benchmem -benchtime=100ms

.PHONY: test-tui-editor-viewport
test-tui-editor-viewport: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-editor-viewport.mjs

test-tui-search: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-search.mjs

test-tui-regular: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-regular.mjs

test-tui-outcomes: build
	$(BUN) scripts/test-tui-outcomes.mjs

test-tui-reading: build
	$(BUN) scripts/test-tui-reading.mjs

test-tui-input-history: build
	FEATURE_FILE=$(abspath features/tui/input_history.feature) ARTIFACT_DIR=$(abspath $(TEST_RESULTS))/tui-input-history TEST_DIR=$(abspath $(TUI_TEST_DIR))-input-history bash scripts/test-tui-gherkin.sh

test-tui-sessions: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-sessions.mjs

.PHONY: test-tui-pending-media
.PHONY: test-tui-durable-draft test-tui-durable-draft-pty
test-tui-durable-draft-pty: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-durable-draft.mjs

test-tui-durable-draft:
	$(GO) test -race -count=3 ./internal/tui -run DurableDraft

.PHONY: test-tui-text-journal test-concurrent-session-submit
test-concurrent-session-submit:
	$(GO) test -race -count=10 ./internal/turn -run '^TestConcurrentSubmitDifferentSessionsRunsConcurrently$$'

test-tui-text-journal:
	$(GO) test -race -count=3 ./internal/store ./internal/turn -run 'TUITextDraft|TUIComposerDraft|TUIComposerSubmit'

.PHONY: test-tui-media-journal
test-tui-media-journal:
	$(GO) test -race -count=3 ./internal/store ./internal/tui -run 'TUIMediaDraft|PendingMedia|AttachCommand|PasteImage'

.PHONY: test-held-retry
test-held-retry:
	$(GO) test -race -count=3 ./internal/store ./internal/turn ./internal/tui -run 'TUIRetry|RetryHeld|HeldRetry|HoldAndResolve|HoldResolution|SkipHeld'

.PHONY: test-tui-unselected-model test-tui-submit-guards
test-tui-submit-guards:
	$(GO) test -race -count=3 ./internal/tui -run 'TUIUnselectedModel|PendingMediaCommandsLimitsAndNoModelDraft'

test-tui-unselected-model: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-unselected-model.mjs

.PHONY: test-tui-retry-commands
test-tui-retry-commands: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-retry-commands.mjs

.PHONY: test-tui-queue-commands test-terminal-queue
test-terminal-queue:
	$(GO) test -race -count=3 ./internal/tui ./internal/store ./internal/turn -run 'TUIQueue|QueueSteer|QueuedTurn|QueueOrder'

test-tui-queue-commands: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-queue-commands.mjs

test-tui-pending-media: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-pending-media.mjs

.PHONY: test-tui-session-picker
test-tui-session-picker: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-session-picker.mjs

test-tui-smoke: build
	chmod +x scripts/test-tui-smoke.sh
	ARTIFACT_DIR=$(abspath $(TEST_RESULTS))/tui-smoke TEST_DIR=$(abspath $(TUI_TEST_DIR)) scripts/test-tui-smoke.sh

test-tui-gherkin: build test-tui-markdown test-tui-inline-prose
	chmod +x scripts/test-tui-gherkin.sh
	ARTIFACT_DIR=$(abspath $(TEST_RESULTS))/tui-gherkin TEST_DIR=$(abspath $(TUI_TEST_DIR))-gherkin scripts/test-tui-gherkin.sh

# ── Cleanup ─────────────────────────────────────────────────────────────

clean:
	rm -rf $(RUN_DIR) $(BIN_DIR) $(TEST_DIR) $(TUI_TEST_DIR) $(TEST_RESULTS)
	rm -f gi gi-tui

.PHONY: pixel-install pixel-baseline pixel-compare test-pixel-helpers
# Build-time screenshot decoding only; not part of the native runtime.
pixel-install:
	$(BUN) install

# Exits nonzero until repeated captures are stable and cross-host pixels match.
# PICLAW_PIXEL_ROOT must point to the pinned installed runtime tree.
pixel-baseline: build-web
	xvfb-run -a -s '-screen 0 1920x1200x24' $(BUN) scripts/compose-pixel-baseline.mjs

pixel-compare:
	$(BUN) scripts/compose-pixel-compare.mjs

test-pixel-helpers:
	$(BUN) test tests/ux/support/pixel-*.test.mjs

.PHONY: test-context-control-helpers
test-context-control-helpers:
	$(BUN) test tests/ux/support/context-usage.test.ts tests/ux/support/compaction-state.test.ts

.PHONY: test-notification-helpers test-ux-notifications
test-notification-helpers:
	$(BUN) test tests/ux/support/notifications.test.ts

test-ux-notifications: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_NOTIFICATIONS=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/notifications.spec.mjs

.PHONY: test-voice-input-helpers test-ux-voice-input
test-voice-input-helpers:
	$(BUN) test tests/ux/support/voice-input.test.ts tests/ux/support/voice-adapter.test.ts

test-ux-voice-input: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_VOICE_INPUT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/voice-input.spec.mjs

.PHONY: test-ux-session-panel test-session-panel-helpers
test-session-panel-helpers:
	$(BUN) test tests/ux/support/session-panel.test.ts

test-ux-session-panel: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_SESSION_PANEL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/session-panel.spec.mjs

.PHONY: test-ux-model-panel test-model-panel-helpers
test-model-panel-helpers:
	$(BUN) test tests/ux/support/model-panel.test.ts tests/ux/support/model-picker.test.ts tests/ux/support/model-accessibility.test.ts tests/ux/support/model-thinking.test.ts

.PHONY: test-piclaw-quick-action-prefill test-piclaw-typeahead-exclusions test-piclaw-upload-state test-piclaw-copy-controls
test-piclaw-quick-action-prefill:
	$(BUN) tests/ux/oracle/piclaw-quick-action-prefill-probe.mjs

test-piclaw-typeahead-exclusions:
	$(BUN) tests/ux/oracle/piclaw-typeahead-exclusions-probe.mjs

test-piclaw-upload-state:
	$(BUN) tests/ux/oracle/piclaw-upload-state-probe.mjs

test-piclaw-copy-controls:
	$(BUN) tests/ux/oracle/piclaw-copy-ui-probe.mjs

.PHONY: test-piclaw-stale-terminal
test-piclaw-stale-terminal:
	$(BUN) tests/ux/oracle/piclaw-stale-terminal-probe.mjs

.PHONY: test-piclaw-speech-ui
test-piclaw-speech-ui:
	$(BUN) tests/ux/oracle/piclaw-speech-ui-probe.mjs

.PHONY: test-piclaw-concurrent-tools
test-piclaw-concurrent-tools:
	$(BUN) tests/ux/oracle/piclaw-concurrent-tools-probe.mjs

.PHONY: test-piclaw-svg-adversarial
test-piclaw-svg-adversarial:
	$(BUN) tests/ux/oracle/piclaw-svg-adversarial-probe.mjs

.PHONY: test-piclaw-theme-handler
test-piclaw-theme-handler:
	PICLAW_DB_IN_MEMORY=1 $(BUN) tests/ux/oracle/piclaw-theme-handler-probe.mjs

.PHONY: test-piclaw-theme-combined
test-piclaw-theme-combined:
	PICLAW_DB_IN_MEMORY=1 $(BUN) tests/ux/oracle/piclaw-theme-combined-probe.mjs

.PHONY: test-piclaw-message-retrieval
test-piclaw-message-retrieval:
	PICLAW_DB_IN_MEMORY=1 $(BUN) tests/ux/oracle/piclaw-message-retrieval-probe.mjs

.PHONY: test-piclaw-compose-escape test-ux-compose-escape
test-piclaw-compose-escape:
	$(BUN) tests/ux/oracle/piclaw-compose-escape-probe.mjs

test-ux-compose-escape: build-web
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_COMPOSE_ESCAPE=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/compose-escape.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-piclaw-picker-thinking
test-piclaw-picker-thinking:
	$(BUN) tests/ux/oracle/piclaw-picker-thinking-probe.mjs

.PHONY: test-ux-picker-thinking
test-ux-picker-thinking: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PICKER_THINKING=1 GI_UX_THINKING=1 GI_UX_THINKING_DEFAULT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/picker-thinking.spec.mjs $(UX_PARITY_ARGS)

test-ux-model-panel: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_MODEL_PANEL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/model-panel.spec.mjs

.PHONY: test-ux-compose-surface test-compose-surface-helpers
test-compose-surface-helpers:
	$(BUN) test tests/ux/support/compose-surface.test.ts tests/ux/support/accent-contrast.test.ts tests/ux/support/theme-text-contrast.test.ts

test-ux-compose-surface: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_COMPOSE_SURFACE=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/compose-surface.spec.mjs

.PHONY: test-ux-workspace-tabs
test-ux-workspace-tabs:
	$(MAKE) test-ux-parity UX_PARITY_ARGS='tests/ux/workspace-tabs.spec.mjs'
	cp test-results/ux-parity/results.json test-results/ux-parity/workspace-tabs-results.json

.PHONY: test-ux-slash
test-ux-slash: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_SLASH=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(BUN) x playwright test --config=playwright.ux.config.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-picker-geometry
test-ux-picker-geometry: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PICKER_GEOMETRY=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(BUN) x playwright test --config=playwright.ux.config.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-journey
test-ux-journey: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_JOURNEY=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(BUN) x playwright test --config=playwright.ux.config.mjs $(UX_PARITY_ARGS)

.PHONY: test-browser-auth test-ux-auth
test-browser-auth:
	$(GO) test ./internal/web ./internal/auth -run 'Test(Browser|ProtectedEndpoint|Auth)' -count=1

.PHONY: test-passkey-criteria test-ux-passkeys
.PHONY: test-passkey-cancel-repeat
test-passkey-cancel-repeat: build-web test-passkey-criteria
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PASSKEYS=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config playwright.ux.config.mjs tests/ux/passkeys.spec.mjs --project=chromium-phone --project=chromium-tablet --project=chromium-desktop --grep 'Settings enrolls two passkeys|passkey login cancellation owns' --repeat-each=3

.PHONY: test-passkey-login-boundary
test-passkey-login-boundary: build-web test-passkey-criteria
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PASSKEYS=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config playwright.ux.config.mjs tests/ux/passkeys.spec.mjs --project=chromium-phone --project=chromium-tablet --project=chromium-desktop --grep 'passkey login cancellation owns'

test-passkey-criteria:
	$(BUN) test tests/ux/support/passkey-criteria.test.ts tests/ux/support/passkey-lifecycle.test.ts

test-ux-passkeys: build-web test-passkey-criteria
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PASSKEYS=1 GI_UX_SERVER_BIN=$(UX_LOCAL_BIN) $(BUN) x playwright test --config playwright.ux.config.mjs tests/ux/passkeys.spec.mjs --project=chromium-phone --project=chromium-tablet --project=chromium-desktop

test-ux-auth: build-web
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_AUTH=1 GI_UX_SERVER_BIN=$(UX_LOCAL_BIN) $(BUN) x playwright test --config playwright.ux.config.mjs tests/ux/auth.spec.mjs

.PHONY: test-browser-auth-race
test-browser-auth-race:
	$(GO) test -race ./internal/web ./internal/auth -count=3

.PHONY: test-auth-state
test-auth-state:
	$(GO) test -race ./internal/auth -count=10

# Compare actual tmux cells against ANSI frame diffs while tables enter/leave
# the viewport. No provider, sqlite3 CLI, Bun, or running Gi instance required.
.PHONY: test-tui-table-scroll
test-tui-table-scroll:
	GI_TABLE_SCROLL_PTY=1 $(GO) test -count=1 ./internal/tui -run '^TestMarkdownTableScrollTerminal$$' -v

# Run interactively in the affected Ghostty window, without tmux.
.PHONY: probe-tui-ghostty
probe-tui-ghostty:
	$(GO) run ./tests/tui-ghostty-probe

# Joker codemode prototype (explicit CLI; no running-instance mutations).
.PHONY: deps-mcp test-codemode codemode
deps-mcp:
	$(GO) get github.com/modelcontextprotocol/go-sdk@v1.8.0

test-codemode:
	$(GO) test ./internal/codemode ./cmd/gi -run 'TestCodemode|TestMCP' -count=1

codemode: build
	$(BIN) codemode -config "$(MCP_CONFIG)" -script "$(CODEMODE_SCRIPT)"
