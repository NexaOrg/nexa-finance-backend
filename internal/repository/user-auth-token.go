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

type UserAuthenticationTokenRepository struct {
	collection *mongo.Collection
}

func NewUserAuthenticationTokenRepository(client *mongo.Client, dbName, collectionName string) *UserAuthenticationTokenRepository {
	return &UserAuthenticationTokenRepository{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

func (ua *UserAuthenticationTokenRepository) FindTokenByUserID(userID primitive.ObjectID) (*model.UserAuthenticationToken, error) {
	var userAuthenticationToken model.UserAuthenticationToken

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{"userID": userID}

	result := ua.collection.FindOne(ctx, filter)

	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get: %w", err)
	}

	if err := result.Decode(&userAuthenticationToken); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return &userAuthenticationToken, nil
}

func (ua *UserAuthenticationTokenRepository) IncrementFails(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$inc": bson.M{"fails": 1}}

	result := ua.collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		return fmt.Errorf("failed to increment authentication fail: %s", result.Err())
	}

	return nil
}

func (ua *UserAuthenticationTokenRepository) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	result, err := ua.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete: %s", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no validation found with ID %s", id.Hex())
	}

	return nil
}

func (ua *UserAuthenticationTokenRepository) Insert(UserAuthenticationToken *model.UserAuthenticationToken) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	result, err := ua.collection.InsertOne(ctx, UserAuthenticationToken)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to insert: %w", err)
	}

	return result.InsertedID.(primitive.ObjectID), nil
}
