package tcp

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"talknet/def"
	"talknet/operator"
	"time"
)

var Client = make(map[uint32]*Connection)
var ClientMutex = sync.Mutex{}

// 将用户的连接信息添加到全局Client
func AddClient(uuid uint32, cli *Connection) {
	ClientMutex.Lock()
	defer ClientMutex.Unlock()

	_, ok := Client[uuid]
	if ok {
		Client[uuid].Reconnect <- true
	}
	Client[uuid] = cli
}

// 将用户的连接信息从全局Client中删除
func DeleteClient(uuid uint32) {
	ClientMutex.Lock()
	defer ClientMutex.Unlock()
	delete(Client, uuid)
}

// 获取uuid的连接，若不存在则返回nil
func GetConnectionUse(uuid uint32) *Connection {
	cli, ok := Client[uuid]
	if !ok {
		return nil
	}
	return cli
}

func GetRemoteIpUseUUID(uuid uint32) string {
	cli := GetConnectionUse(uuid)
	if cli == nil {
		return ""
	}
	return strings.Split(cli.Connection.RemoteAddr().String(), ":")[0]
}

func GetRemoteIpUseConn(conn net.Conn) string {
	return strings.Split(conn.RemoteAddr().String(), ":")[0]
}

// 监听端口，等待连接请求
func ListenTCP(address string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ListenTCP recover", err)
		}
	}()

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Println("Failed to ResolveTCPAddr:", err.Error())
		return
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println("Failed to ListenTCP:", err.Error())
		return
	}

	log.Printf("TCP Server is listening [%s], "+
		"waiting for client to connect...\n", address)

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			log.Println("Accept client connection exception:", err.Error())
			continue
		}
		//log.Println("Client connection comes from:", conn.RemoteAddr().String())
		go Connect(conn)
	}
}

// 为连接提供服务
func Connect(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			_ = conn.Close()
			log.Println("Connect recover", err)
		}
	}()

	var (
		err  error
		data = make([]byte, LengthHeadPackage+10)
		n    int
		pkg  = NewPackage()
		cli  *Connection
	)

	// 接收心跳请求
	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Read(data)
	if err != nil || n != LengthHeadPackage {
		log.Println("Error: Receive connection request", err)
		_ = conn.Close()
		return
	}
	pkg = ConvertToPackage(data[:n])
	if !pkg.CheckHeadCheckSum(){
		log.Println("Error: Check connection request")
		_ = conn.Close()
		log.Println(data)
		return
	}
	if pkg.GetRequestCode() != def.HeartbeatRequest {
		if pkg.GetRequestCode() != def.RegisterRequest {
			log.Println("Error: Check connection request")
			_ = conn.Close()
			log.Println(data)
			return
		}
		username, password := UnwrapLoginData(pkg.GetHeadData())
		ok := operator.AddUser(username, username, password)
		//log.Println("Register", ok)
		pkg.SetRequestCode(def.RegisterRespond)
		pkg.SetHeadData([]byte(fmt.Sprintf("Register %v", ok)))
		pkg.SetHeadCheckSum()
		_, _ = conn.Write(pkg.Data())
		_ = conn.Close()
		return
	}

	// 回应心跳请求
	pkg.ClearExceptSeq()
	pkg.SetRequestCode(def.HeartbeatRespond)
	pkg.SetACK(pkg.GetSEQ())
	pkg.SetSEQ(100)
	pkg.SetHeadCheckSum()
	err = conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Write(pkg.Data())
	if err != nil || n != LengthHeadPackage {
		log.Println("Error: Send connection response")
		_ = conn.Close()
		return
	}

	// 建立连接成功，等待用户登录
	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Read(data)
	if err != nil || n != LengthHeadPackage {
		log.Println(data)
		log.Println("Error: Receive login data")
		_ = conn.Close()
		return
	}
	pkg = ConvertToPackage(data[:n])
	if !pkg.CheckHeadCheckSum() || pkg.GetRequestCode() != def.Login {
		log.Println("Error: Check login data")
		_ = conn.Close()
		return
	}
	username, password := UnwrapLoginData(pkg.GetHeadData())

	// 处理登录请求
	pkg.ClearExceptSeq()
	pkg.SetACK(pkg.GetSEQ())
	pkg.SetSEQ(101)
	u, ok := operator.Login(username, password)
	if ok {
		cli = NewConnection(u.Uuid, u.Username, u.Nickname, conn)
		AddClient(u.Uuid, cli)
		log.Printf("Accept Login: [%d]%s", u.Uuid, username)
		pkg.SetRequestCode(def.LoginSuccessful)
		pkg.SetHeadData(WrapUserInfo(u.Uuid, u.Username, u.Nickname))
		pkg.SetHeadCheckSum()
		err = conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
		if err != nil {
			_ = conn.Close()
			return
		}
		n, err = conn.Write(pkg.Data())
		if err != nil || n != LengthHeadPackage {
			_ = conn.Close()
			return
		}
	} else {
		pkg.SetRequestCode(def.LoginFailure)
		pkg.SetHeadData([]byte("Wrong password"))
		pkg.SetHeadCheckSum()
		err = conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
		if err != nil {
			_ = conn.Close()
			return
		}
		_, _ = conn.Write(pkg.Data())
		_ = conn.Close()
		return
	}

	go Terminator(cli)
	go Sender(cli)
	go Receiver(cli)
	go Heartbeat(cli)

	friends := operator.GetFriends(u.Uuid)
	for _, to := range friends {
		cliTo := GetConnectionUse(to)
		if cliTo == nil {
			continue
		}
		go AutoAdaptSendMessage(def.Message, cliTo, 0,
			fmt.Sprintf("User [%d][%s] Logined", u.Uuid, u.Username), pkg.GetSEQ())
	}

	select {
	case <-cli.WorkerReq.Server:
		cli.WorkerRes.Server <- true
		return
	}
}

