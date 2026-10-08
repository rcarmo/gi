.DEFAULT_GOAL := help

# Resolve once, before redirecting child temporary paths. This repository owns
# the resolver; CI does not depend on a host /workspace helper.
GO ?= go
export GOTOOLCHAIN := go1.27.1
export PROJECT_NAME := gi
export PROJECT_ORIGINAL_TMPDIR := $(if $(filter undefined,$(origin PROJECT_ORIGINAL_TMPDIR)),$(TMPDIR),$(PROJECT_ORIGINAL_TMPDIR))
_PROJECT_TMP_RESOLVED := $(shell $(if $(filter-out undefined,$(origin PROJECT_TMP_BASE)),PROJECT_TMP_BASE='$(PROJECT_TMP_BASE)') $(if $(filter-out undefined,$(origin PROJECT_TMP_ROOT)),PROJECT_TMP_ROOT='$(PROJECT_TMP_ROOT)') PROJECT_NAME=gi PROJECT_ORIGINAL_TMPDIR='$(PROJECT_ORIGINAL_TMPDIR)' CI='$(CI)' RUNNER_TEMP='$(RUNNER_TEMP)' bash scripts/project-tmp.sh)
ifeq ($(_PROJECT_TMP_RESOLVED),)
$(error Cannot resolve a usable project-owned temporary root)
endif
export PROJECT_TMP_ROOT := $(_PROJECT_TMP_RESOLVED)
export GI_WORKTREE := $(notdir $(CURDIR))
ifndef GI_TEST_RUN_ID
GI_TEST_RUN_ID := $(shell date -u +%Y%m%dT%H%M%SZ)-$(shell echo $$$$)
endif
export GI_TEST_RUN_ID
ifndef GI_TEST_RUN_ROOT
GI_TEST_RUN_ROOT := $(PROJECT_TMP_ROOT)/runs/tests/$(GI_WORKTREE)/$(GI_TEST_RUN_ID)
endif
export GI_TEST_RUN_ROOT
_TEST_PATH_VALID := $(shell PROJECT_NAME=gi PROJECT_TMP_ROOT='$(PROJECT_TMP_ROOT)' GI_TEST_RUN_ROOT='$(GI_TEST_RUN_ROOT)' bash -c 'source scripts/project-test-env.sh && echo ok')
ifneq ($(_TEST_PATH_VALID),ok)
$(error Unsafe project test run root)
endif
export TMPDIR := $(GI_TEST_RUN_ROOT)/tmp
export TMP := $(TMPDIR)
export TEMP := $(TMPDIR)
export BUN_INSTALL_CACHE_DIR := $(PROJECT_TMP_ROOT)/cache/bun
export npm_config_cache := $(PROJECT_TMP_ROOT)/cache/npm
export XDG_CACHE_HOME := $(PROJECT_TMP_ROOT)/cache/xdg
export PLAYWRIGHT_BROWSERS_PATH ?= $(PROJECT_TMP_ROOT)/cache/ms-playwright
export PROFILING ?= 0
export PROFILE_KEEP ?= 0

# ── CPU throttling ──────────────────────────────────────────────────────
# Every recipe (builds, tests, the dev server) runs niced and pinned to a
# CPU subset so the machine stays usable and runs are reproducible. Tune per
# invocation, e.g. `make test CPU_SET=0-3 CPU_PROCS=4` or `CPU_NICE=0`.
#   CPU_NICE   nice level for every recipe
#   CPU_SET    taskset CPU list every recipe is pinned to (skipped where
#              taskset is unavailable, e.g. macOS, which has no CPU pinning)
#   CPU_PROCS  GOMAXPROCS and Go build parallelism (-p)
# Make itself is .NOTPARALLEL, and test* targets run one Go package at a
# time (-p=1), so test suites always execute sequentially.
CPU_NICE ?= 10
CPU_SET ?= 0-1
CPU_PROCS ?= 2
.NOTPARALLEL:
# Arguments go in SHELL itself, not .SHELLFLAGS: make 3.81 (macOS's
# /usr/bin/make) ignores .SHELLFLAGS and runs `$(SHELL) -c '<recipe>'`.
TASKSET := $(shell command -v taskset 2>/dev/null)
SHELL := /usr/bin/env nice -n $(CPU_NICE) $(if $(TASKSET),$(TASKSET) -c $(CPU_SET)) bash
export GOMAXPROCS := $(CPU_PROCS)
export GOFLAGS += -p=$(CPU_PROCS)
test%: export GOFLAGS := $(filter-out -p=%,$(GOFLAGS)) -p=1

