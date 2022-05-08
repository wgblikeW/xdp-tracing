#include <sys/resource.h>
#include <bpf/libbpf.h>
#include <bpf/bpf.h>
#include <net/if.h>
#include <linux/if_link.h>
#include <common_defines.h>
#include "xdp-proxy.skel.h"
#include "load-bpf.h"

/* Attach to ens33 by default */
#define DEV_NAME "ens33"

/*
    @param idx: the network interface controller index that you want attach to
    @param fd: the BPF program index when it is loaded
    @param name: the BPF program name
    @param name: XDP attached flag (different attach mode) 
    @return error: error info when doing attach, 0 if there is no error
 */
static int do_attach(int idx, int fd, const char *name, __u32 xdp_flags)
{
    struct bpf_prog_info info = {};
    __u32 info_len = sizeof(info);
    int err;
    /* I have no idea what this use for PATCH GIT: [https://www.spinics.net/lists/bpf/msg53136.html] */
    struct bpf_xdp_attach_opts xdp_attach_opts = {10, 0};

    err = bpf_xdp_attach(idx, fd, xdp_flags, &xdp_attach_opts);
    if (err < 0) {
        printf("ERROR: failed to attach program to %s\n", name);
        return err;
    }

    err = bpf_obj_get_info_by_fd(fd, &info, &info_len);
    if (err) {
        printf("can't get prog info - %s\n", strerror(errno));
        return err;
    }
    prog_id = info.id;

    return err;
}

int execute_bpf_prog() {
    struct xdp_proxy_bpf *obj;
    int err = 0;

    struct rlimit rlim_new = {RLIM_INFINITY, RLIM_INFINITY};
    err = setrlimit(RLIMIT_MEMLOCK, &rlim_new);
    if (err) {
        fprintf(stderr, "failed to change rlimit\n");
        return 1;
    }

    struct config cfg = {
        .xdp_flags = XDP_FLAGS_UPDATE_IF_NOEXIST | XDP_FLAGS_SKB_MODE,
        .ifindex = -1,
    } 
    
    unsigned int ifindex = if_nametoindex(DEV_NAME);
    if (ifindex == 0) {
        fprintf(stderr, "failed to find interface %s\n", DEV_NAME);
        return 1;
    }

    obj = xdp_proxy_bpf__open();
    if (!obj)
    {
        fprintf(stderr, "failed to open and/or load BPF object\n");
        return 1;
    }

    err = xdp_proxy_bpf__load(obj);
    if (err)
    {
        fprintf(stderr, "failed to load BPF object %d\n", err);
        goto cleanup;
    }

    /* Attach the XDP program to ens33 */
    int prog_id = bpf_program__fd(obj->progs.xdp_proxy);
    /* I have no idea what this use for PATCH GIT: [https://www.spinics.net/lists/bpf/msg53136.html] */
    struct bpf_xdp_attach_opts xdp_attach_opts = {10, 0};

    err = do_attach(ifindex, prog_id, xdp_flags);
    if (err)
    {
        fprintf(stderr, "failed to attach BPF programs\n");
        goto cleanup;
    }
    

    printf("Successfully run! Tracing /sys/kernel/debug/tracing/trace_pipe.\n");
    system("cat /sys/kernel/debug/tracing/trace_pipe");

    

cleanup:
    /* detach and free XDP program on exit */
    bpf_xdp_detach(ifindex, xdp_flags, &xdp_attach_opts);
    xdp_proxy_bpf__destroy(obj);
    return err != 0;
}


// clang -g -O2 -Wall -I. -c xdp-proxy.c -o xdp-proxy.o
// clang -Wall -O2 -g xdp-proxy.o -static -lbpf -lelf -lz -o xdp-proxy