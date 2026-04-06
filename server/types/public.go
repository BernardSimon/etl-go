package types

type ResponseModel struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseWithData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorData struct {
	Errors []FieldError `json:"errors,omitempty"`
}

type ServiceError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *ServiceError) Error() string {
	return e.Message
}

// IDUri 用于只含路径参数 :id 的接口（如删除、启动、停止等）
type IDUri struct {
	Id string `uri:"id"`
}

type Mission struct {
	Spec string
	Func func() error
}

type RequestLog struct {
	Method   string            `json:"method"`
	Ip       string            `json:"ip"`
	Path     string            `json:"path"`
	Body     string            `json:"body"`
	Headers  map[string]string `json:"headers"`
	Response *ResponseModel    `json:"response"`
}

type MissionLog struct {
	Name string
	Spec string
}
