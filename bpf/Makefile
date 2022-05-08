all: vmlinux.h bpf_target bpf_skeleton go_target

vmlinux.h:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h

bpf_target: xdp-proxy.bpf.c
	clang -g -O2 -c -target bpf -o xdp-proxy.bpf.o xdp-proxy.bpf.c

bpf_skeleton:
	bpftool gen skeleton xdp-proxy.bpf.o > xdp-proxy.skel.h

go_target: xdp-proxy.o main.go
	CC=clang CGO_CFLAGS="-I/usr/include/bpf" CGO_LDFLAGS="/usr/lib64/libbpf.so.0.7.0" go build -buildvcs=false -o xdp-proxy

clean:
	rm xdp-proxy.skel.h xdp-proxy vmlinux.h xdp-proxy.bpf.o xdp-proxy.o