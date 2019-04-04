package service

import (
	"database/sql"
	"github.com/crueltycute/tech-db-forum/models"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"strings"
)

func UsersCreate(db *sql.DB, params operations.UserCreateParams) middleware.Responder {
	user := params.Profile
	user.Nickname = params.Nickname

	_, err := db.Exec(queryAddUser, &user.Nickname, &user.Fullname, &user.Email, &user.About)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			rows, err := db.Query(queryGetUserByNickOrEmail, &user.Email, &user.Nickname)
			defer rows.Close()

			if err != nil {
				panic(err)
			}

			existingUsers := models.Users{}

			for rows.Next() {
				existingUser := &models.User{}
				err = rows.Scan(&existingUser.Nickname, &existingUser.Fullname, &existingUser.Email, &existingUser.About)

				if err != nil {
					panic(err)
				}

				existingUsers = append(existingUsers, existingUser)
			}

			return operations.NewUserCreateConflict().WithPayload(existingUsers)
		}
		panic(err)
	}

	return operations.NewUserCreateCreated().WithPayload(user)
}


func UsersGetOne(db *sql.DB, params operations.UserGetOneParams) middleware.Responder {
	nickname := params.Nickname

	rows, _ := db.Query(queryGetUserByNick, nickname)
	defer rows.Close()

	if rows.Next() {
		user := &models.User{}
		rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		return operations.NewUserGetOneOK().WithPayload(user)
	}

	return operations.NewUserGetOneNotFound().WithPayload(&models.Error{ Message: "user not found" })
}


func UsersUpdate(db *sql.DB, params operations.UserUpdateParams) middleware.Responder {
	_, err := db.Exec(queryUpdateUser, &params.Profile.Fullname,
				      &params.Profile.Email, &params.Profile.About,
					  &params.Nickname)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return operations.NewUserUpdateConflict().WithPayload(&models.Error{"cannot update user"})
		}
		return operations.NewUserUpdateNotFound().WithPayload(&models.Error{ Message: "user not found" })
	}

	rows, _ := db.Query(queryGetUserByNick, params.Nickname)
	defer rows.Close()

	if rows.Next() {
		user := &models.User{}
		rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		return operations.NewUserUpdateOK().WithPayload(user)
	}

	return operations.NewUserUpdateNotFound().WithPayload(&models.Error{ Message: "updated user not found" })
}