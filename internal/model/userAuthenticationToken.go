package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAuthenticationToken struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
	Code      string             `json:"code" bson:"code"`
	ExpiresAt time.Time          `json:"expiresAt" bson:"expiresAt"`
	Fails     int                `json:"fails" bson:"fails"`
}

func (t *UserAuthenticationToken) HasExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}
