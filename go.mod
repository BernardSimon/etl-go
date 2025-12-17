module github.com/BernardSimon/etl-go

go 1.24.4

require (
	github.com/BernardSimon/etl-go/components/datasource/doris v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/datasource/mysql v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/datasource/postgre v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/executor/sql v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/processors/convertType v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/processors/filterRows v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/processors/maskData v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/processors/renameColumn v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/processors/selectColumns v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/sinks/csv v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/sinks/doris v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/sinks/json v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/sinks/sql v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/sources/csv v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/sources/json v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/sources/sql v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/components/variable/sql v0.0.0-00010101000000-000000000000
	github.com/BernardSimon/etl-go/etl/core v0.0.0-00010101000000-000000000000
	github.com/gin-contrib/cors v1.7.6
	github.com/gin-gonic/gin v1.11.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/nicksnyder/go-i18n/v2 v2.6.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.9.3
	go.uber.org/zap v1.27.1
	golang.org/x/text v0.32.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.2 // indirect
	github.com/bytedance/sonic/loader v0.4.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/gabriel-vasile/mimetype v1.4.11 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.28.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.19.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.57.1 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.1 // indirect
	go.uber.org/mock v0.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.23.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace (
	github.com/BernardSimon/etl-go/components/datasource/doris => ./components/datasource/doris
	github.com/BernardSimon/etl-go/components/datasource/mysql => ./components/datasource/mysql
	github.com/BernardSimon/etl-go/components/datasource/postgre => ./components/datasource/postgre
	github.com/BernardSimon/etl-go/components/executor/sql => ./components/executor/sql
	github.com/BernardSimon/etl-go/components/processors/convertType => ./components/processors/convertType
	github.com/BernardSimon/etl-go/components/processors/filterRows => ./components/processors/filterRows
	github.com/BernardSimon/etl-go/components/processors/maskData => ./components/processors/maskData
	github.com/BernardSimon/etl-go/components/processors/renameColumn => ./components/processors/renameColumn
	github.com/BernardSimon/etl-go/components/processors/selectColumns => ./components/processors/selectColumns
	github.com/BernardSimon/etl-go/components/sinks/csv => ./components/sinks/csv
	github.com/BernardSimon/etl-go/components/sinks/doris => ./components/sinks/doris
	github.com/BernardSimon/etl-go/components/sinks/json => ./components/sinks/json
	github.com/BernardSimon/etl-go/components/sinks/sql => ./components/sinks/sql
	github.com/BernardSimon/etl-go/components/sources/csv => ./components/sources/csv
	github.com/BernardSimon/etl-go/components/sources/json => ./components/sources/json
	github.com/BernardSimon/etl-go/components/sources/sql => ./components/sources/sql
	github.com/BernardSimon/etl-go/components/variable/sql => ./components/variable/sql
	github.com/BernardSimon/etl-go/etl/core => ./etl/core
)
