package service

import (
	"database/sql"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"strings"
)

const insertUser = `
	INSERT INTO Users (nickname, fullname, email, about) 
	VALUES ($1, $2, $3, $4)`

const getUserByNickOrEmail = `
	SELECT nickname, fullname, email, about 
	FROM Users WHERE email = $1 OR nickname = $2`

const getUserByNick = `
	SELECT nickname, fullname, email, about
	FROM Users WHERE nickname = $1`

const updateUser = `
	UPDATE Users SET fullname = COALESCE(NULLIF($1, ''), fullname), 
	email = COALESCE(NULLIF($2, ''), email), about = COALESCE(NULLIF($3, ''), about)
	WHERE nickname = $4`


func UsersCreate(db *sql.DB, params operations.UserCreateParams) middleware.Responder {
	var createdUser *models.User
	createdUser = params.Profile
	createdUser.Nickname = params.Nickname

	_, err := db.Exec(insertUser, &createdUser.Nickname, &createdUser.Fullname, &createdUser.Email, &createdUser.About)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			rows, err := db.Query(getUserByNickOrEmail, &createdUser.Email, &createdUser.Nickname)
			defer rows.Close()

			if err != nil {
				panic(err)
			}

			existingUsers := models.Users{}

			for rows.Next() {
				existingUser := &models.User{}
				rows.Scan(&existingUser.Nickname, &existingUser.Fullname, &existingUser.Email, &existingUser.About)
				existingUsers = append(existingUsers, existingUser)
			}

			return operations.NewUserCreateConflict().WithPayload(existingUsers)
		} else {
			panic(err)
		}
	}

	return operations.NewUserCreateCreated().WithPayload(createdUser)
}

func GetUserByNick(db *sql.DB, params operations.UserGetOneParams) middleware.Responder {
	nickname := params.Nickname

	rows, _ := db.Query(getUserByNick, nickname)
	defer rows.Close()

	if rows.Next() {
		user := &models.User{}
		rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		return operations.NewUserGetOneOK().WithPayload(user)
	}

	return operations.NewUserGetOneNotFound().WithPayload(&models.Error{ Message: "user not found" })
}

func UsersUpdate(db *sql.DB, params operations.UserUpdateParams) middleware.Responder {
	_, err := db.Exec(updateUser, &params.Profile.Fullname,
						&params.Profile.Email, &params.Profile.About,
						&params.Nickname)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return operations.NewUserUpdateConflict().WithPayload(&models.Error{"cannot update user"})
		}
		return operations.NewUserUpdateNotFound().WithPayload(&models.Error{ Message: "user not found" })
	}

	rows, _ := db.Query(getUserByNick, params.Nickname)
	defer rows.Close()

	if rows.Next() {
		user := &models.User{}
		rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		return operations.NewUserUpdateOK().WithPayload(user)
	}

	return operations.NewUserUpdateNotFound().WithPayload(&models.Error{ Message: "updated user not found" })
}