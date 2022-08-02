TOOLS ?=$(BLOCKER_TOOLS) $(CRITICAL_TOOLS) $(TRIVIAL_TOOLS)

.PHONY: tools.install
tools.install: $(addprefix tools.install., $(TOOLS))

.PHONY: tools.install.%
tools.install.%:
	@echo -e "\033[32m ===========> Installing $* \033[0m"
	@$(MAKE) install.$*

.PHONY: tools.verify.%
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi

.PHONY: install.addlicense
install.addlicense:
	@$(GO) install github.com/google/addlicense@latest

.PHONY: install.golines
install.golines:
	@$(GO) install github.com/segmentio/golines@latest

.PHONY: install.goimports
install.goimports:
	@$(GO) install golang.org/x/tools/cmd/goimports@latest

.PHONY: install.revive
install.revive:
	@$(GO) install github.com/mgechev/revive@latest