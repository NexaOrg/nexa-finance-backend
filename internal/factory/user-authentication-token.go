package factory

import (
	"nexa/internal/model"
	"time"
)

type UserAuthenticationTokenFactory struct{}

func NewUserAuthenticationTokenFactory() *UserAuthenticationTokenFactory {
	return &UserAuthenticationTokenFactory{}
}

func (f *UserAuthenticationTokenFactory) CreateUserAuthenticationToken(userID, code string, ttlMinutes int) *model.UserAuthenticationToken {
	return &model.UserAuthenticationToken{
		UserID:    userID,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Duration(ttlMinutes) * time.Minute),
		Fails:     0,
	}
}
