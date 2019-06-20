package service

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"io/ioutil"
	"net/http"
	"strings"
)

func UsersCreate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	profile := models.User{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = profile.UnmarshalJSON(body)

	user := profile
	user.Nickname = nickname

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

			//return operations.NewUserCreateConflict().WithPayload(existingUsers)
			models.ResponseObject(res, http.StatusConflict, existingUsers)
			return
		}
		panic(err)
	}

	//return operations.NewUserCreateCreated().WithPayload(user)
	models.ResponseObject(res, http.StatusCreated, user)
	return
}

func UsersGetOne(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	//nickname := params.Nickname

	rows, _ := db.Query(queryGetUserByNick, nickname)
	defer rows.Close()

	if rows.Next() {
		user := &models.User{}
		_ = rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		//return operations.NewUserGetOneOK().WithPayload(user)
		models.ResponseObject(res, http.StatusOK, user)
		return
	}

	//return operations.NewUserGetOneNotFound().WithPayload(&internal.Error{Message: "user not found"})
	models.ErrResponse(res, http.StatusNotFound, "user not found")
	return
}

func UsersUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	u := models.User{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = u.UnmarshalJSON(body)

	_, err := db.Exec(queryUpdateUser, &u.Fullname, &u.Email, &u.About, &nickname)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			//return operations.NewUserUpdateConflict().WithPayload(&internal.Error{"cannot update user"})
			models.ErrResponse(res, http.StatusConflict, "cannot update user")
			return
		}
		//return operations.NewUserUpdateNotFound().WithPayload(&internal.Error{Message: "user not found"})
		models.ErrResponse(res, http.StatusNotFound, "user not found")
		return
	}

	rows, _ := db.Query(queryGetUserByNick, nickname)
	defer rows.Close()

	if rows.Next() {
		user := &models.User{}
		_ = rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		//return operations.NewUserUpdateOK().WithPayload(user)
		models.ResponseObject(res, http.StatusOK, user)
		return
	}

	//return operations.NewUserUpdateNotFound().WithPayload(&internal.Error{Message: "updated user not found"})
	models.ErrResponse(res, http.StatusNotFound, "thread not found")
	return
}
