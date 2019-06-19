package models

//easyjson:json
type PostFull struct {
	// author
	Author *User `json:"author,omitempty"`

	// forum
	Forum *Forum `json:"forum,omitempty"`

	// post
	Post *Post `json:"post,omitempty"`

	// thread
	Thread *Thread `json:"thread,omitempty"`
}
