package repository

import (
	"context"
	"fmt"
	"nexa/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
)

type UserAuthenticationTokenRepository struct {
	db     *pgx.Conn
	schema string
	table  string
}

func NewUserAuthenticationTokenRepository(db *pgx.Conn, schema, table string) *UserAuthenticationTokenRepository {
	return &UserAuthenticationTokenRepository{
		db:     db,
		schema: schema,
		table:  table,
	}
}

func (r *UserAuthenticationTokenRepository) tableFQN() string {
	return fmt.Sprintf("%s.%s", r.schema, r.table)
}

func (r *UserAuthenticationTokenRepository) FindTokenByUserID(userID string) (*model.UserAuthenticationToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf("SELECT id, user_id, code, expires_at, fails FROM %s WHERE user_id = $1 LIMIT 1", r.tableFQN())

	var token model.UserAuthenticationToken
	err := r.db.QueryRow(ctx, query, userID).Scan(&token.ID, &token.UserID, &token.Code, &token.ExpiresAt, &token.Fails)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get token by user_id: %w", err)
	}

	return &token, nil
}

func (r *UserAuthenticationTokenRepository) Insert(token *model.UserAuthenticationToken) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Inserimos os campos e retornamos o id gerado
	query := fmt.Sprintf("INSERT INTO %s (user_id, code, expires_at, fails) VALUES ($1, $2, $3, $4) RETURNING id", r.tableFQN())

	var id string
	err := r.db.QueryRow(ctx, query, token.UserID, token.Code, token.ExpiresAt, token.Fails).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert token: %w", err)
	}

	return id, nil
}

func (r *UserAuthenticationTokenRepository) IncrementFails(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf("UPDATE %s SET fails = fails + 1 WHERE id = $1", r.tableFQN())

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment fails: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no token found to increment fails for id %s", id)
	}

	return nil
}

func (r *UserAuthenticationTokenRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableFQN())

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no token found with id %s", id)
	}

	return nil
}
