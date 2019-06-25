package service

import (
	"fmt"
	"github.com/jackc/pgx"
	"strconv"

	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
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
	err := db.QueryRow(queryGetUserNickByNick, forum.User).Scan(&nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "forum author not found")
			return
		}
		panic(err)
	}

	_, err = db.Exec(queryAddForum, forum.Title, nickname, forum.Slug)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingForum := &models.Forum{}
			err = db.QueryRow(queryGetForumBySlug, forum.Slug).Scan(&existingForum.Title, &existingForum.User, &existingForum.Slug)

			existingForum.User = nickname

			models.ResponseObject(res, http.StatusConflict, existingForum)
			return
		}
		panic(err)
	}

	newForum := &models.Forum{}
	err = db.QueryRow(queryGetForumBySlug, forum.Slug).Scan(&newForum.Title, &newForum.User, &newForum.Slug)

	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusCreated, newForum)
	return
}

func ForumGetOne(res http.ResponseWriter, req *http.Request) {
	slug := req.URL.Query().Get(":slug")
	db := db2.Connection

	forum, err := getForumBySlug(db, slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "forum author not found")
			return
		}
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, forum)
	return
}

func ForumGetThreads(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")
	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if limit == 0 {
		limit = 100
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	sinceStr := ""
	if since != "" {
		if desc {
			sinceStr = fmt.Sprintf("and created <= '%s'::timestamptz", since)
		} else {
			sinceStr = fmt.Sprintf("and created >= '%s'::timestamptz", since)
		}
	}

	queryStatement := fmt.Sprintf(`
		SELECT id, title, author, forum, message, coalesce(slug, ''), created, votes
		FROM thread
		WHERE forum = $1 %s
		ORDER BY created %s
		LIMIT $2`, sinceStr, order)

	rows, err := db.Query(queryStatement, slugName, limit)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	threads := models.Threads{}
	for rows.Next() {
		thread := &models.Thread{}

		err := rows.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Slug, &thread.Created, &thread.Votes)

		if err != nil {
			panic(err)
		}

		threads = append(threads, thread)
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}

	if len(threads) == 0 && !forumExists(db, slugName) {
		models.ErrResponse(res, http.StatusNotFound, "forum not found")
		return
	}

	models.ResponseObject(res, http.StatusOK, threads)
	return
}

func ForumGetUsers(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")

	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if exists := forumExists(db, slugName); !exists {
		models.ErrResponse(res, http.StatusNotFound, "forum not found")
		return
	}

	if limit == 0 {
		limit = 10000
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	sinceStr := ""
	if since != "" {
		comparisonSign := ">"
		if desc {
			comparisonSign = "<"
		}
		sinceStr = fmt.Sprintf("and ff.nickname %s '%s'", comparisonSign, since)
	}
	queryStatement := fmt.Sprintf(`
		SELECT f.nickname, f.fullname, f.about, f.email
		FROM ForumUser AS ff
		JOIN Users as f ON ff.nickname = f.nickname
		WHERE ff.slug = $1 %s
		ORDER BY f.nickname %s
		LIMIT $2`, sinceStr, order)

	rows, err := db.Query(queryStatement, slugName, limit)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	users := models.Users{}
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	models.ResponseObject(res, http.StatusOK, users)
	return
}
