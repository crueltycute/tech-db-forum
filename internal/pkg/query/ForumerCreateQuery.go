package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"io/ioutil"
	"net/http"
	"strings"
)

func ForumerCreate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	profile := models.User{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = profile.UnmarshalJSON(body)

	profile.Nickname = nickname

	_, err := db.Exec("INSERT INTO Forumer(nickname, fullname, about, email) VALUES ($1, $2, $3, $4)",
		profile.Nickname, profile.Fullname, profile.About, profile.Email)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			rows, err := db.Query("SELECT nickname, fullname, about, email FROM Forumer WHERE nickname = $1 OR email = $2",
				profile.Nickname, profile.Email)
			if err != nil {
				panic(err)
			}
			defer rows.Close()
			forumers := models.Users{}
			for rows.Next() {
				forumer := &models.User{}
				err := rows.Scan(&forumer.Nickname, &forumer.Fullname, &forumer.About, &forumer.Email)
				if err != nil {
					panic(err)
				}
				forumers = append(forumers, forumer)
			}
			models.ResponseObject(res, http.StatusConflict, forumers)
			return
		}
		panic(err)
	}

	models.ResponseObject(res, http.StatusCreated, profile)
	return
}
