package dao

import (
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type AfterCheckAccIdExists func(map[uint64]bool, error)

func IsAccIdExists(accIdList []uint64, cbFunc AfterCheckAccIdExists) {
	if cbFunc == nil {
		return
	}

	if len(accIdList) == 0 {
		cbFunc(nil, errors.New("accId is empty"))
		return
	}

	if len(accIdList) > 1024 {
		cbFunc(nil, errors.New("accId count is too much"))
		return
	}

	rdb := GetRedisRDB(AccId_RedisName, 0)
	if rdb == nil {
		cbFunc(nil, errors.New("get redis client failed"))
		return
	}

	callback := func(result map[uint64]bool, err error) {
		utils.WorkerPool.AddCallback("BatchCheckAccIdExists", func() {
			cbFunc(result, err)
		})
	}

	utils.WorkerPool.Do(0, "BatchCheckAccIdExists", func() {
		pipe := rdb.Pipeline()
		defer pipe.Close()

		cmds := make(map[uint64]*redis.IntCmd)
		for _, accId := range accIdList {
			key := fmt.Sprintf("DontDelUsedAccId:%d", accId)
			cmds[accId] = pipe.Exists(rdb.Context(), key)
		}
		_, err := pipe.Exec(rdb.Context())
		if err != nil && err != redis.Nil {
			logger.Log.Error("redis pipeline exec failed. err: %s", err.Error())
			callback(nil, err)
			return
		}
		resultMap := make(map[uint64]bool, len(accIdList))
		for accId, v := range cmds {
			value, err := v.Result()
			if err != nil && err != redis.Nil {
				logger.Log.Error("redis check accId exist failed. accId: %d, err: %s", accId, err.Error())
				callback(nil, err)
				return
			}
			if value != 0 {
				resultMap[accId] = true
			} else {
				resultMap[accId] = false
			}
		}
		callback(resultMap, nil)
	})
}
