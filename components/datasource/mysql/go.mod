module github.com/BernardSimon/etl-go/components/datasource/mysql

replace github.com/BernardSimon/etl-go/etl/core => ../../../etl/core

go 1.24.4

require (
	github.com/BernardSimon/etl-go/etl/core v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.9.3
)

require filippo.io/edwards25519 v1.1.0 // indirect
