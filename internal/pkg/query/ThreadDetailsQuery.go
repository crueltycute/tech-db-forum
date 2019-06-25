package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"
)

func ThreadDetails(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugOrID := req.URL.Query().Get(":slug_or_id")

	thread, err := getThreadBySlugOrId(db, slugOrID)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "thread does not exist")
			return
		}
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, thread)
	return
}

func getThreadBySlugOrId(db txOrDb, slugOrId string) (*models.Thread, error) {
	if id, err := strconv.Atoi(slugOrId); err == nil {
		return getThreadById(db, id)
	}
	return getThreadBySlug(db, slugOrId)
}

func getThreadById(db txOrDb, id int) (*models.Thread, error) {
	thread := &models.Thread{}
	row := db.QueryRow(`
SELECT id, title, author, forum, message, coalesce(slug, ''), created, votes
FROM thread
WHERE id = $1`,
		id)

	//var created pgtype.Timestamptz
	err := row.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum,
		&thread.Message, &thread.Slug, &thread.Created, &thread.Votes)
	if err != nil && err != pgx.ErrNoRows {
		panic(err)
	}
	//t := strfmt.DateTime(created.Time)
	//thread.Created = &t
	return thread, err
}

func getThreadBySlug(db txOrDb, slug string) (*models.Thread, error) {
	thread := &models.Thread{}
	row := db.QueryRow(`
SELECT id, title, author, forum, message, coalesce(slug, ''), created, votes
FROM thread
WHERE slug = $1`,
		slug)

	//var created pgtype.Timestamptz
	err := row.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum,
		&thread.Message, &thread.Slug, &thread.Created, &thread.Votes)
	//t := strfmt.DateTime(created.Time)
	//thread.Created = &t
	if err != nil && err != pgx.ErrNoRows {
		panic(err)
	}
	return thread, err
}
