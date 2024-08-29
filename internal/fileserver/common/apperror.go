package common

type AppError struct {
	Error   error `json:"error"`
	Message string `json:"msg"`
	Code    int `json:"code"`
	Status  bool `json:"status"`
}
