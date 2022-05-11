package handler

import (
	"errors"
	"net"
	"reflect"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	SPACE = " "
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
	Payload    []byte
	PayloadLen uint32
}

type TCP_IP_Handler struct {
	TcpFlagsS string
	// Timestamp
	// IP Header Field
	SrcIP net.IP
	DstIP net.IP
	TTL   uint8
	// TCP Header Field
	SrcPort      layers.TCPPort
	DstPort      layers.TCPPort
	PayloadExist bool
	*PayloadMeta
}

func NewTCPFlags(tcp *layers.TCP) *tcpFlags {
	return &tcpFlags{FIN: tcp.FIN, ACK: tcp.ACK, RST: tcp.RST,
		PSH: tcp.PSH, URG: tcp.URG, ECE: tcp.ECE, CWR: tcp.CWR, NS: tcp.NS, SYN: tcp.SYN}
}

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

func (handler *TCP_IP_Handler) hasTCPLayerAndRetrieve(packet gopacket.Packet) (*layers.TCP, error) {
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		return tcpLayer.(*layers.TCP), nil
	}
	return nil, errors.New("no valid TCP layers found")
}

func (handler *TCP_IP_Handler) hasIPLayerAndRetrieve(packet gopacket.Packet) (gopacket.Layer, int, error) {
	if ipV4Layer := packet.Layer(layers.LayerTypeIPv4); ipV4Layer != nil {
		return ipV4Layer, IPv4Packet, nil
	} else if ipV6Layer := packet.Layer(layers.LayerTypeIPv6); ipV6Layer != nil {
		return ipV6Layer, IPv6Packet, nil
	}
	return nil, None, errors.New("no valid IP layers found")
}

func (handler *TCP_IP_Handler) resolveIPv4Field(ipLayer *layers.IPv4) {
	handler.SrcIP = ipLayer.SrcIP
	handler.DstIP = ipLayer.DstIP
	handler.TTL = ipLayer.TTL
}

//TODO: Adding support IPv6
func (handler *TCP_IP_Handler) resolveIPv6Field(ipLayer *layers.IPv6) {

}

func (handler *TCP_IP_Handler) resolveTCPField(tcpLayer *layers.TCP) {
	handler.SrcPort = tcpLayer.SrcPort
	handler.DstPort = tcpLayer.DstPort
	// resolve TCP Flags
	tcpFlags := NewTCPFlags(tcpLayer)
	handler.TcpFlagsS = parseFlagsToString(tcpFlags)
}

func (handler *TCP_IP_Handler) Handle(packet gopacket.Packet) (interface{}, error) {
	//TODO: Adding support for IPv6
	ipLayer, version, err := handler.hasIPLayerAndRetrieve(packet)
	if err != nil {
		return nil, err
	}

	tcpLayer, err := handler.hasTCPLayerAndRetrieve(packet)
	if err != nil {
		return nil, err
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
		handler.PayloadMeta = &PayloadMeta{
			Payload:    payload,
			PayloadLen: uint32(len(payload)),
		}
	} else {
		handler.PayloadExist = false
	}

	return handler, nil
}
