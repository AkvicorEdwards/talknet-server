package dam

import (
	"encoding/json"
	"log"
	"time"
)

// 向 user 表添加记录
func AddUser(username, nickname, password string) (uint32, bool) {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()
	user := User{
		Uuid:       GetInc("user") + 1,
		Username:   username,
		Nickname:   nickname,
		Password:   password,
		Friends:    make([]uint32, 0),
		Groups:     make([]uint32, 0),
		Unaccepted: make([]uint32, 0),
		Deleted:    0,
	}
	userDB := user.Transfer()

	if err := db.Table("user").Create(&userDB).Error; err != nil {
		log.Println(err)
		return 0, false
	}

	UpdateInc("user", user.Uuid)

	return user.Uuid, true
}

// 将申请人的uuid添加到被申请人的好友请求列表
func AddFriendRequest(inviter, invitee uint32) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()
	user := GetUser(invitee)
	if user.Uuid == 0 {
		return false
	}
	for _, v := range user.Friends {
		if v == inviter {
			return false
		}
	}
	for _, v := range user.Unaccepted {
		if v == inviter {
			return false
		}
	}
	user.Unaccepted = append(user.Unaccepted, inviter)

	res := db.Table("user").Where("uuid=?", invitee).Update(
		"unaccepted", user.TransferUnaccepted())
	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}
	return true
}

// 将申请人的uuid添加到被申请人的好友请求列表
func AddGroup(uuid, guid uint32) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	user := GetUser(uuid)
	if user.Uuid == 0 {
		return false
	}

	for _, v := range user.Groups {
		if v == guid {
			return true
		}
	}

	user.Groups = append(user.Groups, guid)

	res := db.Table("user").Where("uuid=?", uuid).Update(
		"groups", user.TransferGroups())
	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}
	return true
}

// 将申请人的uuid添加到被申请人的好友请求列表
func AcceptFriendRequest(uuid, accept uint32) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	// 被申请人
	user := GetUser(uuid)
	if user.Uuid == 0 {
		return false
	}
	for _, v := range user.Friends {
		if v == accept {
			return false
		}
	}
	for k, v := range user.Unaccepted {
		if v == accept {
			for i := k; i < len(user.Unaccepted)-1; i++ {
				user.Unaccepted[i] = user.Unaccepted[i+1]
			}
			user.Unaccepted = user.Unaccepted[:len(user.Unaccepted)-1]
			break
		}
	}
	user.Friends = append(user.Friends, accept)

	res := db.Table("user").Where("uuid=?", user.Uuid).Updates(map[string]interface{}{
		"friends":   user.TransferFriends(),
		"unaccepted":   user.TransferUnaccepted(),
	})

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}

	// 申请人
	user = GetUser(accept)
	for _, v := range user.Friends {
		if v == uuid {
			return false
		}
	}
	for k, v := range user.Unaccepted {
		if v == uuid {
			for i := k; i < len(user.Unaccepted)-1; i++ {
				user.Unaccepted[i] = user.Unaccepted[i+1]
			}
			user.Unaccepted = user.Unaccepted[:len(user.Unaccepted)-1]
			break
		}
	}
	user.Friends = append(user.Friends, uuid)
	res = db.Table("user").Where("uuid=?", user.Uuid).Updates(map[string]interface{}{
		"friends":   user.TransferFriends(),
		"unaccepted":   user.TransferUnaccepted(),
	})

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}

	return true
}

