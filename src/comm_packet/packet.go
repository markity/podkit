package commpacket

import "encoding/json"

// 里面包含了与shim client交互的json数据结构体

type MsgType int

const (
	// 客户端连接上后的请求
	// 请求开启交互式命令
	TypePacketClientExecInteractiveRequest MsgType = iota
	// 告知客户端交互式命令已经开启或找不到这个命令
	TypePacketServerExecInteractiveResponse
	// 交互式命令的输入输出
	TypePacketClientSendPtyInput
	TypePacketServerSendPtyOutput
	// 告知客户端交互式命令正常结束, 在告知客户端已经开启命令后, 会发这个包让客户端结束运行
	// 但是也能发送ServerNotifyInteractiveExecContainerClosed通知这个容器已经关闭
	TypePacketServerNotifyExecInteractiveExited
	// 通知正在进行交互式命令的客户端容器已经被关闭了
	TypePacketServerNotifyExecInteractiveContainerClosed

	// 请求开启守护进程命令
	TypePacketClientExecBackgroundRequest
	// 告知客户端守护进程命令已经开启或者已经结束, 守护进程的stdin/out/err都是null设备
	TypePacketServerExecBackgroundResponse

	// 请求关闭容器
	TypePacketClientCloseContainerRequest
	// 通知连接已经成功关闭
	TypePacketServerContainerClosedOK
)

// 标识包的类型
type MsgHeader struct {
	Type MsgType `json:"type"`
}

// 请求1, 客户端请求交互式执行
type PacketClientExecInteractiveRequest struct {
	MsgHeader
	Command string `json:"cmd"`
	Rows    int    `json:"rows"`
	Cols    int    `json:"cols"`
}

func (p *PacketClientExecInteractiveRequest) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketClientExecInteractiveRequest
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 用来做持续的交互
type PacketClientSendPtyInput struct {
	MsgHeader
	Data string
}

func (p *PacketClientSendPtyInput) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketClientSendPtyInput
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 用来做持续的交互
type PacketServerSendPtyOutput struct {
	MsgHeader
	Data string
}

func (p *PacketServerSendPtyOutput) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketServerSendPtyOutput
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 请求2, 客户端发来的消息是用来exec -d
type PacketClientExecBackgroundRequest struct {
	MsgHeader
	Command string `json:"cmd"`
}

func (p *PacketClientExecBackgroundRequest) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketClientExecBackgroundRequest
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 请求2, 客户端发来的消息是用来exec -d
type PacketServerExecBackgroundResponse struct {
	MsgHeader
	CommandExists bool `json:"command_exists"`
}

func (p *PacketServerExecBackgroundResponse) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketServerExecBackgroundResponse
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 请求3, 客户端要求关闭容器
type PacketClientCloseContainerRequest struct {
	MsgHeader
}

func (p *PacketClientCloseContainerRequest) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketClientCloseContainerRequest
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 告知用户容器已经关闭
type PacketServerNotifyExecInteractiveContainerClosed struct {
	MsgHeader
}

func (p *PacketServerNotifyExecInteractiveContainerClosed) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketServerNotifyExecInteractiveContainerClosed
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 告知用户容器已经关闭
type PacketServerContainerClosedOK struct {
	MsgHeader
}

func (p *PacketServerContainerClosedOK) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketServerContainerClosedOK
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

type PacketServerExecInteractiveResponse struct {
	MsgHeader
	CommandExists bool `json:"command_exists"`
}

func (p *PacketServerExecInteractiveResponse) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketServerExecInteractiveResponse
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

type PacketServerNotifyExecInteractiveExited struct {
	MsgHeader
	CommandExists bool `json:"command_exists"`
}

func (p *PacketServerNotifyExecInteractiveExited) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypePacketServerNotifyExecInteractiveExited
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}
