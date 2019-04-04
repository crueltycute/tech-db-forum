package service

import (
	"database/sql"
	"fmt"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"strings"
)

func ForumCreate(db *sql.DB, params operations.ForumCreateParams) middleware.Responder {
	var usersNickname string
	err := db.QueryRow(queryGetUserNickByNick, params.Forum.User).Scan(&usersNickname)

	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewForumCreateNotFound().WithPayload(&models.Error{ Message: "forum author not found" })
		}
		panic(err)
	}

	_, err = db.Exec(queryAddForum, &params.Forum.Slug, &usersNickname, &params.Forum.Title)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingForum := &models.Forum{}
			err := db.QueryRow(queryGetForumBySlug, &params.Forum.Slug).Scan(&existingForum.Slug, &existingForum.User, &existingForum.Title)

			if err != nil {
				panic(err)
			}

			return operations.NewForumCreateConflict().WithPayload(existingForum)
		}
		return operations.NewForumCreateNotFound().WithPayload(&models.Error{ Message: "forum author not found" })
	}

	createdForum := &models.Forum{}
	err = db.QueryRow(queryGetForumBySlug, &params.Forum.Slug).Scan(&createdForum.Slug, &createdForum.User, &createdForum.Title)

	if err != nil {
		panic(err)
	}

	return operations.NewForumCreateCreated().WithPayload(createdForum)
}


func ForumGetOne(db *sql.DB, params operations.ForumGetOneParams) middleware.Responder {
	forum := &models.Forum{}
	err := db.QueryRow(queryGetFullForumBySlug, params.Slug).Scan(&forum.Slug, &forum.User, &forum.Title, &forum.Posts, &forum.Threads)
	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewForumGetOneNotFound().WithPayload(&models.Error{ Message: "forum author not found" })
		}
	}
	return operations.NewForumGetOneOK().WithPayload(forum)
}


func ForumGetThreads(db *sql.DB, params operations.ForumGetThreadsParams) middleware.Responder {
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

	threads := models.Threads{}
	for rows.Next() {
		thread := &models.Thread{}
		err = rows.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Slug, &thread.Created)

		if err != nil {
			panic(err)
		}

		threads = append(threads, thread)
	}

	if contains := forumIsInDB(db, &params.Slug); !contains && len(threads) == 0 {
		return operations.NewForumGetThreadsNotFound().WithPayload(&models.Error{Message: "forum not found"})
	}

	return operations.NewForumGetThreadsOK().WithPayload(threads)
}


func ForumGetUsers(db *sql.DB, params operations.ForumGetUsersParams) middleware.Responder {
	if contains := forumIsInDB(db, &params.Slug); !contains {
		return operations.NewForumGetUsersNotFound().WithPayload(&models.Error{Message: "forum not found"})
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

	users := models.Users{}
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	return operations.NewForumGetUsersOK().WithPayload(users)
}