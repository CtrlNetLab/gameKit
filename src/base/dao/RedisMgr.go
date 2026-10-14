package dao

import (
	"gamekit/src/base/logger"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gamekit/src/base/utils"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	AddrList     []string          `yaml:"AddrList"`
	Passwd       string            `yaml:"Passwd"`
	DB           int               `yaml:"DB"`
	PoolSize     int               `yaml:"PoolSize"`
	MinIdleConns int               `yaml:"MinIdleConns"`
	MaxRetries   int               `yaml:"MaxRetries"`
	Scripts      map[string]string `yaml:"Scripts"`
}

type redisObj struct {
	rdb    *redis.Client
	script map[string]*redis.Script
}

const (
	// 玩家在线状态redis, 地区不共享
	Online_RedisName string = "Online"
	// AccId相关redis， 地区不共享
	AccId_RedisName string = "AccId"
	// 游戏业务缓存redis， 地区不共享
	GameCache_RedisName string = "GameCache"
	// 全局业务缓存redis， 地区共享
	Global_RedisName string = "Global"
	// 玩家在线状态Global副本，地区写，全局读
	OnlineGlobal_RedisName string = "OnlineGlobal"

	// Redis Lua 脚本名(与 run/redislua/ 下的 .lua 文件对应, 新增脚本时在此登记)
	Online_Script          string = "onlineCheck"
	Offline_Script         string = "offline"
	CreateAccId_Script     string = "createAccId"
	SaveToken_Script       string = "saveToken"
	UpdatePlayerIndex_Script string = "updatePlayerIndex"
	FindPlayer_Script      string = "findPlayer"
)

var (
	ctx            = context.Background()
	redisMap       map[string][]*redisObj
	RedisConfigMap map[string]*Redis
)

func (this *redisObj) GetRDB() *redis.Client {
	return this.rdb
}

func GetRedisRDB(name string, primary uint64) *redis.Client {
	rObj := getRedisObj(name, primary)
	if rObj == nil {
		logger.Log.Error("can't find redis config. name: %s, primary: %d", name, primary)
		return nil
	}
	return rObj.rdb
}

// 暂存各个进程redis配置信息， yaml配置热更后需要调用此方法更新
func InitRedisConfigMap(redisConfigMap map[string]*Redis) {
	RedisConfigMap = redisConfigMap
}

func InitRedisMgr() {
	redisMap = make(map[string][]*redisObj)
	for name, rc := range RedisConfigMap {
		rlen := len(rc.AddrList)
		redisMap[name] = make([]*redisObj, rlen, rlen)

		for idx := range rc.AddrList {
			initRedis(name, uint64(idx))
		}
	}
}

func initRedis(name string, primary uint64) *redisObj {
	var (
		ok       bool
		rc       *Redis
		rdb      *redis.Client
		rObj     *redisObj
		rObjList []*redisObj
	)

	if rc, ok = RedisConfigMap[name]; !ok {
		logger.Log.Error("not found redis name config. Name: %s, Primary: %d", name, primary)
		return nil
	}

	if rObjList, ok = redisMap[name]; !ok || len(rObjList) <= 0 {
		logger.Log.Error("not found redis name object. Name: %s, Primary: %d", name, primary)
		return nil
	}

	if len(rc.AddrList) != len(rObjList) {
		logger.Log.Error("redis config and object length not equal. Addr.len: %d, Obj.len: %d, Name: %s, Primary: %d", len(rc.AddrList), len(rObjList), name, primary)
		return nil
	}

	rlen := uint64(len(rc.AddrList))
	idx := int(primary % rlen)

	addr := rc.AddrList[idx]

	// @TODO ryder pool/timeout 相关配置???
	rdb = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     rc.Passwd,
		DB:           rc.DB,
		MinIdleConns: rc.MinIdleConns,
		PoolSize:     rc.PoolSize,
		MaxRetries:   rc.MaxRetries,
	})

	rObj = &redisObj{rdb: rdb, script: make(map[string]*redis.Script)}

	for scriptName, scriptFile := range rc.Scripts {
		bytes, err := os.ReadFile(scriptFile)
		if err != nil {
			logger.Log.Error("read script file failed. Name: %s, File: %s, Err: %#v,errMsg=%s", name, scriptFile, err, err.Error())
			return nil
		}

		script := redis.NewScript(string(bytes))
		sha, serr := script.Load(rObj.rdb.Context(), rObj.rdb).Result()
		if serr != nil {
			logger.Log.Error("load script file failed. Name: %s, File: %s, Err: %#v", name, scriptFile, serr)
			return nil
		}

		rObj.script[scriptName] = script

		logger.Log.Error("init script success. RedisName: %s, ScriptFile: %s, SHA: %s", name, scriptFile, sha)
	}

	rObjList[idx] = rObj

	logger.Log.Error("init redis success. Name: %s, Addr: %s", name, addr)

	return rObj
}

