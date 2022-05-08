package config

const (
	XDP_FLAGS_UPDATE_IF_NOEXIST uint32 = 1 << iota
	XDP_FLAGS_SKB_MODE
	XDP_FLAGS_DRV_MODE
	XDP_FLAGS_HW_MODE
	XDP_FLAGS_REPLACE
)

type Config struct {
	Xdp_flags uint32
	IfIndex   int
	Ifname    string
}

// NewConfig returns a Config struct with the default value
func NewConfig() *Config {
	return &Config{
		Xdp_flags: XDP_FLAGS_UPDATE_IF_NOEXIST | XDP_FLAGS_SKB_MODE,
		IfIndex:   -1,
		Ifname:    "eth0",
	}
}
