package discovery

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/aashu10sh/p2pchat/internal/db"
	"github.com/aashu10sh/p2pchat/internal/events"
	"github.com/aashu10sh/p2pchat/internal/peer"
	"github.com/hashicorp/mdns"
)

const (
	ServiceName = "_p2pchat._tcp"
	Domain      = "local."
)

type MDNSService struct {
	server      *mdns.Server
	peerManager *peer.Manager
	database    *db.Database
	eventBus    *events.EventBus
	myPeerID    string
	myUsername  string
	grpcPort    int
}

func NewMDNSService(
	peerManager *peer.Manager,
	database *db.Database,
	eventBus *events.EventBus,
	myPeerID string,
	myUsername string,
	grpcPort int,
) *MDNSService {
	return &MDNSService{
		peerManager: peerManager,
		database:    database,
		eventBus:    eventBus,
		myPeerID:    myPeerID,
		myUsername:  myUsername,
		grpcPort:    grpcPort,
	}
}

// StartServer announces this peer on the network
func (m *MDNSService) StartServer() error {
	// Get local IP address
	host, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("failed to get local IP: %w", err)
	}

	// Create service info
	info := []string{
		"peer_id=" + m.myPeerID,
		"username=" + m.myUsername,
	}

	service, err := mdns.NewMDNSService(
		m.myPeerID,                  // instance name
		ServiceName,                 // service type
		Domain,                      // domain
		"",                          // hostname (empty = use local hostname)
		m.grpcPort,                  // port
		[]net.IP{net.ParseIP(host)}, // IPs
		info,                        // TXT records
	)
	if err != nil {
		return fmt.Errorf("failed to create mDNS service: %w", err)
	}

	// Create and start the mDNS server
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return fmt.Errorf("failed to create mDNS server: %w", err)
	}

	m.server = server
	log.Printf("mDNS server started, announcing as %s on %s:%d", m.myUsername, host, m.grpcPort)

	return nil
}

// StartDiscovery continuously discovers peers on the network
func (m *MDNSService) StartDiscovery(ctx context.Context) {
	log.Println("Starting mDNS peer discovery...")

	discoveryTicker := time.NewTicker(10 * time.Second)
	defer discoveryTicker.Stop()

	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	// Do initial discovery
	m.discoverPeers()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping mDNS discovery...")
			return
		case <-discoveryTicker.C:
			m.discoverPeers()
		case <-statusTicker.C:
			m.updatePeerStatus()
		}
	}
}

// updatePeerStatus marks peers as offline if not seen recently
func (m *MDNSService) updatePeerStatus() {
	peers, err := m.database.GetAllPeers()
	if err != nil {
		log.Printf("Failed to get peers for status update: %v", err)
		return
	}

	threshold := time.Now().Add(-2 * time.Minute)
	for _, peer := range peers {
		if peer.LastSeen.Before(threshold) && peer.IsOnline {
			peer.IsOnline = false
			if err := m.database.UpdatePeer(peer); err != nil {
				log.Printf("Failed to update peer status: %v", err)
			} else {
				log.Printf("Peer %s marked as offline", peer.UserName)
			}
		}
	}
}

func (m *MDNSService) discoverPeers() {
	entriesCh := make(chan *mdns.ServiceEntry, 10)

	go func() {
		for entry := range entriesCh {
			m.handleDiscoveredPeer(entry)
		}
	}()

	// Lookup peers
	params := &mdns.QueryParam{
		Service:             ServiceName,
		Domain:              Domain,
		Timeout:             5 * time.Second,
		Entries:             entriesCh,
		WantUnicastResponse: false,
	}

	if err := mdns.Query(params); err != nil {
		log.Printf("mDNS query error: %v", err)
	}

	close(entriesCh)
}

func (m *MDNSService) handleDiscoveredPeer(entry *mdns.ServiceEntry) {
	// Parse TXT records
	peerID := ""
	username := ""

	for _, txt := range entry.InfoFields {
		if len(txt) > 8 && txt[:8] == "peer_id=" {
			peerID = txt[8:]
		}
		if len(txt) > 9 && txt[:9] == "username=" {
			username = txt[9:]
		}
	}

	// Skip if it's ourselves
	if peerID == m.myPeerID {
		return
	}

	if peerID == "" || username == "" {
		log.Printf("Discovered peer with incomplete info: %v", entry)
		return
	}

	// Get the address
	address := ""
	if entry.AddrV4 != nil {
		address = fmt.Sprintf("%s:%d", entry.AddrV4.String(), entry.Port)
	} else if entry.AddrV6 != nil {
		address = fmt.Sprintf("[%s]:%d", entry.AddrV6.String(), entry.Port)
	} else {
		log.Printf("No valid address for peer %s", peerID)
		return
	}

	log.Printf("Discovered peer: %s (%s) at %s", username, peerID, address)

	// Check if peer already exists in database
	existingPeer, err := m.database.GetPeerByID(peerID)
	if err != nil {
		// Peer doesn't exist, create new
		newPeer := &db.Peer{
			PeerId:   peerID,
			UserName: username,
			Address:  address,
			IsOnline: true,
			LastSeen: time.Now(),
		}

		if err := m.database.CreatePeer(newPeer); err != nil {
			log.Printf("Failed to create peer: %v", err)
			return
		}

		log.Printf("New peer added to database: %s", username)
	} else {
		// Update existing peer
		existingPeer.Address = address
		existingPeer.IsOnline = true
		existingPeer.LastSeen = time.Now()
		existingPeer.UserName = username

		if err := m.database.UpdatePeer(existingPeer); err != nil {
			log.Printf("Failed to update peer: %v", err)
			return
		}
	}

	// Connect to peer via gRPC
	if err := m.peerManager.ConnectToPeer(peerID, address, username); err != nil {
		log.Printf("Failed to connect to peer %s: %v", username, err)
		return
	}

	// Publish event
	m.eventBus.Publish(events.Event{
		Type: "peer_discovered",
		Data: map[string]interface{}{
			"peer_id":  peerID,
			"username": username,
			"address":  address,
		},
	})
}

// Shutdown stops the mDNS server
func (m *MDNSService) Shutdown() error {
	if m.server != nil {
		if err := m.server.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown mDNS server: %w", err)
		}
		log.Println("mDNS server stopped")
	}
	return nil
}

// getLocalIP returns the non-loopback local IP of the host
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no local IP address found")
}

// ParseGRPCPort extracts port number from gRPC address
func ParseGRPCPort(addr string) (int, error) {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(portStr)
}
