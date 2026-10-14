package utils

import (
	"gamekit/src/base/logger"
	"math"
	"strings"
	"time"
)

var MonthDayList [][]int = [][]int{
	[]int{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31},
	[]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31},
}

// 2099年
const LONG_TIME_FUTURE float64 = 4070880000
const LONG_TIME_MS_FUTURE float64 = 4070880000000
const ONE_DAY_TIME_STAMP_MS int64 = 86400000
const DATE_YMD_HIS_FORMAT = "2006-01-02 15:04:05"
const DATE_YMD_FORMAT = "2006-01-02"
const HOURTOSEC uint32 = 3600

// 三个月过期时间
const THREE_MONTH_TO_SEC int64 = 7776000
const ONE_DAY_TIME_STAMP int64 = 86400
const FIVE_CLOCK_OFFSET int64 = 18000
const FOUR_CLOCK_OFFSET int64 = 14400

var (
	ZONE_AsiaShanghai = "Asia/Shanghai" //中国时区
	ZONE_AsiaIndia    = "Asia/Kolkata"  //印度时区
)
var (
	gNow                 time.Time
	gLastTime            time.Time
	gUnixTime            int64
	gUnixMillisecond     int64
	gLastUnixMillisecond int64
	gDeltaNano           int64
	gNowTimeStr          string
	gLoc                 *time.Location = nil
)

func InitTime(reset bool) {
	now := time.Now()
	if reset {
		gLastTime = gNow
		gDeltaNano = now.UnixNano() - gNow.UnixNano()
	} else {
		gLastTime = now
		gDeltaNano = 0
	}
	gNow = now
	if gUnixTime != gNow.Unix() {
		gNowTimeStr = now.Format(time.DateTime)
	}
	gUnixTime = gNow.Unix()
	gUnixMillisecond = gNow.UnixNano() / 1e6
	gLastUnixMillisecond = gLastTime.UnixNano() / 1e6

	if gLoc == nil {
		gLoc = now.Location()
	}

	// StatMgr.StatFuncCall("InitTime", TimeSince(now))
}

func ResetTime() {
	InitTime(true)
}

func GetTime() int64 {
	return gUnixTime
}

func GetTimeMS() int64 {
	return gUnixMillisecond
}

func GetTimeMSPointer() *int64 {
	return &gUnixMillisecond
}

func GetNextXDaysTime(days int64) int64 {
	if days < 0 {
		return time.Now().Unix()
	}
	return time.Now().Unix() + days*86400
}

func GetLastTimeMS() int64 {
	return gLastUnixMillisecond
}

func GetDeltaMS() int64 {
	return int64(gDeltaNano / 1e6)
}

func IsWeekDay(dayList []uint32) bool {
	week := uint32(time.Now().Weekday())
	if week == 0 {
		week = 7
	}

	for _, day := range dayList {
		if day == week {
			return true
		}
	}
	return false
}

// 判断是否在时间范围内
func InTimeSpan(start, end, check time.Time) bool {
	start, end = start.UTC(), end.UTC()
	if start.After(end) {
		start, end = end, start
	}
	check = check.UTC()
	return !check.Before(start) && !check.After(end)
}

// 判断是否在时间范围内
func InTimestampSpan(startTime, endTime, checkTime int64) bool {
	start := time.Unix(startTime/1000, 0)
	end := time.Unix(endTime/1000, 0)
	check := time.Unix(checkTime/1000, 0)
	start, end = start.UTC(), end.UTC()
	if start.After(end) {
		start, end = end, start
	}
	check = check.UTC()
	return !check.Before(start) && !check.After(end)
}

func GetMonthDay(year, month uint32) int {
	var leap int = 1

	if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
		leap = 0
	}
	return MonthDayList[leap][month]
}

// 毫秒时间戳时间属于那个月份
func GetMonth(timestamp int64) int32 {
	t := time.Unix(0, timestamp).In(time.Local)
	return int32(t.Month())
}

func GetTimeYearMonthDay(t time.Time) uint32 {
	return uint32(t.Year()*10000 + int(t.Month())*100 + t.Day())
}

func GetNowYearMonth() (uint32, uint32) {
	nowTime := time.Now()
	return uint32(nowTime.Year()), uint32(nowTime.Month())
}

func GetLastMonth() int {
	res := int(time.Now().Month()) - 1
	if res == 0 {
		res = 12
	}
	return res
}

