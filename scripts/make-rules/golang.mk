GO := go
GO_SUPPORTED_VERSIONS ?= 1.13|1.14|1.15|1.16|1.17|1.18

.PHONY: go.build.verify
go.build.verify:
ifneq ($(shell $(GO) version | grep -q -E 'go($(GO_SUPPORTED_VERSIONS))' && echo 0 || echo 1), 0) 
	$(error unsupported go version. Please make insall one of the following supported version: '$(GO_SUPPORTED_VERSIONS)')
endif

.PHONY: go.build.xdp-tracing
go.build.xdp-tracing:
	@echo -e "\033[32m ============> Build Binary <============ \033[0m"

.PHONY: go.build
go.build: go.build.verify go.build.xdp-tracing