// 信息发送服务
// 监听 Connection.DataSend 中的信息并发送
func Sender(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Sender recover", err)
		}
	}()

	for {
		select {
		case <-cli.WorkerReq.Sender:
			cli.WorkerRes.Sender <- true
			return
		case d := <-cli.DataSend:
			// 设置SEQ
			cli.SEQMutex.Lock()
			if cli.HeartbeatSEQ != 0 {
				cli.HeartbeatSEQ = cli.SEQ
			}
			d.SetSEQ(cli.SEQ)
			cli.SEQ++
			cli.SEQMutex.Unlock()
			// 设置时间戳
			d.SetTime(uint64(time.Now().UnixNano()))
			d.SetHeadCheckSum()
			PrintPackage(d, false, false)
			_ = cli.Connection.SetWriteDeadline(time.Now().Add(20 * time.Second))
			n, err := cli.Connection.Write(d.Data())
			if err != nil {
				log.Printf("UUID:[%v] Error sending data:"+
					" failed to send. [%s]\n", cli.UUID, err.Error())
				continue
			}
			if n != LengthHeadPackage {
				log.Printf("UUID:[%v] Error sending data:"+
					" the length of the sent data does not match. "+
					"%d bytes sent. actual length %d bytes\n",
					cli.UUID, n, LengthHeadPackage)
				continue
			}
			if cli.HeartbeatSEQ == 0 {
				cli.ResetHeartbeat <- true
			}
		}
	}
}

// 接收并处理信息
func Receiver(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Receiver recover", err)
		}
	}()

	dataT := make([]byte, LengthHeadPackage+10)
	down := false
	go func() {
		for {
			select {
			case <-cli.WorkerReq.Receiver:
				down = true
				return
			}
		}
	}()

	for {
		if down {
			cli.WorkerRes.Receiver <- true
			return
		}

		_ = cli.Connection.SetReadDeadline(time.Now().Add(20 * time.Second))
		n, err := cli.Connection.Read(dataT)
		if err != nil || n != LengthHeadPackage {
			continue
		}

		data := ConvertToPackage(dataT[:n])
		if !data.CheckHeadCheckSum() || time.Now().UnixNano()-int64(data.GetTime()) >= int64(10 * time.Second) {
			fmt.Println("******** !!!Broken!!! *******")
			PrintPackage(&data, true,true)
			continue
		}

		if cli.HeartbeatSEQ == 0 {
			cli.ResetHeartbeat <- true
		}

		PrintPackage(&data, false,true)
		ProcessPackage(&data, cli)
	}
}

