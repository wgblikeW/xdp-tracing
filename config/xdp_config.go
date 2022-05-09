package config

// this Struct should be alias with input_args in load-bpf.h
type XDPConfig struct {
	Xdp_flags uint32
	Ifname    string
	Filename  string
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
		Filename:  "xdp-proxy.bpf.o",
	}
}
