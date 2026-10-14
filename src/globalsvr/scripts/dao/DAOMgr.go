package dao

import (
	"fmt"
	"time"

	gdao "gamekit/src/base/dao"
	"gamekit/src/base/logger"
	"globalsvr/scripts/config"

	_ "github.com/go-sql-driver/mysql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

const (
	GlobalDataName string = "GlobalData"
	FriendDataName string = "FriendData"
)

type DAOMgr struct {
	// 数据库对应的 gorm.DB 句柄
	DBMap map[string][]*gorm.DB
}

var (
	daomgr *DAOMgr

	newLogger glog.Interface = glog.New(
		logger.GormWriter{}, // io writer
		glog.Config{
			SlowThreshold: 200 * time.Millisecond, // 慢 SQL 阈值
			Colorful:      false,                  // 禁用彩色打印
			LogLevel:      glog.Warn,              // Log level
		},
	)
)

func InitMySQL() {
	daomgr = new(DAOMgr)
	daomgr.DBMap = make(map[string][]*gorm.DB)

	mySQLMap := config.Conf.MySQLMap
	for k, v := range mySQLMap {
		initDatabase(k, &v)
	}
}

func initDatabase(name string, mySQL *config.MySQL) {
	dbList := make([]*gorm.DB, 0)

	for _, v := range mySQL.HostList {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mySQL.User, mySQL.Passwd, v, mySQL.Port, mySQL.DBName, mySQL.CharSet)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
		if err != nil {
			panic(fmt.Sprintf("Open MySQL failed. Name: %s, Host: %s, DNS: %s, Err: %v", name, v, dsn, err))
		}

		sqlDB, err := db.DB()
		if err != nil {
			panic(err)
		}

		sqlDB.SetMaxIdleConns(mySQL.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(mySQL.ConnMaxLifetime) * time.Second)
		sqlDB.SetMaxOpenConns(mySQL.MaxOpenConns)

		if mySQL.Debug {
			dbList = append(dbList, db.Debug())
		} else {
			dbList = append(dbList, db)
		}
	}

	daomgr.DBMap[name] = dbList
}

/***
 * @brief 数据库句柄
 *
 * @param[in] dbName 数据库名字
 * @param[in] primary 数据库主键
 *
 * @return 返回数据库句柄
 */
func GetDB(dbName string, primary uint64) *gorm.DB {
	dbList, ok := daomgr.DBMap[dbName]
	if !ok || len(dbList) <= 0 {
		logger.Log.Error("can't find %s database handle", dbName)
		return nil
	}

	dbLen := len(dbList)
	mySQL, mok := config.Conf.MySQLMap[dbName]
	if !mok {
		logger.Log.Error("can't find %s database config", dbName)
		return nil
	}

	if dbLen != len(mySQL.HostList) {
		logger.Log.Error("can't %s database, dblen(%d) != mySQL.HostList.len(%d)", dbName, dbLen, len(mySQL.HostList))
		return nil
	}

	splitNum := uint64(mySQL.SplitNum)
	idx := int((primary % splitNum) / (splitNum / uint64(dbLen)))
	if idx < 0 || idx >= dbLen {
		logger.Log.Error("%s database calc idx invalid. idx: %d, dbLen: %d, primary: %d, splitnum: %d",
			dbName, idx, dbLen, primary, mySQL.SplitNum)
		return nil
	}

	return dbList[idx]
}

func reInitHandler(dbName string, primary uint64) *gorm.DB {
	dbList, ok := daomgr.DBMap[dbName]
	if !ok || len(dbList) <= 0 {
		logger.Log.Error("can't find %s database handle", dbName)
		return nil
	}

	dbLen := len(dbList)
	mySQL, mok := config.Conf.MySQLMap[dbName]
	if !mok {
		logger.Log.Error("can't find %s database config", dbName)
		return nil
	}

	if dbLen != len(mySQL.HostList) {
		logger.Log.Error("can't %s database, dblen(%d) != mySQL.HostList.len(%d)", dbName, dbLen, len(mySQL.HostList))
		return nil
	}

	splitNum := uint64(mySQL.SplitNum)
	idx := int((primary % splitNum) / (splitNum / uint64(dbLen)))
	if idx < 0 || idx >= dbLen {
		logger.Log.Error("%s database calc idx invalid. idx: %d, dbLen: %d, primary: %d, splitnum: %d",
			dbName, idx, dbLen, primary, mySQL.SplitNum)
		return nil
	}

	// @TODO ryder 是否需要 close 老句柄???

	host := mySQL.HostList[idx]
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mySQL.User, mySQL.Passwd, host, mySQL.Port, mySQL.DBName, mySQL.CharSet)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		logger.Log.Error("Open MySQL failed. Name: %s, Host: %s, DNS: %s, Err: %v", dbName, host, dsn, err)
		return nil
	}

	if mySQL.Debug {
		dbList[idx] = db.Debug()
	} else {
		dbList[idx] = db
	}

	return db
}

func GetTableName(dbName string, primary uint64) string {
	mySQL, mok := config.Conf.MySQLMap[dbName]
	if !mok {
		logger.Log.Error("can't find %s database config", dbName)
		return "unknown_table"
	}
	splitNum := uint64(mySQL.SplitNum)
	if splitNum <= 1 {
		return mySQL.TablePrefix
	}
	return fmt.Sprintf("%s%d", mySQL.TablePrefix, primary%splitNum)
}

