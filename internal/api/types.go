package api

import "time"

// Request/Response types for REST API

type ProfileResponse struct {
	ID        int64     `json:"id"`
	UserName  string    `json:"user_name"`
	WiFiName  string    `json:"wifi_name"`
	PeerID    string    `json:"peer_id"`
	CreatedAt time.Time `json:"created_at"`
	ImageUrl  string    `json:"image_url"`
}

type CreateProfileRequest struct {
	UserName string `json:"user_name" binding:"required"`
}

type CheckProfileResponse struct {
	HasProfile bool             `json:"has_profile"`
	Profile    *ProfileResponse `json:"profile,omitempty"`
}

type PeerResponse struct {
	PeerID   string    `json:"peer_id"`
	UserName string    `json:"user_name"`
	Address  string    `json:"address"`
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen"`
	ImagUrl  string    `json:"image_url"`
}

type PeerListResponse struct {
	Peers []PeerResponse `json:"peers"`
}

type MessageResponse struct {
	ID           string    `json:"id"`
	FromPeerID   string    `json:"from_peer_id"`
	FromUserName string    `json:"from_user_name"`
	ToPeerID     string    `json:"to_peer_id"`
	Content      string    `json:"content"`
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`
	IsMine       bool      `json:"is_mine"`
}

type MessageListResponse struct {
	Messages []MessageResponse `json:"messages"`
}

type SendMessageRequest struct {
	ToPeerID    string `json:"to_peer_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
	MessageType string `json:"message_type"`
}

type SendMessageResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id"`
	Error     string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
