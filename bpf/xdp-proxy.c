#include <sys/resource.h>
#include <bpf/libbpf.h>
#include <bpf/bpf.h>
#include <net/if.h>
#include <linux/if_link.h>
#include "headers/common_defines.h"
#include "headers/load-bpf.h"

/* Attach to ens33 by default */
#define DEV_NAME "ens33"

enum
{
    /* Error Code */
    OK,
    ERROR_DETACH_PROG_FROM_INTERF,
    ERROR_BPF_GET_LINK_XDP_ID,
    ERROR_NOT_FOUND_BPF_PROG_ON_INTERF,
    ERROR_NOT_EXPECTED_BPF_PROG_FOUND,
    ERROR_ATTACH_PROG_TO_INTERF,
    ERROR_NOT_FOUND_INTERFACE,
    ERROR_OPENING_BPF_OBJECT,
    ERROR_LOADING_BPF_OBJECT,
};

/*  attach XDP type BPF program to interface
    @param idx: the network interface controller index that you want attach to
    @param fd: the BPF program index when it is loaded
    @param name: the if name
    @param xdp_flags: XDP attached flag (different attach mode) 
    @return error: error info when doing attach, 0 if there is no error
 */
int do_attach(int idx, int fd, __u32 xdp_flags)
{
    struct bpf_prog_info info = {};
    __u32 info_len = sizeof(info);
    int err;
    /* I have no idea what this use for PATCH GIT: [https://www.spinics.net/lists/bpf/msg53136.html] */
    struct bpf_xdp_attach_opts xdp_attach_opts = {10, 0};

    err = bpf_xdp_attach(idx, fd, xdp_flags, &xdp_attach_opts);
    if (err < 0) {
        return ERROR_ATTACH_PROG_TO_INTERF;
    }

    // err = bpf_obj_get_info_by_fd(fd, &info, &info_len);
    // if (err) {
    //     printf("can't get prog info - %s\n", strerror(errno));
    //     return err;
    // }

    return OK;
}

/* do_detach detach the BPF program from given interface 
    @param idx: the index of interface
    @param name: the name of interface
    @param tar_prog_id: the index of BPF prog you want to detach [check using bpftool prog]
*/
int do_detach(int idx, int tar_prog_id) {
    __u32 bpf_prog_id = 0;
    int err = 0;
    struct bpf_xdp_attach_opts opts = {10, 0};

    err = bpf_get_link_xdp_id(idx, &bpf_prog_id, 0);
    if (err) {
        return ERROR_BPF_GET_LINK_XDP_ID;
    }

    if (tar_prog_id == bpf_prog_id) {
        // found a expected BPF program, detach it from interface
        err = bpf_xdp_detach(idx, 0, &opts);
        if (err < 0) {
            return ERROR_DETACH_PROG_FROM_INTERF;
        }
    } else if (!bpf_prog_id) {
        // couldn't find BPF program on given interface
        return ERROR_NOT_FOUND_BPF_PROG_ON_INTERF;
    } else {
        // coundn't find expected BPF program on given interface
        return ERROR_NOT_EXPECTED_BPF_PROG_FOUND;
    }

    return OK;
}

int attach_bpf_prog_to_if(char *if_name, __u32 xdp_flags, char *filename)
{
    struct rlimit r = {RLIM_INFINITY, RLIM_INFINITY};
    struct bpf_prog_load_attr prog_load_attr = {
        .prog_type = BPF_PROG_TYPE_XDP,
    };
    int prog_fd, map_fd;
    struct bpf_object *obj;
    struct bpf_map *map;
    int ret, err, i;
    struct config cfg = {
        .xdp_flags = xdp_flags,
        .ifindex = if_nametoindex(if_name),
    };

    if (cfg.ifindex == 0) {
        return ERROR_NOT_FOUND_INTERFACE;
    }
    return 0;
}

// clang -g -O2 -Wall -I. -c xdp-proxy.c -o xdp-proxy.o
// clang -Wall -O2 -g xdp-proxy.o -static -lbpf -lelf -lz -o xdp-proxy