GO=$(shell which go)
GIT=$(shell which git)

.PHONY: go-check
go-check:
	$(call error-if-empty,$(GO),go)

.PHONY: go-module-version
go-module-version: go-check
	@echo "go get $(shell $(GO) list)@$(shell $(GIT) rev-parse HEAD)"

.PHONY: test
test: go-check
	@$(GO) test --race --cover ./...

define error-if-empty
@if [[ -z $(1) ]]; then echo "$(2) not installed"; false; fi
endef