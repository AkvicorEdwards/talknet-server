package dam

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"sync"
)

var (
	db         *gorm.DB
	Connected  = false
	lockInc    = sync.Mutex{}
	lockUser   = sync.Mutex{}
	lockGroups = sync.Mutex{}
)

// 建立数据库连接
func Connect() {
	if Connected {
		Close()
	}
	var err error
	db, err = gorm.Open("sqlite3", "talknet.db")
	if err != nil {
		log.Println(err)
	}
	Connected = true
}

// 关闭数据库连接
func Close() {
	if !Connected {
		return
	}
	err := db.Close()
	if err != nil {
		log.Println(err)
	}
	Connected = false
}

func Exec(sql string, values ...interface{}) *gorm.DB {
	if !Connected {
		Connect()
	}

	return db.Exec(sql, values...)
}

func Raw(sql string, values ...interface{}) *gorm.DB {
	if !Connected {
		Connect()
	}

	return db.Raw(sql, values...)
}

// 获取自增值
func GetInc(name string) uint32 {
	if !Connected {
		Connect()
	}
	ic := Inc{}
	db.Table("inc").Where("name=?", name).First(&ic)
	return ic.Val
}

// 更新自增值
func UpdateInc(name string, val uint32) bool {
	if !Connected {
		Connect()
	}
	lockInc.Lock()
	defer lockInc.Unlock()
	return db.Table("inc").Where("name=?", name).Update("val", val).Error == nil
}

func LockAll() bool {
	lockInc.Lock()
	lockUser.Lock()
	lockGroups.Lock()
	Close()
	return true
}

func UnlockAll() bool {
	lockInc.Unlock()
	lockUser.Unlock()
	lockGroups.Unlock()
	return true
}
