package query

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
)

func ThreadUpdate(res http.ResponseWriter, req *http.Request) {
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

	t := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = t.UnmarshalJSON(body)

	_, err = db.Exec(`
UPDATE thread 
SET 
title=COALESCE(NULLIF($1, ''), title),
message=COALESCE(NULLIF($2, ''), message)
WHERE id=$3`,
		t.Title, t.Message, thread.ID)

	if err != nil {
		panic(err)
	}

	updatedData, err := getThreadBySlugOrId(db, slugOrID)
	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, updatedData)
	return
}
