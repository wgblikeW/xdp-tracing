package perf

import (
	"time"

	"github.com/p1nant0m/xdp-tracing/handler/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type Empty struct{}

var empty Empty

type Set[T uint32 | int | string | uint64 | uint16 | uint8] struct {
	m map[T]Empty
}

func (s *Set[T]) Add(val T) {
	s.m[val] = empty
}

func (s *Set[T]) Remove(val T) {
	delete(s.m, val)
}

func (s *Set[T]) Len() int {
	return len(s.m)
}

func (s *Set[T]) ToList() []T {
	vals := make([]T, 0, s.Len())

	for v := range s.m {
		vals = append(vals, v)
	}
	return vals
}

type HostInfo struct {
	Hostname string   `json:"hostname"`
	HostIpv4 string   `json:"hostaddr"`
	Platform string   `json:"platform"`
	OpenPort []uint32 `json:"openport"`
	CPUUsage float64  `json:"cpuusage"`
	MemUsage float64  `json:"memusage"`
}

func GetHostPerf() *HostInfo {
	hostInfo, _ := host.Info()
	cpuUsage, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	memUsage := memInfo.UsedPercent
	netStats, _ := net.Connections("tcp")
	var openPort Set[uint32] = Set[uint32]{
		m: make(map[uint32]Empty),
	}
	for _, netStat := range netStats {
		openPort.Add(netStat.Laddr.Port)
	}

	host := &HostInfo{
		Hostname: hostInfo.Hostname,
		HostIpv4: utils.LocalIPObtain(),
		Platform: hostInfo.OS + "-" + hostInfo.Platform + "-" + hostInfo.PlatformVersion,
		OpenPort: openPort.ToList(),
		CPUUsage: cpuUsage[0],
		MemUsage: memUsage,
	}

	return host
}
