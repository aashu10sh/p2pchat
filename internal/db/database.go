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
	db.AutoMigrate(&FileTransfer{})
	db.AutoMigrate(&CallHistory{})
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

	if peerId == "" {
		return nil, errors.New("unknown user name")
	}

	err := d.Db.Where(&Peer{PeerId: peerId}).First(&peer).Error

	if err != nil {
		return nil, err
	}

	return &peer, nil
}

func (d *Database) CreateChat(chat *Chat) error {
	d.Db.Create(chat)

	if chat.ID == 0 {
		return errors.New("could not create chat!")
	}

	return nil
}

func (d *Database) CreateMessage(msg *Message) error {
	d.Db.Create(msg)

	if msg.ID == 0 {
		return errors.New("could not create message!")
	}

	return nil
}

func (d *Database) UpdateChat(chat *Chat) error {
	tx := d.Db.Save(chat)
	return tx.Error
}

func (d *Database) UpdateMessage(msg *Message) error {
	tx := d.Db.Save(msg)
	return tx.Error
}

func (d *Database) GetMessagesByChatID(chatID uint, limit int) ([]*Message, error) {
	var messages []*Message

	query := d.Db.Where("chat_id = ?", chatID).Order("sent_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&messages).Error
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (d *Database) GetChatsByPeerID(peerID string) ([]*Chat, error) {
	var chats []*Chat

	err := d.Db.Where("my_peer_id = ?", peerID).Order("last_message_at DESC").Find(&chats).Error
	if err != nil {
		return nil, err
	}

	return chats, nil
}

func (d *Database) GetChatByID(chatID uint) (*Chat, error) {
	var chat Chat

	err := d.Db.Where("id = ?", chatID).First(&chat).Error
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (d *Database) CreatePeer(peer *Peer) error {
	d.Db.Create(peer)

	if peer.ID == 0 {
		return errors.New("could not create peer!")
	}

	return nil
}

func (d *Database) UpdatePeer(peer *Peer) error {
	tx := d.Db.Save(peer)
	return tx.Error
}

func (d *Database) GetAllPeers() ([]*Peer, error) {
	var peers []*Peer

	err := d.Db.Order("last_seen DESC").Find(&peers).Error
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func (d *Database) CreateFileTransfer(ft *FileTransfer) error {
	return d.Db.Create(ft).Error
}

// GetFileTransfersByPeer returns file transfers between self and a given peer, newest first.
func (d *Database) GetFileTransfersByPeer(selfPeerID, theirPeerID string) ([]*FileTransfer, error) {
	var transfers []*FileTransfer

	err := d.Db.Where(
		"(from_peer_id = ? AND to_peer_id = ?) OR (from_peer_id = ? AND to_peer_id = ?)",
		selfPeerID, theirPeerID, theirPeerID, selfPeerID,
	).Order("created_at DESC").Find(&transfers).Error
	if err != nil {
		return nil, err
	}

	return transfers, nil
}

func (d *Database) CreateCallHistory(ch *CallHistory) error {
	return d.Db.Create(ch).Error
}

func (d *Database) GetCallHistoriesByPeer(selfPeerID, theirPeerID string) ([]*CallHistory, error) {
	var calls []*CallHistory

	err := d.Db.Where(
		"(from_peer_id = ? AND to_peer_id = ?) OR (from_peer_id = ? AND to_peer_id = ?)",
		selfPeerID, theirPeerID, theirPeerID, selfPeerID,
	).Order("created_at DESC").Find(&calls).Error
	
	if err != nil {
		return nil, err
	}

	return calls, nil
}
