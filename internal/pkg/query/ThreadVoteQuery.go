package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	voteInsertQuery = `
INSERT INTO Vote(threadId, nickname, voice) VALUES ($1, $2, $3)
ON CONFLICT ON CONSTRAINT unique_vote DO UPDATE
SET voice = EXCLUDED.voice;`
)

func ThreadVote(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	slugOrID := req.URL.Query().Get(":slug_or_id")

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	thread, err := getThreadBySlugOrId(tx, slugOrID)
	if err != nil {
		if err == pgx.ErrNoRows {
			tx.Rollback()
			models.ErrResponse(res, http.StatusNotFound, "thread not found")
			return
		}
		tx.Rollback()
		panic(err)
	}

	voteToCreate := models.Vote{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = voteToCreate.UnmarshalJSON(body)

	_, err = tx.Exec(voteInsertQuery, thread.ID, voteToCreate.Nickname, voteToCreate.Voice)
	if err != nil {
		if strings.Contains(err.Error(), "vote_nickname_fkey") {
			tx.Rollback()
			models.ErrResponse(res, http.StatusNotFound, "nickname not found")
			return
		}
		tx.Rollback()
		panic(err)
	}

	row := tx.QueryRow(`
SELECT id, title, author, forum, message, votes, coalesce(slug, ''), created
FROM thread
WHERE id = $1`, thread.ID)
	votedThread := &models.Thread{}
	//var created pgtype.Timestamptz
	err = row.Scan(&votedThread.ID, &votedThread.Title, &votedThread.Author, &votedThread.Forum, &votedThread.Message,
		&votedThread.Votes, &votedThread.Slug, &votedThread.Created)
	//t := strfmt.DateTime(created.Time)
	//votedThread.Created = &t
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, votedThread)
	return
}
