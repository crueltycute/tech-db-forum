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


func ThreadCreate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")
	thread := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = thread.UnmarshalJSON(body)

	thread.Forum = slugName

	var nickname string
	err := db.QueryRow(queryGetUserNickByNick, thread.Author).Scan(&nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "forum author not found")
			return
		}
		panic(err)
	}

	var slug string
	err = db.QueryRow(queryGetForumSlugBySlug, slugName).Scan(&slug)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "forum not found th")
			return
		}
		panic(err)
	}

	lastInsertId := 0

	err = db.QueryRow(queryAddThread, thread.Title, nickname,
					  slug, thread.Message, thread.Slug, thread.Created).Scan(&lastInsertId)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingThread := &models.Thread{}
			err := db.QueryRow(queryGetThreadBySlug, thread.Slug).Scan(&existingThread.ID, &existingThread.Title, &existingThread.Author, &existingThread.Forum, &existingThread.Message, &existingThread.Slug, &existingThread.Created)

			if err != nil {
				panic(err)
			}

			models.ResponseObject(res, http.StatusConflict, existingThread)
			return
		}
		panic(err)
	}

	newThread := &models.Thread{}

	err = db.QueryRow(queryGetThreadById, lastInsertId).Scan(&newThread.ID, &newThread.Title, &newThread.Author, &newThread.Forum, &newThread.Message, &newThread.Slug, &newThread.Created)

	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusCreated, newThread)
	return
}

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

	_, err = tx.Exec(queryAddVote, thread.ID, voteToCreate.Nickname, voteToCreate.Voice)
	if err != nil {
		if strings.Contains(err.Error(), "vote_nickname_fkey") {
			tx.Rollback()
			models.ErrResponse(res, http.StatusNotFound, "nickname not found")
			return
		}
		tx.Rollback()
		panic(err)
	}

	votedThread := &models.Thread{}

	err = tx.QueryRow(queryGetThreadWithVotesById, thread.ID).Scan(&votedThread.ID, &votedThread.Title, &votedThread.Author, &votedThread.Forum, &votedThread.Message,
		&votedThread.Votes, &votedThread.Slug, &votedThread.Created)

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

func ThreadGetOne(res http.ResponseWriter, req *http.Request) {
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

func ThreadGetPosts(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	slugOrID := req.URL.Query().Get(":slug_or_id")

	thread, err := getThreadBySlugOrId(db, slugOrID)
	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "thread not found")
			return
		}
		panic(err)
	}

	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since, _ := strconv.Atoi(query.Get("since"))
	sort := query.Get("sort")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if limit == 0 {
		limit = 100
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	sinceStr := ""
	if since != 0 {
		comparisonSign := ">"
		if desc {
			comparisonSign = "<"
		}

		sinceStr = fmt.Sprintf("and post.id %s %d", comparisonSign, since)
		if sort == "tree" {
			sinceStr = fmt.Sprintf("and post.path %s (SELECT tree_post.path FROM post AS tree_post WHERE tree_post.id = %d)", comparisonSign, since)
		}
		if sort == "parent_tree" {
			sinceStr = fmt.Sprintf("and post_roots.path[1] %s (SELECT tree_post.path[1] FROM post AS tree_post WHERE tree_post.id = %d)", comparisonSign, since)
		}
	}

	queryStatement := fmt.Sprintf(`
		SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0)
		FROM post
		WHERE post.thread = $1 %s
		ORDER BY (post.created, post.id) %s
		LIMIT $2`, sinceStr, order)

	if sort == "tree" {
		queryStatement = fmt.Sprintf(`
			SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0)
			FROM post
			WHERE post.thread = $1 %s
			ORDER BY (post.path, post.created) %s
			LIMIT $2`, sinceStr, order)
	} else if sort == "parent_tree" {
		queryStatement = fmt.Sprintf(`
			SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0)
			FROM post
			WHERE post.thread = $1 AND post.path[1] IN (
				SELECT post_roots.id
				FROM post as post_roots
				WHERE post_roots.id = post_roots.path[1] AND post_roots.thread = post.thread %s
				ORDER BY post_roots.path[1] %s
				LIMIT $2
			)
			ORDER BY post.path[1] %s, post.path`, sinceStr, order, order)
	}

	rows, err := db.Query(queryStatement, thread.ID, limit)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	posts := models.Posts{}
	for rows.Next() {
		post := &models.Post{}

		err := rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.Parent)

		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, posts)
	return
}

func ThreadUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	slugOrID := req.URL.Query().Get(":slug_or_id")

	thread, err := getThreadBySlugOrId(db, slugOrID)
	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "thread not found")
			return
		}
		panic(err)
	}

	t := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = t.UnmarshalJSON(body)

	_, err = db.Exec(queryUpdateThread, t.Title, t.Message, thread.ID)

	if err != nil {
		panic(err)
	}

	updatedData, err := getThreadBySlugOrId(db, slugOrID)
	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, updatedData)
	return
}