/***
 * @brief 获取数据库表数据，如果 mysql 句柄失效则会尝试重连一次
 *
 * @param[in] dbName 数据库名
 * @param[int] primaryKey 主键值，用于分表
 * @param[out] out 返回的数据集合
 * @param[in] args 参数列表,
 *            args[0]: where 语句，为 nil 时选整张表
 *            args[1]: offset
 *            args[2]: limit
 */
func GetTableData(dbName string, primary uint64, out interface{}, args ...interface{}) error {
	db := GetDB(dbName, primary)
	if db == nil {
		return fmt.Errorf("Get gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
	}

	tableName := GetTableName(dbName, primary)

	argc := len(args)
	for i := 0; i < 2; i++ {
		var err error
		if argc < 1 {
			// 扫描全表
			err = db.Table(tableName).Find(out).Error
		} else if argc == 1 {
			// 有 where
			if args[0] == nil {
				err = db.Table(tableName).Find(out).Error
			} else {
				err = db.Table(tableName).Where(args[0]).Find(out).Error
			}
		} else if argc == 2 {
			// 有 where offset
			if args[0] == nil {
				err = db.Table(tableName).Offset(args[1].(int)).Find(out).Error
			} else {
				err = db.Table(tableName).Where(args[0]).Offset(args[1].(int)).Find(out).Error
			}
		} else if argc > 2 {
			// 有 where offset limit
			if args[0] == nil {
				err = db.Table(tableName).Offset(args[1].(int)).Limit(args[2].(int)).Find(out).Error
			} else {
				err = db.Table(tableName).Where(args[0]).Offset(args[1].(int)).Limit(args[2].(int)).Find(out).Error
			}
		}

		// 数据库没有错误，则不需要重试
		if err == nil || err == gorm.ErrRecordNotFound {
			return nil
		}

		logger.Log.Error("GetTableData failed. DBName: %s, Primary: %d, err: %#v", dbName, primary, err)

		if gdao.IsNeedReInitDB(err) {
			// 数据库有错误则尝试重连之后再 select
			db = reInitHandler(dbName, primary)
			if db == nil {
				return fmt.Errorf("reInitHandler gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
			}
		}
	}
	return fmt.Errorf("unknown failed. DBName: %s, Primary: %d", dbName, primary)
}

func InsertData(dbName string, primary uint64, out interface{}) error {
	db := GetDB(dbName, primary)
	if db == nil {
		return fmt.Errorf("Get gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
	}

	tableName := GetTableName(dbName, primary)

	for i := 0; i < 2; i++ {
		if err := db.Table(tableName).Create(out).Error; err == nil {
			// 数据库没有错误，不需要重试
			return nil
		} else {
			logger.Log.Error("InsertData failed. DBName: %s, Primary: %d, err: %#v", dbName, primary, err)

			if gdao.IsNeedReInitDB(err) {
				// 重试
				db = reInitHandler(dbName, primary)
				if db == nil {
					return fmt.Errorf("reInitHandler gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
				}
			}
		}
	}

	return fmt.Errorf("unknown failed. DBName: %s, Primary: %d", dbName, primary)
}

func UpdateDataModel(dbName string, primary uint64, where interface{}, m interface{}, updates map[string]interface{}) error {
	db := GetDB(dbName, primary)
	if db == nil {
		return fmt.Errorf("Get gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
	}

	tableName := GetTableName(dbName, primary)

	for i := 0; i < 2; i++ {
		if err := db.Table(tableName).Model(m).Where(where).Updates(updates).Error; err == nil {
			return nil
		} else {
			logger.Log.Error("UpdateData failed. DBName: %s, Primary: %d, err: %#v", dbName, primary, err)
			if gdao.IsNeedReInitDB(err) {
				// 重试
				db = reInitHandler(dbName, primary)
				if db == nil {
					return fmt.Errorf("reInitHandler gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
				}
			}
		}
	}

	return fmt.Errorf("unknown failed. DBName: %s, Primary: %d", dbName, primary)
}

func UpdateData(dbName string, primary uint64, where map[string]interface{}, out interface{}) error {
	db := GetDB(dbName, primary)
	if db == nil {
		return fmt.Errorf("Get gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
	}

	tableName := GetTableName(dbName, primary)

	for i := 0; i < 2; i++ {
		if err := db.Table(tableName).Where(where).Updates(out).Error; err == nil {
			return nil
		} else {
			logger.Log.Error("UpdateData failed. DBName: %s, Primary: %d, err: %#v", dbName, primary, err)

			if gdao.IsNeedReInitDB(err) {
				// 重试
				db = reInitHandler(dbName, primary)
				if db == nil {
					return fmt.Errorf("reInitHandler gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
				}
			}
		}
	}

	return fmt.Errorf("unknown failed. DBName: %s, Primary: %d", dbName, primary)
}

func DeleteData(dbName string, primary uint64, where interface{}, m interface{}) error {
	db := GetDB(dbName, primary)
	if db == nil {
		return fmt.Errorf("Get gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
	}

	tableName := GetTableName(dbName, primary)

	for i := 0; i < 2; i++ {
		if err := db.Table(tableName).Where(where).Delete(m).Error; err == nil {
			return nil
		} else {
			logger.Log.Error("DeleteData failed. DBName: %s, Primary: %d, err: %#v", dbName, primary, err)

			if gdao.IsNeedReInitDB(err) {
				// 重试
				db = reInitHandler(dbName, primary)
				if db == nil {
					return fmt.Errorf("reInitHandler gorm.DB failed. DBName: %s, Primary: %d", dbName, primary)
				}
			}
		}
	}

	return fmt.Errorf("unknown failed. DBName: %s, Primary: %d", dbName, primary)
}
