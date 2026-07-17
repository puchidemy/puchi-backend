package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	pb "github.com/puchidemy/puchi-backend/app/core/api/profile/v1"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
	"github.com/puchidemy/puchi-backend/pkg/apierr"
	auth "github.com/puchidemy/puchi-backend/pkg/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProfileService implements both ProfileServiceServer (gRPC) and ProfileServiceHTTPServer (HTTP).
type ProfileService struct {
	pb.UnimplementedProfileServiceServer
	uc            *biz.ProfileUsecase
	achievementUC *biz.AchievementUsecase
	statsSvc      *StatsService
	cdnBaseURL    string
}

// NewProfileService creates a new ProfileService.
func NewProfileService(uc *biz.ProfileUsecase, achievementUC *biz.AchievementUsecase, statsSvc *StatsService, cdnBaseURL string) *ProfileService {
	return &ProfileService{
		uc:            uc,
		achievementUC: achievementUC,
		statsSvc:      statsSvc,
		cdnBaseURL:    strings.TrimRight(cdnBaseURL, "/"),
	}
}

// GetStats returns the authenticated user's gamification stats.
func (s *ProfileService) GetStats(ctx context.Context, req *emptypb.Empty) (*pb.Stats, error) {
	return s.statsSvc.GetStats(ctx, req)
}

// ListDailyActivity returns daily activity for the authenticated user.
func (s *ProfileService) ListDailyActivity(ctx context.Context, req *pb.ListDailyActivityRequest) (*pb.DailyActivityList, error) {
	return s.statsSvc.ListDailyActivity(ctx, req)
}

// ListWeeklyXP returns weekly XP history for the authenticated user.
func (s *ProfileService) ListWeeklyXP(ctx context.Context, req *pb.ListWeeklyXPRequest) (*pb.WeeklyXPList, error) {
	return s.statsSvc.ListWeeklyXP(ctx, req)
}

// ListAchievements returns the authenticated user's achievements with progress.
func (s *ProfileService) ListAchievements(ctx context.Context, _ *emptypb.Empty) (*pb.AchievementList, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	items, err := s.achievementUC.ListAchievements(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list achievements")
	}

	return achievementItemsToProto(items), nil
}

// GetProfile returns the authenticated user's profile.
// If the user doesn't exist in core.users yet, it is auto-created (lazy sync from auth).
func (s *ProfileService) GetProfile(ctx context.Context, _ *emptypb.Empty) (*pb.User, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	email, _ := auth.EmailFromContext(ctx)
	user, err := s.uc.GetOrCreateProfile(ctx, userID, email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get or create profile")
	}

	return s.userToProto(user), nil
}

