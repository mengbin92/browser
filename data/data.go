package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mengbin92/browser/conf"
	"github.com/mengbin92/browser/db"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
)

func InitData(conf *conf.Data, logger log.Logger) error {
	log := log.NewHelper(logger)
	err := db.Init(conf.Database)
	if err != nil {
		log.Errorf("failed to open local db: %s", err.Error())
		return errors.Wrap(err, "failed to open local db")
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           int(conf.Redis.Db),
		DialTimeout:  conf.Redis.DialTimeout.AsDuration(),
		WriteTimeout: conf.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  conf.Redis.ReadTimeout.AsDuration(),
	})

	if _, err := rdb.Ping(context.TODO()).Result(); err != nil {
		log.Errorf("init redis error: %v", err)
		return errors.Wrap(err, "init redis error")
	}

	return nil
}
