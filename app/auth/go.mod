module github.com/puchidemy/puchi-backend/app/auth

go 1.26.3

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.12.3
	github.com/nats-io/nats.go v1.52.0
	github.com/thecodearcher/limen v0.1.4
	github.com/thecodearcher/limen/adapters/sql v0.1.2
	github.com/thecodearcher/limen/plugins/credential-password v0.1.4
	github.com/thecodearcher/limen/plugins/oauth v0.1.2
	github.com/thecodearcher/limen/plugins/oauth-facebook v0.1.2
	github.com/thecodearcher/limen/plugins/oauth-google v0.1.2
	golang.org/x/oauth2 v0.35.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)