func ProcessPackage(p *Package, cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ProcessPackage recover", err)
		}
	}()

	switch p.GetRequestCode() {
	case def.HeartbeatRespond:
		CheckHeartbeatRespond(p, cli)
	case def.Message:
		ForwardMessage(p, cli)
	case def.AddFriend:
		RespondAddFriend(p, cli)
	case def.AcceptFriendInvitation:
		AcceptFriendInvitation(cli.UUID, BytesToUInt32(p.GetHeadData()))
	case def.ListFriendInvitation:
		SendListFriendInvitation(p, cli)
	case def.ListFriend:
		SendFriendList(p, cli)
	case def.TerminateTheConnection:
		cli.Termination <- true
	case def.DeleteFriend:
		DeleteFriend(cli.UUID, BytesToUInt32(p.GetHeadData()))
	case def.CreateGroup:
		CreateGroup(p, cli)
	case def.JoinGroup:
		JoinGroup(cli.UUID, BytesToUInt32(p.GetHeadData()))
	case def.ListGroup:
		SendGroupList(p, cli)
	case def.ListJoinGroup:
		SendJoinGroupList(p, cli)
	case def.AcceptJoinGroup:
		AcceptJoinGroup(p, cli)
	case def.GroupMessage:
		ForwardGroupMessage(p, cli)
	case def.ListGroupMember:
		SendGroupMemberList(p, cli)
	case def.AppointAdmin:
		AppointGroupAdmin(p, cli)
	case def.ListGroupAdmin:
		SendGroupAdminList(p, cli)
	case def.RevokeAdmin:
		RevokeGroupAdmin(p, cli)
	case def.TransferGroup:
		TransferGroup(p, cli)
	case def.DeleteMember:
		DeleteGroupMember(p, cli)
	case def.SendFile:
		ForwardFile(p, cli)
	case def.SendGroupFile:
		SaveGroupFile(p, cli)
	case def.ListGroupFile:
		SendGroupFileList(p, cli)
	case def.DownloadGroupFile:
		SendGroupFile(p, cli)
	default:
		fmt.Println("******* !!!Rubbish!!! *******")
		PrintPackage(p,true, true)
	}
}

// 心跳监测
// 30s 内没有发送/接收数据时，向 cli 发送心跳请求
// 心跳请求发送后，若 30s 后仍没有接收到心跳回应，则生成终止信号
func Heartbeat(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Heartbeat recover", err)
		}
	}()

	for {
		select {
		case <-cli.WorkerReq.Heartbeat:
			cli.WorkerRes.Heartbeat <- true
			return
		case <-cli.ResetHeartbeat:
			cli.HeartbeatSEQ = 0
		case <-time.After(30 * time.Second):
			data := NewPackage()
			data.SetRequestCode(def.HeartbeatRequest)
			cli.HeartbeatSEQ = 1
			cli.DataSend <- &data
			select {
			case <-time.After(30 * time.Second):
				cli.Termination <- true
			case <-cli.Heartbeat:
				cli.HeartbeatSEQ = 0
			}
		}
	}
}

// 向服务于 cli 的所有routine发送终止服务信号
// 从全局 Client 中删除用户信息
func Terminator(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Terminator recover", err)
		}
	}()

	select {
	case <-cli.Termination:
		DeleteClient(cli.UUID)
	case <-cli.Reconnect:
		p := NewPackage()
		p.SetRequestCode(def.TerminateTheConnection)
		cli.DataSend <- &p
		time.Sleep(1*time.Second)
	}

	log.Printf("UUID:[%v] Kill Signal Generated\n", cli.UUID)

	typ := reflect.TypeOf(*cli.WorkerReq)
	val := reflect.ValueOf(*cli.WorkerReq)
	for k := 0; k < typ.NumField(); k++ {
		val.Field(k).Interface().(chan bool) <- true
	}
	closed := 0
	unclosed := make([]string, 0)
	typ = reflect.TypeOf(*cli.WorkerRes)
	val = reflect.ValueOf(*cli.WorkerRes)
	for k := 0; k < typ.NumField(); k++ {
		select {
		case <-val.Field(k).Interface().(chan bool):
			closed++
			continue
		case <-time.After(1 * time.Minute):
			unclosed = append(unclosed, typ.Field(k).Name)
			continue
		}
	}

	if closed == typ.NumField() {
		log.Printf("UUID:[%v] Connection closed. "+
			"All threads are terminated\n", cli.UUID)
	} else {
		log.Printf("UUID:[%v] Connection closed. "+
			"The following threads are not terminated:%v\n",
			cli.UUID, unclosed)
	}
}
