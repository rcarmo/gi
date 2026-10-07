.PHONY: \
	help bootstrap deps \
	build-web build \
	run start stop restart status logs \
	test vet bun-checks check \
	test-instance-start test-instance-stop test-ux test-ux-parity test-ux-index-config test-web-adapters test-tui-smoke test-tui-gherkin test-tui-input-history test-tui-sessions test-tui-source-copy test-tui-markdown \
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
		"  make test-ux-parity   Alias for pinned shared Classic compliance" \
		"  make fixtures-vibes   Run pinned shared Classic compliance (six projects)" \
		"  make test-web-adapters Run Gi web adapter unit tests" \
		"  make test-web-regression Run Gi-only browser regressions" \
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

bootstrap: deps playwright-browsers
	$(MAKE) --no-print-directory build
	@echo "Bootstrap complete. Run 'make start' or 'make run'."

# Playwright browsers (idempotent; a no-op when already installed). Browsers
# live in ~/.cache/ms-playwright, which cache cleanups may remove.
.PHONY: playwright-browsers
playwright-browsers:
	$(PLAYWRIGHT) install chromium webkit

deps:
	$(call require-command,$(GO),Go is required but not installed or not on PATH)
	$(call require-command,$(BUN),Bun is required but not installed or not on PATH)
	$(GO) mod download
	$(BUN) install --frozen-lockfile
	$(MAKE) -C $(GI_UI) deps

# ── Build ───────────────────────────────────────────────────────────────

build-web:
	$(MAKE) -C $(GI_UI) build

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
test-shared-capability-evidence: fixtures-vibes

.PHONY: test-message-retrieval
test-message-retrieval:
	$(TESTPROFILE) gotest $(RACE) ./internal/store ./internal/tools ./internal/turn -run 'TestMessageRows|TestMessageRetrieval|TestPiclawMessages' -count=3

.PHONY: test-ux-message-retrieval
test-ux-message-retrieval: build-web test-message-retrieval
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_MESSAGE_RETRIEVAL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test -c playwright.web-regression.config.mjs tests/web-regression/message-retrieval.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-shared-message-evidence
test-shared-message-evidence: fixtures-vibes

test-session-thinking:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/inference ./internal/turn ./internal/web -run SessionThinking

test-ux-thinking: build-web test-session-thinking
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_THINKING=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config playwright.web-regression.config.mjs tests/web-regression/session-thinking.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-web-queue-hold
test-web-queue-hold:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/turn ./internal/web -run WebQueueHold

.PHONY: test-web-send-receipts
test-web-send-receipts:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/web -run WebSendReceipt

.PHONY: test-web-http-helpers test-web-basic-send test-web-basic-controls

test-web-basic-controls:
	$(MAKE) test-ux-steer UX_LOCAL_ENV="GI_UX_BASIC_HTTP=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN))" UX_LOCAL_SPEC=tests/web-regression/basic-http-control.spec.mjs
test-web-http-helpers:
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/random-id.test.ts tests/unit/drafts.test.ts tests/unit/send-recovery.test.ts

test-web-basic-send: test-instance-start
	@GI_TEST_URL=http://127.0.0.1:$(TEST_PORT) $(PLAYWRIGHT) test tests/functional/18-basic-http-send.spec.ts tests/functional/19-http-delivery.spec.ts --reporter=line --output=$(TEST_RESULTS)/basic-http-send; \
	status=$$?; $(MAKE) test-instance-stop; exit $$status

# Focused and full Go tests always collect CPU/allocation profiles.
test: $(TESTPROFILE)
	$(TESTPROFILE) go $(if $(TEST_RUN),-run '$(TEST_RUN)') $(TEST_PKGS)

# Same-agent copies, distinct routing identities and explicit peer creation.
# Bounded compatibility checks for the Piclaw v3.3.0 shared UI pin.
.PHONY: test-ui-pin-compatibility
test-ui-pin-compatibility: $(TESTPROFILE)
	$(TESTPROFILE) gotest $(RACE) -count=3 -run 'TestStaticViewerPagesMatchPinnedPolicies|TestSettingsModuleHTTPGraph|TestStaticAssetsServePrecompressedVariants|TestSessionMediaEndpoints|TestMultipartUploadRetriesReuseOnlyExactSessionFile|TestWorkspaceEndpoints|TestWorkspaceFileRejectsSymlinkEscape|TestWorkspaceEditorCompatibility|TestSidePromptHTTP|TestPromptTargetingCurrentAgentStaysInSession|TestForkSessionKeepsSourceAgent|TestCreateSessionForkFromKeepsAgentUnlessExplicitPeer|TestQueueSteer' ./internal/web

.PHONY: test-session-copy-identity bench-session-copy
bench-session-copy: $(TESTPROFILE)
	$(TESTPROFILE) gotest -run '^$$' -bench '^BenchmarkCloneSession$$' -benchtime=2s -benchmem ./internal/store

test-session-copy-identity: $(TESTPROFILE)
	$(TESTPROFILE) gotest $(RACE) -run 'TestCloneSession|TestCloneDefault|TestCloneSessionCommand|TestFork|TestCreateSessionForkFrom|TestTree|TestResolveSessionRef|TestHandleCommandFork' -count=3 ./internal/store ./internal/tui ./internal/web

