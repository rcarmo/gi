PORT ?= 8090
BIND ?= 0.0.0.0
MODEL ?= github-copilot/gpt-5-mini
WORKSPACE ?= /workspace
RUN_DIR ?= .gi-run
BIN_DIR ?= bin
BIN ?= $(BIN_DIR)/gi
DB ?= $(RUN_DIR)/gi.db
LOG ?= $(RUN_DIR)/gi.log
PID ?= $(RUN_DIR)/gi.pid
LISTEN ?=

TEST_PORT ?= 19090
TEST_DIR ?= .gi-test
TEST_DB ?= $(TEST_DIR)/gi.db
TEST_LOG ?= $(TEST_DIR)/gi.log
TEST_PID ?= $(TEST_DIR)/gi.pid
TEST_WORKSPACE ?= $(TEST_DIR)/workspace

.PHONY: build-web build test vet bun-checks start stop restart status logs run clean test-instance-start test-instance-stop test-ux

build-web:
	bun run build:web

build: build-web
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/gi
	go build -o $(BIN_DIR)/gi-tui ./cmd/gi-tui

test:
	go test ./...

vet:
	go vet ./...

bun-checks:
	bun run check:hook-tdz

run: build
	$(BIN) $(if $(LISTEN),-listen $(LISTEN),-bind $(BIND) -port $(PORT)) -model $(MODEL) -db $(DB) -workspace $(WORKSPACE)

start: build
	mkdir -p $(RUN_DIR)
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		echo "Gi already running with PID $$(cat $(PID))"; \
		exit 0; \
	fi
	$(abspath $(BIN)) $(if $(LISTEN),-listen $(LISTEN),-bind $(BIND) -port $(PORT)) \
		-model $(MODEL) \
		-db $(abspath $(DB)) \
		-workspace $(WORKSPACE) \
		-log-file $(abspath $(LOG)) \
		-pid-file $(abspath $(PID)) \
		>/dev/null 2>&1 </dev/null &
	@sleep 2
	@$(MAKE) --no-print-directory status BIND=$(BIND) PORT=$(PORT)

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
		echo "Gi running on $(if $(LISTEN),$(LISTEN),$(BIND):$(PORT)) with PID $$(cat $(PID))"; \
		ss -ltnp | grep ':$(PORT)' || true; \
	else \
		echo "Gi is not running"; \
		exit 1; \
	fi

logs:
	@mkdir -p $(RUN_DIR)
	tail -f $(LOG)

clean:
	rm -rf $(RUN_DIR) $(BIN_DIR) $(TEST_DIR) test-results/

# ── Isolated test instance ──────────────────────────────────────────────

test-instance-start: build
	@mkdir -p $(TEST_DIR)
	@rm -rf $(TEST_WORKSPACE) $(TEST_DB) $(TEST_LOG)
	@mkdir -p $(TEST_WORKSPACE)/.piclaw $(TEST_WORKSPACE)/.pi
	@echo '{"assistant":{"assistantName":"Gi Test"},"user":{"userName":"Test User"}}' > $(TEST_WORKSPACE)/.piclaw/config.json
	@echo '{"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":["test-model"]}' > $(TEST_WORKSPACE)/.pi/settings.json
	@if [ -f $(TEST_PID) ] && kill -0 $$(cat $(TEST_PID)) 2>/dev/null; then \
		kill $$(cat $(TEST_PID)) 2>/dev/null || true; \
		sleep 1; \
	fi
	$(abspath $(BIN)) -bind 127.0.0.1 -port $(TEST_PORT) \
		-model test-model \
		-db $(abspath $(TEST_DB)) \
		-workspace $(abspath $(TEST_WORKSPACE)) \
		-log-file $(abspath $(TEST_LOG)) \
		-pid-file $(abspath $(TEST_PID)) \
		>/dev/null 2>&1 </dev/null &
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
	GI_TEST_URL=http://127.0.0.1:$(TEST_PORT) bunx playwright test tests/base-ux.spec.ts --reporter=line; \
	rc=$$?; \
	$(MAKE) --no-print-directory test-instance-stop; \
	exit $$rc
