package api

import (
	"encoding/json"
	"net/http"

	"github.com/aashu10sh/p2pchat/internal/service"
	"github.com/aashu10sh/p2pchat/internal/utils"
)

type APIHandler struct {
	profileSvc *service.ProfileService
}

func NewAPIHandler(
	profileSvc *service.ProfileService,
) *APIHandler {
	return &APIHandler{
		profileSvc: profileSvc,
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
