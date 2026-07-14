package supertokens

import (
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
)

type Config struct {
	ConnectionURI string
	APIKey        string
	CookieDomain  string
	CookieSecure  bool
}

func Init(cfg Config) error {
	apiBasePath := "/auth"
	websiteBasePath := "/auth"

	return supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: cfg.ConnectionURI,
			APIKey:        cfg.APIKey,
		},
		AppInfo: supertokens.AppInfo{
			AppName:         "puchi",
			APIDomain:       "http://localhost:8000",
			WebsiteDomain:   cfg.CookieDomain,
			APIBasePath:     &apiBasePath,
			WebsiteBasePath: &websiteBasePath,
		},
		RecipeList: []supertokens.Recipe{
			session.Init(&sessmodels.TypeInput{}),
		},
	})
}
