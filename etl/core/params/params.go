package params

type Params struct {
	Key          string `json:"key"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"defaultValue"`
	Description  string `json:"description"`
	Placeholder  string `json:"placeholder,omitempty"`
	Example      string `json:"example,omitempty"`
	Type         string `json:"type,omitempty"`
}
