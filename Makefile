.PHONY: all
all: build

# ====================================================
# Includes

include scripts/make-rules/golang.mk

# ====================================================

# ====================================================
# Targets

## build: Build source code for host platform.
.PHONY: build
build:
	@$(MAKE) go.build