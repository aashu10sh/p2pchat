package db

import (
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func GetDB() *gorm.DB {
	once.Do(func() {
		var err error
		db, err = gorm.Open(sqlite.Open("peers.db"), &gorm.Config{})
		if err != nil {
			panic("error getting database")
		}

		AutoMigrate(db)
	})

	return db
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&Profile{})
	db.AutoMigrate(&Peer{})
	db.AutoMigrate(&Chat{})
	db.AutoMigrate(&Message{})
}

type Database struct {
	Db *gorm.DB
}

func GetDatabase() *Database {
	return &Database{
		Db: GetDB(),
	}
}
