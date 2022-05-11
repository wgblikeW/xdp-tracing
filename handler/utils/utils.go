package utils

import (
	"bytes"
	"encoding/binary"
)

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
