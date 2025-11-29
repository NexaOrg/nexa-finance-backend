package factory

import (
	"nexa/internal/handler"
	"nexa/internal/repository"

	"github.com/jackc/pgc/v5"
)

func buildCategoryHandler(db *pgx.Conn) *handler.CategoryHandler {
	repo := repository.NewCategoryRepository(db)
	categoryHandler := handler.NewCategoryHandler(repo)
	return categoryHandler
}