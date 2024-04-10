package models

type Response struct {
	Message    string `json:"message"`
	HttpStatus int    `json:"http_status"`
	Success    bool   `json:"success"`
}
