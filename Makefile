PROJECT_NAME=build-environment-detector
PACKAGE_NAME:=github.com/MatousJobanek/$(PROJECT_NAME)
CUR_DIR=$(shell pwd)
TMP_PATH=$(CUR_DIR)/tmp
INSTALL_PREFIX=$(CUR_DIR)/bin
VENDOR_DIR=vendor
SOURCE_DIR ?= .
SOURCES := $(shell find $(SOURCE_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)

# This pattern excludes the listed folders while running tests
TEST_PKGS_EXCLUDE_PATTERN = "vendor|app|tool\/environment-detector-cli|design|client|test"

# This is a fix for a non-existing user in passwd file when running in a docker
# container and trying to clone repos of dependencies
GIT_COMMITTER_NAME ?= "user"
GIT_COMMITTER_EMAIL ?= "user@example.com"
export GIT_COMMITTER_NAME
export GIT_COMMITTER_EMAIL

COMMIT=$(shell git rev-parse HEAD 2>/dev/null)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
COMMIT := $(COMMIT)-dirty
endif
BUILD_TIME=`date -u '+%Y-%m-%dT%H:%M:%SZ'`

.DEFAULT_GOAL := help

# Call this function with $(call log-info,"Your message")
define log-info =
@echo "INFO: $(1)"
endef

# -------------------------------------------------------------------
# help!
# -------------------------------------------------------------------

.PHONY: help
help: ## Prints this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

# -------------------------------------------------------------------
# required tools
# -------------------------------------------------------------------

# Find all required tools:
GIT_BIN := $(shell command -v $(GIT_BIN_NAME) 2> /dev/null)

DEP_BIN_DIR := $(TMP_PATH)/bin
DEP_BIN := $(DEP_BIN_DIR)/$(DEP_BIN_NAME)
DEP_VERSION=v0.4.1

GO_BIN := $(shell command -v $(GO_BIN_NAME) 2> /dev/null)

$(INSTALL_PREFIX):
	mkdir -p $(INSTALL_PREFIX)
$(TMP_PATH):
	mkdir -p $(TMP_PATH)

# -------------------------------------------------------------------
# deps
# -------------------------------------------------------------------
$(DEP_BIN_DIR):
	mkdir -p $(DEP_BIN_DIR)


.PHONY: deps
deps: $(DEP_BIN) $(VENDOR_DIR) ## Download build dependencies.

# install dep (see https://golang.github.io/dep/docs/installation.html)
$(DEP_BIN):
	@echo "Installing 'dep' in $(GOPATH)/bin"
	@mkdir -p $(GOPATH)/bin
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

$(VENDOR_DIR): Gopkg.lock Gopkg.toml
	$(DEP_BIN) ensure
	touch $(VENDOR_DIR)


# -------------------------------------------------------------------
# clean
# -------------------------------------------------------------------

# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

CLEAN_TARGETS += clean-artifacts
.PHONY: clean-artifacts
## Removes the ./bin directory.
clean-artifacts:
	-rm -rf $(INSTALL_PREFIX)

CLEAN_TARGETS += clean-object-files
.PHONY: clean-object-files
## Runs go clean to remove any executables or other object files.
clean-object-files:
	go clean ./...

CLEAN_TARGETS += clean-generated
.PHONY: clean-generated
## Removes all generated code.
clean-generated:
	-rm -rf ./app
	-rm -rf ./swagger/

CLEAN_TARGETS += clean-vendor
.PHONY: clean-vendor
## Removes the ./vendor directory.
clean-vendor:
	-rm -rf $(VENDOR_DIR)

CLEAN_TARGETS += clean-tmp
.PHONY: clean-tmp
## Removes the ./vendor directory.
clean-tmp:
	-rm -rf $(TMP_DIR)

# Keep this "clean" target here after all `clean-*` sub tasks
.PHONY: clean
clean: $(CLEAN_TARGETS) ## Runs all clean-* targets.

# -------------------------------------------------------------------
# build the binary executable (to ship in prod)
# -------------------------------------------------------------------
LDFLAGS=-ldflags "-X ${PACKAGE_NAME}/app.Commit=${COMMIT} -X ${PACKAGE_NAME}/app.BuildTime=${BUILD_TIME}"

$(SERVER_BIN): prebuild-check deps ## Build the server
	@echo "building $(SERVER_BIN)..."
	go build -v $(LDFLAGS) -o $(SERVER_BIN)

.PHONY: build
build: $(SERVER_BIN) ## Build the server

.PHONY: test
test: test-deps  ## Executes all tests
	$(eval TEST_PACKAGES:=$(shell go list ./... | grep -v -E $(TEST_PKGS_EXCLUDE_PATTERN)))
	go test -vet off $(TEST_PACKAGES) -v

.PHONY: format ## Removes unneeded imports and formats source code
format:
	@goimports -l -w $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: show-info
show-info:
	$(call log-info,"$(shell go version)")
	$(call log-info,"$(shell go env)")

.PHONY: prebuild-check
prebuild-check: $(TMP_PATH) $(INSTALL_PREFIX) $(CHECK_GOPATH_BIN) show-info
# Check that all tools where found
ifndef GIT_BIN
	$(error The "$(GIT_BIN_NAME)" executable could not be found in your PATH)
endif
ifndef DEP_BIN
	$(error The "$(DEP_BIN_NAME)" executable could not be found in your PATH)
endif
	@$(CHECK_GOPATH_BIN) -packageName=$(PACKAGE_NAME) || (echo "Project lives in wrong location"; exit 1)

$(CHECK_GOPATH_BIN): .make/check_gopath.go
ifndef GO_BIN
	$(error The "$(GO_BIN_NAME)" executable could not be found in your PATH)
endif
ifeq ($(OS),Windows_NT)
	@go build -o "$(shell cygpath --windows '$(CHECK_GOPATH_BIN)')" .make/check_gopath.go
else
	@go build -o $(CHECK_GOPATH_BIN) .make/check_gopath.go
endif

.PHONY: check-go-format
## Exists with an error if there are files whose formatting differs from gofmt's
check-go-format: prebuild-check
	@gofmt -s -l ${SOURCES} 2>&1 \
		| tee /tmp/gofmt-errors \
		| read \
	&& echo "ERROR: These files differ from gofmt's style (run 'make format-go-code' to fix this):" \
	&& cat /tmp/gofmt-errors \
	&& exit 1 \
	|| true

.PHONY: analyze-go-code
## Run a complete static code analysis using the following tools: golint, gocyclo and go-vet.
analyze-go-code: golint gocyclo govet

## Run gocyclo analysis over the code.
golint: $(GOLINT_BIN)
	$(info >>--- RESULTS: GOLINT CODE ANALYSIS ---<<)
	@$(foreach d,$(GOANALYSIS_DIRS),$(GOLINT_BIN) $d 2>&1 | grep -vEf .golint_exclude;)

## Run gocyclo analysis over the code.
gocyclo: $(GOCYCLO_BIN)
	$(info >>--- RESULTS: GOCYCLO CODE ANALYSIS ---<<)
	@$(foreach d,$(GOANALYSIS_DIRS),$(GOCYCLO_BIN) -over 10 $d | grep -vEf .golint_exclude;)

## Run go vet analysis over the code.
govet:
	$(info >>--- RESULTS: GO VET CODE ANALYSIS ---<<)
	@$(foreach d,$(GOANALYSIS_DIRS),go tool vet --all $d/*.go 2>&1;)

.PHONY: format-go-code
## Formats any go file that differs from gofmt's style
format-go-code: prebuild-check
	@gofmt -s -l -w ${SOURCES}


.PHONY: check
check: ## Concurrently runs a whole bunch of static analysis tools
	@gometalinter --enable=misspell --enable=gosimple --enable-gc --vendor --skip=app --skip=client --skip=tool --exclude ^app/test/ --deadline 300s ./...

.PHONY: run
run: build ## runs the service locally
	$(SERVER_BIN)
