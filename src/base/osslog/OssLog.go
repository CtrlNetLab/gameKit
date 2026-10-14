package osslog

import (
	"gamekit/pkg/oss"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

const (
	LocalOSSContentTypeJson   = "json"
	LocalOSSContentTypeBase64 = "base64"
)

type ossLog struct {
	gameName         string
	ServerId         int32
	ServerPath       string
	ServerType       string
	IdentifyStr      string
	EnvName          string
	LocalOssFileType string
	ZoneId           string
	HostName         string
	GitCommit        string
	Branch           string
}

var (
	sanitizeReplacer = regexp.MustCompile(`(playerId|roleId|accId|accountId|uid)\s*[=:]\s*\d+`)
)

type IOSSMessage interface {
	proto.Message
	GetOSSCommon() *oss.OSSCommon
	String() string
}

var (
	OSS *ossLog = &ossLog{}
	// buffer []byte  = make([]byte, 64*1024)
)

func (this *ossLog) OssLogInit(gameName, envName, serverType string, serverId int32, zoneId, hostName, gitCommit, branch string) {
	this.gameName = gameName
	this.ServerPath, _ = os.Getwd()
	this.ServerId = serverId
	this.ServerType = serverType
	this.EnvName = envName
	this.ZoneId = zoneId
	this.HostName = hostName
	this.GitCommit = gitCommit
	this.Branch = branch
	this.IdentifyStr = fmt.Sprintf("%s@%s", logger.GetSysUser(), logger.GetLocalIp())

	isDeveloperMode := os.Getenv("AcDeveloperMode")
	if isDeveloperMode != "" {
		this.LocalOssFileType = LocalOSSContentTypeJson
	} else {
		this.LocalOssFileType = LocalOSSContentTypeBase64
	}
}

// 数据格式: format|logtype|time|game|message|data
// format: 1: json; 2: pb
// time: Ymd H:M:S
// logtype: BI|MT|EK, BI: 统计日志; MT: 监控日志; EK: ELK 日志
// message: pb message 名字
// data: pb 或者 json 的 base64 编码
// e.g: 2|BI|1000_cn|PlayerLogin|XXXXX

func (this *ossLog) Log(obj IOSSMessage) {
	ossCommon := obj.GetOSSCommon()
	gameId := ossCommon.GetGameId()
	stype := ossCommon.GetEvent()
	buffer, err := proto.Marshal(obj)

	if err != nil {
		logger.Log.Error("protobuf marshal failed. Err: %v, Obj: %s", err, obj.String())
		return
	}

	b64 := base64.StdEncoding.EncodeToString(buffer)
	nowStr := utils.GetNowYmdHMS()
	logger.OSS.Error("2|BI|%s|%s|%s|%s", nowStr, gameId, stype, b64)
	if this.LocalOssFileType != LocalOSSContentTypeBase64 {
		jsonObj, _ := json.Marshal(obj)
		logger.LocalOSS.Error("2|BI|%s|%s|%s|%s", nowStr, gameId, stype, string(jsonObj))
	} else {
		logger.LocalOSS.Error("2|BI|%s|%s|%s|%s", nowStr, gameId, stype, b64)
	}
}

// captureCallerStack captures the caller stack using runtime.Callers,
// skipping the specified number of frames to focus on business code.
func captureCallerStack(skip int) string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])
	buf := new(bytes.Buffer)
	for {
		frame, more := frames.Next()
		fmt.Fprintf(buf, "%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return buf.String()
}

func sanitizeMessage(msg string) string {
	return sanitizeReplacer.ReplaceAllString(msg, "${1}=<NUM>")
}

func (this *ossLog) ElkAlarm(enmElk oss.EnmElk, accId uint64, format string, args ...any) {
	formatMsg := sanitizeMessage(fmt.Sprintf(format, args...))
	obj := &oss.ElkAlarm{
		Type:   "ElkAlarm",
		EnmElk: enmElk.String(),
		GameId: this.gameName,
		Time:   utils.GetNowYmdHMS(),
		AccId:  accId,
	}

	hostname := this.HostName
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	var stack string
	if enmElk == oss.EnmElk_Panic_Err {
		stack = string(debug.Stack())
	} else {
		stack = filterStack(string(debug.Stack()))
	}
	sendMsg := fmt.Sprintf("#### Elk告警 \n ###### Game: %s\n ###### EnvName: %s\n ###### Time: %s\n ###### ZoneId: %s\n ###### HostName: %s\n ###### Path: %s\n ###### Identify: %s\n ###### ServiceType: %s \n ###### ServiceId: %d\n ###### GitCommit: %s\n ###### Branch: %s\n ###### EnmElk: %s\n ###### AccId: %d\n ##### Msg: %s\n ###### Stack:\n%s",
		this.gameName, this.EnvName, obj.Time, this.ZoneId, hostname,
		this.ServerPath, this.IdentifyStr, this.ServerType, this.ServerId,
		this.GitCommit, this.Branch, obj.EnmElk, obj.AccId, formatMsg, stack)

	obj.Msg = []byte(sendMsg)
	buffer, err := proto.Marshal(obj)
	if err != nil {
		logger.Log.Error("protobuf marshal failed. Err: %#v, Obj: %s", err, obj.String())
		return
	}
	b64 := base64.StdEncoding.EncodeToString(buffer)
	logger.OSS.Error("2|EK|%s|%s|%s|%s", obj.Time, obj.GameId, obj.Type, b64)
	if this.LocalOssFileType != LocalOSSContentTypeBase64 {
		jsonObj, _ := json.Marshal(obj)
		logger.LocalOSS.Error("2|EK|%s|%s|%s|%s", obj.Time, obj.GameId, obj.Type, string(jsonObj))
	} else {
		logger.LocalOSS.Error("2|EK|%s|%s|%s|%s", obj.Time, obj.GameId, obj.Type, b64)
	}
}
func filterStack(stack string) string {
	frames := strings.Split(strings.TrimRight(stack, "\n"), "\n")
	var kept []string
	var total int
	for _, frame := range frames {
		frame = strings.TrimSpace(frame)
		if frame == "" {
			continue
		}
		total++
		if strings.Contains(frame, "/gameServer/src/") || strings.Contains(frame, "\\gameServer\\src\\") || strings.Contains(frame, "/platform/src/") {
			kept = append(kept, frame)
		}
	}
	if len(kept) == 0 && total > 0 {
		kept = append(kept, strings.TrimSpace(frames[0]))
		omitted := total - 1
		if omitted > 0 {
			kept[0] += fmt.Sprintf("\n... (%d frames omitted)", omitted)
		}
	}
	const maxFrames = 8
	if len(kept) > maxFrames {
		omitted := len(kept) - maxFrames
		kept = kept[:maxFrames]
		kept = append(kept, fmt.Sprintf("... (%d business frames omitted)", omitted))
	}
	return strings.Join(kept, "\n")
}

func (this *ossLog) PanicTrace(err any) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s \n", err)
	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x) \n", file, line, pc)
	}
	return buf.String()
}

func NewOssCommon(event string) *oss.OSSCommon {
	uuid := uuid.NewString()
	accountId := "9000000001"
	game := "ac"
	return &oss.OSSCommon{
		Event:      &event,
		DistinctId: &uuid,
		AccountId:  &accountId,
		GameId:     &game,
	}
}
