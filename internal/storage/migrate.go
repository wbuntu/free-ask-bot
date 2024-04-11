package storage

import (
	"context"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Migrate(ctx context.Context, c *config.Config) error {
	if c.General.EnableDB {
		// TODO：使用redis或其他中间件创建一个分布式锁，确保迁移动作只被一个Pod执行
		// gorm的AutoMigrate只创建或更新数据库表结构
		if err := DB().AutoMigrate(
			&ResourceID{},
			&Cluster{},
			&Task{},
			&TaskLog{},
		); err != nil {
			return errors.Wrap(err, "migrate model")
		}
		// gormigrate可修改或更新数据
		migrationsOptions := gormigrate.DefaultOptions
		migrationsOptions.UseTransaction = true
		m := gormigrate.New(DB(), gormigrate.DefaultOptions, migrations)
		if err := m.Migrate(); err != nil {
			return errors.Wrap(err, "migrate data")
		}
	}
	return nil
}

// migrations 用年月日时分秒作为ID，每个迁移动作都在事务中执行，执行成功一次后记录到数据库，不再执行
var migrations = []*gormigrate.Migration{
	{
		ID: "2024-01-01-00-00-00",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("update clusters set type = ? where type = ?", "k8s", "").Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("update clusters set type = ? where type = ?", "", "k8s").Error
		},
	},
}
