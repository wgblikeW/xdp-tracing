CMD_CLANG ?= clang
CMD_GO ?= go
CMD_AWK ?= awk
CMD_SED ?= sed
CMD_CUT ?= cut

GO_VERSION = $(shell $(CMD_GO) version 2>/dev/null | $(CMD_AWK) '{print $$3}' | $(CMD_SED) 's:go::g' | $(CMD_CUT) -d. -f1,2)
GO_VERSION_MAJ = $(shell echo $(GO_VERSION) | $(CMD_CUT) -d'.' -f1)
GO_VERSION_MIN = $(shell echo $(GO_VERSION) | $(CMD_CUT) -d'.' -f2)

.checkver_$(CMD_GO): \
	| .check_$(CMD_GO)
#
	@if [ ${GO_VERSION_MAJ} -eq 1 ]; then
		if [ ${GO_VERSION_MIN} -lt 16 ]; then
			echo -n "you MUST use golang 1.16 or newer, "
			echo "your current golang version is ${GO_VERSION}"
			exit 1
		fi
	fi
	touch $@
	
all: go_target

go_target: main.go
	CC=$(CMD_CLANG) CGO_CFLAGS="-I/usr/include/bpf" CGO_LDFLAGS="/usr/lib64/libbpf.so.0.7.0" go build -buildvcs=false -o xdp-tracing

clean:
	rm xdp-tracing