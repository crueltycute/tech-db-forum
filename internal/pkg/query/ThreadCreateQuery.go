package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
	"strings"
)

func ThreadCreate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")
	thread := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = thread.UnmarshalJSON(body)

	thread.Forum = slugName

	var nickname string
	row := db.QueryRow("SELECT nickname FROM Forumer WHERE Forumer.nickname = $1", thread.Author)
	err := row.Scan(&nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "forum author not found")
			return
		}
		panic(err)
	}

	var slug string
	row = db.QueryRow("SELECT slug FROM forum WHERE slug = $1", slugName)
	err = row.Scan(&slug)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "forum not found th")
			return
		}
		panic(err)
	}

	lastInsertId := 0
	err = db.QueryRow(`
INSERT INTO thread(title, author, forum, message, slug, created)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)
RETURNING id`, thread.Title, nickname, slug, thread.Message, thread.Slug, thread.Created).Scan(&lastInsertId)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingThread := &models.Thread{}
			row := db.QueryRow("SELECT id, title, author, forum, message, slug, created FROM thread WHERE slug = $1", thread.Slug)
			//var created pgtype.Timestamptz
			err := row.Scan(&existingThread.ID, &existingThread.Title, &existingThread.Author, &existingThread.Forum, &existingThread.Message, &existingThread.Slug, &existingThread.Created)
			//t := strfmt.DateTime(created.Time)
			//existingThread.Created = created.Time
			if err != nil {
				panic(err)
			}

			models.ResponseObject(res, http.StatusConflict, existingThread)
			return
		}
		panic(err)
	}

	newThread := &models.Thread{}
	//var created pgtype.Timestamptz
	row = db.QueryRow("SELECT id, title, author, forum, message, coalesce(slug, ''), created FROM thread WHERE id = $1", lastInsertId)
	err = row.Scan(&newThread.ID, &newThread.Title, &newThread.Author, &newThread.Forum, &newThread.Message, &newThread.Slug, &newThread.Created)
	//t := strfmt.DateTime(created.Time)
	//newThread.Created = &t
	if err != nil {
		panic(err)
	}

	//thread = params.Thread
	//thread.ID = int32(lastInsertId)
	models.ResponseObject(res, http.StatusCreated, newThread)
	return
}
