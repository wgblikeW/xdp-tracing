#ifndef __SOCK_OPS_H__
#define __SOCK_OPS_H__
#define MAXIMUM_LIST_LEN 20
#include <linux/bpf.h>

struct sock_key
{
    __u32 sip;
    __u32 dip;
    __u16 sport;
    __u16 dport;
    // __u32 family;
};

struct tcpInfo {
    unsigned char synflag;
    unsigned char finflag;
    unsigned char rstflag;
    unsigned char pshflag;
    unsigned char ackflag;
    unsigned char urgflag;
    unsigned char payload[5];
};

struct myvalue {
    int counter;
    struct tcpInfo info[MAXIMUM_LIST_LEN];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct sock_key);
    __type(value, struct myvalue);
    __uint(max_entries, 1024);
} hash_map SEC(".maps");

#endif /* __SOCK_OPS_H__ */