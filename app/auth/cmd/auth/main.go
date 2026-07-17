package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/thecodearcher/limen"
	sqladapter "github.com/thecodearcher/limen/adapters/sql"
	credentialpassword "github.com/thecodearcher/limen/plugins/credential-password"
	"github.com/thecodearcher/limen/plugins/oauth"
	oauthfacebook "github.com/thecodearcher/limen/plugins/oauth-facebook"
	oauthgoogle "github.com/thecodearcher/limen/plugins/oauth-google"

	"github.com/puchidemy/puchi-backend/app/auth/internal/config"
	"github.com/puchidemy/puchi-backend/app/auth/internal/events"
	oauthtiktok "github.com/puchidemy/puchi-backend/app/auth/internal/oauth/tiktok"
)

func main() {
	confDir := flag.String("conf", "../../configs", "config directory or file")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfgPath := resolveConfigPath(*confDir)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Error("load config", "err", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.Data.Database.Source)
	if err != nil {
		log.Error("open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Warn("database ping failed (continuing)", "err", err)
	}

	pub, err := events.New(cfg.Data.NATS.URL, log)
	if err != nil {
		log.Error("nats", "err", err)
		os.Exit(1)
	}
	defer pub.Close()

	frontendURL := cfg.Limen.FrontendURL

	httpOpts := []limen.HTTPConfigOption{
		limen.WithHTTPTrustedOrigins(cfg.Limen.TrustedOrigins),
		limen.WithHTTPSessionTransformer(func(user map[string]any, sessionResult *limen.SessionResult) (map[string]any, error) {
			out := map[string]any{"user": user}
			if sessionResult != nil && sessionResult.Token != "" {
				out["token"] = sessionResult.Token
			}
			return out, nil
		}),
	}
	if cfg.Limen.CookieDomain != "" {
		httpOpts = append(httpOpts, limen.WithHTTPCookieCrossSubdomainEnabled(cfg.Limen.CookieDomain))
	} else if !isLocalHTTPBaseURL(cfg.Limen.BaseURL) {
		// SameSite=None; Secure — only safe on HTTPS (prod).
		httpOpts = append(httpOpts, limen.WithHTTPCookieCrossDomainEnabled())
	}
	if isLocalHTTPBaseURL(cfg.Limen.BaseURL) {
		// Limen defaults Secure=true; browsers drop those cookies on http://localhost.
		httpOpts = append(httpOpts, limen.WithHTTPCookieSecure(false))
	}

	auth, err := limen.New(&limen.Config{
		BaseURL:  cfg.Limen.BaseURL,
		Database: sqladapter.NewPostgreSQL(db),
		Secret:   []byte(cfg.Limen.Secret),
		Schema: limen.NewDefaultSchemaConfig(
			limen.WithSchemaIDGenerator(&uuidGenerator{}),
			limen.WithSchemaUser(
				limen.WithUserSerializer(func(u *limen.User) map[string]any {
					raw := u.Raw()
					if raw == nil {
						raw = map[string]any{}
					}
					out := map[string]any{
						"id":    formatLimenUserID(u.ID),
						"email": u.Email,
					}
					if u.EmailVerifiedAt != nil {
						out["email_verified_at"] = u.EmailVerifiedAt
					}
					for _, k := range []string{"username", "first_name", "last_name", "created_at", "updated_at"} {
						if v, ok := raw[k]; ok {
							out[k] = v
						}
					}
					return out
				}),
			),
		),
		Session: limen.NewDefaultSessionConfig(
			limen.WithBearerEnabled(),
		),
		HTTP: limen.NewDefaultHTTPConfig(append(httpOpts,
			limen.WithHTTPHooks(&limen.Hooks{
				After: []*limen.Hook{{
					PathMatcher: func(ctx *limen.HookContext) bool {
						id := ctx.RouteID()
						return id == "signup" || id == "oauth-callback"
					},
					Run: func(ctx *limen.HookContext) bool {
						if ar := ctx.GetAuthResult(); ar != nil && ar.User != nil {
							username := ""
							if raw := ar.User.Raw(); raw != nil {
								if u, ok := raw["username"].(string); ok {
									username = u
								}
							}
							pub.UserCreated(fmt.Sprint(ar.User.ID), ar.User.Email, username)
						}
						return true
					},
				}},
			}),
		)...),
		Email: limen.NewDefaultEmailConfig(
			limen.WithEmailVerification(
				limen.WithSendEmailVerificationMail(func(email, token string) {
					link := fmt.Sprintf("%s/auth/verify-email?token=%s", frontendURL, token)
					pub.SendEmail(email, "email-verify", map[string]any{
						"link":  link,
						"token": token,
					})
				}),
			),
		),
		Plugins: []limen.Plugin{
			credentialpassword.New(
				credentialpassword.WithUsernameSupport(true),
				credentialpassword.WithSendPasswordResetEmail(func(email, token string) {
					link := fmt.Sprintf("%s/auth/reset-password?token=%s", frontendURL, token)
					pub.SendEmail(email, "password-reset", map[string]any{
						"link":  link,
						"token": token,
					})
				}),
			),
			oauth.New(
				oauth.WithProviders(
					oauthgoogle.New(),
					oauthfacebook.New(),
					oauthtiktok.New(),
				),
				oauth.WithMapProfileToUser(func(info *limen.OAuthAccountProfile) map[string]any {
					out := map[string]any{}
					if info.Name != "" {
						out["first_name"] = info.Name
					}
					return out
				}),
			),
		},
	})
	if err != nil {
		log.Error("limen init", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/auth/", auth.Handler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// Cluster-internal session validate for other Go services (Bearer or cookie).
	mux.HandleFunc("GET /internal/session", func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.GetSession(r)
		if err != nil || session == nil || session.User == nil {
			http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		username := ""
		if raw := session.User.Raw(); raw != nil {
			if u, ok := raw["username"].(string); ok {
				username = u
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":  formatLimenUserID(session.User.ID),
			"email":    session.User.Email,
			"username": username,
		})
	})

	srv := &http.Server{
		Addr:              cfg.Server.HTTP.Addr,
		Handler:           withCORS(mux, cfg.Limen.TrustedOrigins),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("auth-service listening", "addr", cfg.Server.HTTP.Addr, "base_url", cfg.Limen.BaseURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func resolveConfigPath(conf string) string {
	info, err := os.Stat(conf)
	if err == nil && !info.IsDir() {
		return conf
	}
	candidates := []string{
		filepath.Join(conf, "config.yaml"),
		filepath.Join(conf, "config.yml"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return filepath.Join(conf, "config.yaml")
}

type uuidGenerator struct{}

func (g *uuidGenerator) GetColumnType() limen.ColumnType { return limen.ColumnTypeUUID }

func (g *uuidGenerator) Generate(context.Context) (any, error) {
	return uuid.New().String(), nil
}

func isLocalHTTPBaseURL(baseURL string) bool {
	u := strings.ToLower(strings.TrimSpace(baseURL))
	return strings.HasPrefix(u, "http://localhost") || strings.HasPrefix(u, "http://127.0.0.1")
}

// formatLimenUserID normalizes Limen user IDs. PG/UUID adapters often return
// []byte; fmt.Sprint([]byte) yields "[48 52 ...]" which breaks core auth context.
func formatLimenUserID(id any) string {
	switch v := id.(type) {
	case string:
		return v
	case []byte:
		if len(v) == 16 {
			if u, err := uuid.FromBytes(v); err == nil {
				return u.String()
			}
		}
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		s := fmt.Sprint(v)
		// Defensive: never return the Go []byte dump form.
		if strings.HasPrefix(s, "[") && strings.Contains(s, " ") {
			if b, ok := id.([]byte); ok {
				return string(b)
			}
		}
		return s
	}
}

func withCORS(next http.Handler, origins []string) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Cookie")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Expose-Headers", "Set-Auth-Token, Set-Refresh-Token")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
