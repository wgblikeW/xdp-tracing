package handler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/p1nant0m/xdp-tracing/handler/utils"
	"github.com/sirupsen/logrus"
)

const (
	SPACE = " "
)

// using in Filter
type PacketStatus int

const (
	DROP = iota
	PASS
)

const (
	None = iota
	IPv4Packet
	IPv6Packet
)

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

type PayloadMeta struct {
	Payload    *[]byte
	PayloadLen uint32
}

// TCP_IP_Handler Struct contains the field that we need in observing
type TCP_IP_Handler struct {
	Timestamp string

	// IP Header Field
	SrcIP net.IP
	DstIP net.IP
	TTL   uint8

	// TCP Header Field
	TcpFlagsS string
	SrcPort   layers.TCPPort
	DstPort   layers.TCPPort

	// Application Payload
	PayloadExist bool
	*PayloadMeta
}

// NewTCPFlags return the pointer of new tcpFlags Struct with TCP flags Settings
func NewTCPFlags(tcp *layers.TCP) *tcpFlags {
	return &tcpFlags{FIN: tcp.FIN, ACK: tcp.ACK, RST: tcp.RST,
		PSH: tcp.PSH, URG: tcp.URG, ECE: tcp.ECE, CWR: tcp.CWR, NS: tcp.NS, SYN: tcp.SYN}
}

// parseFalgsToString parses the TCP Flags settings and returns string contain flags that were set
func parseFlagsToString(flags interface{}) string {
	ret := ""
	v := reflect.ValueOf(flags).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface().(bool) {
			ret += v.Type().Field(i).Name + SPACE
		}
	}
	if len(ret) != 0 {
		return ret[0 : len(ret)-1] // remove the last space
	}

	return ret
}

// NewTCPIPHandler returns the pointer of strcut of TCP_IP_Handler
func NewTCPIPHandler() *TCP_IP_Handler {
	return &TCP_IP_Handler{
		PayloadMeta: &PayloadMeta{},
	}
}

func (handler *TCP_IP_Handler) copy() *TCP_IP_Handler {
	duplicate := &TCP_IP_Handler{
		Timestamp:    handler.Timestamp,
		SrcIP:        handler.SrcIP,
		SrcPort:      handler.SrcPort,
		DstIP:        handler.DstIP,
		DstPort:      handler.DstPort,
		TTL:          handler.TTL,
		TcpFlagsS:    handler.TcpFlagsS,
		PayloadExist: handler.PayloadExist,
		PayloadMeta: &PayloadMeta{
			Payload:    handler.Payload,
			PayloadLen: handler.PayloadLen,
		},
	}
	return duplicate
}

// hasTCPLayerAndRetrieve returns *layers.TCP if it exists in the raw packet
func (handler *TCP_IP_Handler) hasTCPLayerAndRetrieve(packet gopacket.Packet) (*layers.TCP, error) {
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		return tcpLayer.(*layers.TCP), nil
	}
	return nil, errors.New("no valid TCP layers found")
}

// hasIPLayerAndRetrieve returns *layers.IPv4/*layer.IPv6 if it exists in the raw packet
func (handler *TCP_IP_Handler) hasIPLayerAndRetrieve(packet gopacket.Packet) (gopacket.Layer, int, error) {
	if ipV4Layer := packet.Layer(layers.LayerTypeIPv4); ipV4Layer != nil {
		return ipV4Layer, IPv4Packet, nil
	} else if ipV6Layer := packet.Layer(layers.LayerTypeIPv6); ipV6Layer != nil {
		return ipV6Layer, IPv6Packet, nil
	}
	return nil, None, errors.New("no valid IP layers found")
}

// resolveIPv4Field fixs in the field related to IPv4 Header
func (handler *TCP_IP_Handler) resolveIPv4Field(ipLayer *layers.IPv4) {
	handler.SrcIP = ipLayer.SrcIP
	handler.DstIP = ipLayer.DstIP
	handler.TTL = ipLayer.TTL
}

//TODO: Adding support IPv6
func (handler *TCP_IP_Handler) resolveIPv6Field(ipLayer *layers.IPv6) {

}

