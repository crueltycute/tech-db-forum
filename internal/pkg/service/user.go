package service

import (
	"github.com/crueltycute/tech-db-forum/internal"
	"github.com/crueltycute/tech-db-forum/restapi/operations"
	"net/http"
	"strings"
)

func UsersCreate(res http.ResponseWriter, req *http.Request) {
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

			existingUsers := internal.Users{}

			for rows.Next() {
				existingUser := &internal.User{}
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

func UsersGetOne(res http.ResponseWriter, req *http.Request) {
	nickname := params.Nickname

	rows, _ := db.Query(queryGetUserByNick, nickname)
	defer rows.Close()

	if rows.Next() {
		user := &internal.User{}
		rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		return operations.NewUserGetOneOK().WithPayload(user)
	}

	return operations.NewUserGetOneNotFound().WithPayload(&internal.Error{Message: "user not found"})
}

func UsersUpdate(res http.ResponseWriter, req *http.Request) {
	_, err := db.Exec(queryUpdateUser, &params.Profile.Fullname,
		&params.Profile.Email, &params.Profile.About,
		&params.Nickname)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return operations.NewUserUpdateConflict().WithPayload(&internal.Error{"cannot update user"})
		}
		return operations.NewUserUpdateNotFound().WithPayload(&internal.Error{Message: "user not found"})
	}

	rows, _ := db.Query(queryGetUserByNick, params.Nickname)
	defer rows.Close()

	if rows.Next() {
		user := &internal.User{}
		rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		return operations.NewUserUpdateOK().WithPayload(user)
	}

	return operations.NewUserUpdateNotFound().WithPayload(&internal.Error{Message: "updated user not found"})
}
