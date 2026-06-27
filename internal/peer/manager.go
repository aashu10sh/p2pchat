package peer

import (
	"context"
	"fmt"
	"sync"

	"github.com/aashu10sh/p2pchat/internal/events"
	"github.com/aashu10sh/p2pchat/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PeerConnection struct {
	PeerID   string
	Username string
	Address  string
	Client   pb.P2PChatServiceClient
	Conn     *grpc.ClientConn
	IsOnline bool
}

type Manager struct {
	peers    map[string]*PeerConnection
	mu       sync.RWMutex
	eventBus *events.EventBus
}

func NewManager(eventBus *events.EventBus) *Manager {
	return &Manager{
		peers:    make(map[string]*PeerConnection),
		eventBus: eventBus,
	}
}

func (m *Manager) ConnectToPeer(peerID, address string, username string) error {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	client := pb.NewP2PChatServiceClient(conn)

	m.mu.Lock()
	m.peers[peerID] = &PeerConnection{
		PeerID:   peerID,
		Username: username,
		Address:  address,
		Client:   client,
		Conn:     conn,
		IsOnline: true,
	}
	m.mu.Unlock()

	m.eventBus.Publish(events.Event{
		Type: "peer_joined",
		Data: map[string]interface{}{
			"peer_id":  peerID,
			"username": username,
		},
	})

	return nil
}

func (m *Manager) SendMessage(peerID string, msg *pb.Message) error {
	m.mu.RLock()
	peer := m.peers[peerID]
	m.mu.RUnlock()

	if peer == nil {
		return fmt.Errorf("peer not connected")
	}

	// Call peer's gRPC server
	_, err := peer.Client.ReceiveMessage(context.Background(), msg)
	return err
}

func (m *Manager) SendVideoCallOffer(peerID string, offer *pb.VideoCallOffer) error {
	m.mu.RLock()
	peer := m.peers[peerID]
	m.mu.RUnlock()

	if peer == nil {
		return fmt.Errorf("peer not connected")
	}

	_, err := peer.Client.ReceiveVideoCallOffer(context.Background(), offer)
	return err
}

func (m *Manager) SendVideoCallAnswer(peerID string, answer *pb.VideoCallAnswer) error {
	m.mu.RLock()
	peer := m.peers[peerID]
	m.mu.RUnlock()

	if peer == nil {
		return fmt.Errorf("peer not connected")
	}

	_, err := peer.Client.ReceiveVideoCallAnswer(context.Background(), answer)
	return err
}

func (m *Manager) SendVideoCallICECandidate(peerID string, ice *pb.VideoCallICECandidate) error {
	m.mu.RLock()
	peer := m.peers[peerID]
	m.mu.RUnlock()

	if peer == nil {
		return fmt.Errorf("peer not connected")
	}

	_, err := peer.Client.ReceiveVideoCallICECandidate(context.Background(), ice)
	return err
}

func (m *Manager) SendVideoCallHangup(peerID string, hangup *pb.VideoCallHangup) error {
	m.mu.RLock()
	peer := m.peers[peerID]
	m.mu.RUnlock()

	if peer == nil {
		return fmt.Errorf("peer not connected")
	}

	_, err := peer.Client.ReceiveVideoCallHangup(context.Background(), hangup)
	return err
}