# ── Build cache ─────────────────────────────────────────────────────────
# Cache and compiler scratch are project-owned. Do not auto-trim shared caches
# at parse time: another worktree may have an active build.
export GOCACHE := $(PROJECT_TMP_ROOT)/cache/go-build
export GOMODCACHE := $(PROJECT_TMP_ROOT)/cache/go-mod
export GOTMPDIR := $(TMPDIR)/go
_TMP_INIT := $(shell mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)" "$(BUN_INSTALL_CACHE_DIR)" "$(npm_config_cache)" "$(XDG_CACHE_HOME)")

# The race detector needs a ThreadSanitizer-compatible address layout; some
# kernels (e.g. 39/42-bit arm64 VMs) lack it. Probe once and drop -race there.
ifndef RACE
RACE := $(shell f='$(PROJECT_TMP_ROOT)/cache/go-race-1.27.1'; [ -f $$f.ok ] && cat $$f.ok || { printf 'package main\nfunc main(){}\n' > $$f.go; if timeout 60 env GOTOOLCHAIN=go1.27.1 GOCACHE='$(GOCACHE)' GOMODCACHE='$(GOMODCACHE)' GOTMPDIR='$(GOTMPDIR)' TMPDIR='$(TMPDIR)' TMP='$(TMP)' TEMP='$(TEMP)' $(GO) run -race $$f.go >/dev/null 2>&1; then echo -race; fi | tee $$f.ok; rm -f $$f.go; })
endif

# ── Tool commands ───────────────────────────────────────────────────────

GO ?= go
BUN ?= bun
# Web front-end: owned in rcarmo/fixtures-vibes (ui/classic), consumed only through this submodule.
GI_UI := references/fixtures-vibes/ui/classic
PLAYWRIGHT ?= scripts/run-playwright.sh

# ── Runtime defaults ────────────────────────────────────────────────────

PORT ?= 8090
BIND ?= 0.0.0.0
LISTEN ?=
MODEL ?= github-copilot/gpt-5-mini
WORKSPACE ?= /workspace

# ── Local paths ─────────────────────────────────────────────────────────

RUN_DIR ?= .gi-run
BIN_DIR ?= $(PROJECT_TMP_ROOT)/build/$(GI_WORKTREE)
BIN ?= $(BIN_DIR)/gi
DB ?= $(RUN_DIR)/gi.db
LOG ?= $(RUN_DIR)/gi.log
PID ?= $(RUN_DIR)/gi.pid

TEST_PORT ?= 19090
TEST_DIR ?= $(GI_TEST_RUN_ROOT)/instance
TEST_DB ?= $(TEST_DIR)/gi.db
TEST_LOG ?= $(TEST_DIR)/gi.log
TEST_PID ?= $(TEST_DIR)/gi.pid
TEST_WORKSPACE ?= $(TEST_DIR)/workspace
export TEST_RESULTS ?= $(GI_TEST_RUN_ROOT)/results
TUI_TEST_DIR ?= $(GI_TEST_RUN_ROOT)/tui

# ── Derived arguments and data ──────────────────────────────────────────

SERVER_LISTEN_ARGS = -web $(if $(LISTEN),-listen $(LISTEN),-bind $(BIND) -port $(PORT))
SERVER_RUN_ARGS = $(SERVER_LISTEN_ARGS) -model $(MODEL) -db $(DB) -workspace $(WORKSPACE)
SERVER_DAEMON_ARGS = $(SERVER_LISTEN_ARGS) -model $(MODEL) -db $(abspath $(DB)) -workspace $(WORKSPACE) -log-file $(abspath $(LOG)) -pid-file $(abspath $(PID))
SERVER_STATUS_ADDR = $(if $(LISTEN),$(LISTEN),$(BIND):$(PORT))
TEST_SERVER_ARGS = -web -bind 127.0.0.1 -port $(TEST_PORT) -model test-model -db $(abspath $(TEST_DB)) -workspace $(abspath $(TEST_WORKSPACE)) -log-file $(abspath $(TEST_LOG)) -pid-file $(abspath $(TEST_PID))
TEST_PICLAW_CONFIG_JSON = {"assistant":{"assistantName":"Gi Test"},"user":{"userName":"Test User"}}
TEST_ENABLED_MODELS ?= ["test-model"]
TEST_PI_SETTINGS_JSON = {"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":$(TEST_ENABLED_MODELS),"agents":{"list":[{"id":"web","name":"Gi Test","default":true,"model":"test-model"}]}}