// resolveTCPField fixs in the field related to TCP Header
func (handler *TCP_IP_Handler) resolveTCPField(tcpLayer *layers.TCP) {
	handler.SrcPort = tcpLayer.SrcPort
	handler.DstPort = tcpLayer.DstPort
	// resolve TCP Flags
	tcpFlags := NewTCPFlags(tcpLayer)
	handler.TcpFlagsS = parseFlagsToString(tcpFlags)
}

// Handle Implement the Handler Interface and it fills up the field in TCP_IP_Handler struct
func (handler *TCP_IP_Handler) Handle(packet gopacket.Packet) error {
	//TODO: Adding support for IPv6
	ipLayer, version, err := handler.hasIPLayerAndRetrieve(packet)
	if err != nil {
		return err
	}

	tcpLayer, err := handler.hasTCPLayerAndRetrieve(packet)
	if err != nil {
		return err
	}

	// resolve IP Header Field
	switch version {
	case IPv4Packet:
		handler.resolveIPv4Field(ipLayer.(*layers.IPv4))
	case IPv6Packet:
		handler.resolveIPv6Field(ipLayer.(*layers.IPv6))
	}

	// resolve TCP Header Field
	handler.resolveTCPField(tcpLayer)

	// resolve TCP Application Payload
	if appLayer := packet.ApplicationLayer(); appLayer != nil {
		payload := appLayer.Payload()
		handler.PayloadExist = true

		// PayloadMeta Settings
		handler.Payload = &payload
		handler.PayloadLen = uint32(len(payload))
	} else {
		handler.PayloadExist = false
	}

	handler.Timestamp = time.Now().Format("2006-01-02 15:04:05.999999999")

	return nil
}

func find(uint32List []uint32, elem *reflect.Value) int {
	switch elem.Kind() {
	case reflect.Uint16:
		// Process Port Field
		for _, port := range uint32List {
			if (uint16)(port) == (uint16)(elem.Interface().(layers.TCPPort)) {
				// Match the Rules, access the packet
				logrus.Debugf("Port Match Port:%v", port)
				return PASS
			}
		}
	case reflect.Slice:
		// Process Address Field
		for _, address := range uint32List {
			if address == uint32(utils.BytesToUInt32(([]byte)(elem.Interface().(net.IP)))) {
				// Match the Rules, access the packet
				return PASS
			}
		}
	}
	return DROP
}

var support_rules_field = []string{"SrcIP", "DstIP", "SrcPort", "DstPort"}

// Filter Should be called after TCP_IP_Handler Struct is fully constructed(call Handle())
func (handler *TCP_IP_Handler) Filter(rules map[string][]uint32) PacketStatus {
	flag := 0
	for _, field := range support_rules_field {
		if uint32List, ok := rules[field]; ok {
			v := reflect.ValueOf(handler).Elem().FieldByName(field)
			if len(uint32List) == 0 {
				// Empty rule
				continue
			}
			flag |= find(uint32List, &v) // It can pass as it matches one of the rules
		}
	}

	if flag == 1 {
		return PASS
	}

	return DROP
}

func MakeTCPIPRules(rules map[string][]string) map[string][]uint32 {
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

func StartTCPIPHandler(ctx context.Context, rules map[string][]string, observerCh chan<- *TCP_IP_Handler) {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, utils.Htons(syscall.ETH_P_ALL))
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}
	fmt.Println("Listening on Raw Socket")
	defer syscall.Close(fd)
	tcpHandler := NewTCPIPHandler()
	buf := make([]byte, 4096)
	logrus.Debug("In StartTCPIPHandler:274 rulesRaw:%v", rules)
	rulesApplied := MakeTCPIPRules(rules)
	logrus.Debug("In StartTCPIPHandler:274 rulesApplied:%v", rulesApplied)

	for {
		// long-routine
		_, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			panic(err)
		}

		packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
		err = tcpHandler.Handle(packet)
		if tcpHandler.Filter(rulesApplied) == DROP {
			continue
		}

		if err == nil {
			logrus.Debug("TCPHandler Filter Receive Packets")
			observerCh <- tcpHandler.copy()
		}

		select {
		case <-ctx.Done():
			fmt.Println("Stopping Capturing the Packets")
			return
		default:
		}
	}
}
