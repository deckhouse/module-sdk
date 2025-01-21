GO=$(shell which go)
GIT=$(shell which git)
GOLANGCI_LINT=$(shell which golangci-lint)

.PHONY: go-check
go-check:
	$(call error-if-empty,$(GO),go)

.PHONY: git-check
git-check:
	$(call error-if-empty,$(GIT),git)

.PHONY: golangci-lint-check
golangci-lint-check:
	$(call error-if-empty,$(GIT),git)

.PHONY: go-module-version
go-module-version: go-check git-check
	@echo "go get $(shell $(GO) list ./pkg/app)@$(shell $(GIT) rev-parse HEAD)"

.PHONY: test
test: go-check
	@$(GO) test --race --cover ./...

.PHONY: lint
lint: golangci-lint-check
	@$(GOLANGCI_LINT) run ./... --fix

define error-if-empty
@if [[ -z $(1) ]]; then echo "$(2) not installed"; false; fi
endef