# ── Helper macros ───────────────────────────────────────────────────────

define require-command
	@command -v $(1) >/dev/null || { echo "$(2)"; exit 1; }
endef

# ── Opt-in pre-release profiling ────────────────────────────────────────────
TEST_PKGS ?= ./...
TEST_RUN ?=
TEST_PROFILE_DIR ?= $(PROJECT_TMP_ROOT)/runs/profiling/$(GI_WORKTREE)
export GI_TEST_PROFILE_DIR := $(TEST_PROFILE_DIR)
export GO
TESTPROFILE := $(abspath $(BIN_DIR)/testprofile)
_PROFILE_SELECTION = $(strip $(if $(filter-out ./...,$(TEST_PKGS)),pkgs=$(TEST_PKGS)) $(if $(TEST_RUN),run=$(TEST_RUN)) $(if $(PLAYWRIGHT_ARGS),playwright=$(PLAYWRIGHT_ARGS)) $(if $(FIXTURES_SPEC_ARGS),fixtures=$(FIXTURES_SPEC_ARGS)) $(if $(UX_PARITY_ARGS),ux=$(UX_PARITY_ARGS)) $(if $(FEATURE_DIR),features=$(FEATURE_DIR)))
# Indirection keeps make -n from executing the profiling wrapper as a
# recursive-make recipe and recording a dry run as a passing baseline.
PROFILE_MAKE := $(MAKE)

$(TESTPROFILE): $(wildcard scripts/testprofile/*.go)
	mkdir -p $(BIN_DIR)
	$(GO) build -o $@ ./scripts/testprofile

# Wrap the whole lifecycle, not just the test command. The recursive make
# retains CPU limits and command-line overrides; nested suites execute once.
_PROFILE_GOALS := $(if $(filter 1,$(PROFILING)),$(filter-out test-instance-start test-instance-stop test-web-regression-list,$(filter test% fixtures-vibes% bench% profile-tui-complex-tables check,$(MAKECMDGOALS))))
ifneq ($(_PROFILE_GOALS),)
ifeq ($(GI_TEST_PROFILE_ACTIVE),)
# Dispatch all requested goals so mixed invocations (build test) still work.
.PHONY: $(MAKECMDGOALS)
$(MAKECMDGOALS): $(TESTPROFILE)
	$(TESTPROFILE) run -name '$@' $(if $(_PROFILE_SELECTION),-key '$(_PROFILE_SELECTION)') -- $(PROFILE_MAKE) --no-print-directory GI_TEST_PROFILE_ACTIVE=1 $@
else
include scripts/test-targets.mk
endif
else
include scripts/test-targets.mk
endif

.PHONY: test-env
# Shell-quoted configuration for direct CI commands; resolve before changing TMPDIR.
test-env:
	@printf 'export %s=%q\n' PROJECT_NAME gi PROJECT_TMP_ROOT '$(PROJECT_TMP_ROOT)' PROJECT_ORIGINAL_TMPDIR '$(PROJECT_ORIGINAL_TMPDIR)' GI_TEST_RUN_ROOT '$(GI_TEST_RUN_ROOT)' GI_TEST_RUN_ID '$(GI_TEST_RUN_ID)' GI_WORKTREE '$(GI_WORKTREE)' TMPDIR '$(TMPDIR)' TMP '$(TMP)' TEMP '$(TEMP)' GOTMPDIR '$(GOTMPDIR)' GOCACHE '$(GOCACHE)' GOMODCACHE '$(GOMODCACHE)' BUN_INSTALL_CACHE_DIR '$(BUN_INSTALL_CACHE_DIR)' npm_config_cache '$(npm_config_cache)' XDG_CACHE_HOME '$(XDG_CACHE_HOME)' PLAYWRIGHT_BROWSERS_PATH '$(PLAYWRIGHT_BROWSERS_PATH)' TEST_RESULTS '$(TEST_RESULTS)'
