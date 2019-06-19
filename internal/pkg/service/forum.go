package service

import (
	"database/sql"
	"fmt"

	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"io/ioutil"
	"net/http"
	"strings"
)

func ForumCreate(res http.ResponseWriter, req *http.Request) {
	f := models.Forum{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = f.UnmarshalJSON(body)

	db := db2.Connection

	var usersNickname string
	err := db.QueryRow(queryGetUserNickByNick, f.User).Scan(&usersNickname)

	if err != nil {
		if err == sql.ErrNoRows {
			//return operations.NewForumCreateNotFound().WithPayload(&models.Error{Message: "forum author not found"})
		}
		panic(err)
	}

	_, err = db.Exec(queryAddForum, &f.Slug, &usersNickname, &f.Title)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingForum := &models.Forum{}
			err := db.QueryRow(queryGetForumBySlug, &f.Slug).Scan(&existingForum.Slug, &existingForum.User, &existingForum.Title)

			if err != nil {
				panic(err)
			}

			//return operations.NewForumCreateConflict().WithPayload(existingForum)
		}
		//return operations.NewForumCreateNotFound().WithPayload(&internal.Error{Message: "forum author not found"})
	}

	createdForum := &models.Forum{}
	err = db.QueryRow(queryGetForumBySlug, &f.Slug).Scan(&createdForum.Slug, &createdForum.User, &createdForum.Title)

	if err != nil {
		panic(err)
	}

	//return operations.NewForumCreateCreated().WithPayload(createdForum)
}

func ForumGetOne(res http.ResponseWriter, req *http.Request) {
	slug := req.URL.Query().Get("slug")
	db := db2.Connection

	forum := &models.Forum{}
	err := db.QueryRow(queryGetFullForumBySlug, slug).Scan(&forum.Slug, &forum.User, &forum.Title, &forum.Posts, &forum.Threads)
	if err != nil {
		if err == sql.ErrNoRows {
			//return operations.NewForumGetOneNotFound().WithPayload(&internal.Error{Message: "forum author not found"})
		}
	}
	//return operations.NewForumGetOneOK().WithPayload(forum)
}

func ForumGetThreads(res http.ResponseWriter, req *http.Request) {
	order := ""
	if params.Desc != nil {
		if *params.Desc == true {
			order = "DESC"
		} else {
			order = "ASC"
		}
	}

	since := ""
	if params.Since != nil {
		if params.Desc != nil && *params.Desc == true {
			since = fmt.Sprintf("and created <= '%s'::timestamptz", params.Since)
		} else {
			since = fmt.Sprintf("and created >= '%s'::timestamptz", params.Since)
		}
	}

	queryStatement := `SELECT T.id, T.title, T.author, F.slug, T.message, T.slug, T.created
					   FROM Thread as T JOIN Forum as F on T.forum = F.slug
					   WHERE F.slug = $1 %s ORDER BY created %s LIMIT $2`

	query := fmt.Sprintf(queryStatement, since, order)

	rows, err := db.Query(query, &params.Slug, &params.Limit)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	threads := internal.Threads{}
	for rows.Next() {
		thread := &internal.Thread{}
		err = rows.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Slug, &thread.Created)

		if err != nil {
			panic(err)
		}

		threads = append(threads, thread)
	}

	if contains := forumIsInDB(db, &params.Slug); !contains && len(threads) == 0 {
		//return operations.NewForumGetThreadsNotFound().WithPayload(&internal.Error{Message: "forum not found"})
	}

	//return operations.NewForumGetThreadsOK().WithPayload(threads)
}

func ForumGetUsers(res http.ResponseWriter, req *http.Request) {
	if contains := forumIsInDB(db, &params.Slug); !contains {
		//return operations.NewForumGetUsersNotFound().WithPayload(&internal.Error{Message: "forum not found"})
	}

	order := ""
	if params.Desc != nil {
		if *params.Desc == true {
			order = "DESC"
		} else {
			order = "ASC"
		}
	}

	since := ""
	if params.Since != nil {
		comparisonSign := ">"
		if params.Desc != nil && *params.Desc == true {
			comparisonSign = "<"
		}
		since = fmt.Sprintf("and FU.nickname %s '%s'", comparisonSign, *params.Since)
	}

	queryStatement := `SELECT U.nickname, U.fullname, U.about, U.email
					   FROM ForumUser AS FU
					   JOIN Users as U ON FU.nickname = U.nickname
					   WHERE FU.slug = $1 %s
					   ORDER BY U.nickname %s
					   LIMIT $2`

	query := fmt.Sprintf(queryStatement, since, order)

	rows, err := db.Query(query, &params.Slug, &params.Limit)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	users := internal.Users{}
	for rows.Next() {
		user := &internal.User{}
		err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	//return operations.NewForumGetUsersOK().WithPayload(users)
}
