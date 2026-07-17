package service

import (
	"context"
	"net/http"

	pb "github.com/puchidemy/puchi-backend/app/learn/api/learn/v1"
	"github.com/puchidemy/puchi-backend/app/learn/internal/biz"
	"github.com/puchidemy/puchi-backend/app/learn/internal/conf"
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
func (s *LearnService) ClaimGuest(ctx context.Context, _ *pb.ClaimGuestRequest) (*pb.ClaimGuestResponse, error) {
	userID, ok := authpkg.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	guestIDStr, err := s.cookie.guestIDFromRequest(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	guestID, err := uuid.Parse(guestIDStr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid guest cookie")
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
	http.SetCookie(rw, &http.Cookie{
		Name:     w.cookieName(),
		Value:    guestID,
		Path:     "/",
		Domain:   ".puchi.io.vn",
		MaxAge:   w.maxAgeSeconds(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (w defaultGuestCookieWriter) clearGuestCookie(ctx context.Context) error {
	rw, ok := kratoshttp.ResponseWriterFromServerContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "missing response writer")
	}
	http.SetCookie(rw, &http.Cookie{
		Name:     w.cookieName(),
		Value:    "",
		Path:     "/",
		Domain:   ".puchi.io.vn",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
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
