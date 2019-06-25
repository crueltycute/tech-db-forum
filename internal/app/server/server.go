package server

import (
	"github.com/bmizerany/pat"
	"github.com/crueltycute/tech-db-forum/internal/pkg/query"
	"net/http"
)

type ForumServer struct {
	port   string
	router *pat.PatternServeMux
}

func NewForumServer(port string) *ForumServer {
	return &ForumServer{
		port:   ":" + port,
		router: pat.New(),
	}
}

func (fs *ForumServer) EnsureRoutes() {
	fs.router.Post("/api/forum/create", http.HandlerFunc(query.ForumCreate))
	fs.router.Get("/api/forum/:slug/details", http.HandlerFunc(query.ForumDetails))
	fs.router.Post("/api/forum/:slug/create", http.HandlerFunc(query.ThreadCreate))
	fs.router.Get("/api/forum/:slug/users", http.HandlerFunc(query.ForumForumers))
	fs.router.Get("/api/forum/:slug/threads", http.HandlerFunc(query.ForumGetThreads))

	fs.router.Get("/api/post/:id/details", http.HandlerFunc(query.PostDetails))
	fs.router.Post("/api/post/:id/details", http.HandlerFunc(query.PostUpdate))

	fs.router.Post("/api/service/clear", http.HandlerFunc(query.Clear))
	fs.router.Get("/api/service/status", http.HandlerFunc(query.Status))

	fs.router.Post("/api/thread/:slug_or_id/create", http.HandlerFunc(query.PostCreate))
	fs.router.Get("/api/thread/:slug_or_id/details", http.HandlerFunc(query.ThreadDetails))
	fs.router.Post("/api/thread/:slug_or_id/details", http.HandlerFunc(query.ThreadUpdate))
	fs.router.Get("/api/thread/:slug_or_id/posts", http.HandlerFunc(query.ThreadGetPosts))
	fs.router.Post("/api/thread/:slug_or_id/vote", http.HandlerFunc(query.ThreadVote))

	fs.router.Post("/api/user/:nickname/create", http.HandlerFunc(query.ForumerCreate))
	fs.router.Get("/api/user/:nickname/profile", http.HandlerFunc(query.ForumerProfile))
	fs.router.Post("/api/user/:nickname/profile", http.HandlerFunc(query.ForumerUpdate))
}

func (fs *ForumServer) Run() error {
	return http.ListenAndServe(fs.port, fs.router)
}
