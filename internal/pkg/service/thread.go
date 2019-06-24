package service

import (
	"fmt"
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx/pgtype"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

func ThreadCreate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugName := req.URL.Query().Get(":slug")
	t := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = t.UnmarshalJSON(body)

	t.Forum = slugName
	//fmt.Println(t)

	var nickname string
	err := db.QueryRow(queryGetUserNickByNick, t.Author).Scan(&nickname)

	if err != nil {
		//if err == sql.ErrNoRows {
		//	//return operations.NewThreadCreateNotFound().WithPayload(&internal.Error{Message: "forum author not found"})
		//	models.ErrResponse(res, http.StatusNotFound, "forum author not found")
		//	return
		//}
		//panic(err)
		models.ErrResponse(res, http.StatusNotFound, "forum author not found")
		return
	}

	var slug string
	err = db.QueryRow(queryGetForumSlugBySlug, t.Forum).Scan(&slug)

	if err != nil {
		//if err == sql.ErrNoRows {
		//	//return operations.NewThreadCreateNotFound().WithPayload(&internal.Error{Message: "forum not found"})
		//	models.ErrResponse(res, http.StatusNotFound, "forum not found")
		//	return
		//}
		//panic(err)
		fmt.Println(err, slug)
		models.ErrResponse(res, http.StatusNotFound, "forum not found th")
		return
	}

	thread := &models.Thread{}
	err = db.QueryRow(queryAddThread, nickname, slug, t.Slug, t.Title, t.Message, t.Created).Scan(&thread.ID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingThread := &models.Thread{}
			err := db.QueryRow(queryGetThreadBySlug, t.Slug).Scan(&existingThread.ID, &existingThread.Author,
				&existingThread.Forum, &existingThread.Title,
				&existingThread.Slug, &existingThread.Message,
				&existingThread.Created)

			if err != nil {
				panic(err)
			}

			//return operations.NewThreadCreateConflict().WithPayload(existingThread)
			models.ResponseObject(res, http.StatusConflict, existingThread)
			return
		}
	}

	err = db.QueryRow(queryGetThreadById, thread.ID).Scan(&thread.Author, &thread.Forum,
		&thread.Title, &thread.Slug, &thread.Message, &thread.Created)

	//return operations.NewThreadCreateCreated().WithPayload(thread)
	models.ResponseObject(res, http.StatusCreated, thread)
	return
}

