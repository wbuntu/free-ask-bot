package storage

import (
	"context"
	"log"
	"os"
	"time"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	sharedDB   *gorm.DB
	sharedKVDB *redis.Client
)

func DB() *gorm.DB {
	return sharedDB
}

func KVDB() *redis.Client {
	return sharedKVDB
}

func Setup(ctx context.Context, c *config.Config) error {
	var err error
	// 初始化关系型数据库连接池
	if c.General.EnableDB {
		sharedDB, err = NewDB(ctx, c)
		if err != nil {
			return errors.Wrap(err, "NewDB")
		}
	}
	// 初始化键值数据库连接池
	if c.General.EnableKVDB {
		sharedKVDB, err = NewKVDB(ctx, c)
		if err != nil {
			return errors.Wrap(err, "NewKVDB")
		}
	}
	return nil
}

func NewDB(ctx context.Context, c *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch c.DB.Type {
	case "sqlite":
		dialector = sqlite.Open(c.DB.DSN)
	case "mysql":
		dialector = mysql.Open(c.DB.DSN)
	default:
		return nil, errors.Errorf("unsupported db type: %s", c.DB.Type)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "get db")
	}
	sqlDB.SetMaxIdleConns(c.DB.MinIdleConns)
	sqlDB.SetMaxOpenConns(c.DB.MaxActiveConns)
	sqlDB.SetConnMaxLifetime(c.DB.ConnLifetime)
	sqlDB.SetConnMaxIdleTime(c.DB.ConnIdletime)
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "ping db")
	}
	return db, nil
}

func NewKVDB(ctx context.Context, c *config.Config) (*redis.Client, error) {
	redisOpt, err := redis.ParseURL(c.KVDB.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "parse kv_db dsn")
	}
	redisOpt.MinIdleConns = c.KVDB.MinIdleConns
	redisOpt.PoolSize = c.KVDB.MaxActiveConns
	redisOpt.MaxConnAge = c.KVDB.ConnLifetime
	redisOpt.IdleTimeout = c.KVDB.ConnIdletime
	db := redis.NewClient(redisOpt)
	if _, err := db.Ping(ctx).Result(); err != nil {
		return nil, errors.Wrap(err, "ping kv_db")
	}
	return db, nil
}

type Filter struct {
	Query interface{}
	Args  []interface{}
}
