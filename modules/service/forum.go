package service

import (
	"database/sql"
	"fmt"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"strings"
)

const insertForum = `
	INSERT INTO Forum (slug, forumUser, title) 
	VALUES ($1, $2, $3)`

const getForumBySlug = `
	SELECT slug, forumUser, title
	FROM Forum WHERE slug = $1`

const getOneForumBySlug = `
	SELECT slug, forumUser, title, posts, threads
	FROM Forum WHERE slug = $1`

const getUserNickname = `
	SELECT nickname FROM Users WHERE nickname = $1`


func ForumCreate(db *sql.DB, params operations.ForumCreateParams) middleware.Responder {
	var usersNickname string
	err := db.QueryRow(getUserNickname, params.Forum.User).Scan(&usersNickname)

	if err != nil {
		return operations.NewForumCreateNotFound().WithPayload(&models.Error{Message: "forum creator not found"})
	}

	_, err = db.Exec(insertForum, &params.Forum.Slug, &usersNickname, &params.Forum.Title)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingForum := &models.Forum{}
			db.QueryRow(getForumBySlug, &params.Forum.Slug).Scan(&existingForum.Slug, &existingForum.User, &existingForum.Title)
			return operations.NewForumCreateConflict().WithPayload(existingForum)
		}
		return operations.NewForumCreateNotFound().WithPayload(&models.Error{Message: "forum creator not found"})
	}

	createdForum := &models.Forum{}
	err = db.QueryRow(getForumBySlug, &params.Forum.Slug).Scan(&createdForum.Slug, &createdForum.User, &createdForum.Title)

	if err != nil {
		panic(err)
	}

	return operations.NewForumCreateCreated().WithPayload(createdForum)
}


func GetForumBySlug(db *sql.DB, params operations.ForumGetOneParams) middleware.Responder {
	rows, _ := db.Query(getOneForumBySlug, params.Slug)
	defer rows.Close()

	if rows.Next() {
		forum := &models.Forum{}
		rows.Scan(&forum.Slug, &forum.User, &forum.Title, &forum.Posts, &forum.Threads)
		return operations.NewForumGetOneOK().WithPayload(forum)
	}

	return operations.NewForumGetOneNotFound().WithPayload(&models.Error{"forum not found"})
}


func ForumGetThreads(db *sql.DB, params operations.ForumGetThreadsParams) middleware.Responder {
	orderCondition := ""
	sinceCondition := ""

	if *params.Desc == true {
		orderCondition = "ORDER BY created DESC"
		sinceCondition = "AND created <= $3"
	} else {
		orderCondition = "ORDER BY created"
		sinceCondition = "AND created >= $3"
	}

	limitCondition := ""
	if *params.Limit > 0 {
		limitCondition = "LIMIT $2"
	}

	query := fmt.Sprintf(`SELECT T.id, author, F.slug, T.title, T.slug, T.message, T.created
								FROM Thread as T JOIN forum as F on T.forum = F.slug
								WHERE f.slug = $1 %s %s %s`, sinceCondition, orderCondition, limitCondition)

	rows, err := db.Query(query, params.Slug, params.Limit)
	defer rows.Close()

	if err != nil {
		return operations.NewForumGetThreadsNotFound().WithPayload(&models.Error{err.Error()})
	}

	threads := models.Threads{}
	for rows.Next() {
		thread := &models.Thread{}
		rows.Scan(&thread.ID, &thread.Author, &thread.Forum, &thread.Title, &thread.Slug, &thread.Message, &thread.Created)
		threads = append(threads, thread)
	}

	return operations.NewForumGetThreadsOK().WithPayload(threads)
}