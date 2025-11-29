package handler

import (
	"encoding/json"
	"net/http"
	"nexa/internal/model"
	"nexa/internal/repository"
)

type CategoryHandler struct {
	repo *repository.CategoryRepository
}

func NewCategoryHandler(repo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		repo: repo,
	}

type CreateCategoryInput struct {
	Name	string	`json: name`
	Color	*string	`json: color`
	Icon	*string	`json: icon`
	MonthlyLimit	*float64 `json:"monthly_limit"`
}

func (h *CategoryHandler) CreateCategory(w http.Response, r *http.Request) {
	var input CreateCategoryInput
	if err := json.NewDecoder(r.body).decode(&input); err != nil {
		http.Error(w, "Formato de JSON inválido", http.StatusBadRequest)
		return
	}

	// Validação Básica
	if input.Name == "" {
		http.Error(w, "O nome da categoria é obrigatório", http.StatusBadRequest)
		return
	}

	// Pega ID do usuário
	userIDFromContext := r.Context().Value("userID")
	if userIDFromContext == nil {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDFromContext.(string)
	if !ok {
		http.Error(w, "Erro ao processar ID do usuário", http.StatusInternalServerError)
		return
	}
}
}