func getRedisObj(name string, primary uint64) *redisObj {
	rObjList, ok := redisMap[name]
	if !ok || len(rObjList) <= 0 {
		logger.Log.Error("can't find %s redisObj", name)
		return nil
	}

	rc, rok := RedisConfigMap[name]
	if !rok {
		logger.Log.Error("can't find %s redis config", name)
		return nil
	}

	objLen := len(rObjList)
	addrLen := len(rc.AddrList)

	if objLen != addrLen {
		logger.Log.Error("length not equal. Name: %s, objLen: %d, addrLen: %d", name, objLen, addrLen)
		return nil
	}

	idx := int(primary % uint64(objLen))

	return rObjList[idx]
}

func GetFloat32(redisName string, primary uint64, key string) float32 {
	strFloat, err := Get(redisName, primary, key)
	if err != nil || strFloat == "" {
		return 0
	}
	return utils.StrToFloat32(strFloat)
}

func Exists(redisName string, primary uint64, key string) (int64, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("get value not found redisObj. RedisName: %s", redisName)
	}
	exists, err := rObj.rdb.Exists(rObj.rdb.Context(), key).Result()
	if err == nil {
		return exists, err
	}

	logger.Log.Error("get redis value failed. redisName: %s, primary: %d, key: %s, value: %v, err: %v", redisName, primary, key, exists, err)
	return exists, err
}

func IncrBy(redisName string, primary uint64, key string, value int64) (uint64, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("not found redisObj. Name: %s", redisName)
	}

	result, err := rObj.rdb.IncrBy(rObj.rdb.Context(), key, value).Result()
	if err == nil {
		return uint64(result), nil
	}

	logger.Log.Error("redis incr failed. Redis: %s, Key: %s, Err: %v", redisName, key, err)
	return uint64(result), err
}

func Incr(redisName string, primary uint64, key string) (uint64, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("not found redisObj. Name: %s", redisName)
	}

	result, err := rObj.rdb.Incr(rObj.rdb.Context(), key).Result()
	if err == nil {
		return uint64(result), nil
	}

	logger.Log.Error("redis incr failed. Redis: %s, Key: %s, Err: %v", redisName, key, err)
	return uint64(result), err
}

func BatchDel(redisName string, primary uint64, keys ...string) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("hash get not found redisObj. RedisName: %s, keys: %s", redisName, keys)
	}
	_, err := rObj.rdb.Del(rObj.rdb.Context(), keys...).Result()
	if err == nil || err == redis.Nil {
		return nil
	}
	logger.Log.Error("hash BatchDel redis value failed. redisName: %s, primary: %d, keys: %v, err: %v", redisName, primary, keys, err)
	return err
}

func HGet(redisName string, primary uint64, hash string, fields ...string) ([]interface{}, error) {
	var (
		rObj *redisObj
	)
	var none []interface{}
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return none, fmt.Errorf("hash get not found redisObj. RedisName: %s, hash: %s", redisName, hash)
	}

	values, err := rObj.rdb.HMGet(rObj.rdb.Context(), hash, fields...).Result()
	if err == nil || err == redis.Nil {
		return values, nil
	}

	logger.Log.Error("hash get redis value failed. redisName: %s, primary: %d, hash: %v, err: %v", redisName, primary, hash, err)
	return none, err
}

