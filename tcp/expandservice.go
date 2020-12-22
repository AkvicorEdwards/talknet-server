package tcp

import (
	"talknet/operator"
)

func AddFriend(inviter, invitee uint32) bool {
	if inviter == invitee {
		//fmt.Println("Cannot add self as friend")
		return false
	}
	if !operator.AddFriendRequest(inviter, invitee) {
		return false
	}

	return true
}
