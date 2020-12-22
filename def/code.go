package def

const (
	Unknown uint16 = iota
	// 心跳请求
	HeartbeatRequest
	// 心跳回应
	HeartbeatRespond
	// 登录请求
	Login
	// 登陆成功
	LoginSuccessful
	// 登陆失败
	LoginFailure
	// 终止连接
	TerminateTheConnection
	// 信息
	Message
	// 允许长信息
	PermitLongMessage
	// 添加好友
	AddFriend
	// 删除好友
	DeleteFriend
	// 待接收的好友邀请
	ListFriendInvitation
	// 接收好友邀请
	AcceptFriendInvitation
	// 好友列表
	ListFriend
	// 创建组
	CreateGroup
	// 加入组
	JoinGroup
	// 列出已加入组
	ListGroup
	// 列出入组申请
	ListJoinGroup
	// 接受入组申请
	AcceptJoinGroup
	// 任命管理员
	AppointAdmin
	// 转让组
	TransferGroup
	// 撤销管理员
	RevokeAdmin
	// 删除成员
	DeleteMember
	// 群组消息
	GroupMessage
	// 允许长信息
	PermitLongGroupMessage
	// 群成员列表
	ListGroupMember
	// 群管理员
	ListGroupAdmin
	// 发送文件
	SendFile
	// 允许发送文件
	PermitSendFile
	// 发送文件至群
	SendGroupFile
	// 允许发送群文件
	PermitSendGroupFile
	// 下载群文件
	DownloadGroupFile
	// 群文件列表
	ListGroupFile
	// 允许下载群文件
	PermitDownloadGroupFile
	// Register request
	RegisterRequest
	// Register respond
	RegisterRespond
)
