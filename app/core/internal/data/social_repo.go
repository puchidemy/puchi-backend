package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// SocialRepo wraps sqlc-generated queries for social features.
type SocialRepo struct {
	q *gen.Queries
}

// NewSocialRepo creates a new SocialRepo.
func NewSocialRepo(pool *pgxpool.Pool) *SocialRepo {
	return &SocialRepo{q: gen.New(pool)}
}

// Follow adds a follow relationship.
func (r *SocialRepo) Follow(ctx context.Context, followerID, followingID string) error {
	return r.q.Follow(ctx, gen.FollowParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

// Unfollow removes a follow relationship.
func (r *SocialRepo) Unfollow(ctx context.Context, followerID, followingID string) error {
	return r.q.Unfollow(ctx, gen.UnfollowParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

// IsFollowing checks if followerID follows followingID.
func (r *SocialRepo) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	return r.q.IsFollowing(ctx, gen.IsFollowingParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

// ListFollowers returns followers of a user (userID).
// currentUserID is used to check if the current user follows back.
func (r *SocialRepo) ListFollowers(ctx context.Context, userID, currentUserID string) ([]gen.ListFollowersRow, error) {
	return r.q.ListFollowers(ctx, gen.ListFollowersParams{
		FollowingID: userID,
		FollowerID:  currentUserID,
	})
}

// ListFollowing returns users that a user follows.
func (r *SocialRepo) ListFollowing(ctx context.Context, userID string) ([]gen.ListFollowingRow, error) {
	return r.q.ListFollowing(ctx, userID)
}

// SearchUsers searches for users by query string.
// currentUserID is used to check follow status.
func (r *SocialRepo) SearchUsers(ctx context.Context, query, currentUserID string, limit int32) ([]gen.SearchUsersRow, error) {
	return r.q.SearchUsers(ctx, gen.SearchUsersParams{
		Column1:    query,
		FollowerID: currentUserID,
		Limit:      limit,
	})
}

// GetFollowCounts returns following and followers counts.
func (r *SocialRepo) GetFollowCounts(ctx context.Context, userID string) (gen.GetFollowCountsRow, error) {
	return r.q.GetFollowCounts(ctx, userID)
}

// GetWeeklyLeaderboard returns the current weekly leaderboard.
func (r *SocialRepo) GetWeeklyLeaderboard(ctx context.Context, limit int32) ([]gen.GetWeeklyLeaderboardRow, error) {
	return r.q.GetWeeklyLeaderboard(ctx, limit)
}

// UpsertSocialWeeklyXP adds XP to a user's weekly social leaderboard total.
func (r *SocialRepo) UpsertSocialWeeklyXP(ctx context.Context, userID string, weekStart string, xp int32) error {
	t, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return err
	}
	return r.q.UpsertSocialWeeklyXP(ctx, gen.UpsertSocialWeeklyXPParams{
		UserID:    userID,
		WeekStart: pgtype.Date{Time: t, Valid: true},
		WeeklyXp:  xp,
	})
}

// compile-time check that SocialRepo implements biz.SocialRepoInterface.
var _ biz.SocialRepoInterface = (*SocialRepo)(nil)
