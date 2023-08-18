package pg

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	return gorm.Open(postgres.Open(source), &gorm.Config{SkipDefaultTransaction: true, Logger: dblog})
}
