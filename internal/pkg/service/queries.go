package service

import (
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
	"strconv"
)

// service
const queryClearDB = `
	TRUNCATE TABLE
	Users, Forum, ForumUser, 
	Post, Thread, Vote CASCADE`

const queryGetStatus = `
	SELECT
	(SELECT COUNT(*) FROM users),
	(SELECT COUNT(*) FROM forum),
	(SELECT COUNT(*) FROM thread),
	(SELECT COUNT(*) FROM post)`


// forum
const queryAddForum = `
	INSERT INTO Forum (title, forumUser, slug) 
	VALUES ($1, $2, $3)`

const queryGetForumBySlug = `
	SELECT title, forumUser, slug
	FROM Forum WHERE slug = $1`

const queryGetUserNickByNick = `
	SELECT nickname FROM Users WHERE nickname = $1`

const queryGetForumSlugBySlug = `
	SELECT slug FROM Forum WHERE slug = $1`


// user
const queryAddUser = `
	INSERT INTO Users (nickname, fullname, about, email) 
	VALUES ($1, $2, $3, $4)`

const queryGetUserByNickOrEmail = `
	SELECT nickname, fullname, about, email 
	FROM Users WHERE nickname = $1 OR email = $2`

const queryGetUserByNick = `
	SELECT nickname, fullname, about, email
	FROM Users WHERE nickname = $1`

const queryUpdateUser = `
	UPDATE Users SET fullname=COALESCE(NULLIF($1, ''), fullname),
	about=COALESCE(NULLIF($2, ''), about),
	email=COALESCE(NULLIF($3, ''), email)
	WHERE nickname=$4`


// thread
const queryAddThread = `
	INSERT INTO thread(title, author, forum, message, slug, created)
	VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6) RETURNING id`

const queryGetThreadBySlug = `
	SELECT id, title, author, forum, message, slug, created 
	FROM thread WHERE slug = $1`

const queryGetThreadById = `
	SELECT id, title, author, forum, message, coalesce(slug, ''), created 
	FROM thread WHERE id = $1`

const queryAddVote = `
	INSERT INTO Vote(threadId, nickname, voice) VALUES ($1, $2, $3)
	ON CONFLICT ON CONSTRAINT unique_vote DO UPDATE
	SET voice = EXCLUDED.voice;`

const queryGetThreadWithVotesById = `
	SELECT id, title, author, forum, message, votes, coalesce(slug, ''), created
	FROM thread WHERE id = $1`

const queryUpdateThread = `
	UPDATE thread 
	SET title=COALESCE(NULLIF($1, ''), title),
	message=COALESCE(NULLIF($2, ''), message)
	WHERE id=$3`


// post
const queryUpdatePost = `
	UPDATE post
	SET message = COALESCE(NULLIF($1, ''), message)
	WHERE id = $2`


// existence check
func forumExists(db dbOrConn, slug string) bool {
	forum := &models.Forum{}

	err := db.QueryRow("SELECT title, forumUser, slug FROM Forum WHERE slug = $1", slug).Scan(&forum.Title, &forum.User, &forum.Slug)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false
		}
		panic(err)
	}

	return true
}

func userExists(db txOrDb, nickname string) bool {
	user := &models.User{}

	err := db.QueryRow("SELECT nickname FROM Users WHERE nickname = $1", nickname).Scan(&user.Nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func postExistsInThread(db txOrDb, postId, threadId int64) bool {
	var id int64
	err := db.QueryRow(`
		SELECT id FROM post
		WHERE id = $1 and thread = $2`, postId, threadId).Scan(&id)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return false
		}
		panic(err)
	}
	
	return id == postId
}


// getters
func getForumBySlug(db dbOrConn, slug string) (*models.Forum, error) {
	forum := &models.Forum{}

	err := db.QueryRow(`
		SELECT title, forumUser, slug, posts, threads
		FROM Forum WHERE slug = $1`, slug).Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)

	return forum, err
}

func getUserByNickname(db dbOrConn, nickname string) (*models.User, error) {
	user := &models.User{}
	
	err := db.QueryRow(`
		SELECT nickname, fullname, about, email
		FROM Users
		WHERE nickname = $1`, nickname).Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
	
	return user, err
}

func getThreadBySlugOrId(db txOrDb, slugOrId string) (*models.Thread, error) {
	if id, err := strconv.Atoi(slugOrId); err == nil {
		return getThreadById(db, id)
	}
	
	return getThreadBySlug(db, slugOrId)
}

func getThreadById(db txOrDb, id int) (*models.Thread, error) {
	thread := &models.Thread{}
	
	err := db.QueryRow(`
		SELECT id, title, author, forum, message, coalesce(slug, ''), created, votes
		FROM thread WHERE id = $1`, id).Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum,
		&thread.Message, &thread.Slug, &thread.Created, &thread.Votes)

	return thread, err
}

func getThreadBySlug(db txOrDb, slug string) (*models.Thread, error) {
	thread := &models.Thread{}
	
	err := db.QueryRow(`
		SELECT id, title, author, forum, message, coalesce(slug, ''), created, votes
		FROM thread WHERE slug = $1`, slug).Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum,
		&thread.Message, &thread.Slug, &thread.Created, &thread.Votes)

	return thread, err
}

func getPostById(db dbOrConn, id int64) (*models.Post, error) {
	post := &models.Post{}
	
	err := db.QueryRow(`
		SELECT author, created, forum, id, message, thread, coalesce(isedited, FALSE), coalesce(parent, 0)
		FROM Post WHERE id = $1`, id).Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.IsEdited, &post.Parent)

	return post, err
}


func increasePostsCount(tx txOrDb, forumSlug string, count int) error {
	_, err := tx.Exec(`
		UPDATE forum
		SET posts = posts + $1
		WHERE slug = $2;`, count, forumSlug)
	
	return err
}