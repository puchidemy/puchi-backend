package biz

import (
	"context"
	"errors"
	"time"

	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// Domain errors.
var (
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)

// SocialRepoInterface defines the repository contract for social operations.
type SocialRepoInterface interface {
	Follow(ctx context.Context, followerID, followingID string) error
	Unfollow(ctx context.Context, followerID, followingID string) error
	IsFollowing(ctx context.Context, followerID, followingID string) (bool, error)
	ListFollowers(ctx context.Context, userID, currentUserID string) ([]gen.ListFollowersRow, error)
	ListFollowing(ctx context.Context, userID string) ([]gen.ListFollowingRow, error)
	SearchUsers(ctx context.Context, query, currentUserID string, limit int32) ([]gen.SearchUsersRow, error)
	GetFollowCounts(ctx context.Context, userID string) (gen.GetFollowCountsRow, error)
	GetWeeklyLeaderboard(ctx context.Context, limit int32) ([]gen.GetWeeklyLeaderboardRow, error)
	UpsertSocialWeeklyXP(ctx context.Context, userID string, weekStart string, xp int32) error
}

// SocialUsecase handles social operations.
type SocialUsecase struct {
	repo SocialRepoInterface
}

// NewSocialUsecase creates a new SocialUsecase.
func NewSocialUsecase(repo SocialRepoInterface) *SocialUsecase {
	return &SocialUsecase{repo: repo}
}

// Follow adds a follow relationship.
func (uc *SocialUsecase) Follow(ctx context.Context, followerID, followingID string) error {
	if followerID == followingID {
		return ErrCannotFollowSelf
	}
	return uc.repo.Follow(ctx, followerID, followingID)
}

// Unfollow removes a follow relationship.
func (uc *SocialUsecase) Unfollow(ctx context.Context, followerID, followingID string) error {
	if followerID == followingID {
		return ErrCannotFollowSelf
	}
	return uc.repo.Unfollow(ctx, followerID, followingID)
}

// ListFollowing returns users that the given user follows.
func (uc *SocialUsecase) ListFollowing(ctx context.Context, userID string) ([]gen.ListFollowingRow, error) {
	return uc.repo.ListFollowing(ctx, userID)
}

// ListFollowers returns followers of the given user.
func (uc *SocialUsecase) ListFollowers(ctx context.Context, userID string) ([]gen.ListFollowersRow, error) {
	return uc.repo.ListFollowers(ctx, userID, userID)
}

// SearchUsers searches for users by query string.
func (uc *SocialUsecase) SearchUsers(ctx context.Context, query, currentUserID string, limit int32) ([]gen.SearchUsersRow, error) {
	return uc.repo.SearchUsers(ctx, query, currentUserID, limit)
}

// GetWeeklyLeaderboard returns the weekly leaderboard.
func (uc *SocialUsecase) GetWeeklyLeaderboard(ctx context.Context) ([]gen.GetWeeklyLeaderboardRow, error) {
	return uc.repo.GetWeeklyLeaderboard(ctx, 100)
}

// GetWeekStart returns the Monday of the current week in YYYY-MM-DD format.
func GetWeekStart() string {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -int(weekday-time.Monday))
	return monday.Format("2006-01-02")
}
