package repository

import (
	"context"
	"fmt"
	"nexa/internal/model"
	"strings"

	"github.com/jackc/pgx/v5"
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

	validKeys := map[string]bool{
		"id":       true,
		"email":    true,
		"username": true,
		"name":     true,
	}

	if !validKeys[key] {
		return nil, fmt.Errorf("invalid filter key: %s", key)
	}

	query := fmt.Sprintf(`
		SELECT id, name, username, email, password, photo_url, score, created_at, last_login, is_active
		FROM db_nexa.tb_user 
		WHERE %s = $1 
		LIMIT 1
	`, key)

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
		&user.IsActive,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

func (u *UserRepository) UpdateByID(id string, updateData map[string]interface{}) error {
	if len(updateData) == 0 {
		return fmt.Errorf("update data is empty")
	}

	setClauses := []string{}
	values := []interface{}{}
	i := 1

	for column, value := range updateData {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, i))
		values = append(values, value)
		i++
	}

	values = append(values, id)

	query := fmt.Sprintf(
		"UPDATE db_nexa.tb_user SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		i,
	)

	_, err := u.db.Exec(context.Background(), query, values...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
