package service

import (
	"database/sql"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
)

func Clear(db *sql.DB, params operations.ClearParams) middleware.Responder {
	db.QueryRow(queryClearDB)
	return operations.NewClearOK()
}


func Status(db *sql.DB, params operations.StatusParams) middleware.Responder {
	status := &models.Status{}
	err := db.QueryRow(queryGetStatus).Scan(&status.User, &status.Forum, &status.Thread, &status.Post)
	if err != nil {
		panic(err)
	}
	return operations.NewStatusOK().WithPayload(status)
}