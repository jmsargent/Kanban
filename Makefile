# ---------------------------------------------------------------------------
# Tool versions — parsed from cicd/tool-versions (single source of truth).
# Update cicd/tool-versions to change versions across CI and local development.
# ---------------------------------------------------------------------------
GOLANGCI_LINT_VERSION     := $(shell grep '^golangci-lint ' cicd/tool-versions | awk '{print $$2}')
GO_ARCH_LINT_VERSION      := $(shell grep '^go-arch-lint ' cicd/tool-versions | awk '{print $$2}')
GO_SEMVER_RELEASE_VERSION := $(shell grep '^go-semver-release ' cicd/tool-versions | awk '{print $$2}')
GORELEASER_VERSION        := $(shell grep '^goreleaser ' cicd/tool-versions | awk '{print $$2}')
GOTESTSUM_VERSION         := $(shell grep '^gotestsum ' cicd/tool-versions | awk '{print $$2}')

BIN := $(CURDIR)/bin

# Prepend project-local bin/ so all targets use the pinned tool versions.
export PATH := $(BIN):$(PATH)

# Derived from git remote — used by ci-tag and ci-release.
# Override on the command line if needed: make ci-tag GITHUB_OWNER=myorg GITHUB_REPO=myrepo
GITHUB_OWNER ?= $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*[:/]([^/]+)/[^/]+\.git|\1|;s|.*[:/]([^/]+)/[^/]+$$|\1|')
GITHUB_REPO  ?= $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*[:/][^/]+/([^/]+)\.git|\1|;s|.*[:/][^/]+/([^/]+)$$|\1|')

.PHONY: install-tools install-golangci-lint install-go-arch-lint install-go-semver-release \
        install-gotestsum install-goreleaser \
        pre-commit release-snapshot release tag-dry \
        ci-tag ci-checkout-tagged ci-release help \
        ci-arch-check ci-static-analysis ci-unit-tests ci-build \
        ci-set-env ci-e2e-tests ci-fetch-tags ci-tag-and-release \
        ci-smoke-test-binary ci-smoke-test-go-install

# ---------------------------------------------------------------------------
# Tool installation — idempotent, version-aware, installed to bin/
#
# Sentinel file pattern: bin/.toolname-X.Y.Z exists only when that exact
# version is installed. Updating cicd/tool-versions removes the old sentinel,
# triggering reinstall on the next make invocation. Safe to run repeatedly.
# ---------------------------------------------------------------------------
$(BIN):
	@mkdir -p $(BIN)

$(BIN)/.golangci-lint-$(GOLANGCI_LINT_VERSION): | $(BIN)
	@rm -f $(BIN)/.golangci-lint-*
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
	  | sh -s -- -b $(BIN) v$(GOLANGCI_LINT_VERSION)
	@touch $@

$(BIN)/.go-arch-lint-$(GO_ARCH_LINT_VERSION): | $(BIN)
	@rm -f $(BIN)/.go-arch-lint-*
	@GOBIN=$(BIN) go install github.com/fe3dback/go-arch-lint@v$(GO_ARCH_LINT_VERSION)
	@touch $@

$(BIN)/.go-semver-release-$(GO_SEMVER_RELEASE_VERSION): | $(BIN)
	@rm -f $(BIN)/.go-semver-release-*
	@GOBIN=$(BIN) go install github.com/s0ders/go-semver-release/v6@v$(GO_SEMVER_RELEASE_VERSION)
	@touch $@

$(BIN)/.gotestsum-$(GOTESTSUM_VERSION): | $(BIN)
	@rm -f $(BIN)/.gotestsum-*
	@GOBIN=$(BIN) go install gotest.tools/gotestsum@v$(GOTESTSUM_VERSION)
	@touch $@

$(BIN)/.goreleaser-$(GORELEASER_VERSION): | $(BIN)
	@rm -f $(BIN)/.goreleaser-*
	@GOBIN=$(BIN) go install github.com/goreleaser/goreleaser/v2@v$(GORELEASER_VERSION)
	@touch $@

## install-golangci-lint: install golangci-lint to bin/ (version-pinned, idempotent)
install-golangci-lint: $(BIN)/.golangci-lint-$(GOLANGCI_LINT_VERSION)

