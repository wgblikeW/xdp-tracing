package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	FONT_SETTINGS_PREFIX = "\033[1;42;37m"
	FONT_SETTINGS_SUFFIX = "\033[0m"
)

func FontSet(str string) string {
	return FONT_SETTINGS_PREFIX + str + FONT_SETTINGS_SUFFIX
}

func UInt32ToBytes(n uint32) []byte {
	data := n
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func BytesToUInt32(bys []byte) uint32 {
	bytebuff := bytes.NewBuffer(bys)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

func Htons(v uint16) int {
	return int((v << 8) | (v >> 8))
}

// localIPObtain used to obtain the IPv4 Address of the local machine
func LocalIPObtain() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
