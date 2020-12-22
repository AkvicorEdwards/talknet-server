package tcp

import (
	"fmt"
	"log"
	"talknet/dam"
	"talknet/def"
	"talknet/operator"
	"time"
)

func ForwardMessage(p *Package, cliFrom *Connection) {
	from := cliFrom.UUID
	to, filename := UnwrapMessage(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] Forward Message From U[%d] to U[%d]", time.Now().UnixNano(), from, to)
	cliTo := GetConnectionUse(to)
	if cliTo == nil {
		AutoAdaptSendMessage(def.Message, cliFrom, 0, "发送信息失败，用户未登录", p.GetSEQ())
		return
	}

	if p.GetExtendedDataFlag() != 1 {
		// 短消息，直接转发给目标用户
		ReWrapHeadDataMessageUUID(from, p)
		cliTo.DataSend <- p
		return
	} else {
		// 长消息
		// 将信息打包以filename为key存储在 FileMap 中
		log.Printf("FileReceivePrepare [%s]", filename)
		info := FileReceivePrepare(filename, GetRemoteIpUseConn(cliFrom.Connection),
			from, []FileMapInfoTo{{Ip: GetRemoteIpUseUUID(to), uuid: to}}, p.GetExternalDataCheckSum())

		// 配置 Package 1
		// 向发送者发送响应包，代表已准备好接收 存储了长信息的文件
		p.SetRequestCode(def.PermitLongMessage)
		p.SetACK(p.GetSEQ())
		p.SetExtendedDataFlag(0)

		// 判断目标用户是否在线
		if GetConnectionUse(to) == nil {
			FileMapMutex.Lock()
			delete(FileMap, filename)
			FileMapMutex.Unlock()
			return
		}

		// 配置 Package 2，保存在 FileMap[filename] 中
		// 当文件从源客户端接收完毕后，会向目的客户端发送此Package
		// 向目标客户端发送信息，表示有长信息文件需要对方接收
		p2 := NewPackage()
		p2.SetRequestCode(def.Message)
		p2.SetHeadData(p.GetHeadData())
		ReWrapHeadDataMessageUUID(from, &p2)
		p2.SetExtendedDataFlag(1)
		p2.SetExternalDataCheckSum(p.GetExternalDataCheckSum())
		info.Package = &p2

		// 发送 Package 1
		cliFrom.DataSend <- p
		return
	}
}

// TODO ForwardFile
func ForwardFile(p *Package, cliFrom *Connection) {
	from := cliFrom.UUID
	to, filename := UnwrapMessage(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] Forward File From U[%d] to U[%d]", time.Now().UnixNano(), from, to)
	cliTo := GetConnectionUse(to)
	if cliTo == nil {
		AutoAdaptSendMessage(def.Message, cliFrom, 0, "发送文件失败，用户未登录", p.GetSEQ())
		return
	}

	// 将文件信息打包以filename为key存储在 FileMap 中
	info := FileReceivePrepare(filename, GetRemoteIpUseConn(cliFrom.Connection),
		from, []FileMapInfoTo{{Ip: GetRemoteIpUseUUID(to), uuid: to}}, p.GetExternalDataCheckSum())

	// 配置 Package 1
	// 向发送者发送响应包，代表已准备好接收 存储了长信息的文件
	p.SetRequestCode(def.PermitSendFile)
	p.SetACK(p.GetSEQ())
	p.SetExtendedDataFlag(0)

	// 判断目标用户是否在线
	if GetConnectionUse(to) == nil {
		FileMapMutex.Lock()
		delete(FileMap, filename)
		FileMapMutex.Unlock()
		return
	}

	// 配置 Package 2，保存在 FileMap[filename] 中
	// 当文件从源客户端接收完毕后，会向目的客户端发送此Package
	// 向目标客户端发送信息，表示有长信息文件需要对方接收
	p2 := NewPackage()
	p2.SetRequestCode(def.SendFile)
	p2.SetHeadData(p.GetHeadData())
	ReWrapHeadDataMessageUUID(from, &p2)
	p2.SetExtendedDataFlag(1)
	p2.SetExternalDataCheckSum(p.GetExternalDataCheckSum())
	info.Package = &p2

	// 发送 Package 1
	cliFrom.DataSend <- p
	return

}

