package postgres

import (
	"gosilite/pkg/env"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time"`
}

func New() (*gorm.DB, error) {
	return NewWithConfig(&PostgresConfig{
		DSN:             env.GetString("DATABASE_DSN"),
		MaxIdleConns:    env.GetIntD("DATABASE_MAX_IDLE_CONNS", 10),
		MaxOpenConns:    env.GetIntD("DATABASE_MAX_OPEN_CONNS", 100),
		ConnMaxLifetime: env.GetIntD("DATABASE_CONN_MAX_LIFETIME", 5),
		ConnMaxIdleTime: env.GetIntD("DATABASE_CONN_MAX_IDLE_TIME", 5),
	})
}

func NewWithConfig(config *PostgresConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DSN),
		&gorm.Config{
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
		},
	)
	if err != nil {
		panic(err)
	}

	DB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// 连接池配置
	DB.SetMaxIdleConns(config.MaxIdleConns)
	DB.SetMaxOpenConns(config.MaxOpenConns)
	DB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	DB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)

	return db, nil
}
