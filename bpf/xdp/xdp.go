package xdp

import "github.com/p1nant0m/xdp-tracing/config"

type XDP_BPF_Prog struct {
	Name   string
	Config config.XDPConfig
}

func (xdp *XDP_BPF_Prog) Config_XDP_Prog(prog_name string, cfg config.XDPConfig) {

}