func ForwardGroupMessage(p *Package, cliFrom *Connection) {
	from := cliFrom.UUID
	guid, _, filename := UnwrapGroupMessage(p.GetHeadData())
	group := operator.GetGroup(guid)
	// TODO DIS
	log.Printf("[%d] Forward Group Message From U[%d] to G[%d]", time.Now().UnixNano(), from, guid)
	if group.Guid == 0 {
		return
	}

	if !func() bool {
		for _, v := range group.Member {
			if v == from {
				return true
			}
		}
		return false
	}() {
		return
	}
	//log.Println("Forward Group Message", filename)

	// 直接转发
	if p.GetExtendedDataFlag() == 0 {
		for _, to := range group.Member {
			cliTo := GetConnectionUse(to)
			if cliTo != nil {
				cliTo.DataSend <- p
			}
		}
		return
	}

	//log.Println("Forward Long Group Message")

	member := operator.GetGroup(guid).Member

	// 将信息打包以filename为key存储在 FileMap 中
	info := FileReceivePrepare(filename, GetRemoteIpUseConn(cliFrom.Connection),
		from, make([]FileMapInfoTo, 0), p.GetExternalDataCheckSum())
	// 向发送者发送响应包，代表已准备好接收 存储了长信息的文件
	p.SetRequestCode(def.PermitLongGroupMessage)
	p.SetACK(p.GetSEQ())
	p.SetExtendedDataFlag(0)

	for _, v := range member {
		ip := GetRemoteIpUseUUID(v)
		info.To = append(info.To, FileMapInfoTo{
			Ip:   ip,
			uuid: v,
		})
	}

	// 配置 Package 2，保存在 FileMap[filename] 中
	// 当文件从源客户端接收完毕后，会向目的客户端发送此Package
	// 向目标客户端发送信息，表示有长信息文件需要对方接收
	p2 := NewPackage()
	p2.SetRequestCode(def.GroupMessage)
	p2.SetHeadData(p.GetHeadData())
	p2.SetExtendedDataFlag(1)
	p2.SetExternalDataCheckSum(p.GetExternalDataCheckSum())
	info.Package = &p2

	//log.Println("Send PermitLongMessage")
	cliFrom.DataSend <- p
}

// TODO SaveGroupFile
func SaveGroupFile(p *Package, cliFrom *Connection) {
	from := cliFrom.UUID
	guid, filename := UnwrapMessage(p.GetHeadData())
	group := operator.GetGroup(guid)
	// TODO DIS
	log.Printf("[%d] Upload Group File From U[%d] to G[%d]", time.Now().UnixNano(), from, guid)
	if group.Guid == 0 {
		return
	}

	if !func() bool {
		for _, v := range group.Member {
			if v == from {
				return true
			}
		}
		return false
	}() {
		return
	}
	//log.Println("Forward Group Message", filename)

	// 将信息打包以filename为key存储在 FileMap 中
	info := FileReceivePrepare(filename, GetRemoteIpUseConn(cliFrom.Connection),
		from, make([]FileMapInfoTo, 0), p.GetExternalDataCheckSum())
	// 向发送者发送响应包，代表已准备好接收 存储了长信息的文件
	p.SetRequestCode(def.PermitSendGroupFile)
	p.SetACK(p.GetSEQ())
	p.SetExtendedDataFlag(0)
	// TODO FILE

	info.Package = nil

	info.To = append(info.To, FileMapInfoTo{
		Ip:   "",
		uuid: guid,
	})

	// 发送 Package 1
	cliFrom.DataSend <- p
	return

}

