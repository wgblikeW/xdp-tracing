//go:build ignore
#include <stddef.h>
#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include <linux/tcp.h>
#include "headers/sockops.h"



SEC("xdp")
int xdp_proxy(struct xdp_md *ctx)
{   
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    __u64 nh_off = 0;
    __u64 flags = BPF_F_CURRENT_CPU;

    struct ethhdr *eth = data;
    nh_off = sizeof(struct ethhdr);
    /* abort on illegal packets */
    if (data + nh_off > data_end)
    {
        return XDP_ABORTED;
    }

    /* do nothing for non-IP packets */
    if (eth->h_proto != bpf_htons(ETH_P_IP))
    {
        return XDP_PASS;
    }

    struct iphdr *iph = data + nh_off;
    nh_off += sizeof(struct iphdr);
    /* abort on illegal packets */
    if (data + nh_off > data_end)
    {
        return XDP_ABORTED;
    }

    /* do nothing for non-TCP packets */
    if (iph->protocol != IPPROTO_TCP)
    {
        return XDP_PASS;
    }

    struct tcphdr *tcphdr = data + nh_off;
    nh_off += sizeof(struct tcphdr);
    if (data + nh_off > data_end) {
        return XDP_ABORTED;
    }

    __u32 key = iph->saddr;
    char *payload = data + nh_off;

    if ( bpf_map_lookup_elem(&bridge, &key) == NULL) {
        return XDP_PASS;
    }

    return XDP_DROP;
}


char _license[] SEC("license") = "GPL";