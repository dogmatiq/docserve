GO_EMBEDDED_FILES += persistence/schema.sql
GO_EMBEDDED_FILES += $(shell find web -iname '*.html')
GO_EMBEDDED_FILES += $(shell find web/assets)

-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile


.PHONY: run
run: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/browser
	DEBUG=true $<

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"
