package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
	"strings"
)

func ForumCreate(res http.ResponseWriter, req *http.Request) {
	forum := models.Forum{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = forum.UnmarshalJSON(body)

	db := db2.Connection

	var nickname string
	row := db.QueryRow("SELECT nickname FROM Forumer WHERE nickname = $1", forum.User)
	err := row.Scan(&nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "forum author not found")
			return
		}
		panic(err)
	}

	_, err = db.Exec("INSERT INTO Forum(title, forumer, slug) VALUES ($1, $2, $3)", forum.Title, nickname, forum.Slug)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingForum := &models.Forum{}
			row := db.QueryRow("SELECT title, forumer, slug FROM Forum WHERE slug = $1", forum.Slug)
			_ = row.Scan(&existingForum.Title, &existingForum.User, &existingForum.Slug)

			existingForum.User = nickname

			models.ResponseObject(res, http.StatusConflict, existingForum)
			return
		}
		panic(err)
	}

	newForum := &models.Forum{}
	row = db.QueryRow("SELECT title, forumer, slug FROM Forum WHERE slug = $1", forum.Slug)
	err = row.Scan(&newForum.Title, &newForum.User, &newForum.Slug)

	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusCreated, newForum)
	return
}