// TODO SendGroupFile
func SendGroupFile(p *Package, cli *Connection) {
	guid, fuid := UnwrapGuidUuid(p.GetHeadData())
	log.Printf("U[%d] Download F[%d] From G[%d]\n", cli.UUID, fuid, guid)
	group := operator.GetGroup(guid)

	// 检查是否是群成员
	if !func() bool {
		for _, v := range group.Member {
			if cli.UUID == v {
				return true
			}
		}
		return false
	}(){
		return
	}
	var f dam.Files
	// 检查文件是否存在
	if !func() bool {
		for _, v := range group.Files {
			if fuid == v.Fuid {
				f = v
				return true
			}
		}
		return false
	}(){
		return
	}
	// TODO 文件发送

	_, _, hashVal := CalculateFileHashValue(def.GroupDir+f.Filename)
	if hashVal != f.Hash {
		return
	}

	*p = NewPackage()
	p.SetExtendedDataFlag(1)
	//data, _ := WrapMessage(0, f.Filename)
	//log.Println("Permit send file", f.Filename, data)
	//p.SetHeadData(data)
	//log.Println(p.GetHeadData())
	p.SetExternalDataCheckSum(hashVal)
	p.SetRequestCode(def.PermitDownloadGroupFile)
	PrepareSendFile(p, cli, def.GroupDir, f.Filename)
}

// 检查心跳回应Package
func CheckHeartbeatRespond(p *Package, cli *Connection) {
	// TODO DIS
	//log.Printf("[%d] Heartbeat Respond [%d] ACK [%d]\n", time.Now().UnixNano(), cli.UUID, p.GetACK())
	if p.GetACK() == cli.HeartbeatSEQ {
		cli.Heartbeat <- true
	} else {
		log.Printf("Error: Heartbeat need:[%d] got:[%d]\n", cli.HeartbeatSEQ, p.GetACK())
	}
}

// 响应 添加好友
func RespondAddFriend(p *Package, cli *Connection) {
	user, ok := operator.GetUser(BytesToUInt32(p.GetHeadData()))
	if !ok {
		return
	}
	// TODO DIS
	log.Printf("[%d] [%d] Add Friend [%d]\n", time.Now().UnixNano(), cli.UUID, user.Uuid)

	if !AddFriend(cli.UUID, BytesToUInt32(p.GetHeadData())) {
		AutoAdaptSendMessage(def.Message, cli, 0, "添加好友失败", p.GetSEQ())
		return
	}
	cli2 := GetConnectionUse(user.Uuid)
	if cli2 == nil {
		return
	}

	ms := fmt.Sprintf("%s[%d] wants to add you as a friend", cli.Username, cli.UUID)

	AutoAdaptSendMessage(def.Message, cli2, 0, ms, 0)
}

// 发送待接收邀请好友的列表
func SendListFriendInvitation(p *Package, cli *Connection) {
	data := operator.GetUnacceptedFriendsInvitation(cli.UUID)
	AutoAdaptSendMessage(def.ListFriendInvitation, cli, 0, data, p.GetSEQ())
}

// 发送好友列表
func SendFriendList(p *Package, cli *Connection) {
	data := operator.GetFriendList(cli.UUID)
	AutoAdaptSendMessage(def.ListFriendInvitation, cli, 0, data, p.GetSEQ())
}

// 接受好友申请
func AcceptFriendInvitation(uuid, accept uint32) {
	// TODO DIS
	log.Printf("[%d] [%d] Accept Friend [%d]\n", time.Now().UnixNano(), uuid, accept)
	operator.AcceptFriendRequest(uuid, accept)
}

// 删除好友
func DeleteFriend(uuid, delUUID uint32) {
	// TODO DIS
	log.Printf("[%d] [%d] Delete Friend [%d]\n", time.Now().UnixNano(), uuid, delUUID)
	operator.DeleteFriend(uuid, delUUID)
}

