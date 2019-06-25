package service

import (
	db2 "github.com/crueltycute/tech-db-forum/internal/app/db"
	"github.com/crueltycute/tech-db-forum/internal/models"
	"github.com/jackc/pgx"
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

	profile.Nickname = nickname

	_, err := db.Exec(queryAddUser, profile.Nickname, profile.Fullname, profile.About, profile.Email)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			rows, err := db.Query(queryGetUserByNickOrEmail, profile.Nickname, profile.Email)
			defer rows.Close()

			if err != nil {
				panic(err)
			}

			users := models.Users{}
			for rows.Next() {
				user := &models.User{}
				err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
				if err != nil {
					panic(err)
				}
				users = append(users, user)
			}
			models.ResponseObject(res, http.StatusConflict, users)
			return
		}
		panic(err)
	}

	models.ResponseObject(res, http.StatusCreated, profile)
	return
}

func UsersGetOne(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	user, err := getUserByNickname(db, nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			models.ErrResponse(res, http.StatusNotFound, "user not found")
			return
		}
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, user)
	return
}

func UsersUpdate(res http.ResponseWriter, req *http.Request) {
	db := db2.Connection
	nickname := req.URL.Query().Get(":nickname")

	u := models.User{}
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	_ = u.UnmarshalJSON(body)

	rows, err := db.Exec(queryUpdateUser, u.Fullname, u.About, u.Email, nickname)
	//defer rows.Close()

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			models.ErrResponse(res, http.StatusConflict, "cannot update user")
			return
		}
		panic(err)
	}

	count := rows.RowsAffected()
	if count == 0 {
		models.ErrResponse(res, http.StatusNotFound, "user not found")
		return
	}

	updatedData := &models.User{}
	err = db.QueryRow(queryGetUserByNick, nickname).Scan(&updatedData.Nickname, &updatedData.Fullname, &updatedData.About, &updatedData.Email)

	if err != nil {
		panic(err)
	}

	models.ResponseObject(res, http.StatusOK, updatedData)
	return
}
