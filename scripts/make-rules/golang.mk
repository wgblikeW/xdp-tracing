GO := go
GO_SUPPORTED_VERSIONS ?= 1.13|1.14|1.15|1.16|1.17|1.18
GO_LDFLAGS += -X $(VERSION_PACKAGE).GitVersion=$(VERSION) \
	-X $(VERSION_PACKAGE).GitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PACKAGE).GitTreeState=$(GIT_TREE_STATE) \
	-X $(VERSION_PACKAGE).BuildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GO_BUILD_FLAGS += -ldflags "$(GO_LDFLAGS)"

# go.build.verify will check the go version whether matches the requirement for building
.PHONY: go.build.verify
go.build.verify:
ifneq ($(shell $(GO) version | grep -q -E '\bgo($(GO_SUPPORTED_VERSIONS))\b' && echo 0 || echo 1), 0)
	$(error unsupported go version. Please make install one of the following supported version: '$(GO_SUPPORTED_VERSIONS)')
endif

# go.revive will do check syntax and styling of go sources using revive
.PHONY: go.revive
go.revive: tools.verify.revive
	@echo -e "\033[32m===========> Run revive to lint source codes\033[0m"
	@revive -formatter stylish -config $(ROOT_DIR)/.revive.toml $(ROOT_DIR)/...

.PHONY: go.build.static.%
go.build.static.%: bpf.build
	@mkdir -p $(OUTPUT)/cmd/static/$*
	CC=$(CLANG) \
		CGO_CFLAGS=$(CGO_CFLAGS_STATIC) \
		CGO_LDFLAGS=$(CGO_LDFLAGS_STATIC) \
		GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build \
		-o $(OUTPUT)/cmd/static/$* $(ROOT_DIR)/bpf/$*/main.go
	@mv $(OUTPUT)/cmd/static/$*/main $(OUTPUT)/cmd/static/$*/$*
	@echo -e '\n\n'

.PHONY: go.build.dynamic.%
go.build.dynamic.%: bpf.build
	@mkdir -p $(OUTPUT)/cmd/dynamic/$*
	CC=$(CLANG) \
		CGO_CFLAGS=$(CGO_CFLAGS_DYN) \
		CGO_LDFLAGS=$(CGO_LDFLAGS_DYN) \
		$(GO) build \
		-o $(OUTPUT)/cmd/dynamic/$* $(ROOT_DIR)/bpf/$*/main.go
	@mv $(OUTPUT)/cmd/dynamic/$*/main $(OUTPUT)/cmd/dynamic/$*/$*
	@echo -e '\n\n'

COMMANDS_GO = $(wildcard $(BPF_DIR)/*/*.go)
GO_TARGET = $(foreach cmd,$(COMMANDS_GO),$(notdir $(cmd)))

.PHONY: go.build.static
go.build.static: go.mod
	$(MAKE) $(addprefix go.build.static.,$(foreach cmd,$(BPF_TARGET),$(word 1,$(subst ., ,$(cmd)))))

.PHONY: go.build.dynamic
go.build.dynamic: go.mod
	$(MAKE) $(addprefix go.build.dynamic.,$(foreach cmd,$(BPF_TARGET),$(word 1,$(subst ., ,$(cmd)))))

.PHONY: go.mod
go.mod:
	@$(GO) mod tidy