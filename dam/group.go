package dam

import (
	"log"
)

func CreateGroup(name string, owner uint32) (uint32, bool) {
	if !Connected {
		Connect()
	}
	lockGroups.Lock()
	defer lockGroups.Unlock()

	user := GetUser(owner)
	if user.Uuid == 0 {
		return 0, false
	}

	// 创建群组
	group := Group{
		Guid:       GetInc("groups") + 1,
		Name:       name,
		Owner:      owner,
		Admin:      make([]uint32, 0),
		Member:     make([]uint32, 0),
		Unaccepted: make([]uint32, 0),
		Files: make([]Files, 0),
		Deleted:    0,
	}
	group.Member = append(group.Member, owner)
	groupDB := group.Transfer()
	if err := db.Table("groups").Create(&groupDB).Error; err != nil {
		log.Println(err)
		return 0, false
	}
	UpdateInc("groups", group.Guid)

	// 将群添加入用户的Group
	user.Groups = append(user.Groups, group.Guid)

	UpdateUserGroup(owner, user.Groups)

	return group.Guid, true
}

// 将申请人的uuid添加到被申请人的好友请求列表
func JoinGroupRequest(uuid, guid uint32) bool {
	if !Connected {
		Connect()
	}
	lockGroups.Lock()
	defer lockGroups.Unlock()

	user := GetUser(uuid)
	if user.Uuid == 0 {
		return false
	}
	group := GetGroup(guid)
	if group.Guid == 0 {
		return false
	}
	for _, v := range group.Member {
		if v == uuid {
			return false
		}
	}
	for _, v := range group.Unaccepted {
		if v == uuid {
			return false
		}
	}
	group.Unaccepted = append(group.Unaccepted, uuid)

	res := db.Table("groups").Where("guid=?", guid).Update(
		"unaccepted", group.Transfer().Unaccepted)
	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}
	return true
}

func AddFileInfoToGroup(guid, uuid uint32, filename, realName string, hash uint32) bool {
	if !Connected {
		Connect()
	}
	lockGroups.Lock()
	defer lockGroups.Unlock()
	group := GetGroup(guid)

	if group.Guid == 0 {
		return false
	}

	if func() bool {
		for _, v := range group.Member {
			if v == uuid {
				return false
			}
		}
		return true
	}() {
		return false
	}

	group.Files = append(group.Files, Files{
		Uuid: uuid,
		Fuid: func() uint32 {
			if len(group.Files) == 0 {
				return 1
			}
			return group.Files[len(group.Files)-1].Fuid+1
		}(),
		Filename: filename,
		RealName: realName,
		Hash:     hash,
	})

	res := db.Table("groups").Where("guid=?", guid).Update(
		"files", group.Transfer().Files)
	if res.Error != nil {
		log.Println(res.Error)
		return false
	}
	if res.RowsAffected == 0 {
		return false
	}
	return true
}

func DeleteFileInfoFromGroup(guid, uuid uint32, fuid uint32) bool {
	if !Connected {
		Connect()
	}
	lockGroups.Lock()
	defer lockGroups.Unlock()
	group := GetGroup(guid)
	if group.Guid == 0 {
		return false
	}

	if !func() bool {
		// 操作人是群主
		if group.Owner == uuid {
			return true
		}
		// 操作人是管理员
		for _, v := range group.Admin {
			if v == uuid {
				return true
			}
		}
		// 操作者是文件持有者
		for _, v := range group.Files {
			if v.Fuid == fuid {
				if v.Uuid == uuid {
					return true
				}
			}
		}
		return false
	}() {
		return false
	}

	for k, v := range group.Files {
		if v.Fuid == fuid {
			for i := k; i < len(group.Files)-1; i++ {
				group.Files[i] = group.Files[i+1]
			}
			group.Files = group.Files[:len(group.Files)-1]
			break
		}
	}

	res := db.Table("groups").Where("guid=?", guid).Update(
		"files", group.Transfer().Files)
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
func GetGroup(guid uint32) *Group {
	if !Connected {
		Connect()
	}
	group := &GroupDB{}

	db.Table("groups").Where("guid=?", guid).First(group)

	rt := group.Transfer()
	return &rt
}

func GetJoinGroupList(guid uint32) string {
	if !Connected {
		Connect()
	}
	return GetGroup(guid).Transfer().Unaccepted
}

func AcceptJoinGroup(guid, admin, uuid uint32) bool {
	if !Connected {
		Connect()
	}

	lockGroups.Lock()
	defer lockGroups.Unlock()

	group := GetGroup(guid)

	if !func() bool {
		if group.Owner == admin {
			return true
		}
		for _, v := range group.Admin {
			if v == admin {
				return true
			}
		}
		return false
	}() {
		return false
	}

	for k, v := range group.Unaccepted {
		if v == uuid {
			for i := k; i < len(group.Unaccepted)-1; i++ {
				group.Unaccepted[i] = group.Unaccepted[i+1]
			}
			group.Unaccepted = group.Unaccepted[:len(group.Unaccepted)-1]
			break
		}
	}

	group.Member = append(group.Member, uuid)

	g := group.Transfer()

	res := db.Table("groups").Where("guid=?", guid).Updates(map[string]interface{}{
		"member":     g.Member,
		"unaccepted": g.Unaccepted,
	})

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}

	if res.RowsAffected == 0 {
		return false
	}

	if !AddGroup(uuid, guid) {
		return false
	}

	return true
}

