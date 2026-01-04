package main

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aashu10sh/p2pchat/internal/api"
	"github.com/aashu10sh/p2pchat/internal/db"
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

	db := db.GetDatabase()
	eventBus := events.NewEventBus()

	peerManager := peer.NewManager(eventBus)
	profileSvc := service.NewProfileService(db)
	chatSvc := service.NewChatService(db, peerManager, profileSvc, eventBus)

	httpServer := SetupHttpServer(embeddedFiles, profileSvc)

	go func() {
		log.Printf("HTTP server listening on http://localhost:8080")
		log.Printf("Open http://localhost:8080 in your browser")
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", ":5001")
		if err != nil {
			panic(err)
		}
		grpcServer := grpc.NewServer()

		p2pChatServer := ggrpc.NewP2PChatServer(chatSvc, profileSvc, eventBus)
		pb.RegisterP2PChatServiceServer(grpcServer, p2pChatServer)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	// cancel()

}

func SetupHttpServer(
	frontendFiles embed.FS,
	profileSvc *service.ProfileService,
) *http.Server {
	mux := http.NewServeMux()

	handler := api.NewAPIHandler(profileSvc)
	mux.HandleFunc("/api/profile/check", handler.CheckProfile)

	mux.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			panic("not yet implemented")
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

	frontendFS, _ := fs.Sub(frontendFiles, "frontend/build")
	mux.Handle("/", http.FileServer(http.FS(frontendFS)))

	return &http.Server{
		Addr:         ":8080",
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
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
