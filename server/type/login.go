package _type

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginWithCodeRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
