package service

import (
	"fmt"
	"time"

	"github.com/aashu10sh/p2pchat/internal/db"
	"github.com/aashu10sh/p2pchat/internal/events"
	"github.com/aashu10sh/p2pchat/internal/peer"
	"github.com/aashu10sh/p2pchat/pb"
	"github.com/google/uuid"
)

type ChatService struct {
	db         *db.Database
	peerMgr    *peer.Manager
	profileSvc *ProfileService
	eventBus   *events.EventBus
}

func NewChatService(
	database *db.Database,
	peerMgr *peer.Manager,
	profileSvc *ProfileService,
	eventBus *events.EventBus,
) *ChatService {
	return &ChatService{
		db:         database,
		peerMgr:    peerMgr,
		profileSvc: profileSvc,
		eventBus:   eventBus,
	}
}

// GetOrCreateChat - ensures a chat exists for a peer
func (s *ChatService) GetOrCreateChat(theirPeerID string) (*db.Chat, error) {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return nil, err
	}

	// Try to find existing chat
	chat, err := s.db.GetChatByPeers(profile.PeerId, theirPeerID)
	if err == nil {
		return chat, nil
	}

	// Get peer info for denormalization
	peer, err := s.db.GetPeerByID(theirPeerID)
	if err != nil {
		return nil, fmt.Errorf("peer not found: %w", err)
	}

	// Create new chat
	chat = &db.Chat{
		MyPeerId:      profile.PeerId,
		TheirPeerId:   theirPeerID,
		TheirUserName: peer.UserName,
		LastMessageAt: time.Now(),
		UnreadCount:   0,
	}

	if err := s.db.CreateChat(chat); err != nil {
		return nil, err
	}

	return chat, nil
}

// SendMessage - sends message to peer via gRPC and saves locally
func (s *ChatService) SendMessage(toPeerID, content, msgType string) (string, error) {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return "", err
	}

	// Get or create chat
	chat, err := s.GetOrCreateChat(toPeerID)
	if err != nil {
		return "", err
	}

	// Create message
	messageID := uuid.New().String()
	msg := &pb.Message{
		Id:           messageID,
		FromPeerId:   profile.PeerID,
		FromUserName: profile.UserName,
		ToPeerId:     toPeerID,
		Content:      content,
		Timestamp:    time.Now().Unix(),
		Type:         pb.MessageType(pb.MessageType_value[msgType]),
	}

	// Save to local database first
	dbMsg := &db.Message{
		ChatID:      chat.ID,
		FromPeerID:  msg.FromPeerId,
		ToPeerID:    msg.ToPeerId,
		Content:     msg.Content,
		MessageType: msgType,
		SentAt:      time.Unix(msg.Timestamp, 0),
	}

	if err := s.db.CreateMessage(dbMsg); err != nil {
		return "", fmt.Errorf("failed to save message: %w", err)
	}

	// Update chat last message time
	chat.LastMessageAt = dbMsg.SentAt
	s.db.UpdateChat(chat)

	// Send via gRPC to peer
	if err := s.peerMgr.SendMessage(toPeerID, msg); err != nil {
		// Message saved locally but failed to send
		return messageID, fmt.Errorf("saved locally but failed to send: %w", err)
	}

	// Mark as delivered (optimistic)
	now := time.Now()
	dbMsg.DeliveredAt = &now
	s.db.UpdateMessage(dbMsg)

	// Publish event for local UI update
	s.eventBus.Publish(events.Event{
		Type: "message_sent",
		Data: dbMsg,
	})

	return messageID, nil
}

// SaveIncomingMessage - saves message received from peer via gRPC
func (s *ChatService) SaveIncomingMessage(msg *pb.Message) error {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return err
	}

	// Get or create chat
	chat, err := s.GetOrCreateChat(msg.FromPeerId)
	if err != nil {
		return err
	}

	// Save message
	dbMsg := &db.Message{
		ChatID:      chat.ID,
		FromPeerID:  msg.FromPeerId,
		ToPeerID:    msg.ToPeerId,
		Content:     msg.Content,
		MessageType: msg.Type.String(),
		SentAt:      time.Unix(msg.Timestamp, 0),
	}

	// Mark as delivered immediately
	now := time.Now()
	dbMsg.DeliveredAt = &now

	if err := s.db.CreateMessage(dbMsg); err != nil {
		return err
	}

	// Update chat
	chat.LastMessageAt = dbMsg.SentAt
	chat.UnreadCount++
	s.db.UpdateChat(chat)

	// Set computed field
	dbMsg.IsSentByMe = dbMsg.FromPeerID == profile.PeerID

	return nil
}

// GetMessages - retrieves messages for a chat
func (s *ChatService) GetMessages(peerID string, limit int) ([]*db.Message, error) {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return nil, err
	}

	chat, err := s.db.GetChatByPeers(profile.PeerID, peerID)
	if err != nil {
		return nil, err
	}

	messages, err := s.db.GetMessagesByChatID(chat.ID, limit)
	if err != nil {
		return nil, err
	}

	// Set computed field
	for _, msg := range messages {
		msg.IsSentByMe = msg.FromPeerID == profile.PeerID
	}

	return messages, nil
}

// GetChats - retrieves all chats for current user
func (s *ChatService) GetChats() ([]*db.Chat, error) {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return nil, err
	}

	return s.db.GetChatsByPeerID(profile.PeerID)
}

// MarkChatAsRead - resets unread count
func (s *ChatService) MarkChatAsRead(chatID uint) error {
	chat, err := s.db.GetChatByID(chatID)
	if err != nil {
		return err
	}

	chat.UnreadCount = 0
	return s.db.UpdateChat(chat)
}
