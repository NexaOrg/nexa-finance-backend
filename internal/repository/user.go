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

	// ⚠️ Proteção contra SQL Injection
	validKeys := map[string]bool{
		"id":       true,
		"email":    true,
		"username": true,
		"name":     true,
	}

	if !validKeys[key] {
		return nil, fmt.Errorf("invalid filter key: %s", key)
	}

	query := fmt.Sprintf(`SELECT id, name, username, email, password, photo_url, score, created_at, last_login 
		FROM db_nexa.tb_user WHERE %s = $1 LIMIT 1`, key)

	err := u.db.QueryRow(context.Background(), query, value).Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.PhotoUrl,
		&user.Score,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

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
