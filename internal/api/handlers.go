package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aashu10sh/p2pchat/internal/db"
	"github.com/aashu10sh/p2pchat/internal/events"
	"github.com/aashu10sh/p2pchat/internal/service"
	"github.com/aashu10sh/p2pchat/internal/utils"
)

type APIHandler struct {
	profileSvc *service.ProfileService
	peerSvc    *service.PeerService
	database   *db.Database
	chatSvc    *service.ChatService
	eventBus   *events.EventBus
}

func NewAPIHandler(
	profileSvc *service.ProfileService,
	database *db.Database,
	chatSvc *service.ChatService,
	peerSvc *service.PeerService,
	eventBus *events.EventBus,
) *APIHandler {
	return &APIHandler{
		profileSvc: profileSvc,
		database:   database,
		chatSvc:    chatSvc,
		peerSvc:    peerSvc,
		eventBus:   eventBus,
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

func (h *APIHandler) GetPeerById(w http.ResponseWriter, r *http.Request) {
	peerID := r.URL.Query().Get("peer_id")
	log.Println("peerid is" + peerID)
	if peerID == "" {
		respondError(w, http.StatusNotFound, "peer not found")
		return
	}
	peer, err := h.peerSvc.GetPeerByID(peerID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, 200, peer)
}

func (h *APIHandler) HandleSignaling(w http.ResponseWriter, r *http.Request) {
	var req SignalingMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

}

func (h *APIHandler) CheckProfile(w http.ResponseWriter, r *http.Request) {

	profile, err := h.profileSvc.GetCurrentProfile()

	if err != nil {
		respondError(w, http.StatusNotFound, "Profile not found")
		return
	}

	respondJSON(w, http.StatusOK, ProfileResponse{
		ID:        int64(profile.ID),
		PeerID:    profile.PeerId,
		UserName:  profile.UserName,
		WiFiName:  profile.WifiName,
		ImageUrl:  profile.ImageUrl,
		CreatedAt: profile.CreatedAt,
	})

}

func (h *APIHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req CreateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserName == "" {
		respondError(w, http.StatusBadRequest, "Username is required")
		return
	}

	wifiName, err := utils.GetCurrentSSID()

	if err != nil {
		panic("error, could not get current wifi")
	}

	profile, err := h.profileSvc.CreateProfile(
		wifiName,
		req.UserName,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, ProfileResponse{
		ID:        int64(profile.ID),
		UserName:  profile.UserName,
		WiFiName:  profile.WifiName,
		PeerID:    profile.PeerId,
		CreatedAt: profile.CreatedAt,
	})
}

func (h *APIHandler) GetCurrentWifiName(w http.ResponseWriter, r *http.Request) {
	wifiName, err := utils.GetCurrentSSID()

	if err != nil {
		respondError(w, http.StatusNotFound, "could not find wifi")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"wifiName": wifiName,
	})

}

func (h *APIHandler) GetPeers(w http.ResponseWriter, r *http.Request) {
	peers, err := h.database.GetAllPeers()

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch peers")
		return
	}

	respondJSON(w, http.StatusOK, peers)
}

func (h *APIHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE - must be set before any writes
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Write status code explicitly
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Send initial connection event
	if err := respondEvent(w, flusher, "connected", map[string]string{"status": "connected"}); err != nil {
		return
	}

	// Subscribe to event bus
	eventChan := h.eventBus.Subscribe()
	defer h.eventBus.Unsubscribe(eventChan)

	// Keep-alive ticker to prevent timeout
	keepAlive := time.NewTicker(30 * time.Second)
	defer keepAlive.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			// Forward event bus events to SSE immediately
			switch event.Type {
			case "message_received", "message_sent":
				if err := respondEvent(w, flusher, event.Type, event.Data); err != nil {
					return
				}
			case "peer_discovered", "peer_joined":
				// Send full peer list update on peer changes
				peers, err := h.database.GetAllPeers()
				if err != nil {
					continue
				}
				if err := respondEvent(w, flusher, "peers_update", peers); err != nil {
					return
				}
			}
		case <-keepAlive.C:
			// Send keep-alive comment to prevent connection timeout
			if _, err := fmt.Fprintf(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *APIHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	chats, err := h.chatSvc.GetChats()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch chats")
		return
	}

	respondJSON(w, http.StatusOK, chats)
}

func (h *APIHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	peerID := r.URL.Query().Get("peer_id")
	if peerID == "" {
		respondError(w, http.StatusBadRequest, "peer_id is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	messages, err := h.chatSvc.GetMessages(peerID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}

	respondJSON(w, http.StatusOK, messages)
}

func (h *APIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ToPeerID == "" || req.Content == "" {
		respondError(w, http.StatusBadRequest, "to_peer_id and content are required")
		return
	}

	if req.MessageType == "" {
		req.MessageType = "TEXT"
	}

	messageID, err := h.chatSvc.SendMessage(req.ToPeerID, req.Content, req.MessageType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message_id": messageID,
		"status":     "sent",
	})
}

func respondEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Write event type
	if _, err := fmt.Fprintf(w, "event: %s\n", eventType); err != nil {
		return err
	}

	// Write data
	if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
		return err
	}

	// Flush immediately
	flusher.Flush()
	return nil
}