// UpdateProfile updates the authenticated user's profile.
func (s *ProfileService) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.User, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	user, err := s.uc.UpdateProfile(ctx, userID, biz.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Bio:       req.Bio,
		AgeRange:  req.AgeRange,
	})
	if err != nil {
		if err == biz.ErrUsernameTaken {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.userToProto(user), nil
}

// UpdateAvatar stores the media object key and returns the profile with CDN avatar_url.
func (s *ProfileService) UpdateAvatar(ctx context.Context, req *pb.UpdateAvatarRequest) (*pb.User, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	user, err := s.uc.UpdateAvatar(ctx, userID, req.AvatarKey)
	if err != nil {
		if err == biz.ErrInvalidAvatarKey {
			return nil, status.Error(codes.InvalidArgument, "avatar_key must start with avatar/")
		}
		if err == biz.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.userToProto(user), nil
}

// GetProfileByUsername returns a user's public profile by username.
func (s *ProfileService) GetProfileByUsername(ctx context.Context, req *pb.GetProfileByUsernameRequest) (*pb.User, error) {
	user, err := s.uc.GetProfileByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// If user is logged in and it's their own profile, show email
	currentUserID, isLoggedIn := auth.UserIDFromContext(ctx)
	userProto := s.userToProto(user)
	if !isLoggedIn || currentUserID != user.ID {
		userProto.Email = ""
	}
	return userProto, nil
}

// CompleteOnboarding completes onboarding and saves profile + answers.
// If the user doesn't exist in core.users yet, it is auto-created first (lazy sync).
func (s *ProfileService) CompleteOnboarding(ctx context.Context, req *pb.CompleteOnboardingRequest) (*pb.User, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	if req.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "first_name is required")
	}
	if req.LastName == "" {
		return nil, status.Error(codes.InvalidArgument, "last_name is required")
	}
	validAgeRanges := map[string]bool{
		"13-17": true,
		"18-24": true,
		"25-34": true,
		"35-44": true,
		"45-54": true,
		"55+":   true,
	}
	if !validAgeRanges[req.AgeRange] {
		return nil, status.Error(codes.InvalidArgument, "invalid age_range value")
	}

	// Ensure user exists in core.users (lazy creation if first request)
	email, _ := auth.EmailFromContext(ctx)
	if _, err := s.uc.GetOrCreateProfile(ctx, userID, email); err != nil {
		return nil, status.Error(codes.Internal, "failed to get or create profile")
	}

	user, err := s.uc.CompleteOnboarding(ctx, userID, biz.OnboardingInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AgeRange:  req.AgeRange,
		Username:  req.Username,
		HowHeard:  req.HowHeard,
		WhyLearn:  req.WhyLearn,
		Level:     req.Level,
	})
	if err != nil {
		if err == biz.ErrUsernameTaken {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.userToProto(user), nil
}

// GetLinkedAccounts returns linked third-party accounts.
// TODO: Implement via Zitadel user info API when needed.
func (s *ProfileService) GetLinkedAccounts(ctx context.Context, _ *emptypb.Empty) (*pb.LinkedAccountsResponse, error) {
	_, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Linked accounts are now managed by Zitadel.
	// This endpoint returns an empty list until Zitadel user info API integration is added.
	return &pb.LinkedAccountsResponse{Accounts: []*pb.LinkedAccount{}}, nil
}

// HandleMergeGuest handles POST /v1/profile/merge-guest
func (s *ProfileService) HandleMergeGuest(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		apierr.Unauthorized(w, "UNAUTHORIZED")
		return
	}

	var req struct {
		LessonsCompleted   int      `json:"lessons_completed"`
		TotalCorrect       int      `json:"total_correct"`
		TotalXp            int      `json:"total_xp"`
		CompletedLessonIDs []string `json:"completed_lesson_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
		return
	}

	if err := s.uc.MergeGuestProgress(r.Context(), userID, biz.MergeGuestInput{
		LessonsCompleted:   req.LessonsCompleted,
		TotalCorrect:       req.TotalCorrect,
		TotalXp:            req.TotalXp,
		CompletedLessonIDs: req.CompletedLessonIDs,
	}); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "merge failed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "guest progress merged"})
}

// userToProto converts a gen.CoreUser to a proto User, resolving avatar CDN URL.
func (s *ProfileService) userToProto(u *gen.CoreUser) *pb.User {
	return &pb.User{
		Id:                  u.ID,
		Username:            u.Username,
		FirstName:           u.FirstName,
		LastName:            u.LastName,
		Email:               u.Email,
		AvatarUrl:           resolveAvatarURL(s.cdnBaseURL, safePtr(u.AvatarKey)),
		Bio:                 safePtr(u.Bio),
		CreatedAt:           timestamppb.New(u.CreatedAt),
		UpdatedAt:           timestamppb.New(u.UpdatedAt),
		OnboardingCompleted: u.OnboardingCompleted,
		AgeRange:            u.AgeRange,
	}
}

func resolveAvatarURL(cdnBase, key string) string {
	if key == "" {
		return ""
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	if cdnBase == "" {
		return key
	}
	return cdnBase + "/" + strings.TrimLeft(key, "/")
}

// achievementItemsToProto converts biz achievement items to proto AchievementList.
func achievementItemsToProto(items []biz.AchievementItem) *pb.AchievementList {
	pbItems := make([]*pb.Achievement, 0, len(items))
	for _, item := range items {
		a := &pb.Achievement{
			Id:            item.ID,
			Title:         item.Title,
			Description:   item.Description,
			Icon:          item.Icon,
			Color:         item.Color,
			Progress:      item.Progress,
			ProgressLabel: item.ProgressLabel,
			Unlocked:      item.Unlocked,
		}
		if item.UnlockedAt.Valid {
			a.UnlockedAt = timestamppb.New(item.UnlockedAt.Time)
		}
		pbItems = append(pbItems, a)
	}
	return &pb.AchievementList{Items: pbItems}
}

// safePtr dereferences a string pointer, returning empty string for nil.
func safePtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
