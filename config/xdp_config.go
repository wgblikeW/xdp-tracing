// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

/* */
import "C"

// this Struct should be alias with input_args in load-bpf.h
type XDPConfig struct {
	Xdp_flags uint32
	Ifname    string
	Filename  string
}

type C__XDP_Config struct {
	Xdp_flags C.uint
	Filename  *C.char
	if_name   *C.char
}

const (
	XDP_FLAGS_UPDATE_IF_NOEXIST uint32 = 1 << iota
	XDP_FLAGS_SKB_MODE
	XDP_FLAGS_DRV_MODE
	XDP_FLAGS_HW_MODE
	XDP_FLAGS_REPLACE
)

// NewConfig returns a Config struct with the default value
func NewXDPConfig() *XDPConfig {
	return &XDPConfig{
		Xdp_flags: XDP_FLAGS_UPDATE_IF_NOEXIST | XDP_FLAGS_SKB_MODE,
		Ifname:    "eth0",
		Filename:  "xdp_proxy_kern.o",
	}
}
