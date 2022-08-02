IGNORE_ACTIVE := active

ifeq ($(origin IGNORE_ACTIVE),undefined)
IGNORE_FLAG := --ignore
EXCLUDED_DIR := output/**
endif

.PHONY: copyright.verify
copyright.verify: tools.verify.addlicense
	@echo -e "\033[32m ===========> Verifying the boilerplate headers for all files \033[0m"
	@addlicense --check -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) $(IGNORE_FLAG) $(EXCLUDED_DIR)

.PHONY: copyright.add
copyright.add: tools.verify.addlicense
	@addlicense -v -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) $(IGNORE_FLAG) $(EXCLUDED_DIR)