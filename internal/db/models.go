package db

import (
	"time"

	"gorm.io/gorm"
)

type Profile struct {
	gorm.Model
	PeerId   string `json:"peer_id" gorm:"uniqueIndex"`
	UserName string `json:"username" gorm:"uniqueIndex:idx_username_wifiname"`
	WifiName string `json:"wifiname" gorm:"uniqueIndex:idx_username_wifiname;index:idx_wifiname;uniqueIndex"`
	ImageUrl string `json:"image_url"`
}

type Peer struct {
	gorm.Model
	PeerId   string    `json:"peer_id" gorm:"uniqueIndex"`
	UserName string    `json:"username"`
	WifiName string    `json:"wifiname" gorm:"index"`
	Address  string    `json:"address"`
	IsOnline bool      `json:"is_online" gorm:"default:true"`
	ImageUrl string    `json:"image_url"`
	LastSeen time.Time `json:"last_seen"`
}

type Chat struct {
	gorm.Model
	MyPeerId    string `gorm:"index" json:"my_peer_id"`
	TheirPeerId string `gorm:"uniqueIndex:idx_chat_peers;index" json:"their_peer_id"`

	TheirUserName string    `json:"their_username"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `gorm:"default:0" json:"unread_count"`
}

type Message struct {
	gorm.Model
	ChatID uint `gorm:"index" json:"chat_id"`
	Chat   Chat `gorm:"foreignKey:ChatID"`

	FromPeerID string `gorm:"index" json:"from_peer_id"`
	ToPeerID   string `gorm:"index" json:"to_peer_id"`

	Content     string `json:"content"`
	MessageType string `gorm:"default:'text'" json:"type"` // text, image, file

	SentAt      time.Time  `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`

	// computed field
	IsSentByMe bool `gorm:"-" json:"is_sent_by_me"`
}