// 将申请人的uuid添加到被申请人的好友请求列表
func DeleteFriend(uuid, delUUID uint32) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	user := GetUser(uuid)
	if user.Uuid == 0 {
		return false
	}

	for k, v := range user.Friends {
		if v == delUUID {
			for i := k; i < len(user.Friends)-1; i++ {
				user.Friends[i] = user.Friends[i+1]
			}
			user.Friends = user.Friends[:len(user.Friends)-1]
			break
		}
	}

	res := db.Table("user").Where("uuid=?", user.Uuid).Updates(map[string]interface{}{
		"friends":   user.TransferFriends(),
	})

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}

	return true

}
func DeleteGroup(uuid, delGUID uint32) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	user := GetUser(uuid)
	if user.Uuid == 0 {
		return false
	}

	for k, v := range user.Groups {
		if v == delGUID {
			for i := k; i < len(user.Groups)-1; i++ {
				user.Groups[i] = user.Groups[i+1]
			}
			user.Groups = user.Groups[:len(user.Groups)-1]
			break
		}
	}

	res := db.Table("user").Where("uuid=?", user.Uuid).Updates(map[string]interface{}{
		"groups":   user.TransferGroups(),
	})

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}

	return true

}

// 将 user 表中目标记录的deleted置为当前时间戳
func DeleteUser(u interface{}) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	query := ""
	switch u.(type) {
	case int, int64, int32, int16, int8, uint, uint32, uint16, uint8:
		query = "uuid=? AND deleted=0"
	case string:
		query = "username=? AND deleted=0"
	default:
		return false
	}

	res := db.Table("user").Where(query, u).Update("deleted", time.Now().Unix())
	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}
	return true
}

// 删除 user 表中的一条记录
func DeleteUsers(uuid []int) bool {
	if !Connected {
		Connect()
	}

	lockUser.Lock()
	defer lockUser.Unlock()

	res := db.Table("user").Where("uuid IN (?) AND deleted=0", uuid).Update("deleted", time.Now().Unix())
	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}
	return true

}

// 更新 user 表中的一条记录
// uuid和username不会被更新
func UpdateUser(user User) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	res := db.Table("user").Where("uuid=?", user.Uuid).Updates(map[string]interface{}{
		"nickname":   user.Nickname,
		"password":   user.Password,
		"deleted":    user.Deleted,
	})

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}

	if res.RowsAffected == 0 {
		return false
	}

	return true
}

// 更新用户的密码
func UpdatePassword(u interface{}, password, newPassword string) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	query := ""
	switch u.(type) {
	case int, int64, int32, int16, int8, uint, uint32, uint16, uint8:
		query = "uuid=? AND password=?"
	case string:
		query = "username=? AND password=?"
	default:
		return false
	}

	res := db.Table("user").Where(query, u, password).Update("password", newPassword)

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}

	if res.RowsAffected == 0 {
		return false
	}

	return true
}

// 从 user 表中获取一条记录
func GetUser(u interface{}) *User {
	if !Connected {
		Connect()
	}
	user := &UserDB{}
	query := ""
	switch u.(type) {
	case int, int64, int32, int16, int8, uint, uint32, uint16, uint8:
		query = "uuid=?"
	case string:
		query = "username=?"
	default:
		return &User{}
	}
	db.Table("user").Where(query, u).First(user)

	rt := user.Transfer()
	return &rt
}

// 获取所有用户
func GetAllUser() []User {
	if !Connected {
		Connect()
	}
	users := make([]UserDB, 0)
	db.Table("user").Find(&users)
	user := make([]User, len(users))
	for k, v := range users {
		user[k] = v.Transfer()
	}
	return user
}

func GetUnacceptedFriendsInvitation(uuid uint32) string {
	if !Connected {
		Connect()
	}
	return GetUser(uuid).TransferUnaccepted()
}

func GetFriendList(uuid uint32) string {
	if !Connected {
		Connect()
	}
	return GetUser(uuid).TransferFriends()
}

func GetFriends(uuid uint32) []uint32 {
	if !Connected {
		Connect()
	}
	return GetUser(uuid).Friends
}

func GetGroupList(uuid uint32) string {
	if !Connected {
		Connect()
	}
	return GetUser(uuid).TransferGroups()
}


func UpdateUserGroup(uuid uint32, groups []uint32) bool {
	if !Connected {
		Connect()
	}
	lockUser.Lock()
	defer lockUser.Unlock()

	group, err := json.Marshal(groups)
	if err != nil {
		return false
	}

	res := db.Table("user").Where("uuid=?", uuid).Update(
		"groups", string(group))
	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}

	return true
}
