package service

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"net/http"
)

func Clear(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	_, _ = db.Exec(queryClearDB)

	//if err != nil {
	//	panic(err)
	//}

	models.ErrResponse(res, http.StatusOK, "ok")
	return
}

func Status(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	status := &models.Status{}
	row := db.QueryRow(queryGetStatus)
	_ := row.Scan(&status.User, &status.Forum, &status.Thread, &status.Post)
	//if err != nil {
	//	panic(err)
	//}

	models.ResponseObject(res, http.StatusOK, status)
	return
}