func ZRangeByScore(redisName string, primary uint64, key string, start, end int64) ([]string, error) {
	var (
		rObj *redisObj
	)
	var none []string
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return none, fmt.Errorf("ZRangeByScore get not found redisObj. RedisName: %s, hash: %s", redisName, key)
	}
	r := &redis.ZRangeBy{Min: fmt.Sprintf("%d", start), Max: fmt.Sprintf("%d", end)}
	values, err := rObj.rdb.ZRangeByScore(rObj.rdb.Context(), key, r).Result()
	if err == nil || err == redis.Nil {
		return values, nil
	}
	logger.Log.Error("ZRangeByScore get redis value failed. redisName: %s, primary: %d, hash: %v, err: %v", redisName, primary, key, err)
	return values, err
}

// 批量hash incr
func HIncrBatch(redisName string, primary uint64, hash string, expire int32, incr int64, fields ...string) (map[string]int64, error) {
	var (
		rObj *redisObj
	)
	var incrMap = make(map[string]int64, 0)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return incrMap, fmt.Errorf("hash HIncrBatch not found redisObj. RedisName: %s, hash: %s", redisName, hash)
	}
	pipe := rObj.rdb.Pipeline()
	defer func(pipe redis.Pipeliner) {
		if pipe == nil {
			return
		}
		err := pipe.Close()
		if err != nil {
			logger.Log.Error("HIncrBatch pipe close err=%s", err.Error())
		}
	}(pipe)
	cmds := make(map[string]*redis.IntCmd)
	for _, field := range fields {
		cmds[field] = pipe.HIncrBy(rObj.rdb.Context(), hash, field, incr)
	}
	if expire > 0 {
		pipe.Expire(rObj.rdb.Context(), hash, time.Duration(expire)*time.Second)
	}
	_, err := pipe.Exec(rObj.rdb.Context())
	if err != nil && err != redis.Nil {
		logger.Log.Error("%d|redis pipeline exec failed,primary=%d err: %s", primary, err.Error())
		return nil, err
	}
	//if len(cmds) != len(fields) {
	//	logger.Log.Error("redis HIncrBatch len(cmds) != len(fields)")
	//	return nil, err
	//}
	for sid, v := range cmds {
		value, err := v.Result()
		if err != nil && err != redis.Nil {
			logger.Log.Error("%d|redis HIncrBatch  failed, err: %s", err.Error())
			continue
		}
		if value != 0 {
			incrMap[sid] = value
		}
	}
	return incrMap, err
}

func HMSet(redisName string, primary uint64, hash string, fields ...interface{}) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("hash get not found redisObj. RedisName: %s, hash: %s", redisName, hash)
	}

	_, err := rObj.rdb.HMSet(rObj.rdb.Context(), hash, fields...).Result()
	if err == nil || err == redis.Nil {
		return nil
	}

	logger.Log.Error("hash get redis value failed. redisName: %s, primary: %d, hash: %v, err: %v", redisName, primary, hash, err)
	return err
}

// 批量新增zset
func ZAddBatch(redisName string, primary uint64, key string, fields map[string]float64) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("ZAddBatch get not found redisObj. RedisName: %s, hash: %s", redisName, key)
	}
	if len(fields) == 0 {
		return fmt.Errorf("ZAddBatch get not found redisObj. RedisName: %s, hash: %s", redisName, key)
	}
	var arr = make([]redis.Z, 0)
	for field, value := range fields {
		arr = append(arr, redis.Z{Member: field, Score: value})
	}

	zAddArgs := redis.ZAddArgs{
		Members: arr,
	}
	_, err := rObj.rdb.ZAddArgs(rObj.rdb.Context(), key, zAddArgs).Result()
	if err == nil || err == redis.Nil {
		return nil
	}
	logger.Log.Error("ZAddBatch value failed. redisName: %s, primary: %d, hash: %v, err: %v", redisName, primary, key, err)
	return err
}
func HDel(redisName string, primary uint64, hash string, fields ...string) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("hash hdel not found redisObj. RedisName: %s, hash: %s", redisName, hash)
	}

	_, err := rObj.rdb.HDel(rObj.rdb.Context(), hash, fields...).Result()
	if err == nil || err == redis.Nil {
		return nil
	}
	logger.Log.Error("hash hdel redis value failed. redisName: %s, primary: %d, hash: %v, err: %v", redisName, primary, hash, err)
	return err
}

