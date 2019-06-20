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

func PostsCreate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	slugOrID := req.URL.Query().Get(":slug_or_id")

	var exists bool
	var threadId int32
	var forumSlug string
	if exists, threadId, forumSlug = threadIsInDB(db, slugOrID); !exists {
		//return operations.NewPostsCreateNotFound().WithPayload(&internal.Error{Message: "slug or id not found"})
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
		}

		if exists := userIsInDB(db, post.Author); !exists {
			//return operations.NewPostsCreateNotFound().WithPayload(&internal.Error{Message: "author not found"})
		}

		queryStatement += fmt.Sprintf("(NULLIF($%d, 0), $%d, $%d, $%d),", (rowIndex*4)+1,
			(rowIndex*4)+2,
			(rowIndex*4)+3,
			(rowIndex*4)+4)
		vals = append(vals, post.Parent, post.Author, post.Message, threadId)
		rowIndex++
	}

	if rowIndex == 0 {
		//posts := models.Posts{}
		//return operations.NewPostsCreateCreated().WithPayload(posts)
	}

	queryStatement = queryStatement[0 : len(queryStatement)-1]
	queryStatement += " RETURNING author, created, id, message, thread, coalesce(parent, 0)"

	postsAdded := models.Posts{}
	rows, err := db.Query(queryStatement, vals...)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.Author, &post.Created, &post.ID, &post.Message, &post.Thread, &post.Parent)

		if err != nil {
			panic(err)
		}

		post.Forum = forumSlug
		postsAdded = append(postsAdded, post)
	}

	//return operations.NewPostsCreateCreated().WithPayload(postsAdded)
}

func PostGetOne(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	ID := req.URL.Query().Get(":id")
	query := req.URL.Query()
	related := strings.Split(query.Get("related"), ",")

	post := &models.Post{}
	err := db.QueryRow(queryGetPostById, ID).Scan(&post.Author, &post.Created,
		&post.Forum, &post.ID, &post.Message,
		&post.Thread, &post.IsEdited)

	if err != nil {
		if err == sql.ErrNoRows {
			//return operations.NewPostGetOneNotFound().WithPayload(&internal.Error{Message: "post not found"})
		}
		panic(err)
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

			err := db.QueryRow(queryGetThreadAndVoteCountByIdOrSlug, int64(post.Thread)).Scan(&thread.ID, &thread.Title,
				&thread.Author, &thread.Forum,
				&thread.Message, &thread.Slug,
				&thread.Created, &thread.Votes)

			if err != nil {
				if err != sql.ErrNoRows {
					panic(err)
				}
			}

			fullPost.Thread = thread
		}
	}

	//return operations.NewPostGetOneOK().WithPayload(fullPost)
}

func PostUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection

	ID := req.URL.Query().Get(":id")

	post := &models.Post{}
	err := db.QueryRow(queryGetPostById, ID).Scan(&post.Author, &post.Created,
		&post.Forum, &post.ID, &post.Message,
		&post.Thread, &post.IsEdited)

	if err != nil {
		if err == sql.ErrNoRows {
			//return operations.NewPostUpdateNotFound().WithPayload(&internal.Error{Message: "post not found"})
		}
		panic(err)
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
		}
		panic(err)
	}

	//return operations.NewPostUpdateOK().WithPayload(updatedPost)
}
