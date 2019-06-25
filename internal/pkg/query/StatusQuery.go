package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"net/http"
)

func Status(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	status := &models.Status{}
	row := db.QueryRow(`
SELECT
(SELECT COUNT(*) FROM forumer),
(SELECT COUNT(*) FROM forum),
(SELECT COUNT(*) FROM thread),
(SELECT COUNT(*) FROM post)`)
	err := row.Scan(&status.User, &status.Forum, &status.Thread, &status.Post)
	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, status)
	return
}
