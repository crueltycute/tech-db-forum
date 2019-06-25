package query

import (
	"fmt"
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"net/http"
	"strconv"
)

func ForumGetThreads(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")
	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if limit == 0 {
		limit = 100
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	sinceStr := ""
	if since != "" {
		if desc {
			sinceStr = fmt.Sprintf("and created <= '%s'::timestamptz", since)
		} else {
			sinceStr = fmt.Sprintf("and created >= '%s'::timestamptz", since)
		}
	}

	queryStatement := fmt.Sprintf(`
SELECT id, title, author, forum, message, coalesce(slug, ''), created, votes
FROM thread
WHERE forum = $1 %s
ORDER BY created %s
LIMIT $2
`, sinceStr, order)

	rows, err := db.Query(queryStatement, slugName, limit)
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	threads := models.Threads{}
	for rows.Next() {
		thread := &models.Thread{}
		//var created pgtype.Timestamptz
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Slug, &thread.Created, &thread.Votes)
		//t := strfmt.DateTime(created.Time)
		//thread.Created = &t
		if err != nil {
			panic(err)
		}
		threads = append(threads, thread)
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}

	if len(threads) == 0 && !forumExists(db, slugName) {
		models.ErrResponse(res, http.StatusNotFound, "forum not found")
		return
	}

	models.ResponseObject(res, http.StatusOK, threads)
	return
}
