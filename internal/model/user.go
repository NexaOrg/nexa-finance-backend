package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	IDUser    primitive.ObjectID `json:"idUser,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty"`
	Nickname  string             `json:"nickname,omitempty" bson:"nickname,omitempty"`
	Email     string             `json:"email,omitempty" bson:"email,omitempty"`
	Password  string             `json:"password,omitempty" bson:"password,omitempty"`
	IsPremium bool               `json:"isPremium,omitempty" bson:"isPremium,omitempy"`
	IsActive  bool               `json:"isActive,omitempty" bson:"isActive,omitempy"`
	Bio       string             `json:"bio,omitempty" bson:"bio,omitempty"`
	PhotoUrl  string             `json:"photoUrl,omitempty" bson:"photoUrl,omitempty"`
	Banner    string             `json:"banner,omitempty" bson:"banner,omitempty"`
}
