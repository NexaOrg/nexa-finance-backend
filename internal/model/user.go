package model

import "time"

type User struct {
	ID        int       `json:"id,omitempty"`
	Username  string    `json:"username,omitempty"`
	FirstName string    `json:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	PhotoUrl  string    `json:"photoUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	LastLogin time.Time `json:"lastLogin"`
}
