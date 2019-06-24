package service

import (
	"fmt"
	"github.com/jackc/pgx/pgtype"
	"log"
	"strconv"

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
		//if err == sql.ErrNoRows {
		//	//return operations.NewForumCreateNotFound().WithPayload(&models.Error{Message: "forum author not found"})
		//	models.ErrResponse(res, http.StatusNotFound, "forum author not found")
		//	return
		//}
		//panic(err)
		models.ErrResponse(res, http.StatusNotFound, "forum author not found")
		return
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
			models.ResponseObject(res, http.StatusConflict, existingForum)
			return
		}
		//return operations.NewForumCreateNotFound().WithPayload(&internal.Error{Message: "forum author not found"})
		models.ErrResponse(res, http.StatusNotFound, "forum author not found")
		return
	}

	createdForum := &models.Forum{}
	err = db.QueryRow(queryGetForumBySlug, &f.Slug).Scan(&createdForum.Slug, &createdForum.User, &createdForum.Title)

	if err != nil {
		panic(err)
	}

	//return operations.NewForumCreateCreated().WithPayload(createdForum)
	models.ResponseObject(res, http.StatusCreated, createdForum)
	return
}

func ForumGetOne(res http.ResponseWriter, req *http.Request) {
	slug := req.URL.Query().Get(":slug")
	db := db2.Connection

	forum := &models.Forum{}
	err := db.QueryRow(queryGetFullForumBySlug, slug).Scan(&forum.Slug, &forum.User, &forum.Title, &forum.Posts, &forum.Threads)
	if err != nil {
		//if err == sql.ErrNoRows {
		//	//return operations.NewForumGetOneNotFound().WithPayload(&internal.Error{Message: "forum author not found"})
		//	models.ErrResponse(res, http.StatusNotFound, "forum author not found")
		//	return
		//}
		models.ErrResponse(res, http.StatusNotFound, "forum author not found")
		log.Println("ForumGetOne", err)
		return
	}
	//return operations.NewForumGetOneOK().WithPayload(forum)
	models.ResponseObject(res, http.StatusOK, forum)
	return
}

func ForumGetThreads(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")
	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	orderDB := ""
	if desc {
		orderDB = "DESC"
	} else {
		orderDB = "ASC"
	}

	sinceDB := ""
	if since != "" {
		if desc == true {
			sinceDB = fmt.Sprintf("and created <= '%s'", since)
		} else {
			sinceDB = fmt.Sprintf("and created >= '%s'", since)
		}
	}

	queryStatement := `SELECT T.id, T.title, T.author, F.slug, T.message, T.slug, T.created
					   FROM Thread as T JOIN Forum as F on T.forum = F.slug
					   WHERE F.slug = $1 %s ORDER BY created %s`

	if limit > 0 {
		queryStatement = fmt.Sprintf("%s LIMIT %d", queryStatement, limit)
	}

	queryDB := fmt.Sprintf(queryStatement, sinceDB, orderDB)

	rows, err := db.Query(queryDB, slugName)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	threads := models.Threads{}
	nullSlug := pgtype.Text{}
	for rows.Next() {
		thread := &models.Thread{}
		err = rows.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &nullSlug, &thread.Created)
		thread.Slug = nullSlug.String
		if err != nil {
			panic(err)
		}

		threads = append(threads, thread)
	}

	if contains := forumIsInDB(db, &slugName); !contains && len(threads) == 0 {
		//return operations.NewForumGetThreadsNotFound().WithPayload(&internal.Error{Message: "forum not found"})
		models.ErrResponse(res, http.StatusNotFound, "forum not found")
		return
	}

	//return operations.NewForumGetThreadsOK().WithPayload(threads)
	models.ResponseObject(res, http.StatusOK, threads)
	return
}

func ForumGetUsers(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")

	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if contains := forumIsInDB(db, &slugName); !contains {
		//return operations.NewForumGetUsersNotFound().WithPayload(&internal.Error{Message: "forum not found"})
		models.ErrResponse(res, http.StatusNotFound, "forum not found")
		return
	}

	order := ""
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}

	sinceQuery := ""
	if since != "" {
		comparisonSign := ">"
		if desc == true {
			comparisonSign = "<"
		}
		sinceQuery = fmt.Sprintf("and FU.nickname %s '%s'", comparisonSign, since)
	}

	queryStatement := `SELECT U.nickname, U.fullname, U.about, U.email
					   FROM ForumUser AS FU
					   JOIN Users as U ON FU.nickname = U.nickname
					   WHERE FU.slug = $1 %s
					   ORDER BY U.nickname %s`

	if limit > 0 {
		queryStatement = fmt.Sprintf("%s LIMIT %d", queryStatement, limit)
	}

	queryDB := fmt.Sprintf(queryStatement, sinceQuery, order)

	rows, err := db.Query(queryDB, slugName)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	users := models.Users{}
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	//return operations.NewForumGetUsersOK().WithPayload(users)
	models.ResponseObject(res, http.StatusOK, users)
	return
}
