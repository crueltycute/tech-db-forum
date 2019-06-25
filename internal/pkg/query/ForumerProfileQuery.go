package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"net/http"
)

func ForumerProfile(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	forumer, err := getForumerByNickname(db, nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "user not found")
			return
		}
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, forumer)
	return
}

func forumerExists(db txOrDb, nickname string) bool {
	forumer := &models.User{}
	row := db.QueryRow("SELECT nickname FROM Forumer WHERE nickname = $1", nickname)
	err := row.Scan(&forumer.Nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func getForumerByNickname(db dbOrConn, nickname string) (*models.User, error) {
	forumer := &models.User{}
	row := db.QueryRow(`
SELECT nickname, fullname, about, email
FROM Forumer
WHERE nickname = $1`, nickname)
	err := row.Scan(&forumer.Nickname, &forumer.Fullname, &forumer.About, &forumer.Email)
	return forumer, err
}
