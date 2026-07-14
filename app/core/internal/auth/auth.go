package auth

import (
	"github.com/puchidemy/puchi-backend/app/core/internal/auth/supertokens"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
)

// InitSupertokens initializes the Supertokens SDK from auth config.
// Call this in main.go before wireApp.
func InitSupertokens(cfg *conf.Auth) error {
	stCfg := supertokens.Config{
		ConnectionURI: cfg.Supertokens.ConnectionUri,
		APIKey:        cfg.Supertokens.ApiKey,
		CookieDomain:  cfg.Supertokens.CookieDomain,
		CookieSecure:  cfg.Supertokens.CookieSecure,
	}
	return supertokens.Init(stCfg)
}
