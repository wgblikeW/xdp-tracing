package main

import (
	"encoding/hex"
	"fmt"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	handler "github.com/p1nant0m/xdp-tracing/handler/1"
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

	for {
		buf := make([]byte, 4096)
		_, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			panic(err)
		}

		packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
		tcpHandler := &handler.TCP_IP_Handler{}
		tcpInfoInF, err := tcpHandler.Handle(packet)
		tcpInfo := tcpInfoInF.(*handler.TCP_IP_Handler)
		if err == nil {
			fmt.Printf("%s:%d -> %s:%d [%s] TTL:%d\n", tcpInfo.SrcIP, tcpInfo.SrcPort, tcpInfo.DstIP, tcpInfo.DstPort, tcpInfo.TcpFlagsS, tcpInfo.TTL)
			if tcpInfo.PayloadExist {
				fmt.Println(hex.Dump(tcpHandler.Payload))
			}
		}
	}
}
