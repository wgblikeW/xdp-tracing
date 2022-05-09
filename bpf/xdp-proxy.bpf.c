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

    // filter TCP Steam without port of 1080
    if (key.dport == 22) {
        return XDP_PASS;
    }
    struct myvalue *value = bpf_map_lookup_elem(&hash_map, &key);
    struct tcpInfo curInfo = {tcphdr->syn, tcphdr->fin, tcphdr->rst, tcphdr->psh, tcphdr->ack, tcphdr->urg};
    if (data + nh_off + 5 < data_end) {
        __builtin_memcpy(curInfo.payload, payload, 5);
    }

    if (value)
    {
        // check boundary of the memory in order to pass the JIT check
        if (value->counter + 1 < MAXIMUM_LIST_LEN && value->counter + 1 > 0) {
            value->info[value->counter + 1] = curInfo;
        }
        __sync_fetch_and_add(&value->counter, 1);
    }
    else
    {
        struct myvalue val = {0};
        val.info[0] = curInfo;
        bpf_map_update_elem(&hash_map, &key, &val, BPF_ANY);
    }

    return XDP_PASS;
}


char _license[] SEC("license") = "GPL";