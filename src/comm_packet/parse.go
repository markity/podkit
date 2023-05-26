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
	case TypeServerNotifyInteractiveExecContainerClosed:
		p := ServerNotifyInteractiveExecContainerClosed{}
		json.Unmarshal(data, &p)
		return &p
	case TypeServerSendPtyOutput:
		p := ServerSendPtyOutput{}
		json.Unmarshal(data, &p)
		return &p
	case TypeServerExecBackgroundResp:
		p := ServerExecBackgroundResp{}
		json.Unmarshal(data, &p)
		return &p
	case TypeServerNotifyContainerClosedSuccesfully:
		p := ServerNotifyContainerClosedSuccesfully{}
		json.Unmarshal(data, &p)
		return &p
	case TypeServerInteractiveCommandExited:
		p := ServerInteractiveCommandExited{}
		json.Unmarshal(data, &p)
		return &p
	case TypeServerInteractiveCommandResp:
		p := ServerInteractiveCommandResp{}
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
	case TypeClientRequestCloseContainer:
		p := ClientRequestCloseContainer{}
		json.Unmarshal(data, &p)
		return &p
	case TypeClientRequestExecBackground:
		p := ClientRequestExecBackground{}
		json.Unmarshal(data, &p)
		return &p
	case TypeClientRequestExecInteractive:
		p := ClientRequestExecInteractive{}
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
