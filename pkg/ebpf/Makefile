CLANG = clang

INCLUDE_DIR = /usr/include
LIB_DIR = /usr/lib64
LIBBPF_OBJ = $(LIB_DIR)/libbpf.a

CGO_CFLAGS_STATIC = "-I$(abspath $(INCLUDE_DIR))"
CGO_LDFLAGS_STATIC = "-lelf -lz $(LIBBPF_OBJ)"
GO = go

# Set a specific PLATFORM
ifeq ($(origin PLATFORM), undefined)
	ifeq ($(origin GOOS), undefined)
		GOOS := $(shell go env GOOS)
	endif
	ifeq ($(origin GOARCH), undefined)
		GOARCH := $(shell go env GOARCH)
	endif
	PLATFORM := $(GOOS)_$(GOARCH)
else
	GOOS := $(word 1, $(subst _, ,$(PLATFORM)))
	GOARCH := $(word 2, $(subst _, ,$(PLATFORM)))
	IMAGE_PLAT := $(PLATFORM)
endif

all: ebpf.test

.PHONY: ebpf.test
ebpf.test:
	@CC=$(CLANG) \
		CGO_CFLAGS=$(CGO_CFLAGS_STATIC) \
		CGO_LDFLAGS=$(CGO_LDFLAGS_STATIC) \
		GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) test -v 2>/dev/null