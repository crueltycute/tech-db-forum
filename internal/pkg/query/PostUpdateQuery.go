package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
	"strconv"
)

func PostUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	ID, _ := strconv.Atoi(req.URL.Query().Get(":id"))

	_, err := getPostById(db, int64(ID))
	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "post not found")
			return
		}
		panic(err)
	}

	newPost := models.Post{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = newPost.UnmarshalJSON(body)

	_, err = db.Exec(`
UPDATE post
SET
message = COALESCE(NULLIF($1, ''), message)
WHERE id = $2`,
		newPost.Message, ID)

	if err != nil {
		panic(err)
	}

	updatedPost, err := getPostById(db, int64(ID))
	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, updatedPost)
	return
}
