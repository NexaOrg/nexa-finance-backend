package model

type Category struct {
	ID	string	`json:"id,omitempty"`
	UserId	string `json:"user_id,omitempty"`
	Name	string 	`json:"name,omitempty"`
	Color	*string	`json:"color,omitempty"`
	Icon	*string	`json:"icon,omitempty"`
	MonthlyLimit	*float64 `json:"monthly-limit,omitempty"`
}

type CreateCategoryInput struct {
	Name         string  `json:"name"`
	Color        string  `json:"color"`
	Icon         string  `json:"icon"`
	MonthlyLimit float64 `json:"monthly_limit"`
}