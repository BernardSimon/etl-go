module github.com/BernardSimon/etl-go/components/sinks/redis

replace (
	github.com/BernardSimon/etl-go/components/datasource/redis => ../../datasource/redis
	github.com/BernardSimon/etl-go/etl/core => ../../../etl/core
)

go 1.24.4

require (
	github.com/BernardSimon/etl-go/components/datasource/redis v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/etl/core v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.7.3
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)
