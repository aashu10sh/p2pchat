package service

import (
	"errors"
	"fmt"

	"github.com/aashu10sh/p2pchat/internal/db"
	"github.com/aashu10sh/p2pchat/internal/utils"
	"github.com/google/uuid"
)

type ProfileService struct {
	db *db.Database
}

func NewProfileService(db *db.Database) *ProfileService {
	return &ProfileService{
		db: db,
	}
}

func (s *ProfileService) GetCurrentProfile() (*db.Profile, error) {
	var profile *db.Profile

	wifiName, _ := utils.GetCurrentSSID()

	result := s.db.Db.Find(&profile).Where(&db.Profile{
		WifiName: wifiName,
	})

	if result.Error != nil {
		return nil, result.Error
	}

	if profile == nil || profile.ID == 0 {
		return nil, errors.New("no profile for the current wifi")
	}

	return profile, nil
}

func (s *ProfileService) CreateProfile(wifiName, userName string) (*db.Profile, error) {

	var profile = db.Profile{
		PeerId:   uuid.NewString(),
		UserName: userName,
		WifiName: wifiName,
		ImageUrl: fmt.Sprintf("https://robohash.org/%s", userName),
	}

	result := s.db.Db.Create(&profile)

	if result.Error != nil {
		return nil, result.Error
	}

	return &profile, nil
}
