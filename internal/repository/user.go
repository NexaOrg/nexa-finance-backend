package repository

import (
	"context"
	"fmt"
	"synap/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(client *mongo.Client, dbName, collectionName string) *UserRepository {
	return &UserRepository{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

func (b *UserRepository) InsertUser(user *model.User) error {
	_, err := b.collection.InsertOne(context.TODO(), user)
	return err
}

// MÉTODO MODIFICADO AQUI
func (u *UserRepository) FindByFilter(key string, value interface{}) (*model.User, error) {
	var user model.User

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{key: value}
	result := u.collection.FindOne(ctx, filter)

	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Usuário não encontrado não é mais tratado como erro
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if err := result.Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

func (u *UserRepository) UpdateByID(id primitive.ObjectID, updateData map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	res := u.collection.FindOneAndUpdate(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateData}},
	)

	return res.Err()
}
