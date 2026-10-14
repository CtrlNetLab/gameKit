package dao

import (
	"gamekit/src/base/logger"
	"errors"

	"github.com/go-sql-driver/mysql"
)

func IsNeedReInitDB(dbErr error) bool {
	// MySQL数据库错误， 不需要重连
	var sqlError *mysql.MySQLError
	if errors.As(dbErr, &sqlError) {
		logger.Log.Error("mysql error. errCode: %d, errMsg: %s", sqlError.Number, sqlError.Message)
		return false
	}

	// 默认需要重连
	return true
}
