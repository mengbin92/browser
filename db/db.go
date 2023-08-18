package db

import (
	"fmt"
	"sync"

	"github.com/mengbin92/browser/conf"
	"github.com/mengbin92/browser/db/mysql"
	"github.com/mengbin92/browser/db/pg"
	"github.com/mengbin92/browser/db/sqlite3"

	"gorm.io/gorm"
)

var (
	gdb      *gorm.DB
	initOnce sync.Once
)

// Init inits the database connection only once
func Init(conf *conf.Database) error {
	var err error

	initOnce.Do(func() {
		if conf.Driver == "postgre" {
			gdb, err = pg.InitDB(conf.Source)
		} else if conf.Driver == "sqlite" {
			gdb, err = sqlite3.InitDB(conf.Source)
		} else {
			gdb, err = mysql.InitDB(conf.Source) // MySQL is default
		}
	})

	sqlDB, err := gdb.DB()
	if err != nil {
		panic(fmt.Sprintf("set connection error: %s", err.Error()))
	}
	sqlDB.SetMaxIdleConns(int(conf.MaxIdleConn))
	sqlDB.SetMaxOpenConns(int(conf.MaxOpenConn))
	sqlDB.SetConnMaxLifetime(conf.GetConnMaxLifetime().AsDuration())
	return err
}

func Get() *gorm.DB {
	if gdb == nil {
		panic("db is nil")
	}

	return gdb
}
