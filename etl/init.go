package etl

import (
	"errors"

	dorisDatasource "github.com/BernardSimon/etl-go/components/datasource/doris"
	mysqlDatasource "github.com/BernardSimon/etl-go/components/datasource/mysql"
	postgreDatasource "github.com/BernardSimon/etl-go/components/datasource/postgre"
	sqliteDatasource "github.com/BernardSimon/etl-go/components/datasource/sqlite"
	sqlExecutor "github.com/BernardSimon/etl-go/components/executor/sql"
	convertTypeProcessor "github.com/BernardSimon/etl-go/components/processors/convertType"
	filterRowsProcessor "github.com/BernardSimon/etl-go/components/processors/filterRows"
	maskDataProcessor "github.com/BernardSimon/etl-go/components/processors/maskData"
	renameColumnProcessor "github.com/BernardSimon/etl-go/components/processors/renameColumn"
	selectColumnsProcessor "github.com/BernardSimon/etl-go/components/processors/selectColumns"
	csvSink "github.com/BernardSimon/etl-go/components/sinks/csv"
	dorisSink "github.com/BernardSimon/etl-go/components/sinks/doris"
	httpSink "github.com/BernardSimon/etl-go/components/sinks/http"
	jsonSink "github.com/BernardSimon/etl-go/components/sinks/json"
	sqlSink "github.com/BernardSimon/etl-go/components/sinks/sql"
	csvSource "github.com/BernardSimon/etl-go/components/sources/csv"
	httpSource "github.com/BernardSimon/etl-go/components/sources/http"
	jsonSource "github.com/BernardSimon/etl-go/components/sources/json"
	sqlSource "github.com/BernardSimon/etl-go/components/sources/sql"
	sqlVariable "github.com/BernardSimon/etl-go/components/variable/sql"
	"github.com/BernardSimon/etl-go/etl/factory"
	"go.uber.org/zap"
)

func init() {
	if err := RegisterComponents(); err != nil {
		zap.L().Fatal("Failed to register ETL components", zap.Error(err))
	}
}

// RegisterComponents registers all built-in ETL components.
// Returns the first registration error encountered.
func RegisterComponents() error {
	var errs []error

	// 注册数据源 (must be first, as other components depend on them)
	errs = append(errs, factory.RegisterDataSource(dorisDatasource.DatasourceCreator))
	errs = append(errs, factory.RegisterDataSource(mysqlDatasource.DatasourceCreator))
	errs = append(errs, factory.RegisterDataSource(postgreDatasource.DatasourceCreator))
	errs = append(errs, factory.RegisterDataSource(sqliteDatasource.DatasourceCreator))

	// 注册变量执行器
	errs = append(errs, factory.RegisterVariable(sqlVariable.VariableCreatorMysql))
	errs = append(errs, factory.RegisterVariable(sqlVariable.VariableCreatorPostgre))
	errs = append(errs, factory.RegisterVariable(sqlVariable.VariableCreatorSqlite))

	// 注册执行器
	errs = append(errs, factory.RegisterExecutor(sqlExecutor.ExecutorCreatorMysql))
	errs = append(errs, factory.RegisterExecutor(sqlExecutor.ExecutorCreatorPostgre))
	errs = append(errs, factory.RegisterExecutor(sqlExecutor.ExecutorCreatorSqlite))

	// 注册数据输入
	errs = append(errs, factory.RegisterSource(sqlSource.SourceCreatorMysql))
	errs = append(errs, factory.RegisterSource(sqlSource.SourceCreatorPostgre))
	errs = append(errs, factory.RegisterSource(csvSource.SourceCreator))
	errs = append(errs, factory.RegisterSource(jsonSource.SourceCreator))
	errs = append(errs, factory.RegisterSource(httpSource.SourceCreator))
	errs = append(errs, factory.RegisterSource(sqlSource.SourceCreatorSqlite))
	errs = append(errs, factory.RegisterSource(sqlSource.SourceCreatorDoris))

	// 注册数据输出
	errs = append(errs, factory.RegisterSink(sqlSink.SinkCreatorMysql))
	errs = append(errs, factory.RegisterSink(sqlSink.SinkCreatorPostgre))
	errs = append(errs, factory.RegisterSink(csvSink.SinkCreator))
	errs = append(errs, factory.RegisterSink(jsonSink.SinkCreator))
	errs = append(errs, factory.RegisterSink(dorisSink.SinkCreator))
	errs = append(errs, factory.RegisterSink(httpSink.SinkCreator))
	errs = append(errs, factory.RegisterSink(sqlSink.SinkCreatorSqlite))

	// 注册处理器
	errs = append(errs, factory.RegisterProcessor(convertTypeProcessor.ProcessorCreator))
	errs = append(errs, factory.RegisterProcessor(filterRowsProcessor.ProcessorCreator))
	errs = append(errs, factory.RegisterProcessor(maskDataProcessor.ProcessorCreator))
	errs = append(errs, factory.RegisterProcessor(renameColumnProcessor.ProcessorCreator))
	errs = append(errs, factory.RegisterProcessor(selectColumnsProcessor.ProcessorCreator))

	return errors.Join(errs...)
}
