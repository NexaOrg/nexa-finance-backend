package repository

import (
	"context"
	"fmt"
	"nexa/internal/model"
	"strings"
	"github.com/jackc/pgx/v5"
)

type CategoryRepository struct {
	db *pgx.Conn
}

func NewCategoryRepository(conn *pgx.Conn) *CategoryRepository {
	return &CategoryRepository{
		db: conn,
	}
}

func (r *CategoryRepository) InsertCategory(category *model.Category) error {
	_, err := r.db.Exec(context.Background(), "INSERT INTO db_nexa.tb_category (user_id, name, color, icon, monthly_limit) VALUES ($1, $2, $3, $4, $5)",
		category.UserID, category.Name, category.Color, category.Icon, category.MonthlyLimit)

		return err
}

func (r *CategoryRepository) FindByFilter(key string, value any) (*model.Category, error) {
	var category model.Category

	validKeys := map[string]bool {
		"id":	true,
		"user_id": true,
		"name": true,
	}

	if !validKeys[key] {
		return nil, fmt.Errorf("invalid filter key: %s", key)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, name, color, icon, monthly_limit
		FROM db_nexa.tb_category 
		WHERE %s = $1 
		LIMIT 1
	`, key)

	err := r.db.QueryRow(context.Background(), query, value).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.Color,
		&category.Icon,
		&category.MonthlyLimit,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find category: %w", err)
	}

	return &category, nil
}