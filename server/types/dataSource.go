package types

import params2 "github.com/BernardSimon/etl-go/etl/core/params"

type NewDataSourceRequest struct {
	ID   string    `json:"id"`
	Name string    `json:"name" binding:"required"`
	Type string    `json:"type" binding:"required"`
	Data KeyValues `json:"data" binding:"required"`
	Edit bool      `json:"edit"`
}

type TestDataSourceRequest struct {
	Type string    `json:"type" binding:"required"`
	Data KeyValues `json:"data" binding:"required"`
}

type GetDataSourceTypeListResponse struct {
	Type   string           `json:"type"`
	Params []params2.Params `json:"params"`
}

// ── Schema 发现 ───────────────────────────────────────────────────────────────

type GetDataSourceSchemaResponse struct {
	Tables []DataSourceTable `json:"tables"`
}

type DataSourceTable struct {
	Name    string                `json:"name"`
	Columns []DataSourceColumn    `json:"columns"`
}

type DataSourceColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}
