package model

// 登陆状态
type LoginState int32

const (
	// 未登陆
	LoginState_Init LoginState = 0
	// 登陆成功
	LoginState_Success LoginState = 1
)

// 玩家下线原因
type LogoutReason int32

const (
	// 正常下线
	LogoutReason_Normal LogoutReason = 0
	// 连接关闭
	LogoutReason_CloseSock LogoutReason = 1
	// 重复登录被踢下线
	LogoutReason_DupLogin LogoutReason = 2
)
