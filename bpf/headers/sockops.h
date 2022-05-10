#define MAXIMUM_LIST_LEN 20
#define MAX_CPUS 128
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
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(int));
    __uint(value_size, sizeof(__u32));
    __uint(max_entries, MAX_CPUS);
} bridge SEC(".maps");