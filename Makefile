.PHONY: all
all: build

# ====================================================
# Includes

include scripts/make-rules/golang.mk
include scripts/make-rules/bpf.mk
include scripts/make-rules/copyright.mk
include scripts/make-rules/tools.mk
include scripts/make-rules/common.mk
# ====================================================

# ====================================================
# Targets

## build: Build source code for host platform.
.PHONY: build
build:
	@$(MAKE) go.build

## add-copyright: Ensures source code files have copyright license headers.
.PHONY: add-copyright
add-copyright:
	@$(MAKE) copyright.add