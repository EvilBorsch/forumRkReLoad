package repository

import (
	umodel "go-server-server-generated/src/user/models"
	"go-server-server-generated/src/utills"
)

func SaveUser(user umodel.User) error {
	conn := utills.GetConnection()
	query := `INSERT INTO "user" (nickname, fullname, about, email) VALUES ($1,$2,$3,$4)`
	_, err := conn.Exec(query, user.Nickname, user.Fullname, user.About, user.Email)
	return err
}

func GetUserByNickname(nickname string) (umodel.User, error) {
	conn := utills.GetConnection()
	query := `SELECT * from "user" WHERE nickname=$1`
	var user umodel.User
	err := conn.Get(&user, query, nickname)
	return user, err
}

func GetUserByNicknameAndEmail(nickname string, email string) ([]umodel.User, error) {
	conn := utills.GetConnection()
	query := `SELECT * from "user" WHERE nickname=$1 OR email=$2`
	var userlist []umodel.User
	err := conn.Select(&userlist, query, nickname, email)
	return userlist, err
}

func UpdateUser(oldNickname string, newUser umodel.User) (umodel.User, error) {
	conn := utills.GetConnection()
	query := `UPDATE "user" SET 
                about=COALESCE(NULLIF($1, ''), about),
                email=COALESCE(NULLIF($2, ''), email),
                fullname=COALESCE(NULLIF($3, ''), fullname) 
WHERE LOWER(nickname) = LOWER($4) RETURNING *`
	if newUser.Email == "" && newUser.About == "" && newUser.Fullname == "" {
		return GetUserByNickname(oldNickname)
	}

	var updatedUser umodel.User
	err := conn.Get(&updatedUser, query, newUser.About, newUser.Email, newUser.Fullname, oldNickname)
	return updatedUser, err

}

func GetUserByEmail(email string) (umodel.User, error) {
	conn := utills.GetConnection()
	query := `SELECT * from "user" WHERE email=$1`
	var user umodel.User
	err := conn.Get(&user, query, email)
	return user, err
}
