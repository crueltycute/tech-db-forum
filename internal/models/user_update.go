package models

import (
	"github.com/go-openapi/strfmt"
)

//easyjson:json
type UserUpdate struct {

	// Описание пользователя.
	About string `json:"about,omitempty"`

	// Почтовый адрес пользователя (уникальное поле).
	// Format: email
	Email strfmt.Email `json:"email,omitempty"`

	// Полное имя пользователя.
	Fullname string `json:"fullname,omitempty"`
}
