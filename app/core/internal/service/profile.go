package service

import (
	"context"

	pb "github.com/puchidemy/puchi-backend/app/core/api/profile/v1"
	"github.com/puchidemy/puchi-backend/app/core/internal/auth"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"

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
}

// NewProfileService creates a new ProfileService.
func NewProfileService(uc *biz.ProfileUsecase, achievementUC *biz.AchievementUsecase, statsSvc *StatsService) *ProfileService {
	return &ProfileService{uc: uc, achievementUC: achievementUC, statsSvc: statsSvc}
}

// GetStats returns the authenticated user's gamification stats.
func (s *ProfileService) GetStats(ctx context.Context, req *emptypb.Empty) (*pb.Stats, error) {
	return s.statsSvc.GetStats(ctx, req)
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
func (s *ProfileService) GetProfile(ctx context.Context, _ *emptypb.Empty) (*pb.User, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	user, err := s.uc.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return userToProto(user), nil
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

	return userToProto(user), nil
}

// GetProfileByUsername returns a user's public profile by username.
func (s *ProfileService) GetProfileByUsername(ctx context.Context, req *pb.GetProfileByUsernameRequest) (*pb.User, error) {
	user, err := s.uc.GetProfileByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// If user is logged in and it's their own profile, show email
	currentUserID, isLoggedIn := auth.UserIDFromContext(ctx)
	userProto := userToProto(user)
	if !isLoggedIn || currentUserID != user.ID {
		userProto.Email = ""
	}
	return userProto, nil
}

// CompleteOnboarding completes onboarding and saves profile + answers.
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
		"13-17":  true,
		"18-24":  true,
		"25-34":  true,
		"35-44":  true,
		"45-54":  true,
		"55+":    true,
	}
	if !validAgeRanges[req.AgeRange] {
		return nil, status.Error(codes.InvalidArgument, "invalid age_range value")
	}

	user, err := s.uc.CompleteOnboarding(ctx, userID, biz.OnboardingInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AgeRange:  req.AgeRange,
		HowHeard:  req.HowHeard,
		WhyLearn:  req.WhyLearn,
		Level:     req.Level,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return userToProto(user), nil
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

// userToProto converts a gen.CoreUser to a proto User.
func userToProto(u *gen.CoreUser) *pb.User {
	return &pb.User{
		Id:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		AvatarUrl: safePtr(u.AvatarKey),
		Bio:       safePtr(u.Bio),
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
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
