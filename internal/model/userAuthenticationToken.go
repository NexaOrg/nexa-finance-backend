package model

import "time"

type UserAuthenticationToken struct {
	ID        string    `json:"id,omitempty"`
	UserID    string    `json:"userID"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
	Fails     int       `json:"fails"`
}

func (t *UserAuthenticationToken) HasExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}
