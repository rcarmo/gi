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
	test-instance-start test-instance-stop test-ux test-ux-parity test-ux-index-config ux-parity-inventory test-tui-smoke test-tui-gherkin test-tui-sessions test-tui-source-copy \
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
test-ux-reconnect: build-web
	@mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_RECONNECT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/reconnect.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-status-swipes test-ux-thoughts
test-ux-thoughts:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_THOUGHTS=1 UX_LOCAL_SPEC=tests/ux/thoughts.spec.mjs

test-ux-status-swipes:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_STATUS_SWIPES=1 UX_LOCAL_SPEC=tests/ux/status-swipes.spec.mjs

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
	$(GO) test -race -count=3 ./internal/store ./internal/web -run 'ToolActivity|ToolPreview|SessionActivity'

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

.PHONY: test-ux-links
test-ux-links:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_LINKS=1' UX_LOCAL_SPEC='tests/ux/remote-links.spec.mjs tests/ux/rendering.spec.mjs tests/ux/lightbox.spec.mjs' UX_LOCAL_FUNCTIONAL=tests/functional/12-remote-links.spec.ts
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

.PHONY: test-ux-speech-contract
test-ux-speech-contract:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_SPEECH=1' UX_LOCAL_SPEC='tests/ux/speech-contract.spec.mjs tests/ux/speech.spec.mjs tests/ux/message-copy.spec.mjs tests/ux/rendering.spec.mjs'
	$(MAKE) ux-parity-report UX_PARITY_REPORT_ARGS=test-results/ux-parity/results.json

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

test-tui-search: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-search.mjs

test-tui-regular: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-regular.mjs

test-tui-outcomes: build
	$(BUN) scripts/test-tui-outcomes.mjs

test-tui-reading: build
	$(BUN) scripts/test-tui-reading.mjs

test-tui-sessions: build
	$(BUN) scripts/test-tui-sessions.mjs

.PHONY: test-tui-pending-media
test-tui-pending-media: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-pending-media.mjs

.PHONY: test-tui-session-picker
test-tui-session-picker: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-session-picker.mjs

test-tui-smoke: build
	chmod +x scripts/test-tui-smoke.sh
	ARTIFACT_DIR=$(abspath $(TEST_RESULTS))/tui-smoke TEST_DIR=$(abspath $(TUI_TEST_DIR)) scripts/test-tui-smoke.sh

test-tui-gherkin: build
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

.PHONY: test-ux-session-panel test-session-panel-helpers
test-session-panel-helpers:
	$(BUN) test tests/ux/support/session-panel.test.ts

test-ux-session-panel: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_SESSION_PANEL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/session-panel.spec.mjs

.PHONY: test-ux-model-panel test-model-panel-helpers
test-model-panel-helpers:
	$(BUN) test tests/ux/support/model-panel.test.ts tests/ux/support/model-picker.test.ts

test-ux-model-panel: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_MODEL_PANEL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.ux.config.mjs tests/ux/model-panel.spec.mjs

.PHONY: test-ux-compose-surface test-compose-surface-helpers
test-compose-surface-helpers:
	$(BUN) test tests/ux/support/compose-surface.test.ts

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
test-passkey-criteria:
	$(BUN) test tests/ux/support/passkey-criteria.test.ts

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
