package def

const (
	LengthUsername = 50
	LengthPassword = 130
	LengthNickname = 50

	// LoginData
	LoginDataOffsetUsernameLength = 0
	LoginDataOffsetUsername       = 2
	LoginDataLengthUsername       = 50
	LoginDataOffsetPasswordLength = 1
	LoginDataOffsetPassword       = 52
	LoginDataLengthPassword       = 130

	// Message
	MessageOffsetMessageLength = 0
	MessageLengthMessageLength = 1
	MessageOffsetUUID = 1
	MessageLengthUUID = 4
	MessageOffsetMessage = 5
	MessageLengthMessage = 178

	MessageOffsetGroupMessageLength = 0
	MessageLengthGroupMessageLength = 1
	MessageOffsetGroupGUID = 1
	MessageLengthGroupGUID = 4
	MessageOffsetGroupUUID = 5
	MessageLengthGroupUUID = 4
	MessageOffsetGroupMessage = 9
	MessageLengthGroupMessage = 174


	// UserInfo uuid username nickname
	UserInfoOffsetUsernameLength = 0
	UserInfoLengthUsernameLength = 1
	UserInfoOffsetNicknameLength = 1
	UserInfoLengthNicknameLength = 1
	UserInfoOffsetUUID = 2
	UserInfoLengthUUID = 4
	UserInfoOffsetUsername = 6
	UserInfoLengthUsername = 50
	UserInfoOffsetNickname = 56
	UserInfoLengthNickname = 50
)
