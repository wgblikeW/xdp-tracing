// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package handler

import "github.com/google/gopacket"

type PacketHandler interface {
	Handle(gopacket.Packet) error
}
