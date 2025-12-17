module github.com/BernardSimon/etl-go/etl/components/datasource/postgre

replace github.com/BernardSimon/etl-go/etl/core => ../../../etl/core

go 1.24.4

require (
	github.com/BernardSimon/etl-go/etl/core v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
)
