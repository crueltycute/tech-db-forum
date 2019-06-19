package models

import (
	"github.com/go-openapi/strfmt"
)

//easyjson:json
type Thread struct {

	// Пользователь, создавший данную тему.
	// Required: true
	Author string `json:"author"`

	// Дата создания ветки на форуме.
	// Format: date-time
	Created *strfmt.DateTime `json:"created,omitempty"`

	// Форум, в котором расположена данная ветка обсуждения.
	// Read Only: true
	Forum string `json:"forum,omitempty"`

	// Идентификатор ветки обсуждения.
	// Read Only: true
	ID int32 `json:"id,omitempty"`

	// Описание ветки обсуждения.
	// Required: true
	Message string `json:"message"`

	// Человекопонятный URL (https://ru.wikipedia.org/wiki/%D0%A1%D0%B5%D0%BC%D0%B0%D0%BD%D1%82%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B8%D0%B9_URL).
	// В данной структуре slug опционален и не может быть числом.
	//
	// Read Only: true
	// Pattern: ^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$
	Slug string `json:"slug,omitempty"`

	// Заголовок ветки обсуждения.
	// Required: true
	Title string `json:"title"`

	// Кол-во голосов непосредственно за данное сообщение форума.
	// Read Only: true
	Votes int32 `json:"votes,omitempty"`
}
