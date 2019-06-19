package service

import (
	"github.com/crueltycute/tech-db-forum/internal"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"net/http"
)

func Clear(res http.ResponseWriter, req *http.Request) {
	db.QueryRow(queryClearDB)
	return operations.NewClearOK()
}

func Status(res http.ResponseWriter, req *http.Request) {
	status := &internal.Status{}
	err := db.QueryRow(queryGetStatus).Scan(&status.User, &status.Forum, &status.Thread, &status.Post)
	if err != nil {
		panic(err)
	}
	return operations.NewStatusOK().WithPayload(status)
}