func GetNextMonth() int {
	res := int(time.Now().Month()) + 1
	if res == 13 {
		res = 1
	}
	return res
}

func GetNextDayZero() int64 {
	nextDay := time.Now().AddDate(0, 0, 1)
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, time.Local).UnixMilli()
}

func GetPlusDayZero(day int) int64 {
	nextDate := time.Now().AddDate(0, 0, day)
	return time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), 0, 0, 0, 0, time.Local).UnixMilli()
}

func GetNowMonth() int {
	return int(time.Now().Month())
}

// 获取当前时间的“2006-01-02 15:04:05”格式字符串
func GetNowYmdHMS() string {
	return gNowTimeStr
}

func GetYmdHMS(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func GetYmdHMSM(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

func DirFormatYmdH(t time.Time) string {
	return t.Format("2006/01/02/15")
}

func GetYearMonthDay(year, month, day uint32) uint32 {
	return year*10000 + month*100 + day
}

func ConvertDateToUnixTime(date string, srclayout string, srczone string) (int64, error) {
	srcloc, err := time.LoadLocation(srczone)
	if err != nil {
		return 0, err
	}
	t, err := time.ParseInLocation(srclayout, date, srcloc)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func ConvertStringToUnixTime(str string) int64 {
	if gLoc == nil {
		gLoc = time.Now().Location()
	}
	formatTm, err := time.ParseInLocation("2006-1-2 15:04:05", str, gLoc)
	if err != nil {
		return 0
	}

	return formatTm.Unix()
}

// 判断指定时间戳是否在指定格式时间内格式:2026-2-14 00:00:00
func CheckTimeDateInRange(nowTime int64, startDate string, endDate string) bool {
	startTime := ConvertStringToUnixTime(startDate)
	endTime := ConvertStringToUnixTime(endDate)
	return startTime <= nowTime && nowTime < endTime
}

func ConvertYMDStringToUnixTime(str string) int64 {
	if gLoc == nil {
		gLoc = time.Now().Location()
	}
	formatTm, err := time.ParseInLocation("2006-1-2", str, gLoc)
	if err != nil {
		return 0
	}

	return formatTm.Unix()
}

func FormatUnixTime(v int64) string {
	return GetYmdHMS(time.Unix(v, 0))
}

func IsSameDay(t1 int64, t2 int64) bool {
	y1, m1, d1 := time.Unix(t1, 0).Date()
	y2, m2, d2 := time.Unix(t2, 0).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsSameDayMS(t1 int64, t2 int64) bool {
	return IsSameDay(t1/1000, t2/1000)
}

func IsSameMonth(t1 int64, t2 int64) bool {
	_, m1, _ := time.Unix(t1, 0).Date()
	_, m2, _ := time.Unix(t2, 0).Date()
	return m1 == m2
}

func IsSameMonthMS(t1 int64, t2 int64) bool {
	return IsSameMonth(t1/1000, t2/1000)
}

func IsSameWeek(timestamp1 int64, timestamp2 int64) bool {
	t1 := time.Unix(timestamp1, 0)
	t2 := time.Unix(timestamp2, 0)
	year1, week1 := t1.ISOWeek()
	year2, week2 := t2.ISOWeek()
	return year1 == year2 && week1 == week2
}

func IsSameWeekMS(t1 int64, t2 int64) bool {
	logger.Log.Debug("IsSameWeekMS execute t1: %d, t2: %d.", t1, t2)
	return IsSameWeek(t1/1000, t2/1000)
}

func GetZeroTimeStamp(t1 int64, offset int64) int64 {
	t := time.Unix(t1, 0)
	if gLoc == nil {
		gLoc = time.Now().Location()
	}
	zeroTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, gLoc)
	return zeroTime.Unix() + offset*ONE_DAY_TIME_STAMP
}

func GetZeroTimeStampByLastMonthDay(day int) int64 {
	lastMonthTime := time.Unix(GetZeroTimeStampByMonthDay(1)-1, 0)
	if gLoc == nil {
		gLoc = time.Now().Location()
	}
	targetDate := time.Date(lastMonthTime.Year(), lastMonthTime.Month(), day, 0, 0, 0, 0, gLoc)
	return targetDate.Unix()
}

func GetZeroTimeStampByMonthDay(day int) int64 {
	if gLoc == nil {
		gLoc = time.Now().Location()
	}
	targetDate := time.Date(time.Now().Year(), time.Now().Month(), day, 0, 0, 0, 0, gLoc)
	return targetDate.Unix()
}

func GetZeroTimeStampByWeekDay(weekDay int) int64 {
	if gLoc == nil {
		gLoc = time.Now().Location()
	}

	now := time.Now()
	unixTime := GetZeroTimeStamp(now.Unix(), 0)
	var days int64
	w := now.Weekday()
	if w == 0 {
		days = 6
	} else {
		days = int64(w - 1)
	}
	unixTime -= (days * 86400)
	return unixTime + int64((weekDay-1)*86400)
}

func GetZeroTimeStampByClock(clock int) int64 {
	if gLoc == nil {
		gLoc = time.Now().Location()
	}

	targetDate := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), clock, 0, 0, 0, gLoc)
	return targetDate.Unix()
}

// The number of seconds until tomorrow's zero
func GetUntilTomorrowZero() int64 {
	// 获取当前服务器时间
	now := time.Now()

	// 获取明天零点的时间
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	// 计算距离明天零点还相差的秒数
	return int64(tomorrow.Sub(now).Seconds())
}
func GetHMSTime(hms string) int64 {
	fields := strings.Split(hms, ":")
	if len(fields) != 3 {
		return 0
	}
	h := StrToInt64(fields[0])
	m := StrToInt64(fields[1])
	s := StrToInt64(fields[2])
	return h*3600 + m*60 + s
}

// 反向取值，值越大，则时间越早，目前用于排行榜的对于redis中sorted set类型的score分值做排序计算
func AbsUTCUnixTimeDifference() float64 {
	differenceTimestamp := LONG_TIME_FUTURE - float64(GetTime())
	return differenceTimestamp
}

func GetDiffDays(left int64, right int64) int32 {
	leftDate := time.Unix(left, 0)
	rightDate := time.Unix(right, 0)

	return int32(leftDate.Sub(rightDate).Hours() / 24)
}

// CalcDiffDay 自然日的相差天数
func CalcDiffDay(t1, t2 int64) int {
	time1 := time.Unix(t1, 0)
	time2 := time.Unix(t2, 0)
	zeroTimt1 := time.Date(time1.Year(), time1.Month(), time1.Day(), 0, 0, 0, 0, time.Local)
	zeroTimt2 := time.Date(time2.Year(), time2.Month(), time2.Day(), 0, 0, 0, 0, time.Local)

	return int(math.Abs(zeroTimt1.Sub(zeroTimt2).Hours() / 24))
}

func CalcDiffDayMS(t1, t2 int64) int {
	return CalcDiffDay(t1/1000, t2/1000)
}

func GetNextMonth1thTime() int64 {
	year, month, _ := time.Now().Date()
	if (month+1)%12 == 1 {
		month = time.January
		year++
	}

	return time.Date(year, month, 1, 0, 0, 0, 0, time.Local).UnixMilli()
}

func GetNextDayAppointTime(sTime string) int64 {
	sHourMin := strings.Split(sTime, ":")
	if len(sHourMin) != 2 {
		//logger.Log.Error("Invalid time config, %s", sTime)
		return 0
	}

	hour := StrToInt64(sHourMin[0])
	minute := StrToInt64(sHourMin[1])
	return GetNextDayZero() + hour*3600 + minute*60
}

func TimeSince(t time.Time) time.Duration {
	now := time.Now()
	return now.Sub(t)
}

// 判断某个时间与当前时间是否处于同一周
func JudgeSameWeek(timestamp int64) bool {
	timeJudge := time.UnixMilli(timestamp)
	weekdayJudge := timeJudge.Weekday()
	var sunDayTimestampJudge int64
	if int(weekdayJudge) != 0 {
		daysUntilSundayA := 7 - int(weekdayJudge)
		sunDayTimestampJudge = timeJudge.AddDate(0, 0, daysUntilSundayA).UnixNano() / int64(time.Millisecond)
	} else {
		sunDayTimestampJudge = timestamp
	}
	weekNow := time.Now().Weekday()
	daysUntilSundayB := 7 - int(weekNow)
	sunDayTimestampNow := time.Now().AddDate(0, 0, daysUntilSundayB).UnixNano() / int64(time.Millisecond)
	// 判断两个时间戳是否处于同一个周中
	if math.Abs(float64(sunDayTimestampNow-sunDayTimestampJudge)) < float64(ONE_DAY_TIME_STAMP_MS) {
		return true
	} else {
		return false
	}
}

// 传入指定时间格式返回当天0点的时间戳
func GetDayZeroTimestamps(dateStr, format string) int64 {
	// "2006-1-2"
	t, err := time.ParseInLocation(format, dateStr, time.Local)
	if err != nil {
		return 0
	}
	// 获取当日0点时间
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return startTime.UnixMilli()
}

// 传入指定时间格式返回当天最后一秒的时间戳
func GetDayLastTimestamps(dateStr, format string) int64 {
	// "2006-1-2"
	t, err := time.ParseInLocation(format, dateStr, time.Local)
	if err != nil {
		return 0
	}
	// 获取当日0点时间
	endTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
	return endTime.UnixMilli()
}

// 获取下一个周一零点时间
func NextMondayZero(now time.Time) time.Time {
	weekday := int(now.Weekday())
	//0=周日，1=周一，..., 6=周六
	daysUntilNextMonday := (8 - weekday) % 7
	if daysUntilNextMonday == 0 {
		daysUntilNextMonday = 7
	}

	nextMonday := now.AddDate(0, 0, daysUntilNextMonday)
	midnight := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, now.Location())
	return midnight
}

