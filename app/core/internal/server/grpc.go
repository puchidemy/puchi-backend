package server

import (
	profilepb "github.com/puchidemy/puchi-backend/app/core/api/profile/v1"
	socialpb "github.com/puchidemy/puchi-backend/app/core/api/social/v1"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
	"github.com/puchidemy/puchi-backend/app/core/internal/service"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, _ *conf.Auth, profileService *service.ProfileService, socialService *service.SocialService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	profilepb.RegisterProfileServiceServer(srv, profileService)
	socialpb.RegisterSocialServiceServer(srv, socialService)
	return srv
}
