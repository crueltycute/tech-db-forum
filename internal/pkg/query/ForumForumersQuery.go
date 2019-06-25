package query

import (
	"fmt"
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"net/http"
	"strconv"
)

func ForumForumers(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")

	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if exists := forumExists(db, slugName); !exists {
		models.ErrResponse(res, http.StatusNotFound, "forum not found")
		return
	}

	if limit == 0 {
		limit = 10000
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	sinceStr := ""
	if since != "" {
		comparisonSign := ">"
		if desc {
			comparisonSign = "<"
		}
		sinceStr = fmt.Sprintf("and ff.nickname %s '%s'", comparisonSign, since)
	}
	queryStatement := fmt.Sprintf(`
SELECT f.nickname, f.fullname, f.about, f.email
FROM ForumForumer AS ff
JOIN Forumer as f ON ff.nickname = f.nickname
WHERE ff.slug = $1 %s
ORDER BY f.nickname %s
LIMIT $2
`, sinceStr, order)

	rows, err := db.Query(queryStatement, slugName, limit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	forumers := models.Users{}
	for rows.Next() {
		forumer := &models.User{}
		err := rows.Scan(&forumer.Nickname, &forumer.Fullname, &forumer.About, &forumer.Email)
		if err != nil {
			panic(err)
		}
		forumers = append(forumers, forumer)
	}

	models.ResponseObject(res, http.StatusOK, forumers)
	return
}
