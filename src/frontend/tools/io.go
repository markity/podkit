package tools

import (
	"encoding/binary"
	"io"
)

func ReadPacketWith4BytesLengthHeader(reader io.Reader) ([]byte, error) {
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(reader, lengthBytes)
	if err != nil {
		return nil, err
	}

	data := make([]byte, binary.BigEndian.Uint32(lengthBytes))
	_, err = io.ReadFull(reader, data)
	return data, err
}
