package commpacket

import (
	"encoding/json"
)

func ClientParsePacket(data []byte) interface{} {
	header := MsgHeader{}
	err := json.Unmarshal(data, &header)
	if err != nil {
		return nil
	}

	switch header.Type {
	case TypePacketServerNotifyExecInteractiveContainerClosed:
		p := PacketServerNotifyExecInteractiveContainerClosed{}
		json.Unmarshal(data, &p)
		return &p
	case TypePacketServerSendPtyOutput:
		p := PacketServerSendPtyOutput{}
		json.Unmarshal(data, &p)
		return &p
	case TypePacketServerExecBackgroundResponse:
		p := PacketServerExecBackgroundResponse{}
		json.Unmarshal(data, &p)
		return &p
	case TypePacketServerContainerClosedOK:
		p := PacketServerContainerClosedOK{}
		json.Unmarshal(data, &p)
		return &p
	case TypePacketServerNotifyExecInteractiveExited:
		p := PacketServerNotifyExecInteractiveExited{}
		json.Unmarshal(data, &p)
		return &p
		// 共6个
	case TypePacketServerExecInteractiveResponse:
		p := PacketServerExecInteractiveResponse{}
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
	case TypePacketClientCloseContainerRequest:
		p := PacketClientCloseContainerRequest{}
		json.Unmarshal(data, &p)
		return &p
	case TypePacketClientExecBackgroundRequest:
		p := PacketClientExecBackgroundRequest{}
		json.Unmarshal(data, &p)
		return &p
	case TypePacketClientExecInteractiveRequest:
		p := PacketClientExecInteractiveRequest{}
		json.Unmarshal(data, &p)
		return &p
		// 共4个
	case TypePacketClientSendPtyInput:
		p := PacketClientSendPtyInput{}
		json.Unmarshal(data, &p)
		return &p
	default:
		return nil
	}
}
