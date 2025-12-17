package etl

import (
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
	jsonSink "github.com/BernardSimon/etl-go/components/sinks/json"
	sqlSink "github.com/BernardSimon/etl-go/components/sinks/sql"
	csvSource "github.com/BernardSimon/etl-go/components/sources/csv"
	jsonSource "github.com/BernardSimon/etl-go/components/sources/json"
	sqlSource "github.com/BernardSimon/etl-go/components/sources/sql"
	sqlVariable "github.com/BernardSimon/etl-go/components/variable/sql"
	"github.com/BernardSimon/etl-go/etl/factory"
)

func init() {
	// 注册数据源
	factory.RegisterDataSource(dorisDatasource.DatasourceCreator)
	factory.RegisterDataSource(mysqlDatasource.DatasourceCreator)
	factory.RegisterDataSource(postgreDatasource.DatasourceCreator)
	factory.RegisterDataSource(sqliteDatasource.DatasourceCreator)

	//注册变量执行器
	factory.RegisterVariable(sqlVariable.VariableCreatorMysql)
	factory.RegisterVariable(sqlVariable.VariableCreatorPostgre)
	factory.RegisterVariable(sqlVariable.VariableCreatorSqlite)

	//注册执行器
	factory.RegisterExecutor(sqlExecutor.ExecutorCreatorMysql)
	factory.RegisterExecutor(sqlExecutor.ExecutorCreatorPostgre)
	factory.RegisterExecutor(sqlExecutor.ExecutorCreatorSqlite)

	//注册数据输入
	factory.RegisterSource(sqlSource.SourceCreatorMysql)
	factory.RegisterSource(sqlSource.SourceCreatorPostgre)
	factory.RegisterSource(csvSource.SourceCreator)
	factory.RegisterSource(jsonSource.SourceCreator)
	factory.RegisterSource(sqlSource.SourceCreatorSqlite)

	//注册数据输出
	factory.RegisterSink(sqlSink.SinkCreatorMysql)
	factory.RegisterSink(sqlSink.SinkCreatorPostgre)
	factory.RegisterSink(csvSink.SinkCreator)
	factory.RegisterSink(jsonSink.SinkCreator)
	factory.RegisterSink(dorisSink.SinkCreator)
	factory.RegisterSink(sqlSink.SinkCreatorSqlite)

	//注册处理器
	factory.RegisterProcessor(convertTypeProcessor.ProcessorCreator)
	factory.RegisterProcessor(filterRowsProcessor.ProcessorCreator)
	factory.RegisterProcessor(maskDataProcessor.ProcessorCreator)
	factory.RegisterProcessor(renameColumnProcessor.ProcessorCreator)
	factory.RegisterProcessor(selectColumnsProcessor.ProcessorCreator)

}
