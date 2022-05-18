#define MAX_ENTRIES 1024
#include <linux/bpf.h>

struct sock_key
{
    __u32 sip;
    __u16 dport;
    // __u32 family;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(key_size, sizeof(struct sock_key));
    __uint(value_size, sizeof(__u32));
    __uint(max_entries, MAX_ENTRIES);
} bridge SEC(".maps");