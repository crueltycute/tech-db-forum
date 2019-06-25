package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"net/http"
)

func ForumDetails(res http.ResponseWriter, req *http.Request) {
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

func forumExists(db dbOrConn, slug string) bool {
	forum := &models.Forum{}
	row := db.QueryRow("SELECT title,forumer, slug FROM Forum WHERE Forum.slug = $1", slug)
	err := row.Scan(&forum.Title, &forum.User, &forum.Slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func getForumBySlug(db dbOrConn, slug string) (*models.Forum, error) {
	forum := &models.Forum{}
	row := db.QueryRow(`
SELECT title, forumer, slug, posts, threads
FROM Forum
WHERE Forum.slug = $1`, slug)
	err := row.Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)
	return forum, err
}
