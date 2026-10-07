package gorm_extra

import (
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	l "gorm.io/gorm/logger"

	"backend_go/pkg/format_errors"
)

var (
	postgresInstance *gorm.DB
	resErr           error
	postgresOnce     sync.Once
)

func GetPostgresConn(cfg *Postgres) (*gorm.DB, error) {
	postgresOnce.Do(func() {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s options='-c lock_timeout=10s'",
			cfg.Host, cfg.User, cfg.Password, cfg.Db, cfg.Port, cfg.Sslmode, cfg.TimeZone)
		var err error
		postgresInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			resErr = format_errors.Wrap(err, "ошибка соединения для "+dsn)
			return
		}
		sqlDB, err := postgresInstance.DB()
		if err != nil {
			resErr = format_errors.Wrap(err, "ошибка соединения для "+dsn)
			return
		}
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(0)
		// l.Info чтобы видеть полные логи gorm. l.Silent -- чтобы заглушить
		postgresInstance.Logger = postgresInstance.Logger.LogMode(l.Silent)
	})
	if resErr != nil {
		return nil, resErr
	}
	return postgresInstance, nil
}
