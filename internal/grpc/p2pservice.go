package grpc

import (
	"context"
	"log"
	"time"

	"github.com/aashu10sh/p2pchat/internal/events"
	"github.com/aashu10sh/p2pchat/internal/service"
	"github.com/aashu10sh/p2pchat/pb"
)

type P2PChatServer struct {
	pb.UnimplementedP2PChatServiceServer
	profileSvc *service.ProfileService
	chatSvc    *service.ChatService
	eventBus   *events.EventBus
}

func NewP2PChatServer(
	chatSvc *service.ChatService,
	profileSvc *service.ProfileService,
	eventBus *events.EventBus,
) *P2PChatServer {
	return &P2PChatServer{
		profileSvc: profileSvc,
		chatSvc:    chatSvc,
		eventBus:   eventBus,
	}
}

func (s *P2PChatServer) ReceiveMessage(ctx context.Context, msg *pb.Message) (*pb.MessageAck, error) {
	//Save to database
	err := s.chatSvc.SaveIncomingMessage(msg)
	if err != nil {
		return &pb.MessageAck{Success: false, Error: err.Error()}, nil
	}

	//Broadcast event to all HTTP SSE listeners
	s.eventBus.Publish(events.Event{
		Type: "message_received",
		Data: msg,
	})

	return &pb.MessageAck{
		Success:   true,
		MessageId: msg.Id,
	}, nil
}

func (s *P2PChatServer) GetPeerInfo(ctx context.Context, _ *pb.Empty) (*pb.PeerInfo, error) {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return nil, err
	}

	return &pb.PeerInfo{
		PeerId:   profile.PeerId,
		UserName: profile.UserName,
		WifiName: profile.WifiName,
	}, nil
}

func (s *P2PChatServer) Ping(ctx context.Context, pingRequest *pb.PingRequest) (*pb.PingResponse, error) {
	self, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		log.Fatal(err)
	}

	return &pb.PingResponse{
		PeerId:    self.PeerId,
		UserName:  self.UserName,
		Timestamp: time.Now().Unix(),
	}, nil
}
