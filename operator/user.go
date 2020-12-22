package operator

import (
	"encoding/json"
	"strconv"
	"talknet/dam"
)

// 添加一个用户
func AddUser(username, nickname, password string) bool {
	_, ok := dam.AddUser(username, nickname, password)
	return ok
}

// 发送好友申请
func AddFriendRequest(inviter, invitee uint32) bool {
	return dam.AddFriendRequest(inviter, invitee)
}

// 删除一组用户
func DeleteUsers(uuids []string) bool {
	uuid := make([]int, len(uuids))
	var err error
	for k, v := range uuids {
		uuid[k], err = strconv.Atoi(v)
		if err != nil {
			return false
		}
	}
	return dam.DeleteUsers(uuid)
}

// 更新一条用户记录
func UpdateUser(uuid uint32, nickname, password string, deleted int64) bool {
	return dam.UpdateUser(dam.User{
		Uuid:     uuid,
		Nickname: nickname,
		Password: password,
		Deleted:  deleted,
	})
}

// 获取所有用户记录的json格式
func GetAllUserJson() string {
	users := dam.GetAllUser()
	data, err := json.Marshal(users)
	if err != nil {
		return ""
	}
	return string(data)
}

// 获取所有用户记录的json格式
func GetAllUser() []dam.User {
	return dam.GetAllUser()
}

// 登录
func Login(username, password string) (*dam.User, bool) {
	user := dam.GetUser(username)
	if user.Deleted != 0 || user.Uuid <= 0 || user.Password != password {
		return &dam.User{}, false
	}
	return user, true
}

// 获取一条用户记录
func GetUser(uuid uint32) (*dam.User, bool) {
	user := dam.GetUser(uuid)
	if user.Deleted != 0 || user.Uuid <= 0 {
		return &dam.User{}, false
	}
	return user, true
}

func GetUnacceptedFriendsInvitation(uuid uint32) string {
	return dam.GetUnacceptedFriendsInvitation(uuid)
}

func GetFriendList(uuid uint32) string {
	return dam.GetFriendList(uuid)
}

func GetFriends(uuid uint32) []uint32 {
	return dam.GetFriends(uuid)
}

func AcceptFriendRequest(uuid, accept uint32) bool {
	return dam.AcceptFriendRequest(uuid, accept)
}

func DeleteFriend(uuid, delUUID uint32) bool {
	return dam.DeleteFriend(uuid, delUUID)
}

func GetGroupList(uuid uint32) string {
	return dam.GetGroupList(uuid)
}