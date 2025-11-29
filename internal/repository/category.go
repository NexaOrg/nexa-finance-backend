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

	func (r *CategoryRepository) UpdateByID(id string, updateData map[string]interface{}) error {
	if len(updateData) == 0 {
		return fmt.Errorf("update data is empty")
	}

	// Verifica se a categoria existe antes de atualizar
	category, err := r.FindByFilter("id", id)
	if err != nil {
		return fmt.Errorf("failed to verify category before update: %w", err)
	}
	if category == nil {
		return fmt.Errorf("category not found")
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
		"UPDATE db_nexa.tb_category SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		i,
	)

	_, err = r.db.Exec(context.Background(), query, values...)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}
}