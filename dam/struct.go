package dam

import "encoding/json"

type Inc struct {
	Name string // 表名
	Val  uint32 // 自增值
}

//type UnacceptedStruct struct {
//	Friends []uint32 `json:"friends"`
//	Groups  []uint32 `json:"groups"`
//}
//
//func (u *UnacceptedStruct) TransferFriends() string {
//	data, _ := json.Marshal(u.Friends)
//	return string(data)
//}
//
//func (u *UnacceptedStruct) TransferGroups() string {
//	data, _ := json.Marshal(u.Groups)
//	return string(data)
//}

type User struct {
	Uuid       uint32           `json:"uuid"`
	Username   string           `json:"username"`
	Nickname   string           `json:"nickname"`
	Password   string           `json:"password"`
	Friends    []uint32         `json:"friend"`
	Groups     []uint32         `json:"groups"`
	Unaccepted []uint32 `json:"unaccepted"`
	Deleted    int64            `json:"deleted"`
}

func (u *User) TransferFriends() string {
	data, _ := json.Marshal(u.Friends)
	return string(data)
}

func (u *User) TransferGroups() string {
	data, _ := json.Marshal(u.Groups)
	return string(data)
}

func (u *User) TransferUnaccepted() string {
	data, _ := json.Marshal(u.Unaccepted)
	return string(data)
}

func (u *User) Transfer() (user UserDB) {
	user = UserDB{
		Uuid:       u.Uuid,
		Username:   u.Username,
		Nickname:   u.Nickname,
		Password:   u.Password,
		Friends:    "",
		Groups:     "",
		Unaccepted: "",
		Deleted:    u.Deleted,
	}
	data, _ := json.Marshal(u.Friends)
	user.Friends = string(data)
	data, _ = json.Marshal(u.Groups)
	user.Groups = string(data)
	data, _ = json.Marshal(u.Unaccepted)
	user.Unaccepted = string(data)
	return
}

type UserDB struct {
	Uuid       uint32
	Username   string
	Nickname   string
	Password   string
	Friends    string
	Groups     string
	Unaccepted string
	Deleted    int64
}

func (u *UserDB) Transfer() (user User) {
	user = User{
		Uuid:       u.Uuid,
		Username:   u.Username,
		Nickname:   u.Nickname,
		Password:   u.Password,
		Friends:    make([]uint32, 0),
		Groups:     make([]uint32, 0),
		Unaccepted: make([]uint32, 0),
		Deleted:    u.Deleted,
	}
	_ = json.Unmarshal([]byte(u.Friends), &user.Friends)
	_ = json.Unmarshal([]byte(u.Groups), &user.Groups)
	_ = json.Unmarshal([]byte(u.Unaccepted), &user.Unaccepted)
	return
}

type Files struct {
	Fuid uint32 `json:"fuid"` // 文件id
	Uuid uint32 `json:"uuid"` // 上传者
	Filename string `json:"filename"`
	RealName string `json:"real_name"`
	Hash uint32 `json:"hash"`
}

type Group struct {
	Guid       uint32
	Name       string
	Owner      uint32
	Admin      []uint32
	Member     []uint32
	Unaccepted []uint32
	Files []Files
	Deleted    int64
}

func (g *Group) Transfer() (group GroupDB) {
	group = GroupDB{
		Guid:       g.Guid,
		Name:       g.Name,
		Owner:      g.Owner,
		Admin:      "",
		Member:     "",
		Unaccepted: "",
		Files: "",
		Deleted:    g.Deleted,
	}
	data, _ := json.Marshal(g.Admin)
	group.Admin = string(data)
	data, _ = json.Marshal(g.Member)
	group.Member = string(data)
	data, _ = json.Marshal(g.Unaccepted)
	group.Unaccepted = string(data)
	data, _ = json.Marshal(g.Files)
	group.Files = string(data)
	return
}

type GroupDB struct {
	Guid       uint32
	Name       string
	Owner      uint32
	Admin      string
	Member     string
	Unaccepted string
	Files string
	Deleted    int64
}

func (g *GroupDB) Transfer() (group Group) {
	group = Group{
		Guid:       g.Guid,
		Name:       g.Name,
		Owner:      g.Owner,
		Admin:      make([]uint32, 0),
		Member:     make([]uint32, 0),
		Unaccepted: make([]uint32, 0),
		Files: make([]Files, 0),
		Deleted:    g.Deleted,
	}
	_ = json.Unmarshal([]byte(g.Admin), &group.Admin)
	_ = json.Unmarshal([]byte(g.Member), &group.Member)
	_ = json.Unmarshal([]byte(g.Unaccepted), &group.Unaccepted)
	_ = json.Unmarshal([]byte(g.Files), &group.Files)
	return
}