.PHONY: test-code-index-guidance
test-code-index-guidance: $(TESTPROFILE)
	$(TESTPROFILE) gotest $(RACE) -run 'TestIndexPanel|TestPlainCodeWorkspaceExplicitRootPolicy|TestConfiguredScanWithoutBuiltinRoots|TestOptionalRootsAbsentAppearAndDisappear' -count=3 ./internal/tui ./internal/search/indexer

# Local fake-provider/MCP acceptance; no live credentials or browser fixture.
.PHONY: test-mcp-provider-auth
test-mcp-provider-auth: $(TESTPROFILE)
	$(TESTPROFILE) gotest $(RACE) -run 'TestProviderAuth|TestProviderToken' -count=3 ./internal/mcp ./internal/inference

.PHONY: bench-hotspot-targets test-hotspot-regressions
bench-hotspot-targets: $(TESTPROFILE)
	$(TESTPROFILE) gotest -run '^$$' -bench 'Benchmark(RuntimeOptions|ListSessions|TranscriptProjection)$$' -benchtime=1s -benchmem ./internal/inference ./internal/store ./internal/tui

test-hotspot-regressions: $(TESTPROFILE)
	$(TESTPROFILE) gotest $(RACE) -run 'TestRuntimeOptions|TestListSessions|TestSession.*Identity|TestLegacySessionIdentity|TestTranscript(Search|Projection|BlocksMemo|Window)' -count=3 ./internal/inference ./internal/store ./internal/tui

.PHONY: test-shell-runtime check-cross-build test-active-steering test-web-terminal bench-web-terminal test-vnc-api

test-vnc-api:
	$(BUN) scripts/test-vnc-api.mjs

.PHONY: test-side-prompt test-ux-side-prompt-api
test-side-prompt:
	$(TESTPROFILE) gotest ./internal/turn ./internal/web -run TestSidePrompt -count=3

test-ux-side-prompt-api: playwright-browsers
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_SIDE_PROMPT=1 GI_WEB_REGRESSION_LIST_ALL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test -c playwright.web-regression.config.mjs tests/web-regression/side-prompt-api.spec.mjs --project=chromium-desktop


test-web-terminal:
	$(TESTPROFILE) gotest ./internal/web -run TestWebTerminal -count=3

bench-web-terminal:
	$(TESTPROFILE) gotest ./internal/web -run '^$$' -bench BenchmarkWebTerminalReplayRing -benchmem -benchtime=300ms


test-active-steering:
	$(TESTPROFILE) gotest $(RACE) -count=50 ./internal/turn -run '^TestSubmitPromptSteersSecondPromptToActiveTurn$$'

