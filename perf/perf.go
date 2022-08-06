// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package perf

import (
	"time"

	"github.com/p1nant0m/xdp-tracing/handler/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

const (
	SAMPLING_PERIOD = 5
)

var (
	previousByteSend  uint64 = 0
	previousByteReve  uint64 = 0
	previousByteRead  uint64 = 0
	previousByteWrite uint64 = 0
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

	// Basic information about the system
	Hostname string   `json:"hostname"`
	HostIpv4 string   `json:"hostaddr"`
	Platform string   `json:"platform"`
	OpenPort []uint32 `json:"openport"`

	// Instant information about the system
	CPUUsage float64 `json:"cpuusage"`
	MemUsage float64 `json:"memusage"`

	// Given specific infomation about system resources in a period of time (SAMPLE_PERIOD)
	BytesSent  uint64 `json:"byteSent"`
	BytesRecv  uint64 `json:"byteRecv"`
	ReadBytes  uint64 `json:"readBytes"`
	WriteBytes uint64 `json:"writeBytes"`
}

func GetHostPerf() *HostInfo {
	hostInfo, _ := host.Info()
	cpuUsage, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	memUsage := memInfo.UsedPercent
	netStats, _ := net.Connections("tcp")
	counterStat, _ := net.IOCounters(false)
	diskCounter, _ := disk.IOCounters("sda3")

	var openPort Set[uint32] = Set[uint32]{
		m: make(map[uint32]Empty),
	}

	for _, netStat := range netStats {
		openPort.Add(netStat.Laddr.Port)
	}

	host := &HostInfo{
		Hostname:   hostInfo.Hostname,
		HostIpv4:   utils.LocalIPObtain(),
		Platform:   hostInfo.OS + "-" + hostInfo.Platform + "-" + hostInfo.PlatformVersion,
		OpenPort:   openPort.ToList(),
		CPUUsage:   cpuUsage[0],
		MemUsage:   memUsage,
		BytesSent:  counterStat[0].BytesSent - previousByteSend,
		BytesRecv:  counterStat[0].BytesRecv - previousByteReve,
		ReadBytes:  diskCounter["sda3"].ReadBytes - previousByteRead,
		WriteBytes: diskCounter["sda3"].WriteBytes - previousByteWrite,
	}

	previousByteSend = counterStat[0].BytesSent
	previousByteReve = counterStat[0].BytesRecv
	previousByteRead = diskCounter["sda3"].ReadBytes
	previousByteReve = diskCounter["sda3"].WriteBytes
	return host
}
