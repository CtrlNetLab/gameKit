package utils

import (
	"errors"
)

var (
	Err_SendMsgToAI    error = errors.New("send message to ai")
	Err_SendMsgOffline error = errors.New("player offline")
)
