package service

import (
	"database/sql"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
)

func CreatePosts(db *sql.DB, params operations.PostsCreateParams) middleware.Responder {

}