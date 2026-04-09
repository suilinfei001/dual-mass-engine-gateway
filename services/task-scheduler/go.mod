module github.com/quality-gateway/task-scheduler

go 1.24.11

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/quality-gateway/shared v0.0.0
)

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/quality-gateway/shared => ../../shared
