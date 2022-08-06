#include "headers/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include "headers/commons.h"

BPF_RINGBUF(perfs, 1024)

SEC("xdp")
int __test_trace_xdp(struct xdp_md *ctx) {
    int *lucknum;
    lucknum = bpf_ringbuf_reserve(&perfs, sizeof(int), 0);
    if (!lucknum) {
        return XDP_PASS;
    }

    *lucknum = 7070;
    bpf_ringbuf_submit(lucknum, 0);
    return XDP_PASS; 
}

char _license[] SEC("license") = "GPL";