func HAll(redisName string, primary uint64, hash string) (map[string]string, error) {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return nil, fmt.Errorf("hash HAll not found redisObj. RedisName: %s, hash: %s", redisName, hash)
	}

	result, err := rObj.rdb.HGetAll(rObj.rdb.Context(), hash).Result()
	if err == nil || err == redis.Nil {
		return result, nil
	}
	logger.Log.Error("hash HAll redis value failed. redisName: %s, primary: %d, hash: %v, err: %v", redisName, primary, hash, err)
	return result, err
}

func HGetGameId(onlineKey string, accId uint64) uint64 {
	primary := uint64(0)
	redisName := Online_RedisName
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0
	}
	key := fmt.Sprintf("%s:%d", onlineKey, accId)
	value, err := rObj.rdb.HGet(rObj.rdb.Context(), key, "gid").Result()
	if err == nil || err == redis.Nil {
		return utils.StrToUint64(value)
	}
	return 0
}

func ExecScript(redisName, scriptName string, primary uint64, args ...interface{}) (interface{}, error) {
	var (
		rObj   *redisObj
		script *redis.Script
		ok     bool
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return nil, fmt.Errorf("not found redisObj. RedisName: %s, ScriptName: %s", redisName, scriptName)
	}

	var (
		none   []string
		err    error
		result interface{}
	)

	if script, ok = rObj.script[scriptName]; !ok {
		return nil, fmt.Errorf("not found redis script or script exception. RedisName: %s, ScriptName: %s", redisName, scriptName)
	}

	result, err = script.EvalSha(rObj.rdb.Context(), rObj.rdb, none, args...).Result()
	if err == nil {
		return result, err
	}

	if errors.Is(err, redis.Nil) {
		return result, nil
	}

	logger.Log.Error("exec redis script failed. RedisName: %s, ScriptName: %s, Args: %v, Err: %v", redisName, scriptName, args, err)

	// 出错重试
	if strings.Contains(err.Error(), "NOSCRIPT No matching script. Please use EVAL.") {
		loadResult, err := script.Load(rObj.rdb.Context(), rObj.rdb).Result()
		if err != nil {
			logger.Log.Error("load redis script failed. RedisName: %s, ScriptName: %s, Err: %v", redisName, scriptName, err)
			return nil, err
		}
		logger.Log.Error("reload redis script success. RedisName: %s, ScriptName: %s, SHA: %s", redisName, scriptName, loadResult)
	}

	result, err = script.EvalSha(rObj.rdb.Context(), rObj.rdb, none, args...).Result()
	if err == nil {
		return result, err
	}

	logger.Log.Error("exec redis script failed again. RedisName: %s, ScriptName: %s, Args: %v, Err: %v", redisName, scriptName, args, err)

	return result, err
}

func LRange(redisName string, primary uint64, key string, start, stop int64) (error, []string) {
	var (
		rObj *redisObj
	)

	var none []string
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key), none
	}

	values, err := rObj.rdb.LRange(rObj.rdb.Context(), key, start, stop).Result()
	if err == nil {
		return nil, values
	}
	return err, none
}

func LPush(redisName string, primary uint64, key string, values ...interface{}) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return
	}

	_, err := rObj.rdb.LPush(rObj.rdb.Context(), key, values).Result()
	if err == nil {
		return
	}

	logger.Log.Error("redis Lpush failed. Redis: %s, Key: %s, Err: %v", redisName, key, err)
}

