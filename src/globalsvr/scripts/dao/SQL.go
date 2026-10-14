package dao

import "fmt"

type DBGlobalData struct {
	Id         uint32 `gorm:"column:Id"`
	GlobalData []byte `gorm:"column:GlobalData"`
}

func (this *DBGlobalData) String() string {
	return fmt.Sprintf("{Global Func Id: %d}", this.Id)
}
