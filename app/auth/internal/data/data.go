package data

import (
	"context"
	"log/slog"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/conf"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/cache"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/publisher"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewCache,
	NewUserRepo,
	NewSessionRepo,
	NewSocialConnectionRepo,
	NewMagicLinkRepo,
	NewPasswordResetTokenRepo,
	NewTOTPRepo,
	NewRoleRepo,
	NewPermissionRepo,
	NewAuditRepo,
	NewOAuthStateRepo,
	NewPublisherProvider,
	NewEmailVerificationRepo,
	wire.Bind(new(biz.UserRepo), new(*UserRepo)),
	wire.Bind(new(biz.EmailVerificationRepo), new(*EmailVerificationRepo)),
	wire.Bind(new(biz.SessionRepo), new(*SessionRepo)),
	wire.Bind(new(biz.SocialConnectionRepo), new(*SocialConnectionRepo)),
	wire.Bind(new(biz.MagicLinkRepo), new(*MagicLinkRepo)),
	wire.Bind(new(biz.PasswordResetTokenRepo), new(*PasswordResetTokenRepo)),
	wire.Bind(new(biz.TOTPRepo), new(*TOTPRepo)),
	wire.Bind(new(biz.RoleRepo), new(*RoleRepo)),
	wire.Bind(new(biz.PermissionRepo), new(*PermissionRepo)),
	wire.Bind(new(biz.EventPublisher), new(*publisher.Publisher)),
	wire.Bind(new(biz.AuditRepo), new(*AuditRepo)),
	wire.Bind(new(biz.SessionCache), new(*cache.Cache)),
	wire.Bind(new(biz.TokenBlacklist), new(*cache.Cache)),
	wire.Bind(new(biz.RateLimiter), new(*cache.Cache)),
)

// NewPublisherProvider extracts the publisher from Data for dependency injection.
func NewPublisherProvider(d *Data) *publisher.Publisher {
	return d.Publisher
}

// Data .
type Data struct {
	Pool      *pgxpool.Pool
	Publisher *publisher.Publisher
	Cache     *cache.Cache
}

// NewData .
func NewData(cfg *conf.Data) (*Data, func(), error) {
	pool, err := pgxpool.New(context.Background(), cfg.Database.Source)
	if err != nil {
		return nil, nil, err
	}

	var pub *publisher.Publisher
	natsURL := ""
	if cfg.Nats != nil {
		natsURL = cfg.Nats.Url
	}
	pub, err = publisher.New(natsURL, pool)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	pub.Start(context.Background())

	cleanup := func() {
		if pub != nil {
			pub.Close()
		}
		pool.Close()
	}
	return &Data{Pool: pool, Publisher: pub}, cleanup, nil
}

// NewCache creates a new Valkey cache client from config.
func NewCache(cfg *conf.Data) (*cache.Cache, func(), error) {
	addr := "localhost:6379"
	password := ""
	db := 0
	poolSize := 10
	if cfg.Valkey != nil {
		addr = cfg.Valkey.Addr
		password = cfg.Valkey.Password
		db = int(cfg.Valkey.Db)
		poolSize = int(cfg.Valkey.PoolSize)
	}
	c := cache.New(addr, password, db, poolSize)
	slog.Info("connected to Valkey", slog.String("addr", addr), slog.Int("db", db))
	cleanup := func() {
		if err := c.Close(); err != nil {
			slog.Warn("close Valkey", slog.Any("error", err))
		}
	}
	return c, cleanup, nil
}
