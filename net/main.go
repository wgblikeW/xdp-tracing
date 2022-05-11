package main

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Convert a uint16 to host byte order (big endian)
func Htons(v uint16) int {
	return int((v << 8) | (v >> 8))
}

type tcpFlags struct {
	FIN bool
	SYN bool
	RST bool
	PSH bool
	ACK bool
	URG bool
	ECE bool
	CWR bool
	NS  bool
}

const (
	SPACE = " "
)

func parseFlagsToString(flags interface{}) string {
	ret := ""
	v := reflect.ValueOf(flags).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface().(bool) {
			ret += v.Type().Field(i).Name + SPACE
		}
	}
	if len(ret) != 0 {
		return ret[0 : len(ret)-1]
	}

	return ret
}

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
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil && (tcpLayer.(*layers.TCP).DstPort == 1080 || tcpLayer.(*layers.TCP).SrcPort == 1080) {
			tcp, _ := tcpLayer.(*layers.TCP)
			ip, _ := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
			flags := &tcpFlags{FIN: tcp.FIN, ACK: tcp.ACK, RST: tcp.RST,
				PSH: tcp.PSH, URG: tcp.URG, ECE: tcp.ECE, CWR: tcp.CWR, NS: tcp.NS, SYN: tcp.SYN}
			stringFlags := parseFlagsToString(flags)
			if len(stringFlags) != 0 {
				fmt.Printf("%s:%d -> %s:%d [%s] TTL:%d\n", ip.SrcIP, tcp.SrcPort, ip.DstIP, tcp.DstPort, stringFlags, ip.TTL)
			}
			if app := packet.ApplicationLayer(); app != nil {
				fmt.Println(hex.Dump(app.Payload()))
			}
		}
	}
}
