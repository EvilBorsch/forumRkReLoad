package delivery

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	umodel "go-server-server-generated/src/user/models"
	urepo "go-server-server-generated/src/user/repository"
	"go-server-server-generated/src/utills"
	"io/ioutil"
	"net/http"
)

var badStaff = errors.New("bad data")

func fetchUser(r *http.Request) (umodel.User, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return umodel.User{}, badStaff
	}
	var user umodel.User
	err = json.Unmarshal(data, &user)
	return user, nil

}

func UserCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	nickname := mux.Vars(r)["nickname"]
	user, err := fetchUser(r)
	if err != nil {
		utills.SendServerError("can Json Data", http.StatusConflict, w)
		return
	}
	user.Nickname = nickname
	err = urepo.SaveUser(user)

	if err != nil {
		userList, err := urepo.GetUserByNicknameAndEmail(user.Nickname, user.Email)
		if err != nil {

			utills.SendServerError("error when try find users with this email and nick", http.StatusConflict, w)

		}
		utills.SendAnswerWithCode(userList, http.StatusConflict, w)
		return
	}

	utills.SendAnswerWithCode(user, http.StatusCreated, w)

}

func UserGetOne(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	nickname := mux.Vars(r)["nickname"]
	user, err := urepo.GetUserByNickname(nickname)

	if err != nil {
		utills.SendServerError("Can't find user by nickname: "+nickname, http.StatusNotFound, w)
		return
	}
	utills.SendOKAnswer(user, w)
}

func addEmptyDataWithOldData(newUser umodel.User, oldUserData umodel.User) umodel.User {
	if newUser.Email == "" {

		newUser.Email = oldUserData.Email
	}

	if newUser.Fullname == "" {
		newUser.Fullname = oldUserData.Fullname
	}

	if newUser.About == "" {
		newUser.About = oldUserData.About
	}
	return newUser
}

func checkIsEmpty(newUser umodel.User) bool {
	if newUser.Email == "" || newUser.Nickname == "" || newUser.About == "" {
		return true
	}
	return false
}

func UserUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	newUser, err := fetchUser(r)
	oldNickname := mux.Vars(r)["nickname"]
	if err != nil {
		utills.SendServerError("can Json Data", http.StatusConflict, w)
		return
	}
	updatedUser, err := urepo.UpdateUser(oldNickname, newUser)

	if err == sql.ErrNoRows {
		utills.SendServerError("cant find user with this nick: "+oldNickname, http.StatusNotFound, w)
		return
	}
	if err != nil {
		userWithThisEmail, err := urepo.GetUserByEmail(newUser.Email)

		if err != nil {

			utills.SendAnswerWithCode("err : "+oldNickname, http.StatusNotFound, w)
			return
		}
		utills.SendServerError("cant find user with this nick: "+userWithThisEmail.Nickname, http.StatusConflict, w)

		return
	}
	utills.SendOKAnswer(updatedUser, w)
}
