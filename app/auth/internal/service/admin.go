package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
)

// AdminService provides admin HTTP handlers for RBAC management.
type AdminService struct {
	rbacUC  *biz.RBACUsecase
	authUC  *biz.AuthUsecase
	tokenUC *biz.TokenUsecase
}

// NewAdminService creates a new AdminService.
func NewAdminService(rbacUC *biz.RBACUsecase, authUC *biz.AuthUsecase, tokenUC *biz.TokenUsecase) *AdminService {
	return &AdminService{
		rbacUC:  rbacUC,
		authUC:  authUC,
		tokenUC: tokenUC,
	}
}

// HandleListRoles returns all roles.
func (s *AdminService) HandleListRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	roles, err := s.rbacUC.ListRoles(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list roles")
		return
	}

	type roleResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		IsSystem    bool   `json:"is_system"`
	}

	resp := make([]roleResponse, len(roles))
	for i, role := range roles {
		resp[i] = roleResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleListPermissions returns all permissions.
func (s *AdminService) HandleListPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	perms, err := s.rbacUC.ListPermissions(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list permissions")
		return
	}

	type permResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	resp := make([]permResponse, len(perms))
	for i, p := range perms {
		resp[i] = permResponse{
			ID:          p.ID.String(),
			Name:        p.Name,
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleAssignRole assigns a role to a user.
// POST /admin/users/{id}/roles
func (s *AdminService) HandleAssignRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := extractUserIDFromPath(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Extract the admin who is granting the role from the JWT
	grantedByID := uuid.Nil
	if claims, err := s.tokenUC.VerifyAccessToken(extractBearerToken(r)); err == nil {
		grantedByID = claims.UserID
	}

	var req struct {
		RoleName string `json:"role_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RoleName == "" {
		writeJSONError(w, http.StatusBadRequest, "role_name is required")
		return
	}

	if err := s.rbacUC.AssignRole(r.Context(), userID, req.RoleName, grantedByID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to assign role: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "role assigned"})
}

// HandleRemoveRole removes a role from a user.
// DELETE /admin/users/{id}/roles
func (s *AdminService) HandleRemoveRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := extractUserIDFromPath(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req struct {
		RoleName string `json:"role_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RoleName == "" {
		writeJSONError(w, http.StatusBadRequest, "role_name is required")
		return
	}

	if err := s.rbacUC.RemoveRole(r.Context(), userID, req.RoleName); err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to remove role: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "role removed"})
}

// HandleGetUserRoles returns all roles assigned to a user.
// GET /admin/users/{id}/roles
func (s *AdminService) HandleGetUserRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := extractUserIDFromPath(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	roles, err := s.rbacUC.GetUserRoles(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to get user roles")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"roles": roles})
}

// HandleGetUserPermissions returns all permissions for a user.
// GET /admin/users/{id}/permissions
func (s *AdminService) HandleGetUserPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := extractUserIDFromPath(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	perms, err := s.rbacUC.GetUserPermissions(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to get user permissions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"permissions": perms})
}

// extractBearerToken extracts the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	return strings.TrimPrefix(auth, "Bearer ")
}

// extractUserIDFromPath extracts a UUID from a path like /admin/users/{id}/roles.
func extractUserIDFromPath(path string) (uuid.UUID, error) {
	// Path format: /admin/users/{id}/roles or /admin/users/{id}/permissions
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return uuid.Nil, fmt.Errorf("invalid path")
	}
	return uuid.Parse(parts[2])
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeJSONError writes a JSON error response.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
