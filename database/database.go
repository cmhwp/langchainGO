package database

import (
	"langchaingo-demo/config"
	"langchaingo-demo/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.DatabaseConfig) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移数据库表
	err = DB.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
	)
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
