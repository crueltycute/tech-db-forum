package models

import (
	"time"
)

//easyjson:json
type Post struct {

	// Автор, написавший данное сообщение.
	// Required: true
	Author string `json:"author"`

	// Дата создания сообщения на форуме.
	// Read Only: true
	// Format: date-time
	Created time.Time `json:"created,omitempty"`

	// Идентификатор форума (slug) данного сообещния.
	// Read Only: true
	Forum string `json:"forum,omitempty"`

	// Идентификатор данного сообщения.
	// Read Only: true
	ID int64 `json:"id,omitempty"`

	// Истина, если данное сообщение было изменено.
	// Read Only: true
	IsEdited bool `json:"isEdited,omitempty"`

	// Собственно сообщение форума.
	// Required: true
	Message string `json:"message"`

	// Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	//
	Parent int64 `json:"parent,omitempty"`

	// Идентификатор ветви (id) обсуждения данного сообещния.
	// Read Only: true
	Thread int32 `json:"thread,omitempty"`
}
