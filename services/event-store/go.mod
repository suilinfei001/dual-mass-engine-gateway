module github.com/quality-gateway/event-store

go 1.24.11

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/google/uuid v1.6.0
	github.com/quality-gateway/shared v1.0.0
)

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/quality-gateway/shared => ../../shared
