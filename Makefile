SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

# ── Tool commands ───────────────────────────────────────────────────────

GO ?= go
BUN ?= bun
PLAYWRIGHT ?= bunx playwright

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
TEST_PI_SETTINGS_JSON = {"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":["test-model"]}

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
	test-instance-start test-instance-stop test-ux test-tui-smoke test-tui-gherkin \
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

vet:
	$(GO) vet ./...

bun-checks:
	$(BUN) run check:hook-tdz

check: test vet build-web bun-checks test-ux

# ── Isolated UX test instance ───────────────────────────────────────────

test-instance-start: build
	@mkdir -p $(TEST_DIR)
	@rm -rf $(TEST_WORKSPACE) $(TEST_DB) $(TEST_LOG)
	@mkdir -p $(TEST_WORKSPACE)/.piclaw $(TEST_WORKSPACE)/.pi
	@printf '%s\n' '$(TEST_PICLAW_CONFIG_JSON)' > $(TEST_WORKSPACE)/.piclaw/config.json
	@printf '%s\n' '$(TEST_PI_SETTINGS_JSON)' > $(TEST_WORKSPACE)/.pi/settings.json
	@if [ -f $(TEST_PID) ] && kill -0 $$(cat $(TEST_PID)) 2>/dev/null; then \
		kill $$(cat $(TEST_PID)) 2>/dev/null || true; \
		sleep 1; \
	fi
	$(abspath $(BIN)) $(TEST_SERVER_ARGS) >/dev/null 2>&1 </dev/null &
	@sleep 2
	@if [ -f $(TEST_PID) ] && kill -0 $$(cat $(TEST_PID)) 2>/dev/null; then \
		echo "Test instance running on 127.0.0.1:$(TEST_PORT) with PID $$(cat $(TEST_PID))"; \
	else \
		echo "Test instance failed to start"; cat $(TEST_LOG) 2>/dev/null; exit 1; \
	fi

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
