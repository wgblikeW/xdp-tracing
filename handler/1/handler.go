package handler

import "github.com/google/gopacket"

type PacketHandler interface {
	Handle(gopacket.Packet) error
}