test-shell-runtime:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tools -run 'RunShellPrompt|KillShellProcess'
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/turn -run 'CancelTurn|Cancelled|CancelQueuedTurn'

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
	PI_TABLE_ORACLE=$(abspath test-results/tui-tables/pi-oracle.json) $(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui -run 'TestMarkdownTable'

.PHONY: test-tool-input-contract
test-tool-input-contract:
	$(BUN) tests/ux/oracle/pi-tool-input-probe.mjs
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/inference -run 'ToolInput'

.PHONY: test-ux-idle-steer
test-ux-idle-steer: test-idle-queue-steer
	$(MAKE) test-ux-parity-regression UX_PARITY_ARGS='tests/web-regression/queue-idle-steer.spec.mjs'

.PHONY: test-ux-ended-steer
test-ux-ended-steer: build-web
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_ENDED_STEER=1 GI_UX_STEER=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/queue-ended-steer.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-idle-queue-steer
test-idle-queue-steer:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/turn ./internal/web -run 'IdleQueue|QueueSteer|WebQueueHold'

.PHONY: test-ux-settings-title

UX_SETTINGS_PROJECTS ?=
test-ux-settings-title:
	$(MAKE) test-ux-parity-regression UX_PARITY_ARGS='tests/web-regression/settings-shell.spec.mjs tests/web-regression/gi-settings.spec.mjs $(UX_SETTINGS_PROJECTS)'

.PHONY: test-conversation-projection
test-conversation-projection:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/web -run 'TestConversation|TestMessagePage'
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/conversation.test.ts

.PHONY: test-provider-retry-oracle test-provider-retry
test-provider-retry-oracle:
	$(BUN) tests/ux/oracle/provider-retry-probe.mjs

test-provider-retry:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/config ./internal/turn ./internal/tui -run 'TestProviderRetry|TestProviderHeaderTimeout|TestTransientProviderFailure'

.PHONY: test-codex-tui-regression
test-codex-tui-regression:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/inference -run 'TestCodex'
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui -run 'TestStreamedErrorAndDurableSystemPost|TestSystemPostDoesNotBecomeAssistant'

vet:
	$(GO) vet ./...

bun-checks:
	$(MAKE) -C $(GI_UI) check

check: test vet build-web bun-checks test-web-adapters test-ux

.PHONY: fixtures-vibes fixtures-vibes-focused test-fixtures-profile-lifecycle
FIXTURES_RUNTIME_PROFILE_DIR ?= $(CURDIR)/test-results/fixtures-runtime-profile

# Startup/shutdown profiling check before a long frozen compliance run.
test-fixtures-profile-lifecycle:
	mkdir -p $(BIN_DIR)
	$(GO) build -tags fixtures_vibes -o $(BIN_DIR)/gi-fixtures-profile-smoke ./cmd/gi
	@set -eu; root=$$(mktemp -d /workspace/tmp/gi-fixture-profile-XXXXXX); \
	trap 'test -z "$${pid:-}" || kill "$${pid}" 2>/dev/null || true' EXIT; \
	mkdir -p "$$root/home" "$$root/workspace"; \
	HOME="$$root/home" PI_OFFLINE=1 GI_FIXTURE_PROFILE_DIR="$$root/profiles" FIXTURE_MODEL_URL=http://127.0.0.1:19999/v1 \
	$(abspath $(BIN_DIR))/gi-fixtures-profile-smoke -web -bind 127.0.0.1 -port 19095 -db "$$root/state.db" -workspace "$$root/workspace" >"$$root/server.log" 2>&1 & pid=$$!; \
	ready=0; for i in $$(seq 1 100); do if curl -fsS http://127.0.0.1:19095/api/sessions >/dev/null 2>&1; then ready=1; break; fi; sleep .1; done; \
	test "$$ready" = 1 || { cat "$$root/server.log"; exit 1; }; \
	kill -TERM $$pid; wait $$pid; pid=; \
	for p in "$$root"/profiles/*; do test -s "$$p/cpu.pprof"; test -s "$$p/mem.pprof"; \
	$(GO) tool pprof -top -cum -nodecount=8 "$$p/cpu.pprof"; \
	$(GO) tool pprof -top -cum -sample_index=alloc_space -nodecount=8 "$$p/mem.pprof"; \
	$(GO) tool pprof -top -cum -sample_index=alloc_objects -nodecount=8 "$$p/mem.pprof"; done; \
	echo "Fixture startup/teardown profiles retained: $$root"
# Focused acceptance does not replace full-run compliance reports.
fixtures-vibes-focused: build-web
	@test -n "$(FIXTURES_SPEC_ARGS)" || { echo 'FIXTURES_SPEC_ARGS is required for a focused run'; exit 2; }
	mkdir -p $(BIN_DIR)
	$(GO) build -tags fixtures_vibes -o $(BIN_DIR)/gi-fixtures-vibes ./cmd/gi
	cd references/fixtures-vibes && GI_FIXTURE_PROFILE_DIR=$(FIXTURES_RUNTIME_PROFILE_DIR) GI_FIXTURE_BIN=$(abspath $(BIN_DIR)/gi-fixtures-vibes) FIXTURES_PROFILE=$(CURDIR)/tests/fixtures-vibes/profile.json node node_modules/@playwright/test/cli.js test -c suite/playwright.config.ts $(FIXTURES_SPEC_ARGS) --reporter=line

.PHONY: fixtures-vibes-report
fixtures-vibes-report:
	$(MAKE) -C references/fixtures-vibes report PROFILE=$(CURDIR)/tests/fixtures-vibes/profile.json SHELL='$(SHELL)'

fixtures-vibes: build-web
	mkdir -p $(BIN_DIR)
	$(GO) build -tags fixtures_vibes -o $(BIN_DIR)/gi-fixtures-vibes ./cmd/gi
	GI_FIXTURE_PROFILE_DIR=$(FIXTURES_RUNTIME_PROFILE_DIR) GI_FIXTURE_BIN=$(abspath $(BIN_DIR)/gi-fixtures-vibes) $(MAKE) -C references/fixtures-vibes deps compliance PROFILE=$(CURDIR)/tests/fixtures-vibes/profile.json $(FIXTURES_VIBES_ARGS)

# ── Isolated UX test instance ───────────────────────────────────────────

test-instance-start: build
	@mkdir -p $(TEST_DIR)
	@if [ -f $(TEST_PID) ] && kill -0 $$(cat $(TEST_PID)) 2>/dev/null; then \
		pid=$$(cat $(TEST_PID)); kill $$pid 2>/dev/null || true; \
		for i in $$(seq 1 100); do kill -0 $$pid 2>/dev/null || break; sleep .1; done; \
		if kill -0 $$pid 2>/dev/null; then kill -KILL $$pid 2>/dev/null || true; fi; \
	fi
	@rm -rf $(TEST_WORKSPACE) $(TEST_DB) $(TEST_DB)-wal $(TEST_DB)-shm $(TEST_LOG) $(TEST_PID)
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
		pid=$$(cat $(TEST_PID)); kill $$pid && echo "Stopping test instance"; \
		for i in $$(seq 1 100); do kill -0 $$pid 2>/dev/null || break; sleep .1; done; \
		if kill -0 $$pid 2>/dev/null; then kill -KILL $$pid 2>/dev/null || true; fi; \
	fi
	@rm -rf $(TEST_DIR)

test-ux: playwright-browsers test-instance-start $(TESTPROFILE)
	mkdir -p $(TEST_RESULTS)
	GI_TEST_URL=http://127.0.0.1:$(TEST_PORT) $(PLAYWRIGHT) test tests/functional/ --reporter=line --output=$(TEST_RESULTS)/playwright $(PLAYWRIGHT_ARGS); \
	rc=$$?; \
	$(MAKE) --no-print-directory test-instance-stop; \
	exit $$rc

# Frozen Vibes/Tau Piclaw Classic corpus, with a separate disposable process.
UX_PARITY_PORT ?= 19091
UX_PARITY_ARGS ?=

# Real local inference checkpoints (no paid provider).
UX_LOCAL_ENV ?= GI_UX_STEER=1
UX_LOCAL_SPEC ?= tests/web-regression/queue-steer.spec.mjs
UX_LOCAL_BIN ?= bin/gi-ux-steer
UX_LOCAL_PORT ?= 19092
.PHONY: test-shared-stop-evidence
test-shared-stop-evidence: fixtures-vibes

test-ux-reconnect: build-web
	@mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_RECONNECT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/reconnect.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-status-swipes test-ux-mobile-exclusions test-ux-thoughts test-ux-direct-reply-steer
test-ux-direct-reply-steer:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_DIRECT_REPLY_STEER=1 UX_LOCAL_SPEC=tests/web-regression/direct-reply-steer.spec.mjs

test-ux-thoughts:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_THOUGHTS=1 UX_LOCAL_SPEC=tests/web-regression/thoughts.spec.mjs

test-ux-status-swipes:
	$(MAKE) test-ux-steer UX_LOCAL_ENV=GI_UX_STATUS_SWIPES=1 UX_LOCAL_SPEC=tests/web-regression/status-swipes.spec.mjs

test-ux-mobile-exclusions:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV='GI_UX_CARD_REJECTION=1 GI_UX_MOBILE_EXCLUSIONS=1' UX_LOCAL_SPEC=tests/web-regression/mobile-exclusions.spec.mjs


test-ux-compaction:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_COMPACTION=1 UX_LOCAL_SPEC=tests/web-regression/compaction.spec.mjs

test-ux-context-meter:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_METER=1 UX_LOCAL_SPEC=tests/web-regression/context-meter.spec.mjs

test-ux-context-fit:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_CONTEXT=1 UX_LOCAL_SPEC=tests/web-regression/context-fit.spec.mjs

test-ux-index-config:
	$(MAKE) --no-print-directory test-ux-steer UX_LOCAL_ENV=GI_UX_INDEX_CONFIG=1 UX_LOCAL_SPEC=tests/web-regression/workspace-index-config.spec.mjs

test-ux-shared-copy-delete: fixtures-vibes

test-ux-steer: build-web
	@mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity/queue-gates
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	@set -e; \
	PATH=$(abspath tests/ux/shell):$$PATH $(UX_LOCAL_ENV) GI_UX_LISTEN=127.0.0.1:$(UX_LOCAL_PORT) GI_UX_QUEUE_GATES=$(abspath test-results/ux-parity/queue-gates) $(UX_LOCAL_BIN) >test-results/ux-parity/steer-server.log 2>&1 & pid=$$!; \
	trap 'kill $$pid 2>/dev/null || true; wait $$pid 2>/dev/null || true' EXIT; \
	ready=0; for i in $$(seq 1 100); do kill -0 $$pid || exit 1; if curl -fsS http://127.0.0.1:$(UX_LOCAL_PORT)/api/runtime/config >/dev/null 2>&1; then ready=1; break; fi; sleep .1; done; \
	test $$ready -eq 1; \
	$(UX_LOCAL_ENV) GI_TEST_URL=http://127.0.0.1:$(UX_LOCAL_PORT) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs $(UX_LOCAL_SPEC) $(UX_PARITY_ARGS); \
	if [ -n "$(UX_LOCAL_FUNCTIONAL)" ]; then $(UX_LOCAL_ENV) GI_TEST_URL=http://127.0.0.1:$(UX_LOCAL_PORT) $(PLAYWRIGHT) test --config=playwright.config.ts --output=test-results/local-functional-artifacts $(UX_LOCAL_FUNCTIONAL); fi
test-web-adapters:
	$(BUN) test ./tests/ux/support/ && $(MAKE) -C $(GI_UI) test

.PHONY: test-web-adapters test-web-regression test-web-regression-list test-ux-parity test-ux-parity-regression
test-ux-parity: fixtures-vibes

# Gi-only race/recovery regressions; no shared-scenario accounting.
test-web-regression:
	$(MAKE) test-ux-parity-regression

test-web-regression-list:
	$(PLAYWRIGHT) test -c playwright.web-regression.config.mjs --list $(UX_PARITY_ARGS)

.PHONY: build-pane-host-fixture
build-pane-host-fixture:
	@mkdir -p test-results
	$(MAKE) -s -C $(GI_UI) deps && $(BUN) $(GI_UI)/scripts/build-pane-host-fixture.mjs

.PHONY: test-web-skills test-ux-skills
.PHONY: test-tool-activity
test-tool-activity:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/web ./internal/turn -run 'ToolActivity|ToolPreview|SessionActivity|ToolTerminal'
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/tool-activity.test.ts

.PHONY: test-ux-tool-terminal
test-ux-tool-terminal: test-tool-activity
	$(MAKE) test-ux-parity-regression UX_PARITY_ARGS='tests/web-regression/tool-activity.spec.mjs'

.PHONY: test-terminal-tool-identity test-tui-tool-timing
.PHONY: test-terminal-links
.PHONY: test-tui-session-actions
test-tui-session-actions: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-session-actions.mjs

.PHONY: test-session-actions
test-session-actions:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui ./internal/store -run 'SessionActions|SessionRename|SessionDisplayCapabilities|SessionPicker'

test-terminal-links:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui -run 'TranscriptLink|TranscriptSelection|TranscriptSearch'

.PHONY: test-tui-links
test-tui-links: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-links.mjs

.PHONY: test-tui-selection-edge
test-tui-selection-edge: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-selection-edge.mjs

test-terminal-tool-identity:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui -run 'ToolRuntime|ToolEndWithout|RenderToolEvent|BuildTranscriptRenderable'

test-tui-tool-timing: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-tool-timing.mjs

test-web-skills:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/web -run 'LoadedWebSkill|WebSkillOpen|QuickActions'

test-ux-skills:
	GI_UX_SKILLS=1 $(MAKE) test-ux-parity-regression TEST_FIXTURES_DIR=tests/ux/fixtures/skills UX_PARITY_ARGS='tests/web-regression/skills.spec.mjs'

.PHONY: test-recovery-marker test-ux-outcomes
test-recovery-marker:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/turn -run 'RecoveryMarker|StartupRecoveryRequeuesCompactingTurn'
.PHONY: test-ux-card-rejection
test-ux-card-rejection:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_CARD_REJECTION=1 GI_UX_RECOVERY_PLACEHOLDERS=1' UX_LOCAL_SPEC='tests/web-regression/card-rejection.spec.mjs tests/web-regression/recovery-placeholders.spec.mjs tests/web-regression/speech.spec.mjs' UX_LOCAL_FUNCTIONAL='tests/functional/16-card-rejection.spec.ts tests/functional/15-recovery-placeholders.spec.ts'


.PHONY: test-ux-recovery-placeholders
test-ux-recovery-placeholders:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_RECOVERY_CONTROLS=1 GI_UX_RECOVERY_PLACEHOLDERS=1' UX_LOCAL_SPEC='tests/web-regression/recovery-placeholders.spec.mjs tests/web-regression/recovery-controls.spec.mjs tests/web-regression/speech.spec.mjs' UX_LOCAL_FUNCTIONAL='tests/functional/14-recovery-controls.spec.ts tests/functional/15-recovery-placeholders.spec.ts'


.PHONY: test-ux-recovery-controls
test-ux-recovery-controls:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_RECOVERY_CONTROLS=1' UX_LOCAL_SPEC='tests/web-regression/recovery-controls.spec.mjs tests/web-regression/message-copy.spec.mjs tests/web-regression/message-delete.spec.mjs' UX_LOCAL_FUNCTIONAL=tests/functional/14-recovery-controls.spec.ts


test-ux-outcomes:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_OUTCOMES=1' UX_LOCAL_SPEC='tests/web-regression/outcomes.spec.mjs tests/web-regression/message-copy.spec.mjs tests/web-regression/speech.spec.mjs' UX_LOCAL_FUNCTIONAL=tests/functional/13-outcomes.spec.ts


.PHONY: test-ux-card-identity
test-ux-card-identity:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_CARD_IDENTITY=1' UX_LOCAL_SPEC='tests/web-regression/card-submission-identity.spec.mjs'

.PHONY: test-ux-btw-mount
test-ux-btw-mount: build-web
	@mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_BTW_MOUNT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs

.PHONY: test-ux-widget-persisted
test-ux-widget-persisted:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_WIDGETS=1' UX_LOCAL_SPEC='tests/web-regression/widget-persisted.spec.mjs'

.PHONY: test-ux-links
test-ux-links:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_LINKS=1' UX_LOCAL_SPEC='tests/web-regression/remote-links.spec.mjs tests/web-regression/rendering.spec.mjs tests/web-regression/lightbox.spec.mjs' UX_LOCAL_FUNCTIONAL=tests/functional/12-remote-links.spec.ts


.PHONY: test-ux-speech-contract
test-ux-speech-contract:
	$(MAKE) test-ux-steer UX_LOCAL_ENV='GI_UX_SPEECH=1' UX_LOCAL_SPEC='tests/web-regression/speech-contract.spec.mjs tests/web-regression/speech.spec.mjs tests/web-regression/message-copy.spec.mjs tests/web-regression/rendering.spec.mjs'


.PHONY: capture-gi-chat-baseline
capture-gi-chat-baseline:
	$(BUN) tests/ux/oracle/gi-chat-baseline-probe.mjs

.PHONY: test-pi-idle-escape-oracle test-pi-tool-boundary-steer test-pi-direct-reply-steer test-pi-steering-modes test-pi-abort-steering test-pi-steering-mode-switch test-pi-preparation-steer

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

.PHONY: test-ux-queue-return

test-ux-queue-return:
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/drafts.test.ts tests/unit/queue-return.test.ts
	$(MAKE) test-ux-parity-regression UX_PARITY_ARGS='tests/web-regression/queue-return.spec.mjs'

.PHONY: diagnose-webkit-unload
diagnose-webkit-unload:
	$(BUN) tests/ux/oracle/webkit-unload-diagnostic.mjs

test-ux-parity-regression:
	@mkdir -p test-results/ux-parity/queue-gates
	PATH="$(abspath tests/ux/shell):$$PATH" GI_UX_QUEUE_GATES="$(abspath test-results/ux-parity/queue-gates)" $(MAKE) --no-print-directory test-instance-start TEST_PORT=$(UX_PARITY_PORT) TEST_DIR=.gi-ux-parity TEST_ENABLED_MODELS='["test-model","bootstrap","test/unavailable-model"]'
	@trap '$(MAKE) --no-print-directory test-instance-stop TEST_DIR=.gi-ux-parity' EXIT; \
		$(MAKE) --no-print-directory build-pane-host-fixture || exit 1; \
		GI_TEST_URL=http://127.0.0.1:$(UX_PARITY_PORT) $(PLAYWRIGHT) test -c playwright.web-regression.config.mjs $(UX_PARITY_ARGS)

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
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui -run 'TestPiEnter|TestPiFollowUp'
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/turn -run 'TestTUIComposerFollowUp|TestTUIComposerRejectsUnknownDelivery|TestTUIComposerSubmit'

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
	$(TESTPROFILE) gotest ./internal/tui -run '^$$' -bench '^BenchmarkEditorLayout$$' -benchmem -benchtime=100ms

.PHONY: bench-tui-transcript-frame
bench-tui-transcript-frame:
	$(TESTPROFILE) gotest ./internal/tui -run '^$$' -bench '^BenchmarkTranscriptFrame$$' -benchmem -benchtime=1s $(BENCH_ARGS)

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
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui -run DurableDraft

.PHONY: test-tui-text-journal test-concurrent-session-submit
test-concurrent-session-submit:
	$(TESTPROFILE) gotest $(RACE) -count=10 ./internal/turn -run '^TestConcurrentSubmitDifferentSessionsRunsConcurrently$$'

test-tui-text-journal:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/turn -run 'TUITextDraft|TUIComposerDraft|TUIComposerSubmit'

.PHONY: test-tui-media-journal
test-tui-media-journal:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/tui -run 'TUIMediaDraft|PendingMedia|AttachCommand|PasteImage'

.PHONY: test-held-retry
test-held-retry:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/store ./internal/turn ./internal/tui -run 'TUIRetry|RetryHeld|HeldRetry|HoldAndResolve|HoldResolution|SkipHeld'

.PHONY: test-tui-unselected-model test-tui-submit-guards
test-tui-submit-guards:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui -run 'TUIUnselectedModel|PendingMediaCommandsLimitsAndNoModelDraft'

test-tui-unselected-model: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-unselected-model.mjs

.PHONY: test-tui-retry-commands
test-tui-retry-commands: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-retry-commands.mjs

.PHONY: test-tui-queue-commands test-terminal-queue
test-terminal-queue:
	$(TESTPROFILE) gotest $(RACE) -count=3 ./internal/tui ./internal/store ./internal/turn -run 'TUIQueue|QueueSteer|QueuedTurn|QueueOrder'

test-tui-queue-commands: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-queue-commands.mjs

test-tui-pending-media: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-pending-media.mjs

.PHONY: test-tui-session-picker
test-tui-session-picker: build
	GI_TUI_BIN=$(abspath $(BIN)) $(BUN) scripts/test-tui-session-picker.mjs

test-tui-smoke: build $(TESTPROFILE)
	chmod +x scripts/test-tui-smoke.sh
	ARTIFACT_DIR=$(abspath $(TEST_RESULTS))/tui-smoke TEST_DIR=$(abspath $(TUI_TEST_DIR)) scripts/test-tui-smoke.sh

test-tui-gherkin: build test-tui-markdown test-tui-inline-prose test-tui-gherkin-features

# Feature files only (no markdown/prose prerequisites). FEATURE_DIR narrows
# the run, e.g. `make test-tui-gherkin-features FEATURE_DIR=/tmp/one-feature`.
.PHONY: test-tui-gherkin-features
test-tui-gherkin-features: build $(TESTPROFILE)
	chmod +x scripts/test-tui-gherkin.sh
	ARTIFACT_DIR=$(abspath $(TEST_RESULTS))/tui-gherkin TEST_DIR=$(abspath $(TUI_TEST_DIR))-gherkin $(if $(FEATURE_DIR),FEATURE_DIR=$(abspath $(FEATURE_DIR))) scripts/test-tui-gherkin.sh

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
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/context-usage.test.ts tests/unit/compaction-state.test.ts

.PHONY: test-notification-helpers test-ux-notifications
test-notification-helpers:
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/notifications.test.ts

test-ux-notifications: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_NOTIFICATIONS=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/notifications.spec.mjs

.PHONY: test-voice-input-helpers test-ux-voice-input
test-voice-input-helpers:
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/voice-input.test.ts tests/unit/voice-adapter.test.ts

test-ux-voice-input: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_VOICE_INPUT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/voice-input.spec.mjs

.PHONY: test-ux-session-panel test-session-panel-helpers
test-session-panel-helpers:
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/session-panel.test.ts

test-ux-session-panel: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_SESSION_PANEL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/session-panel.spec.mjs

.PHONY: test-ux-model-panel test-model-panel-helpers
test-model-panel-helpers:
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/model-panel.test.ts tests/unit/model-picker.test.ts tests/unit/model-accessibility.test.ts tests/unit/model-thinking.test.ts

.PHONY: test-ux-compose-escape

test-ux-compose-escape: build-web
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_COMPOSE_ESCAPE=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/compose-escape.spec.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-picker-thinking
test-ux-picker-thinking: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PICKER_THINKING=1 GI_UX_THINKING=1 GI_UX_THINKING_DEFAULT=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/picker-thinking.spec.mjs $(UX_PARITY_ARGS)

test-ux-model-panel: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_MODEL_PANEL=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/model-panel.spec.mjs

.PHONY: test-ux-compose-surface test-compose-surface-helpers
test-compose-surface-helpers:
	$(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/compose-surface.test.ts tests/unit/accent-contrast.test.ts tests/unit/theme-text-contrast.test.ts

test-ux-compose-surface: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN))
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_COMPOSE_SURFACE=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config=playwright.web-regression.config.mjs tests/web-regression/compose-surface.spec.mjs

.PHONY: test-ux-workspace-tabs
test-ux-workspace-tabs:
	$(MAKE) test-ux-parity-regression UX_PARITY_ARGS='tests/web-regression/workspace-tabs.spec.mjs'
	cp test-results/ux-parity/results.json test-results/ux-parity/workspace-tabs-results.json

.PHONY: test-ux-slash
test-ux-slash: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_SLASH=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(BUN) x playwright test --config=playwright.web-regression.config.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-picker-geometry
test-ux-picker-geometry: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PICKER_GEOMETRY=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(BUN) x playwright test --config=playwright.web-regression.config.mjs $(UX_PARITY_ARGS)

.PHONY: test-ux-journey
test-ux-journey: build-web
	mkdir -p $(dir $(UX_LOCAL_BIN)) test-results/ux-parity
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_JOURNEY=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(BUN) x playwright test --config=playwright.web-regression.config.mjs $(UX_PARITY_ARGS)

.PHONY: test-browser-auth test-ux-auth
test-browser-auth:
	$(TESTPROFILE) gotest ./internal/web ./internal/auth -run 'Test(Browser|ProtectedEndpoint|Auth)' -count=1

.PHONY: test-passkey-criteria test-ux-passkeys
.PHONY: test-passkey-cancel-repeat
test-passkey-cancel-repeat: build-web test-passkey-criteria
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PASSKEYS=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config playwright.web-regression.config.mjs tests/web-regression/passkeys.spec.mjs --project=chromium-phone --project=chromium-tablet --project=chromium-desktop --grep 'Settings enrolls two passkeys|passkey login cancellation owns' --repeat-each=3

.PHONY: test-passkey-login-boundary
test-passkey-login-boundary: build-web test-passkey-criteria
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PASSKEYS=1 GI_UX_SERVER_BIN=$(abspath $(UX_LOCAL_BIN)) $(PLAYWRIGHT) test --config playwright.web-regression.config.mjs tests/web-regression/passkeys.spec.mjs --project=chromium-phone --project=chromium-tablet --project=chromium-desktop --grep 'passkey login cancellation owns'

test-passkey-criteria:
	$(BUN) test tests/ux/support/passkey-criteria.test.ts && ($(MAKE) -s -C $(GI_UI) deps && cd $(GI_UI) && $(BUN) test tests/unit/passkey-lifecycle.test.ts)

test-ux-passkeys: build-web test-passkey-criteria
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_PASSKEYS=1 GI_UX_SERVER_BIN=$(UX_LOCAL_BIN) $(BUN) x playwright test --config playwright.web-regression.config.mjs tests/web-regression/passkeys.spec.mjs --project=chromium-phone --project=chromium-tablet --project=chromium-desktop

test-ux-auth: build-web
	$(GO) build -o $(UX_LOCAL_BIN) ./tests/ux/server
	GI_UX_AUTH=1 GI_UX_SERVER_BIN=$(UX_LOCAL_BIN) $(BUN) x playwright test --config playwright.web-regression.config.mjs tests/web-regression/auth.spec.mjs

.PHONY: test-browser-auth-race
test-browser-auth-race:
	$(TESTPROFILE) gotest $(RACE) ./internal/web ./internal/auth -count=3

.PHONY: test-auth-state
test-auth-state:
	$(TESTPROFILE) gotest $(RACE) ./internal/auth -count=10

.PHONY: test-selected-runtimes test-native-joker-compiler test-joker-wasm-engines
test-joker-wasm-engines: $(TESTPROFILE)
	JOKER_WASM_ENGINE=compiler $(TESTPROFILE) gotest $(RACE) -run 'TestEmbeddedJoker(NativeWASM|WASMEngine|CheckedWASM)' -count=3 ./internal/scripting
	JOKER_WASM_ENGINE=interpreter $(TESTPROFILE) gotest $(RACE) -run 'TestEmbeddedJoker(NativeWASM|WASMEngine|CheckedWASM)' -count=3 ./internal/scripting

test-native-joker-compiler: $(TESTPROFILE)
	$(TESTPROFILE) gotest -run 'TestWasmCompileRetainsModuleBytes|TestWasmRawIntObjectPromotesOutsideNativeRange|TestWasmRawIntRejectsOutOfRangeIndex|TestWasmExecRawIntegerResultUsesNativeRange' -count=1 github.com/rcarmo/go-joker/v42/core

test-selected-runtimes: $(TESTPROFILE)
	$(TESTPROFILE) gotest $(RACE) -run 'TestQuickJS|TestScriptTool(Selects|RuntimeChoice)|TestJavaScriptRuntime|TestEmbeddedJoker(NativeWASM|WASMEngine|CheckedWASM)|TestListRuntimeOptionsMatchesPiCopilot|TestExecuteEmbeddedJoker' -count=3 ./internal/scripting ./internal/tools ./internal/config ./internal/inference

# Compare actual tmux cells against ANSI frame diffs while tables enter/leave
# the viewport. No provider, sqlite3 CLI, Bun, or running Gi instance required.
.PHONY: test-tui-table-scroll
test-tui-table-scroll:
	GI_TABLE_SCROLL_PTY=1 $(GO) test -count=1 ./internal/tui -run '^TestMarkdownTableScrollTerminal$$' -v

# Run interactively in the affected Ghostty window, without tmux.
.PHONY: probe-tui-ghostty
probe-tui-ghostty:
	$(GO) run ./tests/tui-ghostty-probe

.PHONY: fmt-tool-syntax test-tui-tool-syntax
fmt-tool-syntax:
	$(GO) fmt ./internal/tui

test-tui-tool-syntax:
	$(TESTPROFILE) gotest -run 'Test(FileTool|PiTool|ToolSyntax)' ./internal/tui

.PHONY: test-tui-tool-syntax-pty
test-tui-tool-syntax-pty:
	@mkdir -p $(BIN_DIR)
	$(GO) test -c -o $(BIN_DIR)/gi-tool-syntax-test ./internal/tui
	$(BUN) scripts/test-tui-tool-syntax.mjs

.PHONY: fmt-pi-scroll test-pi-scroll test-go-tui-runtime
fmt-pi-scroll:
	$(GO) fmt ./internal/tui
	cd third_party/go-tui && $(GO) fmt .

test-pi-scroll:
	$(TESTPROFILE) gotest $(RACE) ./internal/tui -run 'Test(PiRow|Wheel|TerminalWheel|MarkdownTableScroll)' -count=1

test-go-tui-runtime:
	cd third_party/go-tui && $(TESTPROFILE) gotest $(RACE) . ./internal/... -count=1

# Native controlling-terminal colour negotiation; Bun, no npm dependencies.
.PHONY: test-tui-theme-pty
test-tui-theme-pty:
	mkdir -p $(BIN_DIR)
	$(GO) test -c -o $(BIN_DIR)/gi-theme-test ./internal/tui
	GI_THEME_TEST_BIN=$(abspath $(BIN_DIR)/gi-theme-test) $(BUN) scripts/test-tui-theme.mjs

# Reproducible CPU/allocation profiles; artifacts stay on disk.
TABLE_BENCH ?= BenchmarkComplexMarkdownTable
TABLE_BENCH_TIME ?= 300ms
.PHONY: bench-tui-complex-tables profile-tui-complex-tables profile-tui-complex-tables-report fmt-tui-complex-tables test-tui-complex-tables
bench-tui-complex-tables:
	$(TESTPROFILE) gotest ./internal/tui -run '^$$' -bench '$(TABLE_BENCH)' -benchmem -benchtime=$(TABLE_BENCH_TIME) $(BENCH_ARGS)
profile-tui-complex-tables:
	@mkdir -p $(TEST_RESULTS)/tui-table-perf
	$(GO) test ./internal/tui -run '^$$' -bench '$(TABLE_BENCH)' -benchmem -benchtime=$(TABLE_BENCH_TIME) -o $(TEST_RESULTS)/tui-table-perf/tui.test -cpuprofile=$(TEST_RESULTS)/tui-table-perf/cpu.pprof -memprofile=$(TEST_RESULTS)/tui-table-perf/mem.pprof $(BENCH_ARGS)
	$(MAKE) profile-tui-complex-tables-report
profile-tui-complex-tables-report:
	$(GO) tool pprof -top $(TEST_RESULTS)/tui-table-perf/cpu.pprof
	$(GO) tool pprof -top -alloc_space $(TEST_RESULTS)/tui-table-perf/mem.pprof
fmt-tui-complex-tables:
	$(GO) fmt ./internal/tui
test-tui-complex-tables:
	$(TESTPROFILE) gotest ./internal/tui -count=1 -run 'TestComplexMarkdownTable|TestMarkdownTable|TestTranscriptWindow|TestTranscriptBlocksMemo'

.PHONY: test-unicode-perf bench-unicode-perf fmt-unicode-perf
test-unicode-perf:
	cd third_party/go-tui && $(TESTPROFILE) gotest . -run 'TestUnicodeFast|TestTextMeasurementCache|TestUnwrappedClusters' -count=1
bench-unicode-perf:
	cd third_party/go-tui && $(TESTPROFILE) gotest . -run '^$$' -bench '^BenchmarkUnicodeWidths$$' -benchmem -benchtime=$(TABLE_BENCH_TIME) $(BENCH_ARGS)
fmt-unicode-perf:
	cd third_party/go-tui && $(GO) fmt .

.PHONY: test-tui-complex-tables-pty
test-tui-complex-tables-pty:
	GI_COMPLEX_TABLE_PTY=1 $(TESTPROFILE) gotest ./internal/tui -count=1 -v -run '^TestComplexMarkdownTableTerminal$$'
.PHONY: fmt-tool-paths test-tool-paths
fmt-tool-paths:
	$(GO) fmt ./internal/tools
test-tool-paths:
	$(TESTPROFILE) gotest $(RACE) ./internal/tools -count=1

# Backend-only widget persistence, lookup/auth and turn/SSE tests.
.PHONY: test-dashboard-widgets
test-dashboard-widgets:
	$(TESTPROFILE) gotest $(RACE) ./internal/store ./internal/tools ./internal/turn ./internal/web -run TestDashboardWidget

# Backend-only Plan parser, persistence, tool, API and change events.
.PHONY: test-session-plan
test-session-plan:
	$(TESTPROFILE) gotest $(RACE) ./internal/plan ./internal/store ./internal/tools ./internal/turn ./internal/web -run Plan
