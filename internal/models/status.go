package models

//easyjson:json
type Status struct {

	// Кол-во разделов в базе данных.
	// Required: true
	Forum int32 `json:"forum"`

	// Кол-во сообщений в базе данных.
	// Required: true
	Post int64 `json:"post"`

	// Кол-во веток обсуждения в базе данных.
	// Required: true
	Thread int32 `json:"thread"`

	// Кол-во пользователей в базе данных.
	// Required: true
	User int32 `json:"user"`
}
