module arnobot-core

go 1.24.2

replace arnobot-shared => ../shared

require (
	arnobot-shared v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.7.5
	github.com/nats-io/nats.go v1.42.0
)

require (
	github.com/golang-jwt/jwt/v4 v4.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nicklaw5/helix/v2 v2.31.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)
