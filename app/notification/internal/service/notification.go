package service

import (
	"context"

	pb "github.com/puchidemy/puchi-backend/app/notification/api/notification/v1"
	"github.com/puchidemy/puchi-backend/app/notification/internal/biz"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// NotificationService implements the NotificationService proto.
type NotificationService struct {
	pb.UnimplementedNotificationServiceServer
	uc *biz.NotificationUsecase
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(uc *biz.NotificationUsecase) *NotificationService {
	return &NotificationService{uc: uc}
}

// GetPreferences returns notification preferences for a user.
func (s *NotificationService) GetPreferences(ctx context.Context, req *pb.GetPreferencesRequest) (*pb.Preferences, error) {
	p, err := s.uc.GetPreferences(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return preferenceToProto(p), nil
}

// UpdatePreferences updates notification preferences for a user.
func (s *NotificationService) UpdatePreferences(ctx context.Context, req *pb.UpdatePreferencesRequest) (*pb.Preferences, error) {
	prefs := &biz.Preference{
		UserID: req.GetUserId(),
	}
	if req.PushEnabled != nil {
		prefs.PushEnabled = req.PushEnabled.Value
	}
	if req.StreakReminder != nil {
		prefs.StreakReminder = req.StreakReminder.Value
	}
	if req.FriendActivity != nil {
		prefs.FriendActivity = req.FriendActivity.Value
	}

	p, err := s.uc.UpdatePreferences(ctx, prefs)
	if err != nil {
		return nil, err
	}
	return preferenceToProto(p), nil
}

// Send sends a push notification via Gotify.
func (s *NotificationService) Send(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	sent, err := s.uc.SendNotification(ctx, req.GetUserId(), req.GetCategory(), req.GetTemplateId(), req.GetParams())
	if err != nil {
		return nil, err
	}
	return &pb.SendNotificationResponse{Sent: sent}, nil
}

func preferenceToProto(p *biz.Preference) *pb.Preferences {
	prefs := &pb.Preferences{
		UserId:         p.UserID,
		PushEnabled:    p.PushEnabled,
		EmailEnabled:   p.EmailEnabled,
		StreakReminder: p.StreakReminder,
		FriendActivity: p.FriendActivity,
		Promotions:     p.Promotions,
	}
	if p.QuietHoursStart != nil {
		prefs.QuietHoursStart = wrapperspb.String(*p.QuietHoursStart)
	}
	if p.QuietHoursEnd != nil {
		prefs.QuietHoursEnd = wrapperspb.String(*p.QuietHoursEnd)
	}
	return prefs
}
