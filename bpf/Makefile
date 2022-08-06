# all: vmlinux.h bpf_target bpf_skeleton go_target
all: bpf_target go_target
# vmlinux.h:
# 	bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h

bpf_target: xdp-proxy.bpf.c
	clang -g -O2 -c -target bpf -o xdp_proxy_kern.o xdp-proxy.bpf.c

# bpf_skeleton:
# 	bpftool gen skeleton xdp-proxy.bpf.o > xdp-proxy.skel.h

go_target: xdp-proxy.o libbpfgo_warp.go
	CC=clang CGO_CFLAGS="-I/usr/include/bpf" CGO_LDFLAGS="/usr/lib64/libbpf.so.0.7.0" go build -buildvcs=false -o xdp-proxy

clean:
	rm xdp-proxy.skel.h xdp-proxy vmlinux.h xdp_proxy_kern.o xdp-proxy.o

.PHONY: vmlinux
vmlinux:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > ./headers/vmlinux.h

CLANG ?= clang
CFLAGS = -g -O2 -Wall -fpie
ARCH = x86
OUTPUT = ./output

COMMANDS ?= $(wildcard *.bpf.c)
.PHONY: bpf.build
bpf.build: $(addprefix bpf.build.,$(foreach cmd,$(COMMANDS),$(word 1,$(subst ., ,$(cmd)))))


.PHONY: bpf.build.%
bpf.build.%:
	@mkdir -p $(OUTPUT)
	$(CLANG) $(CFLAGS) -target bpf -D__TARGET_ARCH_$(ARCH) \
	-I./headers \
	-c ./$*.bpf.c -o $(OUTPUT)/$*.bpf.o
	@echo -e '\n\n'

.PHONY: clean
clean:
	@rm -rf ./output