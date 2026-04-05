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

type DeleteDataSourceRequest struct {
	Id string `json:"id" uri:"id" binding:"required"`
}

type GetDatasourceParamsByTypeRequest struct {
	Type string `json:"type" binding:"required"`
}
type GetDataSourceTypeListResponse struct {
	Type   string           `json:"type"`
	Params []params2.Params `json:"params"`
}
