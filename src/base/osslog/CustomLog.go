package osslog

import (
	"encoding/base64"

	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
)

// LogCustom writes a custom log line with the same base format as BI logs.
// format: 1=json, 2=pb
// logType: BI/AD/FB/EK...
func (this *ossLog) LogCustom(format int, logType, gameId, event string, data []byte) {
	nowStr := utils.GetNowYmdHMS()
	b64 := base64.StdEncoding.EncodeToString(data)

	logger.OSS.Error("%d|%s|%s|%s|%s|%s", format, logType, nowStr, gameId, event, b64)
	if this.LocalOssFileType != LocalOSSContentTypeBase64 && format == 1 {
		logger.LocalOSS.Error("%d|%s|%s|%s|%s|%s", format, logType, nowStr, gameId, event, string(data))
	} else {
		logger.LocalOSS.Error("%d|%s|%s|%s|%s|%s", format, logType, nowStr, gameId, event, b64)
	}
}
