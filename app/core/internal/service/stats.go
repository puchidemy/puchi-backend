package service

import (
	"context"

	pb "github.com/puchidemy/puchi-backend/app/core/api/profile/v1"
	auth "github.com/puchidemy/puchi-backend/pkg/auth"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// StatsService handles gamification stats operations.
type StatsService struct {
	uc *biz.StatsUsecase
}

// NewStatsService creates a new StatsService.
func NewStatsService(uc *biz.StatsUsecase) *StatsService {
	return &StatsService{uc: uc}
}

// GetStats returns the authenticated user's gamification stats.
func (s *StatsService) GetStats(ctx context.Context, _ *emptypb.Empty) (*pb.Stats, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	stats, err := s.uc.GetStats(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "stats not found")
	}

	return statsToProto(stats, s.uc.GetXPToNextLevel(ctx, stats.Level)), nil
}

// statsToProto converts a gen.CoreUserStat to a proto Stats.
func statsToProto(stats *gen.CoreUserStat, xpToNextLevel int32) *pb.Stats {
	return &pb.Stats{
		TotalLessons:     stats.TotalLessons,
		CompletedLessons: stats.CompletedLessons,
		TotalMinutes:     stats.TotalMinutes,
		Accuracy:         stats.Accuracy,
		WordsLearned:     stats.WordsLearned,
		CurrentXp:        stats.CurrentXp,
		TotalXp:          stats.TotalXp,
		Level:            stats.Level,
		XpToNextLevel:    xpToNextLevel,
		Streak:           stats.CurrentStreak,
		LongestStreak:    stats.LongestStreak,
		StreakFreezes:    stats.StreakFreezes,
		Crowns:           stats.Crowns,
		Gems:             stats.Gems,
	}
}
