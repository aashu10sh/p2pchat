package service

import "github.com/aashu10sh/p2pchat/internal/db"

type PeerService struct {
	db *db.Database
}

func NewPeerService(db *db.Database) *PeerService {
	return &PeerService{
		db: db,
	}
}

func (s *PeerService) GetPeerByID(peerId string) (*db.Peer, error) {
	peer, err := s.db.GetPeerByID(peerId)
	return peer, err
}
