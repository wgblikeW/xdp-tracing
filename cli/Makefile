all: go_target

go_target: main.go
	CC=clang CGO_CFLAGS="-I/usr/include/bpf" CGO_LDFLAGS="/usr/lib64/libbpf.so.0.7.0" go build -buildvcs=false -o xdp-tracing

clean:
	rm xdp-tracing