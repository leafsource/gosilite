package infra

import (
	"context"
	"fmt"

	"gosilite/application/infra/cache"
	"gosilite/application/infra/cache/impl/redis"
	"gosilite/application/infra/hook"
	"gosilite/application/infra/imagex"
	"gosilite/application/infra/orm/impl/postgres"
	"gosilite/application/infra/storage"
	"gosilite/application/infra/storage/impl/s3"
	"gosilite/conf"

	"gorm.io/gorm"
)

type AppDependencies struct {
	DB       *gorm.DB
	Cache    cache.CacheClient
	Storage  storage.Storage
	ImageOSS imagex.ImageX
}

func Init(ctx context.Context) (*AppDependencies, error) {
	conf := conf.GetConf()
	deps := &AppDependencies{}
	var err error
	// 初始化数据库
	deps.DB, err = postgres.NewWithConfig(&postgres.PostgresConfig{
		DSN:             conf.Database.DSN,
		MaxIdleConns:    conf.Database.MaxIdleConns,
		MaxOpenConns:    conf.Database.MaxOpenConns,
		ConnMaxLifetime: conf.Database.ConnMaxLifetime,
		ConnMaxIdleTime: conf.Database.ConnMaxIdleTime,
	})
	if err != nil {
		return nil, fmt.Errorf("init database failed: %w", err)
	}
	deps.DB.Use(&hook.BaseModelPlugin{})

	// 初始化缓存
	deps.Cache = redis.NewWithConfig(conf.Redis.Address, conf.Redis.Password)
	// 初始化日志
	// 初始化配置

	//初始化Storage
	deps.Storage, err = s3.New(ctx, conf.Storage.AccessKey, conf.Storage.SecretKey, conf.Storage.BucketName, conf.Storage.Endpoint, conf.Storage.Region)
	if err != nil {
		return nil, fmt.Errorf("init storage failed: %w", err)
	}
	deps.ImageOSS, err = s3.NewStorageImagex(ctx, conf.Storage.AccessKey, conf.Storage.SecretKey, conf.Storage.BucketName, conf.Storage.Endpoint, conf.Storage.Region, conf.Storage.Domain)
	if err != nil {
		return nil, fmt.Errorf("init image storage failed: %w", err)
	}
	return deps, nil
}
