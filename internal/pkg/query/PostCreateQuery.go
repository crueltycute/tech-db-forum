package query

import (
	"fmt"
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
	"strings"
)

func PostCreate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugOrID := req.URL.Query().Get(":slug_or_id")

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	var thread *models.Thread
	if thread, err = getThreadBySlugOrId(tx, slugOrID); err != nil {
		if err == pgx.ErrNoRows {
			tx.Rollback()
			models.ErrResponse(res, http.StatusNotFound, "slug or id not found")
			return
		}
		tx.Rollback()
		panic(err)
	}

	postsToCreate := models.Posts{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = postsToCreate.UnmarshalJSON(body)

	var b strings.Builder
	b.WriteString("INSERT INTO post(parent, author, message, thread, forum) VALUES ")
	var vals []interface{}

	rowIndex := 0
	for _, post := range postsToCreate {
		if rowIndex != 0 {
			b.WriteString(",")
		}
		if inThread := isPostInThread(tx, post.Parent, int64(thread.ID)); post.Parent != 0 && !inThread {
			tx.Rollback()
			models.ErrResponse(res, http.StatusConflict, "parent is in another thread")
			return
		}
		if exists := forumerExists(tx, post.Author); !exists {
			tx.Rollback()
			models.ErrResponse(res, http.StatusNotFound, "author not found")
			return
		}
		b.WriteString(fmt.Sprintf("(NULLIF($%d, 0), $%d, $%d, $%d, $%d)", (rowIndex*5)+1, (rowIndex*5)+2, (rowIndex*5)+3, (rowIndex*5)+4, (rowIndex*5)+5))
		vals = append(vals, post.Parent, post.Author, post.Message, thread.ID, thread.Forum)
		rowIndex++
	}
	postsAdded := models.Posts{}
	if rowIndex == 0 {
		if err := tx.Commit(); err != nil {
			panic(err)
		}

		models.ResponseObject(res, http.StatusCreated, postsAdded)
		return
	}

	b.WriteString(" RETURNING author, created, id, message, thread, coalesce(parent, 0)")

	rows, err := tx.Query(b.String(), vals...)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	for rows.Next() {
		post := &models.Post{}
		//var created pgtype.Timestamptz
		err := rows.Scan(&post.Author, &post.Created, &post.ID, &post.Message, &post.Thread, &post.Parent)
		//t := strfmt.DateTime(created.Time)
		//post.Created = &t
		if err != nil {
			tx.Rollback()
			panic(err)
		}

		post.Forum = thread.Forum
		postsAdded = append(postsAdded, post)
	}

	if err := rows.Err(); err != nil {
		tx.Rollback()
		panic(err)
	}
	rows.Close()

	if len(postsAdded) > 0 {
		err = increasePostsCount(tx, postsAdded[0].Forum, len(postsAdded))
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	vals = []interface{}{}
	b.Reset()
	b.WriteString("INSERT INTO ForumForumer(slug, nickname) VALUES ")
	for i, post := range postsAdded {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf("($%d, $%d)", (i*2)+1, (i*2)+2))
		vals = append(vals, post.Forum, post.Author)
	}
	b.WriteString("ON CONFLICT DO NOTHING")

	_, err = tx.Exec(b.String(), vals...)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusCreated, postsAdded)
	return
}

func isPostInThread(db txOrDb, postId, threadId int64) bool {
	var id int64
	row := db.QueryRow(`
SELECT id
FROM post
WHERE id = $1 and thread = $2`, postId, threadId)
	err := row.Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false
		}
		panic(err)
	}
	return id == postId
}

func increasePostsCount(tx txOrDb, forumSlug string, count int) error {
	_, err := tx.Exec(`
UPDATE forum
SET posts = posts + $1
WHERE slug = $2;`, count, forumSlug)
	return err
}
