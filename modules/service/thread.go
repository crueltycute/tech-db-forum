package service

import (
	"database/sql"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"strings"
)

const insertThread = `
	INSERT INTO Thread (author, forum, slug, title, message, created) 
	VALUES ($1, $2, $3 $4, $5, $6) RETURNING id`

const getThreadBySlug = `
	SELECT id, author, forum, slug, title, message, created 
	FROM thread WHERE slug = $1`

func ThreadCreate(db *sql.DB, params operations.ThreadCreateParams) middleware.Responder {
	err := db.QueryRow(getUserNickname, params.Thread.Author)

	if err != nil {
		return operations.NewThreadCreateNotFound().WithPayload(&models.Error{"thread creator not found"})
	}

	err = db.QueryRow(getOneForumBySlug, params.Slug)
	if err != nil {
		return operations.NewThreadCreateNotFound().WithPayload(&models.Error{"thread forum not found"})
	}

	err = db.QueryRow(insertThread, &params.Thread.Author, &params.Thread.Forum,
			&params.Thread.Slug, &params.Thread.Title,
			&params.Thread.Message, &params.Thread.Created)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingThread := &models.Thread{}
			row := db.QueryRow(getThreadBySlug, params.Slug)
			err := row.Scan(&existingThread.ID, &existingThread.Title, &existingThread.Author, &existingThread.Forum, &existingThread.Message, &existingThread.Slug, &existingThread.Created)
			if err != nil {
				panic(err)
			}
			return operations.NewThreadCreateConflict().WithPayload(existingThread)

		}
	}

}
