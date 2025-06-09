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

.PHONY: examples
examples: go-check examples-mod examples-test examples-lint
	@echo "Running examples tests and linting"
	@$(GOLANGCI_LINT) run ./... --fix

.PHONY: examples-mod
examples-mod: go-check
	@for dir in $$(find . -mindepth 2 -name go.mod | sed -r 's/(.*)(go.mod)/\1/g'); do \
		echo "Running go mod tidy in $${dir}"; \
		cd $(CURDIR)/$${dir} && go mod tidy && cd $(CURDIR); \
	done

.PHONY: examples-test
examples-test: go-check
	@for dir in $$(find . -mindepth 2 -name go.mod | sed -r 's/(.*)(go.mod)/\1/g'); do \
		echo "Running tests in $${dir}"; \
		cd $(CURDIR)/$${dir} && $(GO) test --race --cover ./... && cd $(CURDIR); \
	done

.PHONY: examples-lint
examples-lint: golangci-lint-check
	@for dir in $$(find . -mindepth 2 -name go.mod | sed -r 's/(.*)(go.mod)/\1/g'); do \
		echo "Running linter in $${dir}"; \
		cd $(CURDIR)/$${dir} && $(GOLANGCI_LINT) run ./... --fix && cd $(CURDIR); \
	done