func Del(redisName string, primary uint64, key string) error {
	rObj := getRedisObj(redisName, primary)
	if rObj == nil {
		return fmt.Errorf("not found redisObj. Name: %s", redisName)
	}

	err := rObj.rdb.Del(rObj.rdb.Context(), key).Err()
	return err
}

func ZRange(redisName string, primary uint64, key string, start, stop int64) ([]string, error) {
	var (
		rObj *redisObj
	)

	var none []string
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return none, fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}

	values, err := rObj.rdb.ZRange(rObj.rdb.Context(), key, start, stop).Result()
	if err == nil {
		return values, nil
	}

	return none, err
}

func MGetValue(redisName string, primary uint64, keys ...string) ([]interface{}, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return nil, fmt.Errorf("get value not found redisObj. RedisName: %s", redisName)
	}

	value, err := rObj.rdb.MGet(rObj.rdb.Context(), keys...).Result()
	if err == nil || err == redis.Nil {
		return value, nil
	}

	logger.Log.Error("get redis value failed. redisName: %s, primary: %d, key: %s, value: %v, err: %v", redisName, primary, keys, value, err)
	return nil, err
}

func ZIncr(redisName string, primary uint64, key, field string, incr float64) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("ZIncr get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}
	_, err := rObj.rdb.ZIncrBy(rObj.rdb.Context(), key, incr, field).Result()
	if err != nil {
		logger.Log.Error("ZIncr failed. key: %s, field: %s, incr: %f, err: %s", key, field, incr, err.Error())
		return err
	}
	return err
}
func ZRank(key, member string) (int64, error) {
	primary := uint64(0)
	redisName := Online_RedisName
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}

	values, err := rObj.rdb.ZRank(rObj.rdb.Context(), key, member).Result()
	return values, err
}

func Get(redisName string, primary uint64, key string) (string, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return "", fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}

	value, err := rObj.rdb.Get(rObj.rdb.Context(), key).Result()
	if err == nil || err == redis.Nil {
		return value, nil
	}

	logger.Log.Error("get redis value failed. redisName: %s, primary: %d, key: %s, value: %v, err: %v", redisName, primary, key, value, err)
	return "", err
}

func ZAdd(redisName string, primary uint64, key string, member string, score float64) (int64, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}

	zAdd := &redis.Z{
		Score:  score,
		Member: member,
	}

	values, err := rObj.rdb.ZAdd(rObj.rdb.Context(), key, zAdd).Result()
	return values, err
}

func ZRem(redisName string, primary uint64, key string, member string) (int64, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}

	values, err := rObj.rdb.ZRem(rObj.rdb.Context(), key, member).Result()
	return values, err
}

func Set(redisName string, primary uint64, key string, value interface{}, expire int64) (string, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return "", fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}
	result, err := rObj.rdb.Set(rObj.rdb.Context(), key, value, time.Duration(expire)*time.Second).Result()
	if err == nil || err == redis.Nil {
		return result, nil
	}

	logger.Log.Error("get redis value failed. redisName: %s, primary: %d, key: %s, value: %v, err: %v", redisName, primary, key, value, err)
	return "", err
}

func SAdd(redisName string, primary uint64, key string, value ...interface{}) (int64, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("hash get not found redisObj. RedisName: %s, key: %s", redisName, key)
	}
	result, err := rObj.rdb.SAdd(rObj.rdb.Context(), key, value).Result()
	if err == nil || err == redis.Nil {
		return result, nil
	}
	logger.Log.Error(" redis SAdd failed. redisName: %s, primary: %d, key: %s, value: %v, err: %v", redisName, primary, key, value, err)
	return 0, err
}

