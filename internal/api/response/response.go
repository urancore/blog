package response

type BaseResponse struct {
	Status any `json:"status"`
	Error string `json:"error,omitempty"`
}

func Error(status any, msg string) BaseResponse {
	return BaseResponse{
		Status: status,
		Error: msg,
	}
}
