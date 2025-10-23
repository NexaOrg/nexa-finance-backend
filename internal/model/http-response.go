package model

import (
	"time"
)

type UserRegisterSuccessfulResponse struct {
	StatusCode int    `json:"status" bson:"status"`
	IDUser     string `json:"idUser,omitempty" bson:"idUser,omitempty"`
	Message    string `json:"message" bson:"message"`
}

type ErrorResponse struct {
	StatusCode int       `json:"status" bson:"status"`
	Error      string    `json:"error" bson:"error"`
	Message    string    `json:"message" bson:"message"`
	Timestamp  time.Time `json:"timestamp" bson:"timestamp"`
	Path       string    `json:"path" bson:"path"`
	Input      string    `json:"input,omitempty" bson:"input"`
}
