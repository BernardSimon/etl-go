module github.com/BernardSimon/etl-go/components/datasource/kafka

replace github.com/BernardSimon/etl-go/etl/core => ../../../etl/core

go 1.24.4

require (
	github.com/BernardSimon/etl-go/etl/core v0.0.0-00010101000000-000000000000
	github.com/segmentio/kafka-go v0.4.47
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	golang.org/x/text v0.13.0 // indirect
)
