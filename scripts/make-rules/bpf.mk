CLANG ?= clang
CFLAGS = -g -O2 -Wall -fpie

INCLUDE_DIR = /usr/include
LIB_DIR = /usr/lib64
LIBBPF_OBJ = $(LIB_DIR)/libbpf.a
BPF_DIR = $(ROOT_DIR)/bpf

# Static Link Build Settings
CGO_CFLAGS_STATIC = "-I$(abspath $(INCLUDE_DIR))"
CGO_LDFLAGS_STATIC = "-lelf -lz $(LIBBPF_OBJ)"
CGO_EXTLDFLAGS_STATIC = '-w -extldflags "-static"'

# Shared Library Build Settings
CGO_CFGLAGS_DYN = "-I. -I$(INCLUDE_DIR)
CGO_LDFLAGS_DYN = "-lelf -lz -lbpf"

COMMANDS ?= $(wildcard $(BPF_DIR)/*/*.bpf.c)
BPF_TARGET = $(foreach cmd,$(COMMANDS),$(notdir $(cmd)))

.PHONY: gen.vmlinux
gen.vmlinux:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > $(ROOT_DIR)/bpf/headers/vmlinux.h

.PHONY: bpf.build.%
bpf.build.%: gen.vmlinux
	@mkdir -p $(OUTPUT)/bpf
	$(CLANG) $(CFLAGS) -target bpf -D__TARGET_ARCH_$(ARCH) \
	-I$(ROOT_DIR)/bpf/include -I$(BPF_DIR)/$*/ \
	-c $(BPF_DIR)/$*/$*.bpf.c -o $(OUTPUT)/bpf/$*.bpf.o
	@echo -e '\n\n'

.PHONY: bpf.build
bpf.build: $(addprefix bpf.build.,$(foreach cmd,$(BPF_TARGET),$(word 1,$(subst ., ,$(cmd)))))