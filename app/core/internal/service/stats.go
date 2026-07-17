package service

import (
	"context"
	"strings"
	"time"

	pb "github.com/puchidemy/puchi-backend/app/core/api/profile/v1"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
	auth "github.com/puchidemy/puchi-backend/pkg/auth"

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
		return nil, status.Error(codes.Internal, "failed to get stats")
	}

	return statsToProto(stats, s.uc.GetXPToNextLevel(ctx, stats.Level)), nil
}

// ListDailyActivity returns daily activity for the authenticated user.
func (s *StatsService) ListDailyActivity(ctx context.Context, req *pb.ListDailyActivityRequest) (*pb.DailyActivityList, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	from, err := parseDateQuery(req.GetFrom())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid from date, want YYYY-MM-DD")
	}
	to, err := parseDateQuery(req.GetTo())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid to date, want YYYY-MM-DD")
	}

	rows, err := s.uc.ListDailyActivity(ctx, userID, from, to)
	if err != nil {
		if strings.Contains(err.Error(), "from must be") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to list daily activity")
	}

	items := make([]*pb.DailyActivityItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dailyActivityToProto(row))
	}
	return &pb.DailyActivityList{Items: items}, nil
}

// ListWeeklyXP returns weekly XP history for the authenticated user.
func (s *StatsService) ListWeeklyXP(ctx context.Context, req *pb.ListWeeklyXPRequest) (*pb.WeeklyXPList, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	items, err := s.uc.ListWeeklyXP(ctx, userID, int(req.GetWeeks()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list weekly xp")
	}

	pbItems := make([]*pb.WeeklyXPItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, &pb.WeeklyXPItem{
			WeekLabel: item.WeekLabel,
			Xp:        item.XP,
		})
	}
	return &pb.WeeklyXPList{Items: pbItems}, nil
}

func parseDateQuery(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}

func dailyActivityToProto(row gen.CoreDailyActivity) *pb.DailyActivityItem {
	date := ""
	if row.ActivityDate.Valid {
		date = row.ActivityDate.Time.Format("2006-01-02")
	}
	return &pb.DailyActivityItem{
		Date:             date,
		LessonsCompleted: row.LessonsCompleted,
		XpEarned:         row.XpEarned,
		MinutesSpent:     row.MinutesSpent,
	}
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
