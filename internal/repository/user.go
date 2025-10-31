package repository

import (
	"context"
	"fmt"
	"nexa/internal/model"

	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository struct {
	db *pgx.Conn
}

func NewUserRepository(conn *pgx.Conn) *UserRepository {
	return &UserRepository{
		db: conn,
	}
}

func (b *UserRepository) InsertUser(user *model.User) error {
	_, err := b.db.Exec(context.Background(), "INSERT INTO db_nexa.tb_user (name, username, email, password, photo_url, last_login) VALUES ($1, $2, $3, $4, $5, $6)",
		user.Name, user.Username, user.Email, user.Password, user.PhotoUrl, user.LastLogin)

	return err
}

func (u *UserRepository) FindByFilter(key string, value any) (*model.User, error) {
	var user model.User

	// ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	// defer cancel()

	// filter := bson.M{key: value}
	// result := u.collection.FindOne(ctx, filter)

	// if err := result.Err(); err != nil {
	// 	if err == mongo.ErrNoDocuments {
	// 		return nil, nil // Usuário não encontrado não é mais tratado como erro
	// 	}
	// 	return nil, fmt.Errorf("failed to find user: %w", err)
	// }

	// if err := result.Decode(&user); err != nil {
	// 	return nil, fmt.Errorf("failed to decode user: %w", err)
	// }

	return &user, nil
}

func (u *UserRepository) UpdateByID(id primitive.ObjectID, updateData map[string]interface{}) error {
	// ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	// defer cancel()

	// filter := bson.M{"_id": id}

	// res := u.collection.FindOneAndUpdate(
	// 	ctx,
	// 	filter,
	// 	bson.D{{Key: "$set", Value: updateData}},
	// )

	// return res.Err()

	return fmt.Errorf("")
}
