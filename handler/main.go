package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	handler "github.com/p1nant0m/xdp-tracing/handler/1"
	"github.com/p1nant0m/xdp-tracing/handler/utils"
)

// Convert a uint16 to host byte order (big endian)
func Htons(v uint16) int {
	return int((v << 8) | (v >> 8))
}

const (
	SPACE = " "
)

func main() {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, Htons(syscall.ETH_P_ALL))
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}
	fmt.Println("Listening on Raw Socket")
	defer syscall.Close(fd)
	tcpHandler := handler.NewTCPIPHandler()
	rules := make(map[string][]uint32)
	rules["SrcIP"] = make([]uint32, 10)
	rules["SrcIP"] = append(rules["SrcIP"], utils.BytesToUInt32(net.ParseIP("192.168.176.1").To4()))
	buf := make([]byte, 4096)

	for {
		_, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			panic(err)
		}

		packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
		err = tcpHandler.Handle(packet)
		if tcpHandler.Filter(rules) == handler.DROP {
			continue
		}
		if err == nil {
			fmt.Printf("[%s] %s:%d -> %s:%d [%s] TTL:%d\n", tcpHandler.Timestamp, tcpHandler.SrcIP, tcpHandler.SrcPort, tcpHandler.DstIP, tcpHandler.DstPort, tcpHandler.TcpFlagsS, tcpHandler.TTL)
			if tcpHandler.PayloadExist {
				fmt.Println(hex.Dump(*tcpHandler.Payload))
			}
		}
	}
}
