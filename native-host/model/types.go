package model

// RequestMessage 从 Extension 收到的请求消息
type RequestMessage struct {
	ID     string                 `json:"id"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

// ResponseMessage 发送给 Extension 的响应消息
type ResponseMessage struct {
	ID     string      `json:"id"`
	Success bool       `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// APIResponse HTTP API 响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
