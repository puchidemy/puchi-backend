package auth

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v3/middleware"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/puchidemy/puchi-backend/pkg/apierr"
)

// KratosMiddleware validates Limen session (Bearer or limen_session cookie)
// and injects identity into the handler context. Prefer this over Middleware
// (net/http Filter) so values survive Kratos transport wrapping.
func KratosMiddleware(cfg MiddlewareConfig) middleware.Middleware {
	public := append([]string(nil), cfg.PublicPaths...)
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			r, ok := khttp.RequestFromServerContext(ctx)
			if !ok {
				return next(ctx, req)
			}

			isPublic := false
			path := r.URL.Path
			for _, p := range public {
				if p != "" && strings.HasPrefix(path, p) {
					isPublic = true
					break
				}
			}
			if !isPublic && cfg.IsPublic != nil && cfg.IsPublic(r) {
				isPublic = true
			}

			tokenStr := SessionTokenFromRequest(r)
			if isPublic {
				if tokenStr != "" && cfg.Validator != nil {
					if info, err := cfg.Validator.ParseAndValidate(ctx, tokenStr); err == nil {
						ctx = NewContextWithUserID(ctx, info.UserID)
						ctx = NewContextWithEmail(ctx, info.Email)
						ctx = NewContextWithRoles(ctx, info.Roles)
					}
				}
				return next(ctx, req)
			}

			if tokenStr == "" || cfg.Validator == nil {
				return nil, apierr.Unauthenticated("not authenticated")
			}
			info, err := cfg.Validator.ParseAndValidate(ctx, tokenStr)
			if err != nil {
				return nil, apierr.Unauthenticated("not authenticated")
			}
			ctx = NewContextWithUserID(ctx, info.UserID)
			ctx = NewContextWithEmail(ctx, info.Email)
			ctx = NewContextWithRoles(ctx, info.Roles)
			return next(ctx, req)
		}
	}
}
