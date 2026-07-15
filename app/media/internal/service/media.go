package service

import (
	"context"

	pb "github.com/puchidemy/puchi-backend/app/media/api/media/v1"
	"github.com/puchidemy/puchi-backend/app/media/internal/biz"
	"github.com/puchidemy/puchi-backend/app/media/internal/data/sqlc/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MediaService implements both MediaServiceServer (gRPC) and MediaServiceHTTPServer (HTTP).
type MediaService struct {
	pb.UnimplementedMediaServiceServer
	uc *biz.MediaUsecase
}

// NewMediaService creates a new MediaService.
func NewMediaService(uc *biz.MediaUsecase) *MediaService {
	return &MediaService{uc: uc}
}

// RequestUploadURL generates a presigned upload URL for a new media object.
func (s *MediaService) RequestUploadURL(ctx context.Context, req *pb.RequestUploadURLRequest) (*pb.RequestUploadURLResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	obj, uploadURL, objectKey, ttl, err := s.uc.RequestUpload(ctx, userID, req.Category, req.ContentType, req.ContentLength)
	if err != nil {
		return nil, mapBizError(err)
	}

	return &pb.RequestUploadURLResponse{
		UploadUrl:        uploadURL,
		ObjectKey:        objectKey,
		MediaId:          obj.ID,
		ExpiresInSeconds: ttl,
	}, nil
}

// FinalizeUpload marks a media object as ready after upload completes.
func (s *MediaService) FinalizeUpload(ctx context.Context, req *pb.FinalizeUploadRequest) (*pb.MediaObject, error) {
	obj, err := s.uc.FinalizeUpload(ctx, req.MediaId)
	if err != nil {
		return nil, mapBizError(err)
	}

	return mediaObjectToProto(obj, ""), nil
}

// GetMedia retrieves a media object with a download URL.
func (s *MediaService) GetMedia(ctx context.Context, req *pb.GetMediaRequest) (*pb.MediaObject, error) {
	obj, downloadURL, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, mapBizError(err)
	}

	return mediaObjectToProto(obj, downloadURL), nil
}

// DeleteMedia deletes a media object.
func (s *MediaService) DeleteMedia(ctx context.Context, req *pb.DeleteMediaRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteMedia(ctx, req.Id); err != nil {
		return nil, mapBizError(err)
	}
	return &emptypb.Empty{}, nil
}

// mediaObjectToProto converts a gen.MediaObject to a proto MediaObject.
func mediaObjectToProto(obj *gen.MediaObject, url string) *pb.MediaObject {
	return &pb.MediaObject{
		Id:          obj.ID,
		Url:         url,
		SizeBytes:   obj.SizeBytes,
		Width:       derefInt32(obj.Width),
		Height:      derefInt32(obj.Height),
		DurationMs:  derefInt32(obj.DurationMs),
		Category:    obj.Category,
		ContentType: obj.ContentType,
	}
}

// derefInt32 dereferences an *int32 pointer, returning 0 for nil.
func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// extractUserID extracts the user ID from the request context.
// It checks for X-User-ID metadata header (set by auth middleware or passed by client).
func extractUserID(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-user-id"); len(vals) > 0 && vals[0] != "" {
			return vals[0], nil
		}
		// Also check grpcgateway-x-user-id for HTTP gateway compatibility
		if vals := md.Get("grpcgateway-x-user-id"); len(vals) > 0 && vals[0] != "" {
			return vals[0], nil
		}
	}
	return "", status.Error(codes.Unauthenticated, "user not authenticated")
}

// mapBizError converts a biz domain error to a gRPC status error.
func mapBizError(err error) error {
	switch {
	case err == biz.ErrMediaNotFound:
		return status.Error(codes.NotFound, err.Error())
	case err == biz.ErrInvalidCategory, err == biz.ErrInvalidContentType, err == biz.ErrMediaTooLarge:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		if st, ok := status.FromError(err); ok {
			return st.Err()
		}
		return status.Error(codes.Internal, err.Error())
	}
}