## install-go-arch-lint: install go-arch-lint to bin/ (version-pinned, idempotent)
install-go-arch-lint: $(BIN)/.go-arch-lint-$(GO_ARCH_LINT_VERSION)

## install-go-semver-release: install go-semver-release to bin/ (version-pinned, idempotent)
install-go-semver-release: $(BIN)/.go-semver-release-$(GO_SEMVER_RELEASE_VERSION)

## install-gotestsum: install gotestsum to bin/ (version-pinned, idempotent)
install-gotestsum: $(BIN)/.gotestsum-$(GOTESTSUM_VERSION)

## install-goreleaser: install goreleaser to bin/ (version-pinned, idempotent)
install-goreleaser: $(BIN)/.goreleaser-$(GORELEASER_VERSION)

## install-tools: install all pipeline tools to bin/ (version-pinned, idempotent)
install-tools: install-golangci-lint install-go-arch-lint install-go-semver-release install-gotestsum install-goreleaser

# ---------------------------------------------------------------------------
# Quality gates
# ---------------------------------------------------------------------------

## pre-commit: run the same quality gates as the CI pipeline, in the same order
pre-commit: install-tools
	@make ci-arch-check
	@make ci-static-analysis
	@make ci-unit-tests
	@make ci-build
	@make ci-e2e-tests

# ---------------------------------------------------------------------------
# Release
# ---------------------------------------------------------------------------

## release-snapshot: build all cross-compile targets locally without publishing
release-snapshot: install-goreleaser
	@goreleaser release --snapshot --clean --config cicd/goreleaser.yml

## release: publish a release via goreleaser (requires GITHUB_TOKEN)
release: install-goreleaser
	@goreleaser release --clean --config cicd/goreleaser.yml

## tag-dry: dry-run semantic version tagging
tag-dry: install-go-semver-release
	@go-semver-release release --dry-run

## ci-tag: compute and push semver tag (mirrors CI tag-and-release job; requires GITHUB_TOKEN)
ci-tag: install-go-semver-release
	@if git log -1 --pretty=%B | grep -qi '\[skip release\]'; then \
	  echo "[skip release] detected in commit message — skipping tag"; \
	  exit 0; \
	fi
	@if git describe --tags --exact-match HEAD >/dev/null 2>&1; then \
	  echo "Commit $$(git rev-parse --short HEAD) is already tagged. Skipping."; \
	  exit 0; \
	fi
	@go-semver-release release "https://$${GITHUB_TOKEN}@github.com/$(GITHUB_OWNER)/$(GITHUB_REPO).git" \
	  --tag-prefix v \
	  --access-token $${GITHUB_TOKEN} \
	  --branches '[{"name": "main"}]'

## ci-checkout-tagged: checkout the commit the latest tag points to
ci-checkout-tagged:
	@git fetch --tags
	@LATEST_TAG=$$(git tag --sort=-version:refname | grep '^v' | head -1); \
	TAG_SHA=$$(git rev-list -n1 "$$LATEST_TAG"); \
	CURRENT_SHA=$$(git rev-parse HEAD); \
	if [ "$$TAG_SHA" != "$$CURRENT_SHA" ]; then \
	  echo "Tag $$LATEST_TAG points to $$TAG_SHA, not $$CURRENT_SHA — checking out tagged commit"; \
	  git checkout "$$TAG_SHA"; \
	else \
	  echo "Tag $$LATEST_TAG correctly points to current HEAD $$CURRENT_SHA"; \
	fi

## ci-release: publish release via goreleaser (mirrors CI release step; requires GITHUB_TOKEN)
ci-release: install-goreleaser
	@if git log -1 --pretty=%B | grep -qi '\[skip release\]'; then \
	  echo "[skip release] detected in commit message — skipping release"; \
	  exit 0; \
	fi
	@GITHUB_REPOSITORY_OWNER=$(GITHUB_OWNER) goreleaser release --config cicd/goreleaser.yml --clean

# ---------------------------------------------------------------------------
# CI job steps — thin wrappers invoked by cicd/config.yml run steps.
# Each CI run step delegates to exactly one make target.
# ---------------------------------------------------------------------------

