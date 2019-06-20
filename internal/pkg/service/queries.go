package service

import (
	"database/sql"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
)

// service
const queryClearDB = `
	TRUNCATE TABLE Users CASCADE;
	TRUNCATE TABLE Forum CASCADE;
	TRUNCATE TABLE ForumUser CASCADE;
	TRUNCATE TABLE Post CASCADE;
	TRUNCATE TABLE Thread CASCADE;
	TRUNCATE TABLE VoteCount CASCADE;
	TRUNCATE TABLE Vote CASCADE;`

const queryGetStatus = `
 SELECT
  (SELECT COUNT(*) FROM Users),
  (SELECT COUNT(*) FROM Forum),
  (SELECT COUNT(*) FROM Thread),
  (SELECT COUNT(*) FROM Post);`

// forum
const queryAddForum = `
	INSERT INTO Forum (slug, forumUser, title) 
	VALUES ($1, $2, $3)`

const queryGetForumBySlug = `
	SELECT slug, forumUser, title
	FROM Forum WHERE slug = $1`

const queryGetFullForumBySlug = `
	SELECT slug, forumUser, title, posts, threads
	FROM Forum WHERE slug = $1`

const queryGetUserNickByNick = `
	SELECT nickname FROM Users WHERE nickname = $1`

const queryGetForumSlugBySlug = `
	SELECT slug FROM Forum WHERE slug = $1`

// user
const queryAddUser = `
	INSERT INTO Users (nickname, fullname, email, about) 
	VALUES ($1, $2, $3, $4)`

const queryGetUserByNickOrEmail = `
	SELECT nickname, fullname, email, about 
	FROM Users WHERE email = $1 OR nickname = $2`

const queryGetUserByNick = `
	SELECT nickname, fullname, email, about
	FROM Users WHERE nickname = $1`

const queryUpdateUser = `
	UPDATE Users SET fullname = COALESCE(NULLIF($1, ''), fullname), 
	email = COALESCE(NULLIF($2, ''), email), about = COALESCE(NULLIF($3, ''), about)
	WHERE nickname = $4`

// thread
const queryAddThread = `
	INSERT INTO Thread (author, forum, slug, title, message, created) 
	VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6) RETURNING id`

const queryGetThreadBySlug = `
	SELECT id, author, forum, title, slug, message, created
	FROM Thread WHERE slug = $1`

const queryGetThreadById = `
	SELECT author, forum, title, COALESCE(slug, ''), message, created 
	FROM Thread WHERE id = $1`

const queryAddVote = `
	INSERT INTO Vote (threadID, nickname, voice) VALUES ($1, $2, $3)
	ON CONFLICT ON CONSTRAINT unique_vote DO UPDATE
	SET voice = EXCLUDED.voice;`

const queryGetThreadAndVoteCountById = `
	SELECT id, title, author, forum, message, coalesce(voteCount.count, 0), slug, created
	FROM thread
	LEFT JOIN voteCount ON voteCount.threadID = id
	WHERE id = $1`

const queryGetThreadAndVoteCountByIdOrSlug = `
	SELECT id, title, author, forum, message, slug, created, coalesce(voteCount.count, 0)
	FROM thread
	LEFT JOIN voteCount ON voteCount.threadID = thread.id
	WHERE id::citext = $1 or slug = $1`

const queryUpdateThread = `
	UPDATE thread 
	SET 
	title=COALESCE(NULLIF($1, ''), title),
	message=COALESCE(NULLIF($2, ''), message)
	WHERE id=$3`

const queryGetThreadByIdOrSlug = `
	SELECT id, forum FROM thread WHERE id::citext = $1 or slug = $1`

// post
const queryGetPostById = `
	SELECT author, created, forum, id, message, thread, coalesce(isEdited, FALSE)
	FROM Post WHERE id = $1`

const queryUpdatePost = `
	UPDATE post
	SET message = COALESCE(NULLIF($1, ''), message)
	WHERE id = $2`

const queryGetPostByIdAndThread = `
	SELECT id FROM post
	WHERE id = $1 and thread = $2`

// existence check
func forumIsInDB(db *pgx.ConnPool, forumSlug *string) bool {
	scannedSlug := ""
	err := db.QueryRow(queryGetForumSlugBySlug, &forumSlug).Scan(&scannedSlug)

	if err != nil {
		//if err == sql.ErrNoRows {
		//	return false
		//}
		//panic(err)
		return false
	}

	return true
}

func userIsInDB(db *pgx.ConnPool, nickname string) bool {
	var userNickname string
	err := db.QueryRow(queryGetUserNickByNick, nickname).Scan(&userNickname)

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}

	return true
}

func threadIsInDB(db *pgx.ConnPool, slugOrId string) (bool, int32, string) {
	thread := &models.Thread{}
	err := db.QueryRow(queryGetThreadByIdOrSlug, slugOrId).Scan(&thread.ID, &thread.Slug)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, ""
		}
		panic(err)
	}
	return true, thread.ID, thread.Slug
}

func postIsInThread(db *pgx.ConnPool, postId, threadId int64) bool {
	post := &models.Post{}
	err := db.QueryRow(queryGetPostByIdAndThread, postId, threadId).Scan(&post.ID)

	if err != nil {
		//if err == sql.ErrNoRows {
		//	return false
		//}
		//panic(err)
		return false
	}
	return post.ID == postId
}
