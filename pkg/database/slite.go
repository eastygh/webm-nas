package database

import (
	"github.com/eastygh/webm-nas/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqlite(conf *config.DBConfig) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(conf.Filename), &gorm.Config{})
}
