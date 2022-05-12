package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/p1nant0m/xdp-tracing/handler"
	"github.com/p1nant0m/xdp-tracing/handler/utils"
)

const (
	SPACE = " "
)

// Convert a uint16 to host byte order (big endian)
func Htons(v uint16) int {
	return int((v << 8) | (v >> 8))
}

func MakeRules(rules map[string][]string) map[string][]uint32 {
	rulesApplied := make(map[string][]uint32)
	for key := range rules {
		rulesApplied[key] = make([]uint32, 10)
		switch {
		case key == "SrcIP" || key == "DstIP":
			for _, address := range rules[key] {
				rulesApplied[key] = append(rulesApplied[key], utils.BytesToUInt32(net.ParseIP(address).To4()))
			}
		case key == "SrcPort" || key == "DstPort":
			for _, portStr := range rules[key] {
				portInt, _ := strconv.Atoi(portStr)
				rulesApplied[key] = append(rulesApplied[key], uint32(portInt))
			}
		}
	}
	return rulesApplied
}

func main() {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, Htons(syscall.ETH_P_ALL))
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}
	fmt.Println("Listening on Raw Socket")
	defer syscall.Close(fd)
	tcpHandler := handler.NewTCPIPHandler()
	buf := make([]byte, 4096)
	// custom rules
	rules := make(map[string][]string)
	rules["SrcIP"] = append(rules["SrcIP"], "192.168.176.1")
	rules["SrcPort"] = append(rules["SrcPort"], "1080")
	rulesApplied := MakeRules(rules)

	for {
		// long-routine
		_, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			panic(err)
		}

		packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
		err = tcpHandler.Handle(packet)
		if tcpHandler.Filter(rulesApplied) == handler.DROP {
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
