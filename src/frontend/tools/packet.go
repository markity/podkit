package tools

import "encoding/binary"

func DoPackWith4Bytes(data []byte) []byte {
	bs := make([]byte, 4, 4+len(data))
	binary.BigEndian.PutUint32(bs, uint32(len(data)))
	bs = append(bs, data...)
	return bs
}
