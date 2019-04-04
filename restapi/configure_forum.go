// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"database/sql"
	"net/http"

	"github.com/crueltycute/tech-db-forum/modules/service"
	"github.com/crueltycute/tech-db-forum/restapi/operations"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	_ "github.com/lib/pq"
)

//go:generate swagger generate server --target ../../tech-db-forum --name Forum --spec ../swagger.yml

func configureFlags(api *operations.ForumAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ForumAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.BinConsumer = runtime.ByteStreamConsumer()

	api.JSONProducer = runtime.JSONProducer()

	conninfo := "user=nadezda dbname=postgres host=0.0.0.0 sslmode=disable"

	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}

	api.ClearHandler = operations.ClearHandlerFunc(func(params operations.ClearParams) middleware.Responder {
		return service.Clear(db, params)
	})
	api.ForumCreateHandler = operations.ForumCreateHandlerFunc(func(params operations.ForumCreateParams) middleware.Responder {
		return service.ForumCreate(db, params)
	})
	api.ForumGetOneHandler = operations.ForumGetOneHandlerFunc(func(params operations.ForumGetOneParams) middleware.Responder {
		return service.ForumGetOne(db, params)
	})
	api.ForumGetThreadsHandler = operations.ForumGetThreadsHandlerFunc(func(params operations.ForumGetThreadsParams) middleware.Responder {
		return service.ForumGetThreads(db, params)
	})
	api.ForumGetUsersHandler = operations.ForumGetUsersHandlerFunc(func(params operations.ForumGetUsersParams) middleware.Responder {
		return service.ForumGetUsers(db, params)
	})
	api.PostGetOneHandler = operations.PostGetOneHandlerFunc(func(params operations.PostGetOneParams) middleware.Responder {
		return service.PostGetOne(db, params)
	})
	api.PostUpdateHandler = operations.PostUpdateHandlerFunc(func(params operations.PostUpdateParams) middleware.Responder {
		return service.PostUpdate(db, params)
	})
	api.PostsCreateHandler = operations.PostsCreateHandlerFunc(func(params operations.PostsCreateParams) middleware.Responder {
		return service.PostsCreate(db, params)
	})
	api.StatusHandler = operations.StatusHandlerFunc(func(params operations.StatusParams) middleware.Responder {
		return service.Status(db, params)
	})
	api.ThreadCreateHandler = operations.ThreadCreateHandlerFunc(func(params operations.ThreadCreateParams) middleware.Responder {
		return service.ThreadCreate(db, params)
	})
	api.ThreadGetOneHandler = operations.ThreadGetOneHandlerFunc(func(params operations.ThreadGetOneParams) middleware.Responder {
		return service.ThreadGetOne(db, params)
	})
	api.ThreadGetPostsHandler = operations.ThreadGetPostsHandlerFunc(func(params operations.ThreadGetPostsParams) middleware.Responder {
		return service.ThreadGetPosts(db, params)
	})
	api.ThreadUpdateHandler = operations.ThreadUpdateHandlerFunc(func(params operations.ThreadUpdateParams) middleware.Responder {
		return service.ThreadUpdate(db, params)
	})
	api.ThreadVoteHandler = operations.ThreadVoteHandlerFunc(func(params operations.ThreadVoteParams) middleware.Responder {
		return service.ThreadVote(db, params)
	})
	api.UserCreateHandler = operations.UserCreateHandlerFunc(func(params operations.UserCreateParams) middleware.Responder {
		return service.UsersCreate(db, params)
	})
	api.UserGetOneHandler = operations.UserGetOneHandlerFunc(func(params operations.UserGetOneParams) middleware.Responder {
		return service.UsersGetOne(db, params)
	})
	api.UserUpdateHandler = operations.UserUpdateHandlerFunc(func(params operations.UserUpdateParams) middleware.Responder {
		return service.UsersUpdate(db, params)
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