// 判断timestampMS+ttlMinutes是否超过当前时间（是否过期），true表示过期，false表示未过期
func IsExpiredAt(timestampMS int64, ttlMinutes int32, now time.Time) bool {
	if timestampMS <= 0 || ttlMinutes <= 0 {
		return true
	}
	expireAt := time.UnixMilli(timestampMS).Add(time.Duration(ttlMinutes) * time.Minute)
	return !expireAt.After(now)
}

// 获取下个月的最后一天的时间戳（毫秒）
func GetLastDayOfMonthTimestamp(now time.Time) int64 {
	currentMonth := now.Month()
	year := now.Year()
	// 获取下一个月的月份
	nextMonth := currentMonth + 1
	if nextMonth > 12 {
		// 明年1月份
		nextMonth = 1
		// 明年
		year += 1
	}
	return time.Date(year, nextMonth, 0, 23, 59, 59, 0, now.Location()).Unix() * 1000
}

// 获取本周的最后一天的时间戳（毫秒）
func GetWeekLastTimestamp(now time.Time) int64 {
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	endDay := now.AddDate(0, 0, daysUntilSunday)
	end := time.Date(endDay.Year(), endDay.Month(), endDay.Day(), 23, 59, 59, 999_000_000, now.Location())
	return end.UnixMilli()
}

