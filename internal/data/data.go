package data

import (
	"context"
	"mengbin92/browser/internal/conf"
	"mengbin92/browser/internal/db"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAccountRepo)

// Data .
type Data struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(logger)

	if err := db.Init(c.Database); err != nil {
		log.Errorf("init database error: %s", err.Error())
		return nil, nil, errors.Wrap(err, "init database error")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Errorf("init redis error: %v", err)
		return nil, nil, errors.Wrap(err, "init redis error")
	}

	d := &Data{
		db:  db.Get(),
		rdb: rdb,
	}

	cleanup := func() {
		log.Info("closing the data resources")
		if err := d.rdb.Close(); err != nil {
			log.Error(err)
		}
	}
	return d, cleanup, nil
}
