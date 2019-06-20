package models

//easyjson:json
type Error struct {

	// Текстовое описание ошибки.
	// В процессе проверки API никаких проверок на содерижимое данного описание не делается.
	//
	// Read Only: true
	Message string `json:"message,omitempty"`
}
