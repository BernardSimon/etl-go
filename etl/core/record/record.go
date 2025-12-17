package record

// Record 是管道中数据流的基本单元。
// 它将一行数据抽象为一个从列名到值的映射。
// 使用 map[string]interface{} 允许灵活处理各种数据源的记录，无论是来自数据库、CSV 文件还是 JSON API。
type Record map[string]interface{}
