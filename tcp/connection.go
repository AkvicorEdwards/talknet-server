package tcp

import (
	"net"
	"sync"
)

type Connection struct {
	// 用户ID
	UUID uint32
	// 用户名
	Username string
	// 昵称
	Nickname string
	// tcp连接
	Connection net.Conn
	// 待发送数据
	DataSend chan *Package
	// 序列号
	SEQ uint32
	// 序列号锁
	SEQMutex sync.Mutex
	// 心跳
	Heartbeat chan bool
	// 心跳重置
	ResetHeartbeat chan bool
	// 待回应的 Heartbeat SEQ
	HeartbeatSEQ uint32
	// 连接终止
	Termination chan bool
	// 连接重置
	Reconnect chan bool
	// 工作线程
	WorkerReq *WorkerStruct
	WorkerRes *WorkerStruct
}

func NewConnection(uuid uint32, username, nickname string, conn net.Conn) *Connection {
	return &Connection{
		UUID:         uuid,
		Username:     username,
		Nickname:     nickname,
		Connection:   conn,
		DataSend:     make(chan *Package, 2),
		SEQ:          1,
		SEQMutex:     sync.Mutex{},
		Heartbeat:    make(chan bool, 1),
		ResetHeartbeat:    make(chan bool, 1),
		HeartbeatSEQ: 0,
		Termination:  make(chan bool, 1),
		Reconnect:    make(chan bool, 1),
		WorkerReq:    NewWorker(),
		WorkerRes:    NewWorker(),
	}
}

type WorkerStruct struct {
	Server   chan bool
	Sender   chan bool
	Receiver chan bool
	Heartbeat chan bool
}

func NewWorker() *WorkerStruct {
	return &WorkerStruct{
		Server:   make(chan bool, 1),
		Sender:   make(chan bool, 1),
		Receiver: make(chan bool, 1),
		Heartbeat: make(chan bool, 1),
	}
}
