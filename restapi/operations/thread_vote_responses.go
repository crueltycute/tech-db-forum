// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "tech-db-forum/models"
)

// ThreadVoteOKCode is the HTTP code returned for type ThreadVoteOK
const ThreadVoteOKCode int = 200

/*ThreadVoteOK Информация о ветке обсуждения.


swagger:response threadVoteOK
*/
type ThreadVoteOK struct {

	/*
	  In: Body
	*/
	Payload *models.Thread `json:"body,omitempty"`
}

// NewThreadVoteOK creates ThreadVoteOK with default headers values
func NewThreadVoteOK() *ThreadVoteOK {

	return &ThreadVoteOK{}
}

// WithPayload adds the payload to the thread vote o k response
func (o *ThreadVoteOK) WithPayload(payload *models.Thread) *ThreadVoteOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the thread vote o k response
func (o *ThreadVoteOK) SetPayload(payload *models.Thread) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ThreadVoteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ThreadVoteNotFoundCode is the HTTP code returned for type ThreadVoteNotFound
const ThreadVoteNotFoundCode int = 404

/*ThreadVoteNotFound Ветка обсуждения отсутсвует в форуме.


swagger:response threadVoteNotFound
*/
type ThreadVoteNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewThreadVoteNotFound creates ThreadVoteNotFound with default headers values
func NewThreadVoteNotFound() *ThreadVoteNotFound {

	return &ThreadVoteNotFound{}
}

// WithPayload adds the payload to the thread vote not found response
func (o *ThreadVoteNotFound) WithPayload(payload *models.Error) *ThreadVoteNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the thread vote not found response
func (o *ThreadVoteNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ThreadVoteNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
