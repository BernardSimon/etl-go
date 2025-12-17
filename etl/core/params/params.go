package params

type Params struct {
	Key          string `json:"key"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"defaultValue"`
	Description  string `json:"description"`
}
