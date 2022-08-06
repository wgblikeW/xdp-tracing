/*
 * Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

#define MAX_ENTRIES 1024
#include <linux/bpf.h>

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
    __uint(max_entries, MAX_ENTRIES);
} bridge SEC(".maps");