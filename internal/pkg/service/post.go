package service

import (
	"database/sql"
	"fmt"
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx/pgtype"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

func PostsCreate(res http.ResponseWriter, req *http.Request) {
	id := rand.Intn(10000)
	logrus.WithField("id", id).Warn(req.URL)

	db := db2.Connection

	slugOrID := req.URL.Query().Get(":slug_or_id")

	var exists bool
	var threadId int32
	var forumSlug string
	if exists, threadId, forumSlug = threadIsInDB(db, slugOrID); !exists {
		//return operations.NewPostsCreateNotFound().WithPayload(&internal.Error{Message: "slug or id not found"})
		models.ErrResponse(res, http.StatusNotFound, "slug or id not found")
		log.Println("PostsCreate", "slug or id not found")
		return
	}

	postsToCreate := models.Posts{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = postsToCreate.UnmarshalJSON(body)

	queryStatement := "INSERT INTO post (parent, author, message, thread) VALUES "
	var vals []interface{}

	rowIndex := 0
	for _, post := range postsToCreate {
		if inThread := postIsInThread(db, post.Parent, int64(threadId)); post.Parent != 0 && !inThread {
			//return operations.NewPostsCreateConflict().WithPayload(&internal.Error{Message: "parent is in another thread"})
			models.ErrResponse(res, http.StatusConflict, "parent is in another thread")
			log.Println("PostsCreate", "parent is in another thread")
			return
		}

		if exists := userIsInDB(db, post.Author); !exists {
			//return operations.NewPostsCreateNotFound().WithPayload(&internal.Error{Message: "author not found"})
			models.ErrResponse(res, http.StatusNotFound, "author not found")
			log.Println("PostsCreate", "author not found")
			return
		}

		queryStatement += fmt.Sprintf("(NULLIF($%d, 0), $%d, $%d, $%d),", (rowIndex*4)+1,
			(rowIndex*4)+2,
			(rowIndex*4)+3,
			(rowIndex*4)+4)
		vals = append(vals, post.Parent, post.Author, post.Message, threadId)
		rowIndex++
	}

	if rowIndex == 0 {
		posts := models.Posts{}
		//return operations.NewPostsCreateCreated().WithPayload(posts)
		models.ResponseObject(res, http.StatusCreated, posts)
		log.Println("PostsCreate", "rowIndex == 0", "len posts", len(posts))
		return
	}

	queryStatement = queryStatement[0 : len(queryStatement)-1]
	queryStatement += " RETURNING author, created, id, message, thread, coalesce(parent, 0)"

	logrus.WithField("id", id).Info(queryStatement)

	postsAdded := models.Posts{}
	rows, err := db.Query(queryStatement, vals...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.Author, &post.Created, &post.ID, &post.Message, &post.Thread, &post.Parent)

		if err != nil {
			panic(err)
		}

		post.Forum = forumSlug
		postsAdded = append(postsAdded, post)
	}

	//return operations.NewPostsCreateCreated().WithPayload(postsAdded)
	models.ResponseObject(res, http.StatusCreated, postsAdded)
	logrus.WithField("id", id).Println("PostsCreate", "OK", "len = ", len(postsAdded))
	return
}

func PostGetOne(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	ID := req.URL.Query().Get(":id")
	query := req.URL.Query()
	related := strings.Split(query.Get("related"), ",")

	post := &models.Post{}
	err := db.QueryRow(queryGetPostById, ID).Scan(&post.Author, &post.Created,
		&post.Forum, &post.ID, &post.Message, &post.Thread, &post.IsEdited)

	if err != nil {
		//if err == sql.ErrNoRows {
		//	//return operations.NewPostGetOneNotFound().WithPayload(&internal.Error{Message: "post not found"})
		//	models.ErrResponse(res, http.StatusNotFound, "post not found")
		//	return
		//}
		//panic(err)
		models.ErrResponse(res, http.StatusNotFound, "post not found")
		return
	}

	fullPost := &models.PostFull{Post: post}

	for _, relatedStr := range related {
		switch relatedStr {
		case "user":
			user := &models.User{}

			err := db.QueryRow(queryGetUserByNick, &post.Author).Scan(&user.Nickname, &user.Fullname,
				&user.Email, &user.About)

			if err != nil {
				if err != sql.ErrNoRows {
					panic(err)
				}
			}

			fullPost.Author = user
		case "forum":
			forum := &models.Forum{}

			err := db.QueryRow(queryGetFullForumBySlug, &post.Forum).Scan(&forum.Slug, &forum.User,
				&forum.Title, &forum.Posts, &forum.Threads)

			if err != nil {
				if err != sql.ErrNoRows {
					panic(err)
				}
			}

			fullPost.Forum = forum
		case "thread":
			thread := &models.Thread{}
			nullSlug := pgtype.Text{}

			fmt.Println(int64(post.Thread), strconv.Itoa(int(post.Thread)))
			//strconv.Itoa(post.Thread)

			err := db.QueryRow(queryGetThreadAndVoteCountByIdOrSlug, strconv.Itoa(int(post.Thread))).Scan(&thread.ID, &thread.Title,
				&thread.Author, &thread.Forum, &thread.Message, &nullSlug, &thread.Created, &thread.Votes)
			thread.Slug = nullSlug.String
			if err != nil {
				if err != sql.ErrNoRows {
					panic(err)
				}
			}

			fullPost.Thread = thread
		}
	}

	//return operations.NewPostGetOneOK().WithPayload(fullPost)
	models.ResponseObject(res, http.StatusOK, fullPost)
	return
}

func PostUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	ID := req.URL.Query().Get(":id")

	post := &models.Post{}
	err := db.QueryRow(queryGetPostById, ID).Scan(&post.Author, &post.Created,
		&post.Forum, &post.ID, &post.Message, &post.Thread, &post.IsEdited)

	if err != nil {
		//if err == sql.ErrNoRows {
		//	//return operations.NewPostUpdateNotFound().WithPayload(&internal.Error{Message: "post not found"})
		//	models.ErrResponse(res, http.StatusNotFound, "post not found")
		//	return
		//}
		//panic(err)
		models.ErrResponse(res, http.StatusNotFound, "post not found")
		return
	}

	newPost := models.Post{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = newPost.UnmarshalJSON(body)

	_, err = db.Exec(queryUpdatePost, newPost.Message, ID)

	if err != nil {
		panic(err)
	}

	updatedPost := &models.Post{}
	err = db.QueryRow(queryGetPostById, ID).Scan(&updatedPost.Author, &updatedPost.Created,
		&updatedPost.Forum, &updatedPost.ID, &updatedPost.Message,
		&updatedPost.Thread, &updatedPost.IsEdited)

	if err != nil {
		if err == sql.ErrNoRows {
			//return operations.NewPostUpdateNotFound().WithPayload(&internal.Error{Message: "post not found"})
			models.ErrResponse(res, http.StatusNotFound, "post not found")
			return
		}
		panic(err)
	}

	//return operations.NewPostUpdateOK().WithPayload(updatedPost)
	models.ResponseObject(res, http.StatusOK, updatedPost)
	return
}
