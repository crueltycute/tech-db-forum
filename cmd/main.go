package main

import (
	"flag"
	"github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/app/server"
	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

func main() {
	port := flag.String("port", "5000", "forum port")
	flag.Parse()

	logrus.Warn("forum-server will start on port ", *port)

	err := db.EnsureDBConnection(pgx.ConnConfig{
		Host: "localhost",
		Port: 5432,
		//Database: "postgres",
		Database: "docker",
		User:     "docker",
		Password: "docker",
	})
	if err != nil {
		logrus.Fatal(err, "cant connect db")
	}

	fs := server.NewForumServer(*port)
	fs.EnsureRoutes()

	err = fs.Run()
	if err != nil {
		logrus.Fatal(err, "cant run server")
	}
}
