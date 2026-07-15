package auth

import (
	"os"

	"github.com/puchidemy/puchi-backend/app/core/internal/auth/supertokens"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
)

// InitSupertokens initializes the Supertokens SDK from auth config.
// Call this in main.go before wireApp.
func InitSupertokens(cfg *conf.Auth) error {
	apiDomain := os.Getenv("SUPERTOKENS_API_DOMAIN")
	if apiDomain == "" {
		apiDomain = "http://localhost:8000"
	}
	stCfg := supertokens.Config{
		ConnectionURI: cfg.Supertokens.ConnectionUri,
		APIKey:        cfg.Supertokens.ApiKey,
		APIDomain:     apiDomain,
		CookieDomain:  cfg.Supertokens.CookieDomain,
		CookieSecure:  cfg.Supertokens.CookieSecure,
	}
	return supertokens.Init(stCfg)
}
