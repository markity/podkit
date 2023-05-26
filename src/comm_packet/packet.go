package commpacket

import "encoding/json"

// 里面包含了与shim client交互的json数据结构体

type MsgType int

const (
	// 客户端连接上后的请求
	// 请求开启交互式命令
	TypeClientRequestExecInteractive MsgType = iota
	// 告知客户端交互式命令已经开启或找不到这个命令
	TypeServerInteractiveCommandResp
	// 交互式命令的输入输出
	TypeClientSendPtyInput
	TypeServerSendPtyOutput
	// 告知客户端交互式命令正常结束, 在告知客户端已经开启命令后, 会发这个包让客户端结束运行
	// 但是也能发送ServerNotifyInteractiveExecContainerClosed通知这个容器已经关闭
	TypeServerInteractiveCommandExited
	// 通知正在进行交互式命令的客户端容器已经被关闭了
	TypeServerNotifyInteractiveExecContainerClosed

	// 请求开启守护进程命令
	TypeClientRequestExecBackground
	// 告知客户端守护进程命令已经开启或者已经结束, 守护进程的stdin/out/err都是null设备
	TypeServerExecBackgroundResp

	// 请求关闭容器
	TypeClientRequestCloseContainer
	// 通知连接已经成功关闭
	TypeServerNotifyContainerClosedSuccesfully
)

// 标识包的类型
type MsgHeader struct {
	Type MsgType `json:"type"`
}

// 请求1, 客户端请求交互式执行
type ClientRequestExecInteractive struct {
	MsgHeader
	Command string `json:"cmd"`
	Rows    int    `json:"rows"`
	Cols    int    `json:"cols"`
}

func (p *ClientRequestExecInteractive) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeClientRequestExecInteractive
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
type ClientRequestExecBackground struct {
	MsgHeader
	Command string `json:"cmd"`
}

func (p *ClientRequestExecBackground) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeClientRequestExecBackground
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 请求2, 客户端发来的消息是用来exec -d
type ServerExecBackgroundResp struct {
	MsgHeader
	CommandExists bool `json:"command_exists"`
}

func (p *ServerExecBackgroundResp) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeServerExecBackgroundResp
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 请求3, 客户端要求关闭容器
type ClientRequestCloseContainer struct {
	MsgHeader
}

func (p *ClientRequestCloseContainer) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeClientRequestCloseContainer
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 告知用户容器已经关闭
type ServerNotifyInteractiveExecContainerClosed struct {
	MsgHeader
}

func (p *ServerNotifyInteractiveExecContainerClosed) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeServerNotifyInteractiveExecContainerClosed
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

// 告知用户容器已经关闭
type ServerNotifyContainerClosedSuccesfully struct {
	MsgHeader
}

func (p *ServerNotifyContainerClosedSuccesfully) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeServerNotifyContainerClosedSuccesfully
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

type ServerInteractiveCommandResp struct {
	MsgHeader
	CommandExists bool `json:"command_exists"`
}

func (p *ServerInteractiveCommandResp) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeServerInteractiveCommandResp
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}

type ServerInteractiveCommandExited struct {
	MsgHeader
	CommandExists bool `json:"command_exists"`
}

func (p *ServerInteractiveCommandExited) MustMarshalToBytes() []byte {
	p.MsgHeader.Type = TypeServerInteractiveCommandExited
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bs
}
