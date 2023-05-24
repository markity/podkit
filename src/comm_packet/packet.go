package commpacket

import "encoding/json"

// 里面包含了与shim client交互的json数据结构体

type MsgType int

const (
	TypeClientExecInteractive MsgType = iota
	TypeClientSendPtyInput
	TypeServerSendPtyOutput
	TypeClientExecBackground
	TypeClientCloseContainer
	TypeServerNotifyContainerClosed
)

type MsgHeader struct {
	Type MsgType `json:"type"`
}

// 请求1, 客户端发来的消息是用来exec -i
type ClientExecInteractive struct {
	MsgHeader
	CommandPath string `json:"cmd_path"`
}

func (p *ClientExecInteractive) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeClientExecInteractive
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 用来做持续的交互
type ClientSendPtyInput struct {
	MsgHeader
	Data string
}

func (p *ClientSendPtyInput) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeClientSendPtyInput
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 用来做持续的交互
type ServerSendPtyOutput struct {
	MsgHeader
	Data string
}

func (p *ServerSendPtyOutput) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeServerSendPtyOutput
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 请求2, 客户端发来的消息是用来exec -d
type ClientExecBackground struct {
	MsgHeader
	CommandPath string
}

func (p *ClientExecBackground) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeClientExecBackground
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 请求3, 客户端要求关闭容器
type ClientCloseContainer struct {
	MsgHeader
}

func (p *ClientCloseContainer) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeClientCloseContainer
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 告知用户容器已经关闭
type ServerNotifyContainerClosed struct {
	MsgHeader
}

func (p *ServerNotifyContainerClosed) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeServerNotifyContainerClosed
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}
