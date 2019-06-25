package service

import (
	"fmt"
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
	"strconv"
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
	b.WriteString("INSERT INTO post (parent, author, message, thread, forum) VALUES ")
	var vals []interface{}

	rowIndex := 0
	for _, post := range postsToCreate {
		if rowIndex != 0 {
			b.WriteString(",")
		}
		if inThread := postExistsInThread(tx, post.Parent, int64(thread.ID)); post.Parent != 0 && !inThread {
			tx.Rollback()
			models.ErrResponse(res, http.StatusConflict, "parent is in another thread")
			return
		}
		if exists := userExists(tx, post.Author); !exists {
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

		err := rows.Scan(&post.Author, &post.Created, &post.ID, &post.Message, &post.Thread, &post.Parent)

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
	b.WriteString("INSERT INTO ForumUser(slug, nickname) VALUES ")
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

func PostGetOne(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	ID, _ := strconv.Atoi(req.URL.Query().Get(":id"))
	query := req.URL.Query()
	related := strings.Split(query.Get("related"), ",")

	post, err := getPostById(db, int64(ID))
	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "post not found")
			return
		}
		panic(err)
	}
	fullPost := &models.PostFull{Post: post}

	for _, relatedStr := range related {
		switch relatedStr {
		case "user":
			user, err := getUserByNickname(db, post.Author)
			if err != nil {
				if err != pgx.ErrNoRows {
					panic(err)
				}
			}
			fullPost.Author = user
		case "forum":
			forum, err := getForumBySlug(db, post.Forum)
			if err != nil {
				if err != pgx.ErrNoRows {
					panic(err)
				}
			}
			fullPost.Forum = forum
		case "thread":
			thread, err := getThreadById(db, int(post.Thread))
			if err != nil {
				if err != pgx.ErrNoRows {
					panic(err)
				}
			}
			fullPost.Thread = thread
		}
	}

	models.ResponseObject(res, http.StatusOK, fullPost)
	return
}

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

	_, err = db.Exec(queryUpdatePost, newPost.Message, ID)

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
