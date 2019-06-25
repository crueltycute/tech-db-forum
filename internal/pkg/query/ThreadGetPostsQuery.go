package query

import (
	"fmt"
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"
)

func ThreadGetPosts(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	slugOrID := req.URL.Query().Get(":slug_or_id")

	thread, err := getThreadBySlugOrId(db, slugOrID)
	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "thread not found")
			return
		}
		panic(err)
	}

	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since, _ := strconv.Atoi(query.Get("since"))
	sort := query.Get("sort")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if limit == 0 {
		limit = 100
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	sinceStr := ""
	if since != 0 {
		comparisonSign := ">"
		if desc {
			comparisonSign = "<"
		}

		sinceStr = fmt.Sprintf("and post.id %s %d", comparisonSign, since)
		if sort == "tree" {
			sinceStr = fmt.Sprintf("and post.path %s (SELECT tree_post.path FROM post AS tree_post WHERE tree_post.id = %d)", comparisonSign, since)
		}
		if sort == "parent_tree" {
			sinceStr = fmt.Sprintf("and post_roots.path[1] %s (SELECT tree_post.path[1] FROM post AS tree_post WHERE tree_post.id = %d)", comparisonSign, since)
		}
	}

	queryStatement := fmt.Sprintf(`
SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0)
FROM post
WHERE post.thread = $1 %s
ORDER BY (post.created, post.id) %s
LIMIT $2
`, sinceStr, order)

	if sort == "tree" {
		queryStatement = fmt.Sprintf(`
SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0)
FROM post
WHERE post.thread = $1 %s
ORDER BY (post.path, post.created) %s
LIMIT $2
`, sinceStr, order)
	} else if sort == "parent_tree" {
		queryStatement = fmt.Sprintf(`
SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0)
FROM post
WHERE post.thread = $1 AND post.path[1] IN (
	SELECT post_roots.id
	FROM post as post_roots
	WHERE post_roots.id = post_roots.path[1] AND post_roots.thread = post.thread %s
	ORDER BY post_roots.path[1] %s
	LIMIT $2
)
ORDER BY post.path[1] %s, post.path
`, sinceStr, order, order)
	}

	rows, err := db.Query(queryStatement, thread.ID, limit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	posts := models.Posts{}
	for rows.Next() {
		post := &models.Post{}
		//var created pgtype.Timestamptz
		err := rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.Parent)
		//t := strfmt.DateTime(created.Time)
		//post.Created = &t
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, posts)
	return
}
