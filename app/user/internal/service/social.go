package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"os"

	pb "github.com/puchidemy/puchi-backend/app/user/api/social/v1"
	"github.com/puchidemy/puchi-backend/app/user/internal/biz"
	"github.com/puchidemy/puchi-backend/app/user/internal/data/sqlc/gen"

	"github.com/go-kratos/kratos/v3/transport"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SocialService implements both SocialServiceServer (gRPC) and SocialServiceHTTPServer (HTTP).
type SocialService struct {
	pb.UnimplementedSocialServiceServer
	uc *biz.SocialUsecase
}

// NewSocialService creates a new SocialService.
func NewSocialService(uc *biz.SocialUsecase) *SocialService {
	return &SocialService{uc: uc}
}

// userIDFromContext extracts the authenticated user ID from the request context.
// Envoy Gateway injects X-User-ID header after auth verification.
// If X-User-ID-Signature is present, it is verified using HMAC-SHA256
// with the shared secret from HMAC_SECRET env var. If not present,
// the header is trusted directly (dev mode behind Envoy).
func userIDFromContext(ctx context.Context) (string, error) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing transport context")
	}

	userID := tr.RequestHeader().Get("X-User-ID")
	if userID == "" {
		return "", status.Error(codes.Unauthenticated, "missing user id")
	}

	signature := tr.RequestHeader().Get("X-User-ID-Signature")
	if signature != "" {
		secret := os.Getenv("HMAC_SECRET")
		if secret == "" {
			return "", status.Error(codes.Internal, "HMAC_SECRET not configured")
		}

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(userID))
		expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expected)) {
			return "", status.Error(codes.Unauthenticated, "invalid user id signature")
		}
	}

	return userID, nil
}

// Follow follows a user.
func (s *SocialService) Follow(ctx context.Context, req *pb.FollowRequest) (*emptypb.Empty, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.uc.Follow(ctx, userID, req.FollowingId); err != nil {
		if err == biz.ErrCannotFollowSelf {
			return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// Unfollow unfollows a user.
func (s *SocialService) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*emptypb.Empty, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.uc.Unfollow(ctx, userID, req.FollowingId); err != nil {
		if err == biz.ErrCannotFollowSelf {
			return nil, status.Error(codes.InvalidArgument, "cannot unfollow yourself")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// ListFollowing lists users the current user follows.
func (s *SocialService) ListFollowing(ctx context.Context, req *pb.ListFollowingRequest) (*pb.SocialUserList, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.uc.ListFollowing(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SocialUserList{Items: followingRowsToProto(users)}, nil
}

// ListFollowers lists followers of the current user.
func (s *SocialService) ListFollowers(ctx context.Context, req *pb.ListFollowersRequest) (*pb.SocialUserList, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.uc.ListFollowers(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SocialUserList{Items: followerRowsToProto(users)}, nil
}

// SearchUsers searches for users.
func (s *SocialService) SearchUsers(ctx context.Context, req *pb.SearchUsersRequest) (*pb.SocialUserList, error) {
	if len(req.Query) < 2 {
		return nil, status.Error(codes.InvalidArgument, "query must be at least 2 characters")
	}
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pageSize := int32(20)
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	users, err := s.uc.SearchUsers(ctx, req.Query, userID, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SocialUserList{Items: searchUserRowsToProto(users)}, nil
}

// GetWeeklyLeaderboard returns the weekly leaderboard.
func (s *SocialService) GetWeeklyLeaderboard(ctx context.Context, _ *emptypb.Empty) (*pb.LeaderboardList, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	entries, err := s.uc.GetWeeklyLeaderboard(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LeaderboardList{Items: leaderboardRowsToProto(entries, userID)}, nil
}

// -- conversion helpers --

func followingRowsToProto(rows []gen.ListFollowingRow) []*pb.SocialUser {
	items := make([]*pb.SocialUser, len(rows))
	for i, r := range rows {
		items[i] = &pb.SocialUser{
			Id:          r.ID,
			Username:    r.Username,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			AvatarUrl:   safeStr(r.AvatarKey),
			Level:       r.Level,
			Streak:      r.Streak,
			IsFollowing: true,
		}
	}
	return items
}

func followerRowsToProto(rows []gen.ListFollowersRow) []*pb.SocialUser {
	items := make([]*pb.SocialUser, len(rows))
	for i, r := range rows {
		items[i] = &pb.SocialUser{
			Id:          r.ID,
			Username:    r.Username,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			AvatarUrl:   safeStr(r.AvatarKey),
			Level:       r.Level,
			Streak:      r.Streak,
			IsFollowing: r.IsFollowing,
		}
	}
	return items
}

func searchUserRowsToProto(rows []gen.SearchUsersRow) []*pb.SocialUser {
	items := make([]*pb.SocialUser, len(rows))
	for i, r := range rows {
		items[i] = &pb.SocialUser{
			Id:          r.ID,
			Username:    r.Username,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			AvatarUrl:   safeStr(r.AvatarKey),
			Level:       r.Level,
			Streak:      r.Streak,
			IsFollowing: r.IsFollowing,
		}
	}
	return items
}

func leaderboardRowsToProto(rows []gen.GetWeeklyLeaderboardRow, currentUserID string) []*pb.LeaderboardEntry {
	items := make([]*pb.LeaderboardEntry, len(rows))
	for i, r := range rows {
		rank := r.Rank
		items[i] = &pb.LeaderboardEntry{
			Rank:          rank,
			UserId:        r.UserID,
			Username:      r.Username,
			AvatarUrl:     safeStr(r.AvatarKey),
			Level:         r.Level,
			WeeklyXp:      r.WeeklyXp,
			IsCurrentUser: r.UserID == currentUserID,
		}
	}
	return items
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
