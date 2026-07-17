package service

import (
	"context"

	pb "github.com/puchidemy/puchi-backend/app/learn/api/learn/v1"
	"github.com/puchidemy/puchi-backend/app/learn/internal/biz"

	"google.golang.org/protobuf/types/known/emptypb"
)

// LearnService implements both LearnServiceServer (gRPC) and
// LearnServiceHTTPServer (HTTP). Curriculum and guest-progress RPCs are
// added in later tasks of the learn-service reorg.
type LearnService struct {
	pb.UnimplementedLearnServiceServer
	uc *biz.LearnUsecase
}

// NewLearnService creates a new LearnService.
func NewLearnService(uc *biz.LearnUsecase) *LearnService {
	return &LearnService{uc: uc}
}

// Ping is a scaffold no-op RPC, kept until real curriculum RPCs are added.
func (s *LearnService) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
