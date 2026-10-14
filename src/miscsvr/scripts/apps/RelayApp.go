package apps

import (
	ss "gamekit/pkg/ss"
	relay "gamekit/src/base/relay"
)

func init() {
	RegistServiceHandle(ss.EnmCmdValue_ECV_SSRegistRes, relay.HandleServRegistRes)
	RegistServiceHandle(ss.EnmCmdValue_ECV_SSHeartRes, relay.HandleServHeartRes)
}
