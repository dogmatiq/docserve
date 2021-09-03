DOCKER_REPO = ghcr.io/dogmatiq/browser
DOCKER_PLATFORMS += linux/amd64
DOCKER_PLATFORMS += linux/arm64

GO_EMBEDDED_FILES += persistence/schema.sql
GO_EMBEDDED_FILES += $(shell find web -iname '*.html')
GO_EMBEDDED_FILES += $(shell find web/assets)

-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/pkg/docker/v1/Makefile

.PHONY: run
run: $(GO_DEBUG_DIR)/browser
	DEBUG=true $<

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"