func BatchTimeLineMsgZAddAndRem(redisName string, primary uint64, key string, value []byte, score int64, maxCount int64) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("hash BatchSismember not found redisObj. RedisName: %s", redisName)
	}
	pipe := rObj.rdb.Pipeline()
	defer func(pipe redis.Pipeliner) {
		if pipe == nil {
			return
		}
		err := pipe.Close()
		if err != nil {
			logger.Log.Error("BatchZAddAndRem pipe close err=%s", err.Error())
		}
	}(pipe)
	pipe.ZAdd(ctx, key, &redis.Z{Member: value, Score: float64(score)})
	if maxCount > 0 {
		//添加一个zset并且删除掉超过maxCount的
		pipe.ZRemRangeByRank(ctx, key, 0, -maxCount-1)
	} else {
		//删除过期的
		nwoTime := utils.GetTime()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", nwoTime))
	}
	_, err := pipe.Exec(rObj.rdb.Context())
	return err
}

// 批量lpush 并且删除 前N条
func BatchLPushTrim(redisName string, primary uint64, key string, limit int32, delKey string, values ...interface{}) {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return
	}
	pipe := rObj.rdb.Pipeline()
	defer func(pipe redis.Pipeliner) {
		if pipe == nil {
			return
		}
		err := pipe.Close()
		if err != nil {
			logger.Log.Error("BatchLPushTrim pipe close err=%s", err.Error())
		}
	}(pipe)
	if len(delKey) > 0 {
		pipe.LRem(rObj.rdb.Context(), key, 0, delKey)
	}
	pipe.LPush(rObj.rdb.Context(), key, values...)
	pipe.LTrim(rObj.rdb.Context(), key, 0, int64(limit)-1)
	_, err := pipe.Exec(rObj.rdb.Context())
	if err != nil {
		logger.Log.Error("BatchLPushTrim pipe exec err=%s", err.Error())
	}
}

func BatchGetTimeLineMsg(redisName string, primary uint64, maxCountMap map[string]int64) (error, []string) {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("hash BatchGetGiftMsg not found redisObj. RedisName: %s", redisName), nil
	}
	pipe := rObj.rdb.Pipeline()
	defer func(pipe redis.Pipeliner) {
		if pipe == nil {
			return
		}
		err := pipe.Close()
		if err != nil {
			logger.Log.Error("BatchGetGiftMsg pipe close err=%s", err.Error())
		}
	}(pipe)
	results := make(map[string]*redis.StringSliceCmd)
	for key, count := range maxCountMap {
		results[key] = pipe.ZRevRange(ctx, key, 0, count)
	}
	_, err := pipe.Exec(rObj.rdb.Context())
	if err != nil {
		logger.Log.Error("%d|redis pipeline exec failed. err: %s", primary, err.Error())
		return err, nil
	}
	var giftMsgStr []string
	for _, cmds := range results {
		giftMsgStr = append(giftMsgStr, cmds.Val()...)
	}
	return err, giftMsgStr
}

// zset 批量查询
func BatchZRank(redisName string, primary uint64, limit int32, keys ...string) (map[string]*redis.ZSliceCmd, error) {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return nil, fmt.Errorf("hash BatchZRank not found redisObj. RedisName: %s", redisName)
	}
	pipe := rObj.rdb.Pipeline()
	defer func(pipe redis.Pipeliner) {
		if pipe == nil {
			return
		}
		err := pipe.Close()
		if err != nil {
			logger.Log.Error("BatchZRank pipe close err=%s", err.Error())
		}
	}(pipe)
	cmds := make(map[string]*redis.ZSliceCmd)
	for _, key := range keys {
		//查询zet某字段倒叙排序
		cmds[key] = pipe.ZRevRangeWithScores(rObj.rdb.Context(), key, 0, int64(limit))
	}
	_, err := pipe.Exec(rObj.rdb.Context())
	if err != nil && err != redis.Nil {
		logger.Log.Error("%d|redis pipeline exec failed,primary=%d err: %s", primary, err.Error())
		return nil, err
	}
	//for key, v := range cmds {
	//	value, err := v.Result()
	//	if err != nil && err != redis.Nil {
	//		logger.Log.Error("%d|redis HIncrBatch  failed, err: %s", err.Error())
	//		continue
	//	}
	//	existMap[key] = value
	//}
	return cmds, err
}

