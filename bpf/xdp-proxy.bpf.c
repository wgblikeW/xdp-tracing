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
    /* Set up libbpf errors and debug info callback */
   
    
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

    struct sock_key key = {
        .dip = __bpf_ntohl(iph->daddr),
        .sip = __bpf_ntohl(iph->saddr),
        /* convert to network byte order */
        .sport = ((tcphdr->source & 0xff00) >> 8) + ((tcphdr->source & 0x00ff) << 8),
        .dport = ((tcphdr->dest & 0xff00) >> 8) + ((tcphdr->dest & 0x00ff) << 8)};
    char *payload = data + nh_off;

    int ret;
    ret = bpf_perf_event_output(ctx, &bridge, flags, &key, sizeof(key));
    if (ret)
        bpf_printk("perf_event_output failed: %d\n", ret);
    return XDP_PASS;
}


char _license[] SEC("license") = "GPL";