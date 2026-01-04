package db

import (
	"errors"
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

func (d *Database) GetChatByPeers(selfPeerId string, theirPeerId string) (*Chat, error) {
	var chat Chat

	err := d.Db.Where(&Chat{TheirPeerId: theirPeerId, MyPeerId: selfPeerId}).First(&chat).Error

	if err != nil {
		return nil, err
	}

	return &chat, err
}

func (d *Database) GetPeerByID(peerId string) (*Peer, error) {
	var peer Peer

	err := d.Db.Where(&Peer{PeerId: peerId}).First(&peer).Error

	if err != nil {
		return nil, err
	}

	return &peer, nil
}

func (d *Database) CreateChat(chat *Chat) error {
	d.Db.Create(chat)

	if chat.ID == 0 {
		return errors.New("could not create!")
	}

	return nil
}

func (d *Database) CreateMessage(msg *Message) error {

}
