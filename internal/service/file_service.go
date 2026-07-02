package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aashu10sh/p2pchat/internal/db"
	"github.com/aashu10sh/p2pchat/internal/events"
	"github.com/aashu10sh/p2pchat/internal/peer"
	"github.com/aashu10sh/p2pchat/pb"
)

type FileService struct {
	db         *db.Database
	peerMgr    *peer.Manager
	profileSvc *ProfileService
	eventBus   *events.EventBus
	downloadDir string
}

func NewFileService(
	database *db.Database,
	peerMgr *peer.Manager,
	profileSvc *ProfileService,
	eventBus *events.EventBus,
) *FileService {
	
	// Create default download directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	downloadDir := filepath.Join(homeDir, "Downloads", "p2pchat")
	os.MkdirAll(downloadDir, 0755)

	return &FileService{
		db:         database,
		peerMgr:    peerMgr,
		profileSvc: profileSvc,
		eventBus:   eventBus,
		downloadDir: downloadDir,
	}
}

// SaveIncomingFile is called by the gRPC ReceiveFile stream to write chunks to disk
func (s *FileService) SaveIncomingFile(stream pb.P2PChatService_ReceiveFileServer) error {
	var file *os.File
	var filePath string
	var currentFileName string
	var totalSize int64
	var fromPeerID, toPeerID string

	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error receiving file chunk: %w", err)
		}

		if file == nil {
			// First chunk, initialize file
			fromPeerID = chunk.FromPeerId
			toPeerID = chunk.ToPeerId
			currentFileName = chunk.FileName
			totalSize = chunk.FileSize

			// To avoid conflicts, append timestamp
			fileNameWithTimestamp := fmt.Sprintf("%d_%s", time.Now().Unix(), currentFileName)
			filePath = filepath.Join(s.downloadDir, fileNameWithTimestamp)

			file, err = os.Create(filePath)
			if err != nil {
				return fmt.Errorf("could not create file on disk: %w", err)
			}
		}

		_, err = file.Write(chunk.Data)
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}

	if file != nil {
		file.Close() // Ensure file is closed before DB save
		file = nil
	}

	// Save to DB
	ft := &db.FileTransfer{
		FromPeerID: fromPeerID,
		ToPeerID:   toPeerID,
		FileName:   currentFileName,
		FileSize:   totalSize,
		FilePath:   filePath,
		Direction:  "received",
	}

	if err := s.db.CreateFileTransfer(ft); err != nil {
		return fmt.Errorf("failed to save file transfer record: %w", err)
	}

	// Notify UI
	s.eventBus.Publish(events.Event{
		Type: "file_received",
		Data: ft,
	})

	return stream.SendAndClose(&pb.FileTransferAck{
		Success:  true,
		FilePath: filePath,
	})
}

// SendFile reads a local file and streams it to the target peer
func (s *FileService) SendFile(toPeerID string, absolutePath string) error {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return err
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return fmt.Errorf("could not open file to send: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("could not get file info: %w", err)
	}

	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	// 512KB chunk size
	chunkSize := 512 * 1024
	buf := make([]byte, chunkSize)

	// Get a stream from peer manager
	stream, err := s.peerMgr.GetSendFileStream(toPeerID)
	if err != nil {
		return fmt.Errorf("failed to get stream to peer: %w", err)
	}

	var chunkIndex uint32 = 0
	totalChunks := uint32((fileSize + int64(chunkSize) - 1) / int64(chunkSize))

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading file: %w", err)
		}
		if n == 0 {
			break
		}

		chunk := &pb.FileChunk{
			FromPeerId: profile.PeerId,
			ToPeerId:   toPeerID,
			FileName:   fileName,
			FileSize:   fileSize,
			Data:       buf[:n],
			ChunkIndex: chunkIndex,
			TotalChunks: totalChunks,
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
		chunkIndex++
	}

	ack, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to receive ack: %w", err)
	}

	if !ack.Success {
		return fmt.Errorf("peer rejected file transfer: %s", ack.Error)
	}

	// Save to DB
	ft := &db.FileTransfer{
		FromPeerID: profile.PeerId,
		ToPeerID:   toPeerID,
		FileName:   fileName,
		FileSize:   fileSize,
		FilePath:   absolutePath,
		Direction:  "sent",
	}

	if err := s.db.CreateFileTransfer(ft); err != nil {
		return fmt.Errorf("failed to save file transfer record: %w", err)
	}
	
	// Notify UI
	s.eventBus.Publish(events.Event{
		Type: "file_sent",
		Data: ft,
	})

	return nil
}

// GetFileTransfers returns the file transfers for a given peer
func (s *FileService) GetFileTransfers(peerID string) ([]*db.FileTransfer, error) {
	profile, err := s.profileSvc.GetCurrentProfile()
	if err != nil {
		return nil, err
	}

	return s.db.GetFileTransfersByPeer(profile.PeerId, peerID)
}