// 获取当天最后一天的时间戳（毫秒）
func GetTodayEndTimestamp(now time.Time) int64 {
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Unix() * 1000
	return todayEnd
}

// 获取当天开始时间戳（毫秒）
func GetTodayZeroTimestamp(now time.Time) int64 {
	zero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return zero.UnixMilli()
}

// 获取本周开始时间戳（毫秒）
func GetWeekStartTimestamp(now time.Time) int64 {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 周日修正为 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	zero := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())
	return zero.UnixMilli()
}

// 获取本月开始时间戳（毫秒）
func GetMonthStartTimestamp(now time.Time) int64 {
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return first.UnixMilli()
}

// GetEndOfDayAfterNDaysMs 返回 N 天后的 23:59:59.999 的毫秒级时间戳
func GetEndOfDayAfterNDaysMs(n int) int64 {
	if n < 1 {
		n = 0
	} else {
		n -= 1
	}
	now := time.Now()
	target := now.AddDate(0, 0, n)

	endOfDay := time.Date(
		target.Year(),
		target.Month(),
		target.Day(),
		23, 59, 59,
		999*1e6, // 999 毫秒的纳秒数
		target.Location(),
	)

	return endOfDay.UnixNano() / 1e6 // 纳秒 → 毫秒
}

// 获取当天0点时间
func GetDayStart(ts int64) int64 {
	t := time.UnixMilli(ts)
	dayStart := time.Date(
		t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0,
		time.Local,
	)
	return dayStart.UnixMilli()
}

// 获取指定时间戳N天后的时间戳（毫秒）
func GetTimestampAfterDays(timestampMs int64, days int64) int64 {
	// return time.UnixMilli(timestampMs).AddDate(0, 0, days).UnixMilli()
	// 或者
	return timestampMs + int64(days)*24*60*60*1000
}