## ci-arch-check: run architecture lint (CI step)
ci-arch-check:
	@go-arch-lint check --arch-file cicd/go-arch-lint.yml

## ci-static-analysis: run vet, lint, and vulnerability scan (CI step)
ci-static-analysis:
	@go vet ./...
	@golangci-lint run
	@govulncheck ./...

## ci-unit-tests: run unit tests via gotestsum (CI step)
ci-unit-tests:
	@gotestsum --format testname -- ./internal/...

## ci-build: build the kanban binary (CI step)
ci-build:
	@go build -o kanban ./cmd/kanban

## ci-set-env: export KANBAN_BIN for acceptance tests (CI step)
ci-set-env:
	@echo "export KANBAN_BIN=$$PWD/kanban" >> $$BASH_ENV

## ci-e2e-tests: run acceptance tests via gotestsum (CI step)
ci-e2e-tests:
	@KANBAN_BIN="$(CURDIR)/kanban" gotestsum --format testname -- ./tests/acceptance/...

## ci-fetch-tags: fetch full history and tags for release (CI step)
ci-fetch-tags:
	@git fetch --unshallow --tags || git fetch --tags

## ci-tag-and-release: tag, checkout, and release (CI step)
ci-tag-and-release:
	@make ci-tag
	@make ci-checkout-tagged
	@make ci-release

# ---------------------------------------------------------------------------
# Smoke tests — post-release installation verification.
# Each target installs kanban via a distribution channel and runs kanban --help.
# ---------------------------------------------------------------------------

LATEST_TAG ?= $(shell git tag --sort=-version:refname | grep '^v' | head -1)

## ci-smoke-test-binary: download release binary from GitHub and verify it runs
ci-smoke-test-binary:
	@echo "Smoke test: binary download ($(LATEST_TAG))"
	@mkdir -p /tmp/kanban-smoke
	@curl -sSfL "https://github.com/$(GITHUB_OWNER)/$(GITHUB_REPO)/releases/download/$(LATEST_TAG)/kanban_$${LATEST_TAG#v}_linux_amd64.tar.gz" \
	  | tar -xz -C /tmp/kanban-smoke
	@/tmp/kanban-smoke/kanban --help
	@rm -rf /tmp/kanban-smoke
	@echo "PASS: binary smoke test"

## ci-smoke-test-go-install: install via go install and verify it runs
ci-smoke-test-go-install:
	@echo "Smoke test: go install ($(LATEST_TAG))"
	@GOBIN=/tmp/kanban-smoke go install github.com/$(GITHUB_OWNER)/$(GITHUB_REPO)/cmd/kanban@$(LATEST_TAG)
	@/tmp/kanban-smoke/kanban --help
	@rm -rf /tmp/kanban-smoke
	@echo "PASS: go install smoke test"

## help: list available make targets
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Install tools (idempotent — safe to run multiple times):"
	@echo "  install-tools              install all pipeline tools to bin/"
	@echo "  install-golangci-lint      golangci-lint v$(GOLANGCI_LINT_VERSION)"
	@echo "  install-go-arch-lint       go-arch-lint v$(GO_ARCH_LINT_VERSION)"
	@echo "  install-go-semver-release  go-semver-release v$(GO_SEMVER_RELEASE_VERSION)"
	@echo "  install-gotestsum          gotestsum v$(GOTESTSUM_VERSION)"
	@echo "  install-goreleaser         goreleaser v$(GORELEASER_VERSION)"
	@echo ""
	@echo "  Quality gates:"
	@echo "  pre-commit              run CI quality gates locally, in pipeline order"
	@echo ""
	@echo "  Release:"
	@echo "  release-snapshot        build cross-compile targets without publishing"
	@echo "  release                 publish a release via goreleaser"
	@echo "  tag-dry                 dry-run semantic version tagging"
	@echo "  ci-tag                  compute and push semver tag (requires GITHUB_TOKEN)"
	@echo "  ci-checkout-tagged      checkout the commit the latest tag points to"
	@echo "  ci-release              publish release via goreleaser (requires GITHUB_TOKEN)"
	@echo "  help                    list available make targets"
