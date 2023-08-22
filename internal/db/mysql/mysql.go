package mysql

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

func InitDB(source string) (*gorm.DB, error) {
	dblog := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
			SlowThreshold:             time.Second,
		},
	)
	return gorm.Open(mysql.Open(source), &gorm.Config{
		SkipDefaultTransaction:                   true,
		AllowGlobalUpdate:                        false,
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   dblog,
	})
}
