package service

import (
	"database/sql"
	"fmt"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"strings"
)

func ThreadCreate(db *sql.DB, params operations.ThreadCreateParams) middleware.Responder {
	var nickname string
	err := db.QueryRow(queryGetUserNickByNick, params.Thread.Author).Scan(&nickname)

	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewThreadCreateNotFound().WithPayload(&models.Error{Message: "forum author not found"})
		}
		panic(err)
	}

	var slug string
	err = db.QueryRow(queryGetForumSlugBySlug, params.Slug).Scan(&slug)

	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewThreadCreateNotFound().WithPayload(&models.Error{Message: "forum not found"})
		}
		panic(err)
	}

	thread := &models.Thread{}
	err = db.QueryRow(queryAddThread, nickname, slug,
					  params.Thread.Slug, params.Thread.Title,
					  params.Thread.Message, params.Thread.Created).Scan(&thread.ID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			existingThread := &models.Thread{}
			err := db.QueryRow(queryGetThreadBySlug, params.Thread.Slug).Scan(&existingThread.ID, &existingThread.Author,
																         &existingThread.Forum, &existingThread.Title,
																         &existingThread.Slug, &existingThread.Message,
																         &existingThread.Created)

			if err != nil {
				panic(err)
			}

			return operations.NewThreadCreateConflict().WithPayload(existingThread)
		}
	}

	err = db.QueryRow(queryGetThreadById, thread.ID).Scan(&thread.Author, &thread.Forum,
									  				 &thread.Title, &thread.Slug,
													 &thread.Message, &thread.Created)

	return operations.NewThreadCreateCreated().WithPayload(thread)
}


func ThreadVote(db *sql.DB, params operations.ThreadVoteParams) middleware.Responder {
	exists, threadId, _ := threadIsInDB(db, params.SlugOrID)
	if !exists {
		return operations.NewThreadVoteNotFound().WithPayload(&models.Error{Message: "thread not found"})
	}

	_, err := db.Exec(queryAddVote, threadId, &params.Vote.Nickname, &params.Vote.Voice)

	if err != nil {
		if strings.Contains(err.Error(), "vote_nickname_fkey") {
			return operations.NewThreadVoteNotFound().WithPayload(&models.Error{Message: "nickname not found"})
		}
		panic(err)
	}

	votedThread := &models.Thread{}

	err = db.QueryRow(queryGetThreadAndVoteCountById, &threadId).Scan(&votedThread.ID, &votedThread.Title, &votedThread.Author,
																 &votedThread.Forum, &votedThread.Message, &votedThread.Votes,
																 &votedThread.Slug, &votedThread.Created)
	if err != nil {
		panic(err)
	}

	return operations.NewThreadVoteOK().WithPayload(votedThread)
}


func ThreadGetOne(db *sql.DB, params operations.ThreadGetOneParams) middleware.Responder {
	thread := &models.Thread{}

	err := db.QueryRow(queryGetThreadAndVoteCountByIdOrSlug, &params.SlugOrID).Scan(&thread.ID, &thread.Title,
																					&thread.Author, &thread.Forum,
																					&thread.Message, &thread.Slug,
																					&thread.Created, &thread.Votes)

	if err != nil {
		if err == sql.ErrNoRows {
			return operations.NewThreadGetOneNotFound().WithPayload(&models.Error{Message: "thread does not exist"})
		}
		panic(err)
	}

	return operations.NewThreadGetOneOK().WithPayload(thread)
}


func ThreadGetPosts(db *sql.DB, params operations.ThreadGetPostsParams) middleware.Responder {
	exists, threadId, _ := threadIsInDB(db, params.SlugOrID)

	if !exists {
		return operations.NewThreadVoteNotFound().WithPayload(&models.Error{Message: "thread not found"})
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

		since = fmt.Sprintf("and post.id %s %d", comparisonSign, *params.Since)
		if *params.Sort == "tree" {
			since = fmt.Sprintf("and post.path %s (SELECT tree_post.path FROM post AS tree_post WHERE tree_post.id = %d)",
								comparisonSign, *params.Since)
		}
		if *params.Sort == "parent_tree" {
			since = fmt.Sprintf("and post_roots.path[1] %s (SELECT tree_post.path[1] FROM post AS tree_post WHERE tree_post.id = %d)",
								comparisonSign, *params.Since)
		}
	}

	queryStatement := `SELECT post.author, post.created, thread.forum, post.id, 
							  post.message, post.thread, coalesce(post.parent, 0)
							  FROM post
							  JOIN thread ON thread.id = post.thread
							  WHERE post.thread = $1 %s
							  ORDER BY (post.created, post.id) %s
							  LIMIT $2`

	query := fmt.Sprintf(queryStatement, since, order)

	if *params.Sort == "tree" {
		queryStatement = `SELECT post.author, post.created, thread.forum, post.id, 
						  		 post.message, post.thread, coalesce(post.parent, 0)
						  	     FROM post
						  		 JOIN thread ON thread.id = post.thread
						  		 WHERE post.thread = $1 %s
						  		 ORDER BY (post.path, post.created) %s
						  		 LIMIT $2`
		query = fmt.Sprintf(queryStatement, since, order)
	} else if *params.Sort == "parent_tree" {
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
		query = fmt.Sprintf(queryStatement, since, order, order)
	}

	rows, err := db.Query(query, threadId, *params.Limit)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	posts := models.Posts{}
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.Parent)

		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}
	return operations.NewThreadGetPostsOK().WithPayload(posts)
}


func ThreadUpdate(db *sql.DB, params operations.ThreadUpdateParams) middleware.Responder {
	exists, threadId, _ := threadIsInDB(db, params.SlugOrID)

	if !exists {
		return operations.NewThreadUpdateNotFound().WithPayload(&models.Error{Message: "thread not found"})
	}

	_, err := db.Exec(queryUpdateThread, &params.Thread.Title, &params.Thread.Message, &threadId)
	if err != nil {
		panic(err)
	}

	thread := &models.Thread{}
	err = db.QueryRow(queryGetThreadAndVoteCountByIdOrSlug, &params.SlugOrID).Scan(&thread.ID, &thread.Title,
																				   &thread.Author, &thread.Forum,
																				   &thread.Message, &thread.Slug,
																				   &thread.Created, &thread.Votes)

	if err != nil {
		panic(err)
	}

	return operations.NewThreadUpdateOK().WithPayload(thread)
}