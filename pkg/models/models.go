package models

type Response struct {
	Message    string `json:"message"`
	HttpStatus int    `json:"http_status"`
	Success    bool   `json:"success"`
	Data       any    `json:"data"`
}

type User struct {
	ID string `json:"id" binding:"required,alphanum"`
}

type Message struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id" binding:"required,uuid"`
	ServerID string `json:"server_id" binding:"required"`
	Message  string `json:"message" binding:"required"`
}
