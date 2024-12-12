GO=$(shell which go)
GIT=$(shell which git)

.PHONY: go-check
go-check:
	$(call error-if-empty,$(GO),go)

.PHONY: go-module-version
go-module-version: go-check
	@echo "$(shell $(GIT) describe --tags --abbrev=0)-0.$(shell TZ=UTC $(GIT) --no-pager show   --quiet   --abbrev=12   --date='format-local:%Y%m%d%H%M%S'   --format="%cd-%h")"

define error-if-empty
@if [[ -z $(1) ]]; then echo "$(2) not installed"; false; fi
endef