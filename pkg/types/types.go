package types

import (
	"github.com/go-playground/validator/v10"
	"time"
)

type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	Version   int       `json:"version"`
	AccountID string    `json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ItemCreateInput struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value"`
}

type UpdateItemInput struct {
	ID    string `json:"-"`
	Name  string `json:"name" validate:"required"`
	Value string `json:"value"`
}

type Sort string

const (
	ASC  Sort = "asc"
	DESC Sort = "desc"
)

func IsValidSort(fl validator.FieldLevel) bool {
	sort := fl.Field().String()
	if sort == "" {
		// Allow empty value
		return true
	}
	return sort == string(ASC) || sort == string(DESC)
}

type OrderBy string

const (
	OrderByName      OrderBy = "name"
	OrderByCreatedAt OrderBy = "created_at"
	OrderByUpdatedAt OrderBy = "updated_at"
)

func IsValidOrderBy(fl validator.FieldLevel) bool {
	ob := fl.Field().String()
	if ob == "" {
		// Allow empty value
		return true
	}
	return ob == string(OrderByName) || ob == string(OrderByCreatedAt) || ob == string(OrderByUpdatedAt)
}

type ListItemsInput struct {
	Sort    Sort    `json:"sort" validate:"is_valid_sort"`
	OrderBy OrderBy `json:"order_by" validate:"is_valid_orderby"`
	Limit   int     `json:"limit"`
}

type User struct {
	ID        string
	Name      string
	Email     string
	AccountID string
	Token     string
}
