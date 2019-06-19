package models

//easyjson:json
type ThreadUpdate struct {

	// Описание ветки обсуждения.
	Message string `json:"message,omitempty"`

	// Заголовок ветки обсуждения.
	Title string `json:"title,omitempty"`
}
