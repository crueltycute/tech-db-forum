package models

//easyjson:json
type PostUpdate struct {

	// Собственно сообщение форума.
	Message string `json:"message,omitempty"`
}
