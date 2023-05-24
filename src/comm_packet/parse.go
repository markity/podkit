package commpacket

import (
	"encoding/binary"
	"encoding/json"
)

func DoPack(lengthIdentify int, data []byte) []byte {
	bs := make([]byte, lengthIdentify, lengthIdentify+len(data))
	binary.BigEndian.PutUint32(bs, uint32(len(data)))
	bs = append(bs, data...)
	return bs
}

func ClientParsePacket(data []byte) interface{} {
	header := MsgHeader{}
	err := json.Unmarshal(data, &header)
	if err != nil {
		return nil
	}

	switch header.Type {
	case TypeServerNotifyContainerClosed:
		p := ServerNotifyContainerClosed{}
		json.Unmarshal(data, &p)
		return &p
	case TypeServerSendPtyOutput:
		p := ServerSendPtyOutput{}
		json.Unmarshal(data, &p)
		return &p
	default:
		return nil
	}
}

func ServerParsePacket(data []byte) interface{} {
	header := MsgHeader{}
	err := json.Unmarshal(data, &header)
	if err != nil {
		return nil
	}

	switch header.Type {
	case TypeClientCloseContainer:
		p := ClientCloseContainer{}
		json.Unmarshal(data, &p)
		return &p
	case TypeClientExecBackground:
		p := ClientExecBackground{}
		json.Unmarshal(data, &p)
		return &p
	case TypeClientExecInteractive:
		p := ClientExecInteractive{}
		json.Unmarshal(data, &p)
		return &p
	case TypeClientSendPtyInput:
		p := ClientSendPtyInput{}
		json.Unmarshal(data, &p)
		return &p
	default:
		return nil
	}
}
