package service

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	pb "github.com/puchidemy/puchi-backend/app/learn/api/learn/v1"
	"github.com/puchidemy/puchi-backend/app/learn/internal/biz"
	"github.com/puchidemy/puchi-backend/app/learn/internal/conf"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"

	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// LearnService implements both LearnServiceServer (gRPC) and
// LearnServiceHTTPServer (HTTP).
type LearnService struct {
	pb.UnimplementedLearnServiceServer
	uc     *biz.LearnUsecase
	learn  *conf.Learn
	cookie guestCookieWriter
}

type guestCookieWriter interface {
	setGuestCookie(ctx context.Context, guestID string) error
	clearGuestCookie(ctx context.Context) error
	guestIDFromRequest(ctx context.Context) (string, error)
}

// NewLearnService creates a new LearnService.
func NewLearnService(uc *biz.LearnUsecase, learn *conf.Learn) *LearnService {
	return &LearnService{
		uc:     uc,
		learn:  learn,
		cookie: defaultGuestCookieWriter{learn: learn},
	}
}

// Ping is a scaffold no-op RPC.
func (s *LearnService) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// CreateGuestSession creates a server-side guest and sets the HttpOnly cookie.
func (s *LearnService) CreateGuestSession(ctx context.Context, _ *emptypb.Empty) (*pb.GuestSession, error) {
	guestID, err := s.uc.CreateGuestSession(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := s.cookie.setGuestCookie(ctx, guestID.String()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GuestSession{GuestId: guestID.String()}, nil
}

// ClaimGuest merges guest progress into the authenticated user.
// No guest cookie / invalid cookie → idempotent success (0 merged). Callers
// often hit claim after every login even when the user never started a trial.
func (s *LearnService) ClaimGuest(ctx context.Context, _ *pb.ClaimGuestRequest) (*pb.ClaimGuestResponse, error) {
	userID, ok := authpkg.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	guestIDStr, err := s.cookie.guestIDFromRequest(ctx)
	if err != nil {
		return &pb.ClaimGuestResponse{LessonsMerged: 0}, nil
	}
	guestID, err := uuid.Parse(guestIDStr)
	if err != nil {
		return &pb.ClaimGuestResponse{LessonsMerged: 0}, nil
	}

	merged, err := s.uc.ClaimGuest(ctx, userID, guestID)
	if err != nil {
		return nil, mapGuestError(err)
	}
	if err := s.cookie.clearGuestCookie(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ClaimGuestResponse{LessonsMerged: merged}, nil
}

// GetUnit returns a unit with nested skills and lessons.
func (s *LearnService) GetUnit(ctx context.Context, req *pb.GetUnitRequest) (*pb.GetUnitResponse, error) {
	ownerType, ownerID, err := s.resolveOwner(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := s.uc.GetUnit(ctx, ownerType, ownerID, req.GetId(), s.trialUnitID())
	if err != nil {
		return nil, mapCurriculumError(err)
	}
	return unitDetailToProto(unit), nil
}

// GetLesson returns a lesson with exercises (prompts only; answers withheld).
func (s *LearnService) GetLesson(ctx context.Context, req *pb.GetLessonRequest) (*pb.GetLessonResponse, error) {
	ownerType, ownerID, err := s.resolveOwner(ctx)
	if err != nil {
		return nil, err
	}

	lesson, err := s.uc.GetLesson(ctx, ownerType, ownerID, req.GetId(), s.trialUnitID())
	if err != nil {
		return nil, mapCurriculumError(err)
	}
	return lessonDetailToProto(lesson), nil
}

// StartLesson creates a lesson attempt for the resolved owner.
func (s *LearnService) StartLesson(ctx context.Context, req *pb.StartLessonRequest) (*pb.StartLessonResponse, error) {
	ownerType, ownerID, err := s.resolveOwner(ctx)
	if err != nil {
		return nil, err
	}

	attemptID, err := s.uc.StartLesson(ctx, ownerType, ownerID, req.GetId(), s.trialUnitID())
	if err != nil {
		return nil, mapAttemptError(err)
	}
	return &pb.StartLessonResponse{AttemptId: attemptID.String()}, nil
}

// SubmitAnswer grades and stores an answer for an attempt.
func (s *LearnService) SubmitAnswer(ctx context.Context, req *pb.SubmitAnswerRequest) (*pb.SubmitAnswerResponse, error) {
	ownerType, ownerID, err := s.resolveOwner(ctx)
	if err != nil {
		return nil, err
	}

	correct, err := s.uc.SubmitAnswer(ctx, ownerType, ownerID, req.GetAttemptId(), req.GetExerciseId(), json.RawMessage(req.GetPayloadJson()), s.trialUnitID())
	if err != nil {
		return nil, mapAttemptError(err)
	}
	return &pb.SubmitAnswerResponse{Correct: correct}, nil
}

// CompleteLesson finalizes the active attempt and returns session XP.
func (s *LearnService) CompleteLesson(ctx context.Context, req *pb.CompleteLessonRequest) (*pb.CompleteLessonResponse, error) {
	ownerType, ownerID, err := s.resolveOwner(ctx)
	if err != nil {
		return nil, err
	}

	xp, unitCompleted, err := s.uc.CompleteLesson(ctx, ownerType, ownerID, req.GetId(), s.trialUnitID())
	if err != nil {
		return nil, mapAttemptError(err)
	}
	return &pb.CompleteLessonResponse{Xp: xp, UnitCompleted: unitCompleted}, nil
}

// resolveOwner prefers an authenticated user from context, else guest cookie.
func (s *LearnService) resolveOwner(ctx context.Context) (ownerType, ownerID string, err error) {
	if userID, ok := authpkg.UserIDFromContext(ctx); ok {
		return "user", userID, nil
	}
	guestID, err := s.cookie.guestIDFromRequest(ctx)
	if err != nil {
		return "", "", status.Error(codes.Unauthenticated, "authentication required")
	}
	return "guest", guestID, nil
}

func (s *LearnService) trialUnitID() string {
	if s.learn != nil && s.learn.TrialUnitId != "" {
		return s.learn.TrialUnitId
	}
	return "11111111-1111-1111-1111-111111111111"
}

func mapCurriculumError(err error) error {
	switch {
	case errors.Is(err, biz.ErrTrialLimit):
		return status.Error(codes.PermissionDenied, "TRIAL_LIMIT")
	case errors.Is(err, biz.ErrCurriculumNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func mapAttemptError(err error) error {
	switch {
	case errors.Is(err, biz.ErrTrialLimit):
		return status.Error(codes.PermissionDenied, "TRIAL_LIMIT")
	case errors.Is(err, biz.ErrCurriculumNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrAttemptNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrAttemptForbidden), errors.Is(err, biz.ErrExerciseForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, biz.ErrAttemptNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, biz.ErrExerciseNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func unitDetailToProto(unit *biz.UnitDetail) *pb.GetUnitResponse {
	resp := &pb.GetUnitResponse{
		Unit: &pb.Unit{
			Id:       unit.Unit.ID,
			CourseId: unit.Unit.CourseID,
			Position: unit.Unit.Position,
			Title:    unit.Unit.Title,
		},
		UnitStatus: unit.UnitStatus,
	}
	for _, skill := range unit.Skills {
		pbSkill := &pb.Skill{
			Id:       skill.Skill.ID,
			UnitId:   skill.Skill.UnitID,
			Position: skill.Skill.Position,
			Title:    skill.Skill.Title,
		}
		for _, lesson := range skill.Lessons {
			pbSkill.Lessons = append(pbSkill.Lessons, lessonToProto(lesson.Lesson, lesson.Status))
		}
		resp.Skills = append(resp.Skills, pbSkill)
	}
	return resp
}

func lessonDetailToProto(lesson *biz.LessonDetail) *pb.GetLessonResponse {
	resp := &pb.GetLessonResponse{
		Lesson: lessonToProto(lesson.Lesson, ""),
	}
	for _, exercise := range lesson.Exercises {
		resp.Exercises = append(resp.Exercises, exerciseToProto(exercise))
	}
	return resp
}

func lessonToProto(lesson gen.LearnLesson, status string) *pb.Lesson {
	return &pb.Lesson{
		Id:       lesson.ID,
		SkillId:  lesson.SkillID,
		Position: lesson.Position,
		Title:    lesson.Title,
		XpReward: lesson.XpReward,
		Required: lesson.Required,
		Status:   status,
	}
}

func exerciseToProto(exercise gen.LearnExercise) *pb.Exercise {
	return &pb.Exercise{
		Id:         exercise.ID,
		LessonId:   exercise.LessonID,
		Position:   exercise.Position,
		Type:       exercise.Type,
		PromptJson: string(exercise.Prompt),
	}
}

func mapGuestError(err error) error {
	switch err {
	case biz.ErrGuestNotFound:
		return status.Error(codes.NotFound, err.Error())
	case biz.ErrGuestAlreadyClaimed:
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

type defaultGuestCookieWriter struct {
	learn *conf.Learn
}

func (w defaultGuestCookieWriter) cookieName() string {
	if w.learn != nil && w.learn.GuestCookieName != "" {
		return w.learn.GuestCookieName
	}
	return "puchi_guest_id"
}

func (w defaultGuestCookieWriter) maxAgeSeconds() int {
	if w.learn != nil && w.learn.GuestTtlDays > 0 {
		return int(w.learn.GuestTtlDays) * 24 * 60 * 60
	}
	return 30 * 24 * 60 * 60
}

func (w defaultGuestCookieWriter) setGuestCookie(ctx context.Context, guestID string) error {
	rw, ok := kratoshttp.ResponseWriterFromServerContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "missing response writer")
	}
	domain, secure := guestCookieScope(ctx)
	http.SetCookie(rw, &http.Cookie{
		Name:     w.cookieName(),
		Value:    guestID,
		Path:     "/",
		Domain:   domain,
		MaxAge:   w.maxAgeSeconds(),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (w defaultGuestCookieWriter) clearGuestCookie(ctx context.Context) error {
	rw, ok := kratoshttp.ResponseWriterFromServerContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "missing response writer")
	}
	domain, secure := guestCookieScope(ctx)
	http.SetCookie(rw, &http.Cookie{
		Name:     w.cookieName(),
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// guestCookieScope: localhost/dev → host-only + non-Secure (HTTP).
// Prod (api.puchi.io.vn) → .puchi.io.vn + Secure.
func guestCookieScope(ctx context.Context) (domain string, secure bool) {
	req, ok := kratoshttp.RequestFromServerContext(ctx)
	if !ok {
		return ".puchi.io.vn", true
	}
	host := req.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1":
		return "", false
	default:
		return ".puchi.io.vn", true
	}
}

func (w defaultGuestCookieWriter) guestIDFromRequest(ctx context.Context) (string, error) {
	req, ok := kratoshttp.RequestFromServerContext(ctx)
	if !ok {
		return "", status.Error(codes.InvalidArgument, "missing request")
	}
	cookie, err := req.Cookie(w.cookieName())
	if err != nil || cookie.Value == "" {
		return "", status.Error(codes.InvalidArgument, "guest cookie required")
	}
	return cookie.Value, nil
}
