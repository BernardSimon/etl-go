package api

import (
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/params"
)

func normalizeParams(list []params.Params) []params.Params {
	result := make([]params.Params, 0, len(list))
	for _, item := range list {
		if item.Type == "" {
			item.Type = inferParamType(item)
		}
		if item.Placeholder == "" {
			item.Placeholder = inferPlaceholder(item)
		}
		if item.Example == "" {
			item.Example = inferExample(item)
		}
		result = append(result, item)
	}
	return result
}

func inferParamType(item params.Params) string {
	key := strings.ToLower(item.Key)
	switch {
	case strings.Contains(key, "password"):
		return "password"
	case strings.Contains(key, "port"), strings.Contains(key, "limit"), strings.Contains(key, "size"), strings.Contains(key, "rows"):
		return "number"
	case strings.Contains(key, "query"), strings.Contains(key, "sql"), strings.Contains(key, "columns"), strings.Contains(key, "mapping"), strings.Contains(key, "rules"):
		return "textarea"
	case strings.Contains(key, "file"):
		return "file"
	default:
		return "text"
	}
}

func inferPlaceholder(item params.Params) string {
	if inferParamType(item) == "file" {
		return "Select an uploaded file"
	}
	if item.Description != "" {
		return item.Description
	}
	if item.DefaultValue != "" {
		return "Default: " + item.DefaultValue
	}
	return "Please input " + item.Key
}

func inferExample(item params.Params) string {
	key := strings.ToLower(item.Key)
	switch {
	case key == "port":
		return "3306"
	case key == "query":
		return "SELECT * FROM table_name LIMIT 100"
	case key == "delimiter":
		return ","
	case key == "file_ext":
		return "csv"
	case key == "columns":
		return `["id","name","created_at"]`
	case key == "operator":
		return "="
	case key == "value":
		return "100"
	case key == "keys_sample_rows":
		return "100"
	case key == "cron":
		return "0 */5 * * * *"
	case strings.Contains(key, "host"):
		return "127.0.0.1"
	case strings.Contains(key, "database"):
		return "etl"
	case strings.Contains(key, "user"):
		return "root"
	case strings.Contains(key, "password"):
		return "******"
	case strings.Contains(key, "file_id"):
		return ""
	default:
		return item.DefaultValue
	}
}