// 创建小组
func CreateGroup(p *Package, cli *Connection) {
	_, name := UnwrapMessage(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] [%d] Create Group [%s]\n", time.Now().UnixNano(), cli.UUID, name)
	operator.CreateGroup(name, cli.UUID)
}

// 申请加入群组
func JoinGroup(uuid, guid uint32) {
	// TODO DIS
	log.Printf("[%d] [%d] Join Group [%d]\n", time.Now().UnixNano(), uuid, guid)
	if operator.JoinGroup(uuid, guid) {
		log.Println(uuid, "join group", guid)
	}
}

// 发送群组列表
func SendGroupList(p *Package, cli *Connection) {
	data := operator.GetGroupList(cli.UUID)
	AutoAdaptSendMessage(def.ListGroup, cli, 0, data, p.GetSEQ())
}

// 发送申请加入群组列表
func SendJoinGroupList(p *Package, cli *Connection) {
	data := operator.GetJoinGroupList(BytesToUInt32(p.GetHeadData()))
	AutoAdaptSendMessage(def.ListJoinGroup, cli, 0, data, p.GetSEQ())
}

// 允许加入群组
func AcceptJoinGroup(p *Package, cli *Connection) {
	guid, uuid := UnwrapGuidUuid(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] U[%d] Accept U[%d] Join Group G[%d]\n", time.Now().UnixNano(), cli.UUID, uuid, guid)
	operator.AcceptJoinGroup(guid, cli.UUID, uuid)
}

// 发送群组列表
func SendGroupMemberList(p *Package, cli *Connection) {
	data := operator.GetGroup(BytesToUInt32(p.GetHeadData())).Transfer().Member
	AutoAdaptSendMessage(def.ListGroupMember, cli, 0, data, p.GetSEQ())
}

// 发送群组列表
func SendGroupFileList(p *Package, cli *Connection) {
	data := operator.GetGroup(BytesToUInt32(p.GetHeadData())).Transfer().Files
	log.Println(data)
	AutoAdaptSendMessage(def.ListGroupFile, cli, 0, data, p.GetSEQ())
}

// 发送群组列表
func SendGroupAdminList(p *Package, cli *Connection) {
	data := operator.GetGroup(BytesToUInt32(p.GetHeadData())).Transfer().Admin
	AutoAdaptSendMessage(def.ListGroupAdmin, cli, 0, data, p.GetSEQ())
}


func AppointGroupAdmin(p *Package, cli *Connection) {
	guid, uuid := UnwrapGuidUuid(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] U[%d] Appoint U[%d] As Admin of Group G[%d]\n", time.Now().UnixNano(), cli.UUID, uuid, guid)
	operator.AppointAdmin(guid, cli.UUID, uuid)
}

// 允许加入群组
func RevokeGroupAdmin(p *Package, cli *Connection) {
	guid, uuid := UnwrapGuidUuid(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] U[%d] Revoke U[%d] from Admin of Group G[%d]\n", time.Now().UnixNano(), cli.UUID, uuid, guid)
	operator.RevokeAdmin(guid, cli.UUID, uuid)
}

// 允许加入群组
func TransferGroup(p *Package, cli *Connection) {
	guid, uuid := UnwrapGuidUuid(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] U[%d] Appoint U[%d] As Owner of Group G[%d]\n", time.Now().UnixNano(), cli.UUID, uuid, guid)
	operator.TransferGroup(guid, cli.UUID, uuid)
}

// 允许加入群组
func DeleteGroupMember(p *Package, cli *Connection) {
	guid, uuid := UnwrapGuidUuid(p.GetHeadData())
	// TODO DIS
	log.Printf("[%d] U[%d] Delete Member U[%d] of Group G[%d]\n", time.Now().UnixNano(), cli.UUID, uuid, guid)
	operator.DeleteGroupMember(guid, cli.UUID, uuid)
}
