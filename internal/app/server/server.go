package server

import (
	"github.com/bmizerany/pat"
	"github.com/crueltycute/tech-db-forum/internal/pkg/service"
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
	fs.router.Post("/api/forum/create", http.HandlerFunc(service.ForumCreate))
	fs.router.Get("/api/forum/:slug/details", http.HandlerFunc(service.ForumGetOne))
	fs.router.Post("/api/forum/:slug/create", http.HandlerFunc(service.ThreadCreate))
	fs.router.Get("/api/forum/:slug/users", http.HandlerFunc(service.ForumGetUsers))
	fs.router.Get("/api/forum/:slug/threads", http.HandlerFunc(service.ForumGetThreads))

	fs.router.Get("/api/post/:id/details", http.HandlerFunc(service.PostGetOne))
	fs.router.Post("/api/post/:id/details", http.HandlerFunc(service.PostUpdate))

	fs.router.Post("/api/service/clear", http.HandlerFunc(service.Clear))
	fs.router.Get("/api/service/status", http.HandlerFunc(service.Status))

	fs.router.Post("/api/thread/:slug_or_id/create", http.HandlerFunc(service.PostCreate))
	fs.router.Get("/api/thread/:slug_or_id/details", http.HandlerFunc(service.ThreadGetOne))
	fs.router.Post("/api/thread/:slug_or_id/details", http.HandlerFunc(service.ThreadUpdate))
	fs.router.Get("/api/thread/:slug_or_id/posts", http.HandlerFunc(service.ThreadGetPosts))
	fs.router.Post("/api/thread/:slug_or_id/vote", http.HandlerFunc(service.ThreadVote))

	fs.router.Post("/api/user/:nickname/create", http.HandlerFunc(service.UsersCreate))
	fs.router.Get("/api/user/:nickname/profile", http.HandlerFunc(service.UsersGetOne))
	fs.router.Post("/api/user/:nickname/profile", http.HandlerFunc(service.UsersUpdate))
}

func (fs *ForumServer) Run() error {
	return http.ListenAndServe(fs.port, fs.router)
}
