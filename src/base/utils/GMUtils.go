package utils

const (
	CmdType_Control int = 1
	CmdType_Gm      int = 2
	// 按位与 CmdType_All = CmdType_Control | CmdType_Gm
	CmdType_All int = 3
)

const (
	// 没有 GM 权限
	GmType_None int = 0
	// 1 级 GM 权限, 只读权限
	GmType_Lv1 int = 1

	// root GM 权限
	GmType_Root int = 99
)