func ThreadVote(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	SlugOrID := req.URL.Query().Get(":slug_or_id")

	exists, threadId, _ := threadIsInDB(db, SlugOrID)
	if !exists {
		//return operations.NewThreadVoteNotFound().WithPayload(&internal.Error{Message: "thread not found"})
		models.ErrResponse(res, http.StatusNotFound, "thread not found")
		return
	}

	voteToCreate := models.Vote{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = voteToCreate.UnmarshalJSON(body)

	_, err := db.Exec(queryAddVote, threadId, voteToCreate.Nickname, voteToCreate.Voice)

	if err != nil {
		if strings.Contains(err.Error(), "vote_nickname_fkey") {
			//return operations.NewThreadVoteNotFound().WithPayload(&internal.Error{Message: "nickname not found"})
			models.ErrResponse(res, http.StatusNotFound, "nickname not found")
			return
		}
		panic(err)
	}

	votedThread := &models.Thread{}
	nullableSlug := pgtype.Text{}

	err = db.QueryRow(queryGetThreadAndVoteCountById, &threadId).Scan(&votedThread.ID, &votedThread.Title, &votedThread.Author,
		&votedThread.Forum, &votedThread.Message, &votedThread.Votes,
		&nullableSlug, &votedThread.Created)

	votedThread.Slug = nullableSlug.String
	if err != nil {
		panic(err)
	}

	//return operations.NewThreadVoteOK().WithPayload(votedThread)
	models.ResponseObject(res, http.StatusOK, votedThread)
	return
}

func ThreadGetOne(res http.ResponseWriter, req *http.Request) {
	id := rand.Intn(10000)
	logrus.WithField("id", id).Warn(req.URL)

	db := db2.Connection
	slugOrID := req.URL.Query().Get(":slug_or_id")

	thread := &models.Thread{}
	nullSlug := pgtype.Text{}

	err := db.QueryRow(queryGetThreadAndVoteCountByIdOrSlug, slugOrID).Scan(&thread.ID, &thread.Title,
		&thread.Author, &thread.Forum, &thread.Message, &nullSlug, &thread.Created, &thread.Votes)
	thread.Slug = nullSlug.String
	if err != nil {
		//if err == sql.ErrNoRows {
		//	//return operations.NewThreadGetOneNotFound().WithPayload(&internal.Error{Message: "thread does not exist"})
		//	models.ErrResponse(res, http.StatusNotFound, "thread does not exist")
		//	return
		//}
		//panic(err)
		models.ErrResponse(res, http.StatusNotFound, "thread does not exist")
		logrus.WithField("id", id).Warn("thread does not exist")
		return
	}

	//return operations.NewThreadGetOneOK().WithPayload(thread)
	models.ResponseObject(res, http.StatusOK, thread)
	logrus.WithField("id", id).Warn(thread)
	return
}

func ThreadGetPosts(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	slugOrID := req.URL.Query().Get(":slug_or_id")

	exists, threadId, _ := threadIsInDB(db, slugOrID)

	if !exists {
		//return operations.NewThreadVoteNotFound().WithPayload(&internal.Error{Message: "thread not found"})
		models.ErrResponse(res, http.StatusNotFound, "thread not found")
		return
	}

	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since, _ := strconv.Atoi(query.Get("since"))
	sort := query.Get("sort")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	orderDB := ""
	if desc == true {
		orderDB = "DESC"
	} else {
		orderDB = "ASC"
	}

	sinceDB := ""
	if since != 0 {
		comparisonSign := ">"
		if desc == true {
			comparisonSign = "<"
		}

		sinceDB = fmt.Sprintf("and post.id %s %d", comparisonSign, since)
		if sort == "tree" {
			sinceDB = fmt.Sprintf("and post.path %s (SELECT tree_post.path FROM post AS tree_post WHERE tree_post.id = %d)",
				comparisonSign, since)
		}
		if sort == "parent_tree" {
			sinceDB = fmt.Sprintf("and post_roots.path[1] %s (SELECT tree_post.path[1] FROM post AS tree_post WHERE tree_post.id = %d)",
				comparisonSign, since)
		}
	}

	queryStatement := `SELECT post.author, post.created, thread.forum, post.id, 
							  post.message, post.thread, coalesce(post.parent, 0)
							  FROM post
							  JOIN thread ON thread.id = post.thread
							  WHERE post.thread = $1 %s
							  ORDER BY (post.created, post.id) %s
							  LIMIT $2`

	queryDB := fmt.Sprintf(queryStatement, sinceDB, orderDB)

	if sort == "tree" {
		queryStatement = `SELECT post.author, post.created, thread.forum, post.id, 
						  		 post.message, post.thread, coalesce(post.parent, 0)
						  	     FROM post
						  		 JOIN thread ON thread.id = post.thread
						  		 WHERE post.thread = $1 %s
						  		 ORDER BY (post.path, post.created) %s
						  		 LIMIT $2`
		queryDB = fmt.Sprintf(queryStatement, sinceDB, orderDB)
	} else if sort == "parent_tree" {
		queryStatement = `SELECT post.author, post.created, thread.forum, post.id, 
								 post.message, post.thread, coalesce(post.parent, 0)
								 FROM post
								 JOIN thread ON thread.id = post.thread
								 WHERE post.thread = $1 AND post.path[1] IN (
								 	SELECT post_roots.id
									FROM post as post_roots
									WHERE post_roots.id = post_roots.path[1] AND post_roots.thread = post.thread %s
									ORDER BY post_roots.path[1] %s
									LIMIT $2
								 )
								 ORDER BY post.path[1] %s, post.path`
		queryDB = fmt.Sprintf(queryStatement, sinceDB, orderDB, orderDB)
	}

	rows, err := db.Query(queryDB, threadId, limit)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	posts := models.Posts{}
	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.Parent)

		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}

	//return operations.NewThreadGetPostsOK().WithPayload(posts)
	models.ResponseObject(res, http.StatusOK, posts)
	return
}

func ThreadUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	slugOrID := req.URL.Query().Get(":slug_or_id")

	exists, threadId, _ := threadIsInDB(db, slugOrID)

	if !exists {
		//return operations.NewThreadUpdateNotFound().WithPayload(&internal.Error{Message: "thread not found"})
		models.ErrResponse(res, http.StatusNotFound, "thread not found")
		return
	}

	t := models.Thread{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = t.UnmarshalJSON(body)

	_, err := db.Exec(queryUpdateThread, &t.Title, &t.Message, &threadId)
	if err != nil {
		panic(err)
	}

	thread := &models.Thread{}
	err = db.QueryRow(queryGetThreadAndVoteCountByIdOrSlug, slugOrID).Scan(&thread.ID, &thread.Title,
		&thread.Author, &thread.Forum,
		&thread.Message, &thread.Slug,
		&thread.Created, &thread.Votes)

	if err != nil {
		panic(err)
	}

	//return operations.NewThreadUpdateOK().WithPayload(thread)
	models.ResponseObject(res, http.StatusOK, thread)
	return
}
