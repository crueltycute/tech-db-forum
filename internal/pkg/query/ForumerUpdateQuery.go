package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"io/ioutil"
	"net/http"
	"strings"
)

func ForumerUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	u := models.User{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = u.UnmarshalJSON(body)

	rows, err := db.Exec(`
UPDATE forumer SET
fullname=COALESCE(NULLIF($1, ''), fullname),
about=COALESCE(NULLIF($2, ''), about),
email=COALESCE(NULLIF($3, ''), email)
WHERE nickname=$4`,
		u.Fullname, u.About, u.Email, nickname)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			models.ErrResponse(res, http.StatusConflict, "cannot update user")
			return
		}
		panic(err)
	}
	count := rows.RowsAffected()
	if count == 0 {
		models.ErrResponse(res, http.StatusNotFound, "user not found")
		return
	}

	updatedData := &models.User{}
	row := db.QueryRow("SELECT nickname, fullname, about, email FROM Forumer WHERE nickname = $1", nickname)
	err = row.Scan(&updatedData.Nickname, &updatedData.Fullname, &updatedData.About, &updatedData.Email)
	if err != nil {
		panic(err)
	}
	models.ResponseObject(res, http.StatusOK, updatedData)
	return
}
