#include <net/if.h>
#include <linux/types.h>
#include <stdbool.h>

struct config {
    __u32 xdp_flags;
    int ifindex;
    char *ifname;
    char ifname_buf[IF_NAMESIZE];
    int redirect_ifindex;
    char *redirect_ifname;
    char redirect_ifname_buf[IF_NAMESIZE];
    bool do_unload;
    bool reuse_maps;
    char pin_dir[512];
    char filename[512];
    char progsec[32];
    char src_mac[18];
    char dest_mac[18];
    __u16 xsk_bind_flags;
    int xsk_if_queue;
    bool xsk_poll_mode;
};