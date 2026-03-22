# Derived from git remote — used by ci-tag and ci-release.
# Override on the command line if needed: make ci-tag GITHUB_OWNER=myorg GITHUB_REPO=myrepo
GITHUB_OWNER ?= $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*[:/]([^/]+)/[^/]+\.git|\1|;s|.*[:/]([^/]+)/[^/]+$$|\1|')
GITHUB_REPO  ?= $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*[:/][^/]+/([^/]+)\.git|\1|;s|.*[:/][^/]+/([^/]+)$$|\1|')

.PHONY: validate acceptance ci release-snapshot release tag-dry ci-tag ci-checkout-tagged ci-release help

## validate: run static quality gates locally (mirrors CI validate-and-build job)
validate:
	@echo "[0/4] check-versions"
	@cicd/check-versions.sh
	@echo "[1/4] gotestsum ./internal/..."
	@gotestsum --format testname -- ./internal/...
	@echo "[2/4] golangci-lint run"
	@golangci-lint run
	@echo "[3/4] go-arch-lint check"
	@go-arch-lint check
	@echo "[4/4] go build ./..."
	@go build ./...
	@echo "PASS"

## acceptance: build kanban binary and run acceptance tests
acceptance:
	@echo "Building kanban binary..."
	@go build -o kanban ./cmd/kanban
	@echo "Running acceptance tests..."
	@KANBAN_BIN="$(CURDIR)/kanban" gotestsum --format testname -- ./tests/acceptance/...

## ci: run validate then acceptance (mirrors CI pipeline locally)
ci:
	@make validate && make acceptance

## release-snapshot: build all cross-compile targets locally without publishing
release-snapshot:
	@goreleaser release --snapshot --clean --config cicd/goreleaser.yml

## release: publish a release via goreleaser (requires GITHUB_TOKEN)
release:
	@goreleaser release --clean --config cicd/goreleaser.yml

## tag-dry: dry-run semantic version tagging
tag-dry:
	@go-semver-release release --dry-run

## ci-tag: compute and push semver tag (mirrors CI tag job; requires GITHUB_TOKEN)
ci-tag:
	@if git log -1 --pretty=%B | grep -qi '\[skip release\]'; then \
	  echo "[skip release] detected in commit message — skipping tag"; \
	  exit 0; \
	fi
	@if git describe --tags --exact-match HEAD >/dev/null 2>&1; then \
	  echo "Commit $$(git rev-parse --short HEAD) is already tagged. Skipping."; \
	  exit 0; \
	fi
	@go-semver-release release "https://${GITHUB_TOKEN}@github.com/$(GITHUB_OWNER)/$(GITHUB_REPO).git" \
	  --tag-prefix v \
	  --access-token $${GITHUB_TOKEN} \
	  --branches '[{"name": "main"}]'

## ci-checkout-tagged: checkout the commit the latest tag points to (mirrors CI release job pre-step)
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

## ci-release: publish release via goreleaser (mirrors CI release job; requires GITHUB_TOKEN)
ci-release:
	@if git log -1 --pretty=%B | grep -qi '\[skip release\]'; then \
	  echo "[skip release] detected in commit message — skipping release"; \
	  exit 0; \
	fi
	@GITHUB_REPOSITORY_OWNER=$(GITHUB_OWNER) goreleaser release --config cicd/goreleaser.yml --clean

## help: list available make targets
help:
	@echo "Available targets:"
	@echo "  validate        run all quality gates (mirrors CI validate-and-build)"
	@echo "  acceptance      build binary and run acceptance tests"
	@echo "  ci              run validate then acceptance"
	@echo "  release-snapshot build cross-compile targets without publishing"
	@echo "  release         publish a release via goreleaser"
	@echo "  tag-dry         dry-run semantic version tagging"
	@echo "  ci-tag          compute and push semver tag (requires GITHUB_TOKEN)"
	@echo "  ci-checkout-tagged checkout the commit the latest tag points to"
	@echo "  ci-release      publish release via goreleaser (requires GITHUB_TOKEN)"
	@echo "  help            list available make targets"
