package pipeline

import (
	"context"
	"sort"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

// PreviewSink 是用于预览的内存 Sink，不向任何外部系统写入数据。
// 它收集至多 Limit 条记录，供预览接口返回给前端。
type PreviewSink struct {
	Limit   int
	Records []record.Record
	Columns []string // 按顺序排列的列名
	count   int
}

func (s *PreviewSink) Open(_ context.Context, _ map[string]string, columnMapping map[string]string, _ datasource.Datasource) error {
	// 按列名排序，保证顺序稳定
	keys := make([]string, 0, len(columnMapping))
	for k := range columnMapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s.Columns = keys
	return nil
}

func (s *PreviewSink) Write(_ context.Context, _ string, records []record.Record) error {
	for _, r := range records {
		if s.Limit > 0 && s.count >= s.Limit {
			break
		}
		s.Records = append(s.Records, r)
		s.count++
	}
	return nil
}

func (s *PreviewSink) Close() error { return nil }
