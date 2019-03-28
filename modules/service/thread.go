package service

import (
	"database/sql"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"strings"
)

const getForumSlug = `
	SELECT slug FROM Forum 
	WHERE slug = $1`

const insertThread = `
	INSERT INTO Thread (author, forum, slug, title, message, created) 
	VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6) RETURNING id`

const getThreadBySlug = `
	SELECT id, author, forum, title, slug, message, created
	FROM Thread WHERE slug = $1`

const getThreadById = `
	SELECT author, forum, title, COALESCE(slug, ''), message, created 
	FROM Thread WHERE id = $1`


func ThreadCreate(db *sql.DB, params operations.ThreadCreateParams) middleware.Responder {

	var nickname string
	err := db.QueryRow(getUserNickname, params.Thread.Author).Scan(&nickname)
	if err != nil {
		return operations.NewThreadCreateNotFound().WithPayload(&models.Error{Message: "forum creator does not exist"})
	}

	var slug string
	err = db.QueryRow(getForumSlug, params.Slug).Scan(&slug)
	if err != nil {
		return operations.NewThreadCreateNotFound().WithPayload(&models.Error{Message: "forum does not exist"})
	}

	createdThread := &models.Thread{}
	err = db.QueryRow(insertThread, nickname, slug,
						params.Thread.Slug, params.Thread.Title,
						params.Thread.Message, params.Thread.Created).Scan(&createdThread.ID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingThread := &models.Thread{}
			db.QueryRow(getThreadBySlug, params.Thread.Slug).Scan(&existingThread.ID, &existingThread.Author,
																&existingThread.Forum, &existingThread.Title,
																&existingThread.Slug, &existingThread.Message,
																&existingThread.Created)

			return operations.NewThreadCreateConflict().WithPayload(existingThread)
		}
	}

	err = db.QueryRow(getThreadById, createdThread.ID).Scan(&createdThread.Author, &createdThread.Forum,
															&createdThread.Title, &createdThread.Slug,
															&createdThread.Message, &createdThread.Created)

	return operations.NewThreadCreateCreated().WithPayload(createdThread)
}

//func GetThreadBySlugOrId(db* sql.DB, params operations.ThreadGetOneParams) middleware.Responder {
//
//
//}