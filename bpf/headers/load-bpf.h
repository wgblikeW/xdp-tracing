/*
 * Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

#include <linux/types.h>
int do_detach(int, int);
int do_attach(int, int, __u32);

struct input_args
{
    __u32 xdp_flags;
    char *ifname;
    char *filename;
};
int attach_bpf_prog_to_if(struct input_args inputs);
int bpf_update_map(__u32, unsigned int);
int bpf_revoke_map(__u32, unsigned int);