package operator

import "talknet/dam"

func CreateGroup(name string, owner uint32) (uint32, bool) {
	return dam.CreateGroup(name, owner)
}

func JoinGroup(uuid, guid uint32) bool {
	return dam.JoinGroupRequest(uuid, guid)
}

func GetJoinGroupList(guid uint32) string {
	return dam.GetJoinGroupList(guid)
}

func AcceptJoinGroup(guid, admin, uuid uint32) bool {
	return dam.AcceptJoinGroup(guid, admin, uuid)
}

func GetGroup(guid uint32) *dam.Group {
	return dam.GetGroup(guid)
}

func AppointAdmin(guid, admin, uuid uint32) bool {
	return dam.AppointAdmin(guid, admin, uuid)
}

func RevokeAdmin(guid, admin, uuid uint32) bool {
	return dam.RevokeAdmin(guid, admin, uuid)
}

func TransferGroup(guid, admin, uuid uint32) bool {
	return dam.TransferGroup(guid, admin, uuid)
}

func DeleteGroupMember(guid, admin, uuid uint32) bool {
	return dam.DeleteGroupMember(guid, admin, uuid)
}

func AddFileInfoToGroup(guid, uuid uint32, filename, realName string, hash uint32) bool {
	return dam.AddFileInfoToGroup(guid, uuid, filename, realName, hash)
}

func DeleteFileInfoFromGroup(guid, uuid uint32, fuid uint32) bool {
	return dam.DeleteFileInfoFromGroup(guid, uuid, fuid)
}