func AppointAdmin(guid, admin, uuid uint32) bool {
	if !Connected {
		Connect()
	}

	lockGroups.Lock()
	defer lockGroups.Unlock()

	group := GetGroup(guid)

	if admin != group.Owner {
		return false
	}

	if uuid == group.Owner {
		return true
	}

	for _, v := range group.Admin {
		if v == uuid {
			return true
		}
	}

	group.Admin = append(group.Admin, uuid)

	g := group.Transfer()

	res := db.Table("groups").Where("guid=?", guid).Updates(map[string]interface{}{
		"admin":     g.Admin,
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

func RevokeAdmin(guid, admin, uuid uint32) bool {
	if !Connected {
		Connect()
	}

	lockGroups.Lock()
	defer lockGroups.Unlock()

	group := GetGroup(guid)

	if admin != group.Owner {
		return false
	}

	for k, v := range group.Admin {
		if v == uuid {
			for i := k; i < len(group.Member)-1; i++ {
				group.Admin[i] = group.Admin[i+1]
			}
			group.Admin = group.Admin[:len(group.Admin)-1]
			break
		}
	}

	g := group.Transfer()

	res := db.Table("groups").Where("guid=?", guid).Updates(map[string]interface{}{
		"admin":     g.Admin,
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

func TransferGroup(guid, admin, uuid uint32) bool {
	if !Connected {
		Connect()
	}

	lockGroups.Lock()
	defer lockGroups.Unlock()

	group := GetGroup(guid)

	if admin != group.Owner {
		return false
	}

	// 从管理员列表删除
	for k, v := range group.Admin {
		if v == uuid {
			for i := k; i < len(group.Member)-1; i++ {
				group.Admin[i] = group.Admin[i+1]
			}
			group.Admin = group.Admin[:len(group.Admin)-1]
			break
		}
	}

	// 检查是否是群成员
	if !func()bool{
		for _, v := range group.Member {
			if v == uuid {
				return true
			}
		}
		return false
	}() {
		return false
	}

	group.Owner = uuid

	g := group.Transfer()

	res := db.Table("groups").Where("guid=?", guid).Updates(map[string]interface{}{
		"owner":     uuid,
		"admin":     g.Admin,
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

func DeleteGroupMember(guid, admin, uuid uint32) bool {
	if !Connected {
		Connect()
	}

	lockGroups.Lock()
	defer lockGroups.Unlock()

	group := GetGroup(guid)

	// 可以删除用户和管理员
	if admin == group.Owner {

	}

	if !func() bool {
		// 不能删除群主
		if group.Owner == uuid {
			return false
		}
		// 可以删除用户和管理员
		if group.Owner == admin {
			return true
		}
		// 被删除的用户是管理员，非群主不能删除管理员
		for _, v := range group.Admin {
			if v == uuid {
				return false
			}
		}
		// 非管理员不能删除用户
		for _, v := range group.Admin {
			if v == admin {
				return true
			}
		}
		return false
	}() {
		return false
	}

	// 从管理员列表删除
	for k, v := range group.Admin {
		if v == uuid {
			for i := k; i < len(group.Admin)-1; i++ {
				group.Admin[i] = group.Admin[i+1]
			}
			group.Admin = group.Admin[:len(group.Admin)-1]
			break
		}
	}
	// 从群成员列表删除
	for k, v := range group.Member {
		if v == uuid {
			for i := k; i < len(group.Member)-1; i++ {
				group.Member[i] = group.Member[i+1]
			}
			group.Member = group.Member[:len(group.Member)-1]
			break
		}
	}
	// 从用户的Group中删除

	g := group.Transfer()

	res := db.Table("groups").Where("guid=?", guid).Updates(map[string]interface{}{
		"admin":     g.Admin,
		"member":     g.Member,
	})

	if res.Error != nil {
		log.Println(res.Error)
		return false
	}

	if res.RowsAffected == 0 {
		return false
	}

	DeleteGroup(uuid, guid)

	return true
}