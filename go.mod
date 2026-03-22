module github.com/goritskimihail/mudro

go 1.24.0

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/websocket v1.5.3
	github.com/jackc/pgx/v5 v5.5.5
	github.com/minio/minio-go/v7 v7.0.70
	github.com/redis/go-redis/v9 v9.5.1
	github.com/segmentio/kafka-go v0.4.47
	golang.org/x/crypto v0.21.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rs/xid v1.6.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.21.0
	golang.org/x/net => golang.org/x/net v0.21.0
	golang.org/x/sync => golang.org/x/sync v0.6.0
	golang.org/x/sys => golang.org/x/sys v0.17.0
	golang.org/x/text => golang.org/x/text v0.14.0
)
