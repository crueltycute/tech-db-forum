package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"
	"strings"
)

func PostDetails(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	ID, _ := strconv.Atoi(req.URL.Query().Get(":id"))
	query := req.URL.Query()
	related := strings.Split(query.Get("related"), ",")

	post, err := getPostById(db, int64(ID))
	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "post not found")
			return
		}
		panic(err)
	}
	fullPost := &models.PostFull{Post: post}

	for _, relatedStr := range related {
		switch relatedStr {
		case "user":
			forumer, err := getForumerByNickname(db, post.Author)
			if err != nil {
				if err != pgx.ErrNoRows {
					panic(err)
				}
			}
			fullPost.Author = forumer
		case "forum":
			forum, err := getForumBySlug(db, post.Forum)
			if err != nil {
				if err != pgx.ErrNoRows {
					panic(err)
				}
			}
			fullPost.Forum = forum
		case "thread":
			thread, err := getThreadById(db, int(post.Thread))
			if err != nil {
				if err != pgx.ErrNoRows {
					panic(err)
				}
			}
			fullPost.Thread = thread
		}
	}

	models.ResponseObject(res, http.StatusOK, fullPost)
	return
}

func getPostById(db dbOrConn, id int64) (*models.Post, error) {
	post := &models.Post{}
	row := db.QueryRow(`
SELECT author, created, forum, id, message, thread, coalesce(isedited, FALSE), coalesce(parent, 0)
FROM Post
WHERE id = $1`, id)
	//var created pgtype.Timestamptz
	err := row.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.IsEdited, &post.Parent)
	//t := strfmt.DateTime(created.Time)
	//post.Created = &t
	return post, err
}
