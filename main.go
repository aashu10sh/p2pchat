package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aashu10sh/p2pchat/internal/api"
	"github.com/aashu10sh/p2pchat/internal/db"
	"github.com/aashu10sh/p2pchat/internal/discovery"
	"github.com/aashu10sh/p2pchat/internal/events"
	ggrpc "github.com/aashu10sh/p2pchat/internal/grpc"
	"github.com/aashu10sh/p2pchat/internal/peer"
	"github.com/aashu10sh/p2pchat/internal/service"
	"github.com/aashu10sh/p2pchat/pb"
	"google.golang.org/grpc"
)

//go:embed frontend/build/*
var embeddedFiles embed.FS

func main() {
	grpcPort := 5001
	database := db.GetDatabase()
	eventBus := events.NewEventBus()

	peerManager := peer.NewManager(eventBus)

	profileSvc := service.NewProfileService(database)
	chatSvc := service.NewChatService(database, peerManager, profileSvc, eventBus)
	peerSvc := service.NewPeerService(database)
	fileSvc := service.NewFileService(database, peerManager, profileSvc, eventBus)

	// Get current profile for mDNS announcement
	profile, err := profileSvc.GetCurrentProfile()

	if err != nil {
		log.Printf("Warning: No profile found. mDNS will not start. Create a profile first: %v", err)
	}

	httpServer := SetupHttpServer(embeddedFiles, profileSvc, peerSvc, database, chatSvc, fileSvc, eventBus)

	go func() {
		log.Printf("HTTP server listening on http://localhost:8000")
		log.Printf("Open http://localhost:8000 in your browser")
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			panic(err)
		}
		grpcServer := grpc.NewServer()

		p2pChatServer := ggrpc.NewP2PChatServer(chatSvc, profileSvc, fileSvc, eventBus)
		pb.RegisterP2PChatServiceServer(grpcServer, p2pChatServer)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}

	}()

	// Start mDNS service if profile exists
	var mdnsService *discovery.MDNSService
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if profile != nil {
		mdnsService = discovery.NewMDNSService(
			peerManager,
			database,
			eventBus,
			profile.PeerId,
			profile.UserName,
			grpcPort,
		)

		if err := mdnsService.StartServer(); err != nil {
			log.Fatalf("Failed to start mDNS server: %v", err)
		}

		go mdnsService.StartDiscovery(ctx)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	cancel()

	if mdnsService != nil {
		if err := mdnsService.Shutdown(); err != nil {
			log.Printf("Error shutting down mDNS: %v", err)
		}
	}
}

func SetupHttpServer(
	frontendFiles embed.FS,
	profileSvc *service.ProfileService,
	peerSvc *service.PeerService,
	database *db.Database,
	chatSvc *service.ChatService,
	fileSvc *service.FileService,
	eventBus *events.EventBus,
) *http.Server {
	mux := http.NewServeMux()

	handler := api.NewAPIHandler(profileSvc, database, chatSvc, fileSvc, peerSvc, eventBus)

	mux.HandleFunc("/api/profile/check", handler.CheckProfile)

	mux.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetPeerById(w, r)
		case http.MethodPost:
			handler.CreateProfile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/api/current-wifi", handler.GetCurrentWifiName)
	mux.HandleFunc("/api/peers", handler.GetPeers)
	mux.HandleFunc("/api/events", handler.StreamEvents)
	mux.HandleFunc("/api/chats", handler.GetChats)
	mux.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetMessages(w, r)
		case http.MethodPost:
			handler.SendMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Video call signaling routes
	mux.HandleFunc("/api/call/offer", handler.HandleVideoCallOffer)
	mux.HandleFunc("/api/call/answer", handler.HandleVideoCallAnswer)
	mux.HandleFunc("/api/call/ice-candidate", handler.HandleVideoCallICECandidate)
	mux.HandleFunc("/api/call/hangup", handler.HandleVideoCallHangup)
	mux.HandleFunc("/api/call/history", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.HandleGetCallHistories(w, r)
		case http.MethodPost:
			handler.HandleSaveCallHistory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// File transfer routes
	mux.HandleFunc("/api/files/send", handler.HandleSendFile)
	mux.HandleFunc("/api/files", handler.HandleGetFileTransfers)

	frontendFS, _ := fs.Sub(frontendFiles, "frontend/build")
	fileServer := http.FileServer(http.FS(frontendFS))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "."
		}

		if _, err := fs.Stat(frontendFS, path); err != nil {
			// File doesn't exist, fallback to index.html (SPA routing)
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	}))

	return &http.Server{
		Addr:              ":8000",
		Handler:           corsMiddleware(mux),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      0, // No timeout for SSE connections
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
