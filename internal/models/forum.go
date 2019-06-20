package models

//easyjson:json
type Forum struct {

	// Общее кол-во сообщений в данном форуме.
	//
	// Read Only: true
	Posts int64 `json:"posts,omitempty"`

	// Человекопонятный URL (https://ru.wikipedia.org/wiki/%D0%A1%D0%B5%D0%BC%D0%B0%D0%BD%D1%82%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B8%D0%B9_URL), уникальное поле.
	// Required: true
	// Pattern: ^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$
	Slug string `json:"slug"`

	// Общее кол-во ветвей обсуждения в данном форуме.
	//
	// Read Only: true
	Threads int32 `json:"threads,omitempty"`

	// Название форума.
	// Required: true
	Title string `json:"title"`

	// Nickname пользователя, который отвечает за форум.
	// Required: true
	User string `json:"user"`
}
