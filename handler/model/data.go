package model

// 基础模型
type (
	BasicResp struct {
		Status  bool        `json:"status"`
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
)
