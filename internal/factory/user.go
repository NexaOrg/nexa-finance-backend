package factory

import (
	"nexa/internal/model"
)

type UserFactory struct{}

func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

func (userF *UserFactory) CreateUser(name, email, password string) *model.User {
	return &model.User{
		// IDUser:   primitive.NewObjectID(),
		// Name:     name,
		// Email:    email,
		// Password: password,
	}
}
