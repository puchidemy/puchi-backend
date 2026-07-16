package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HandleListSessions returns all active sessions for the authenticated user.
// GET /api/auth/sessions
func (s *AuthService) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	userID, sessionID, _, err := s.extractJWTClaims(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessions, err := s.uc.ListSessions(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}

	type sessionResponse struct {
		ID         string `json:"id"`
		DeviceName string `json:"device_name"`
		DeviceType string `json:"device_type"`
		OS         string `json:"os"`
		IP         string `json:"ip"`
		Location   string `json:"location"`
		LastUsedAt string `json:"last_used_at"`
		CreatedAt  string `json:"created_at"`
		IsCurrent  bool   `json:"is_current"`
	}

	resp := make([]sessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, sessionResponse{
			ID:         s.ID.String(),
			DeviceName: s.DeviceName,
			DeviceType: s.DeviceType,
			OS:         s.OS,
			IP:         s.IPAddress,
			Location:   s.Location,
			LastUsedAt: s.LastUsedAt.Format(time.RFC3339),
			CreatedAt:  s.CreatedAt.Format(time.RFC3339),
			IsCurrent:  s.ID == sessionID,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"sessions": resp})
}

// HandleRevokeAllSessions revokes all sessions except the current one.
// DELETE /api/auth/sessions
func (s *AuthService) HandleRevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID, sessionID, _, err := s.extractJWTClaims(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := s.uc.LogoutAllDevices(r.Context(), userID, sessionID, getClientIPFromReq(r), getUserAgentFromReq(r)); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to revoke sessions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// HandleRevokeSession revokes a specific session by ID.
// DELETE /api/auth/sessions/{id}
func (s *AuthService) HandleRevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, currentSessionID, _, err := s.extractJWTClaims(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract session ID from path: /api/auth/sessions/{id}
	targetID, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/auth/sessions/"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	// Cannot revoke current session through this endpoint (use /api/auth/logout)
	if targetID == currentSessionID {
		writeJSONError(w, http.StatusBadRequest, "cannot revoke current session; use /api/auth/logout instead")
		return
	}

	if err := s.uc.RevokeSession(r.Context(), userID, targetID, getClientIPFromReq(r), getUserAgentFromReq(r)); err != nil {
		if err.Error() == "session not found" {
			writeJSONError(w, http.StatusNotFound, "session not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to revoke session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
