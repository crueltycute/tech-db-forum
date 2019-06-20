package models

//easyjson:json
type Vote struct {

	// Идентификатор пользователя.
	// Required: true
	Nickname string `json:"nickname"`

	// Отданный голос.
	// Required: true
	// Enum: [-1 1]
	Voice int32 `json:"voice"`
}
