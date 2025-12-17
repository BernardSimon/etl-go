package _type

type ResponseModel struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseWithData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
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
