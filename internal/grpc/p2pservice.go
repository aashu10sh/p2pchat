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

	err := s.chatSvc.SaveIncomingMessage(msg)
	if err != nil {
		return &pb.MessageAck{Success: false, Error: err.Error()}, nil
	}

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

// Video call signaling handlers — each publishes the received signaling data
// to the event bus so it reaches the local browser via SSE.

func (s *P2PChatServer) ReceiveVideoCallOffer(ctx context.Context, offer *pb.VideoCallOffer) (*pb.VideoCallAck, error) {
	s.eventBus.Publish(events.Event{
		Type: "video_call_offer",
		Data: map[string]string{
			"from_peer_id": offer.FromPeerId,
			"sdp":          offer.Sdp,
		},
	})
	return &pb.VideoCallAck{Success: true}, nil
}

func (s *P2PChatServer) ReceiveVideoCallAnswer(ctx context.Context, answer *pb.VideoCallAnswer) (*pb.VideoCallAck, error) {
	s.eventBus.Publish(events.Event{
		Type: "video_call_answer",
		Data: map[string]string{
			"from_peer_id": answer.FromPeerId,
			"sdp":          answer.Sdp,
		},
	})
	return &pb.VideoCallAck{Success: true}, nil
}

func (s *P2PChatServer) ReceiveVideoCallICECandidate(ctx context.Context, ice *pb.VideoCallICECandidate) (*pb.VideoCallAck, error) {
	s.eventBus.Publish(events.Event{
		Type: "video_call_ice_candidate",
		Data: map[string]interface{}{
			"from_peer_id":    ice.FromPeerId,
			"candidate":       ice.Candidate,
			"sdp_mid":         ice.SdpMid,
			"sdp_mline_index": ice.SdpMlineIndex,
		},
	})
	return &pb.VideoCallAck{Success: true}, nil
}

func (s *P2PChatServer) ReceiveVideoCallHangup(ctx context.Context, hangup *pb.VideoCallHangup) (*pb.VideoCallAck, error) {
	s.eventBus.Publish(events.Event{
		Type: "video_call_hangup",
		Data: map[string]string{
			"from_peer_id": hangup.FromPeerId,
			"reason":       hangup.Reason,
		},
	})
	return &pb.VideoCallAck{Success: true}, nil
}
