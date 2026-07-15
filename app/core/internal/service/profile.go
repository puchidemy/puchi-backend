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
	uc *biz.ProfileUsecase
}

// NewProfileService creates a new ProfileService.
func NewProfileService(uc *biz.ProfileUsecase) *ProfileService {
	return &ProfileService{uc: uc}
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
	})
	if err != nil {
		if err == biz.ErrUsernameTaken {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return userToProto(user), nil
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

// safePtr dereferences a string pointer, returning empty string for nil.
func safePtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
