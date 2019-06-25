package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"net/http"
)

func Clear(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	_, err := db.Exec(`
TRUNCATE TABLE forumer CASCADE;
TRUNCATE TABLE forum CASCADE;
TRUNCATE TABLE post CASCADE;
TRUNCATE TABLE thread CASCADE;
TRUNCATE TABLE forumForumer CASCADE;
TRUNCATE TABLE vote CASCADE `)

	if err != nil {
		panic(err)
	}

	// по логике не подходит, но мне влом делать новый тип респонсов
	models.ErrResponse(res, http.StatusOK, "ok")
	return
}