func BatchSismember(redisName string, primary uint64, field interface{}, keys ...string) (map[string]bool, error) {
	var (
		rObj *redisObj
	)
	var existMap = make(map[string]bool)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return existMap, fmt.Errorf("hash BatchSismember not found redisObj. RedisName: %s", redisName)
	}
	pipe := rObj.rdb.Pipeline()
	defer func(pipe redis.Pipeliner) {
		if pipe == nil {
			return
		}
		err := pipe.Close()
		if err != nil {
			logger.Log.Error("BatchSismember pipe close err=%s", err.Error())
		}
	}(pipe)
	cmds := make(map[string]*redis.BoolCmd)
	for _, key := range keys {
		cmds[key] = pipe.SIsMember(rObj.rdb.Context(), key, field)
	}
	_, err := pipe.Exec(rObj.rdb.Context())
	if err != nil && err != redis.Nil {
		logger.Log.Error("%d|redis pipeline exec failed,primary=%d err: %s", primary, err.Error())
		return nil, err
	}
	for key, v := range cmds {
		value, err := v.Result()
		if err != nil && err != redis.Nil {
			logger.Log.Error("%d|redis HIncrBatch  failed, err: %s", err.Error())
			continue
		}
		existMap[key] = value
	}
	return existMap, err
}

func ZSetDel(redisName string, primary uint64, key string, keys ...string) (int64, error) {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("ZSetDel  found redisObj. RedisName: %s, key: %s", redisName, key)
	}
	result, err := rObj.rdb.ZRem(rObj.rdb.Context(), key, keys).Result()
	if err == nil || err == redis.Nil {
		return 0, nil
	}
	logger.Log.Error("get redis value failed. redisName: %s, primary: %d, key: %s, err: %v", redisName, primary, key, err)
	return result, err
}

func SIsMember(redisName string, primary uint64, key string, field string) (bool, error) {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return false, fmt.Errorf("HExists  found redisObj. RedisName: %s, key: %s", redisName, key)
	}
	result, err := rObj.rdb.SIsMember(rObj.rdb.Context(), key, field).Result()
	if err == redis.Nil {
		return true, nil
	}
	//logger.Log.Error("SIsMember failed. redisName: %s, primary: %d, key: %s, err: %v", redisName, primary, key, err)
	return result, err
}
func SRem(redisName string, primary uint64, key string, field string) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("HExists  found redisObj. RedisName: %s, key: %s", redisName, key)
	}
	rObj.rdb.SRem(rObj.rdb.Context(), key, field)
	return nil
}

func BatchSRem(redisName string, primary uint64, key string, field ...string) error {
	var (
		rObj *redisObj
	)
	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return fmt.Errorf("BatchSRem  found redisObj. RedisName: %s, key: %s", redisName, key)
	}
	rObj.rdb.SRem(rObj.rdb.Context(), key, field)
	return nil
}

func GetZScore(redisName string, primary uint64, key string) (map[string]int, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return nil, fmt.Errorf("get value not found redisObj. RedisName: %s", redisName)
	}

	pvpRanking := make(map[string]int, 0)

	membersWithScores, err := rObj.rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err == nil || err == redis.Nil {
		for _, memberWithScore := range membersWithScores {
			member := memberWithScore.Member.(string)
			score := memberWithScore.Score
			pvpRanking[member] = int(score)
		}
	}

	return pvpRanking, err
}
func ZScore(redisName string, primary uint64, key string, member string) (float64, error) {
	var (
		rObj *redisObj
	)

	if rObj = getRedisObj(redisName, primary); rObj == nil {
		return 0, fmt.Errorf("ZScore not found redisObj. RedisName: %s", redisName)
	}
	result, err := rObj.rdb.ZScore(ctx, key, member).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}
