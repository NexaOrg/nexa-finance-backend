package factory

import (
	"time"

	"synap/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAuthenticationTokenFactory struct{}

func NewUserAuthenticationTokenFactory() *UserAuthenticationTokenFactory {
	return &UserAuthenticationTokenFactory{}
}

func (*UserAuthenticationTokenFactory) CreateUserAuthenticationToken(userID primitive.ObjectID, token string, duration int) *model.UserAuthenticationToken {
	expiresAt := time.Now().Add(time.Minute * time.Duration(duration))

	return &model.UserAuthenticationToken{
		UserID:    userID,
		Code:      token,
		Fails:     0,
		ExpiresAt: expiresAt,
	}
}
