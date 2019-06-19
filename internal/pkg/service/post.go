package service

import (
	"database/sql"
	"fmt"
	"github.com/crueltycute/tech-db-forum/internal"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"net/http"
)

func PostsCreate(res http.ResponseWriter, req *http.Request) {
	var exists bool
	var threadId int32
	var forumSlug string
	if exists, threadId, forumSlug = threadIsInDB(db, params.SlugOrID); !exists {
		return operations.NewPostsCreateNotFound().WithPayload(&internal.Error{Message: "slug or id not found"})
	}

	queryStatement := "INSERT INTO post (parent, author, message, thread) VALUES "
	vals := []interface{}{}

	rowIndex := 0
	for _, post := range params.Posts {
		if inThread := postIsInThread(db, post.Parent, int64(threadId)); post.Parent != 0 && !inThread {
			return operations.NewPostsCreateConflict().WithPayload(&internal.Error{Message: "parent is in another thread"})
		}

		if exists := userIsInDB(db, post.Author); !exists {
			return operations.NewPostsCreateNotFound().WithPayload(&internal.Error{Message: "author not found"})
		}

		queryStatement += fmt.Sprintf("(NULLIF($%d, 0), $%d, $%d, $%d),", (rowIndex*4)+1,
			(rowIndex*4)+2,
			(rowIndex*4)+3,
			(rowIndex*4)+4)
		vals = append(vals, post.Parent, post.Author, post.Message, threadId)
		rowIndex++
	}

	if rowIndex == 0 {
		posts := internal.Posts{}
		return operations.NewPostsCreateCreated().WithPayload(posts)
	}

	queryStatement = queryStatement[0 : len(queryStatement)-1]
	queryStatement += " RETURNING author, created, id, message, thread, coalesce(parent, 0)"

	postsAdded := internal.Posts{}
	rows, err := db.Query(queryStatement, vals...)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		post := &internal.Post{}
		err := rows.Scan(&post.Author, &post.Created, &post.ID, &post.Message, &post.Thread, &post.Parent)

		if err != nil {
			panic(err)
		}

		post.Forum = forumSlug
		postsAdded = append(postsAdded, post)
	}

	return operations.NewPostsCreateCreated().WithPayload(postsAdded)
}

func PostGetOne(res http.ResponseWriter, req *http.Request) {
	post := &internal.Post{}
	err := db.QueryRow(queryGetPostById, &params.ID).Scan(&post.Author, &post.Created,
		&post.Forum, &post.ID, &post.Message,
		&post.Thread, &post.IsEdited)

	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewPostGetOneNotFound().WithPayload(&internal.Error{Message: "post not found"})
		}
		panic(err)
	}

	fullPost := &internal.PostFull{Post: post}

	for _, relatedStr := range params.Related {
		switch relatedStr {
		case "user":
			user := &internal.User{}

			err := db.QueryRow(queryGetUserByNick, &post.Author).Scan(&user.Nickname, &user.Fullname,
				&user.Email, &user.About)

			if err != nil {
				if err != sql.ErrNoRows {
					panic(err)
				}
			}

			fullPost.Author = user
		case "forum":
			forum := &internal.Forum{}

			err := db.QueryRow(queryGetFullForumBySlug, &post.Forum).Scan(&forum.Slug, &forum.User,
				&forum.Title, &forum.Posts, &forum.Threads)

			if err != nil {
				if err != sql.ErrNoRows {
					panic(err)
				}
			}

			fullPost.Forum = forum
		case "thread":
			thread := &internal.Thread{}

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

	return operations.NewPostGetOneOK().WithPayload(fullPost)
}

func PostUpdate(res http.ResponseWriter, req *http.Request) {
	post := &internal.Post{}
	err := db.QueryRow(queryGetPostById, &params.ID).Scan(&post.Author, &post.Created,
		&post.Forum, &post.ID, &post.Message,
		&post.Thread, &post.IsEdited)

	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewPostUpdateNotFound().WithPayload(&internal.Error{Message: "post not found"})
		}
		panic(err)
	}

	_, err = db.Exec(queryUpdatePost, &params.Post.Message, &params.ID)

	if err != nil {
		panic(err)
	}

	updatedPost := &internal.Post{}
	err = db.QueryRow(queryGetPostById, &params.ID).Scan(&updatedPost.Author, &updatedPost.Created,
		&updatedPost.Forum, &updatedPost.ID, &updatedPost.Message,
		&updatedPost.Thread, &updatedPost.IsEdited)

	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewPostUpdateNotFound().WithPayload(&internal.Error{Message: "post not found"})
		}
		panic(err)
	}

	return operations.NewPostUpdateOK().WithPayload(updatedPost)
}
