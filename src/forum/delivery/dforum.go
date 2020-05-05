package delivery

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	fmodel "go-server-server-generated/src/forum/models"
	frepo "go-server-server-generated/src/forum/repository"
	swagger "go-server-server-generated/src/user/models"
	"go-server-server-generated/src/utills"
	"io/ioutil"
	"net/http"
)

func fetchForum(r *http.Request) (fmodel.Forum, error) {
	var badStaff = errors.New("bad json data")
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmodel.Forum{}, badStaff
	}
	var forum fmodel.Forum
	err = json.Unmarshal(data, &forum)
	return forum, err
}

func ForumCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	forum, err := fetchForum(r)
	newForum, err := frepo.CreateForum(forum)

	if err != nil && (err.Error() == `pq: insert or update on table "forum" violates foreign key constraint "forum_user_nickname_fkey"` || err == sql.ErrNoRows) {
		utills.SendServerError("cant find user", http.StatusNotFound, w)
		return
	}
	if err != nil && err.Error() == `pq: duplicate key value violates unique constraint "forum_pkey"` {
		forumWithThisSlug, _ := frepo.GetForumBySlug(forum.Slug)

		utills.SendAnswerWithCode(forumWithThisSlug, http.StatusConflict, w)
		return
	}
	utills.SendAnswerWithCode(newForum, http.StatusCreated, w)
}

func ForumGetOne(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	slug := mux.Vars(r)["slug"]
	forum, err := frepo.GetForumBySlug(slug)
	if err != nil {
		utills.SendServerError("forum not in", http.StatusNotFound, w)
		return
	}

	utills.SendOKAnswer(forum, w)
}

func ForumGetThreads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	slug := mux.Vars(r)["slug"]
	isDesc := r.FormValue("desc")
	limit := r.FormValue("limit")
	since := r.FormValue("since")
	threads, err, needEmpty := frepo.GetThreadsByForumSlug(slug, isDesc, limit, since)
	if len(threads) == 0 && needEmpty {

		utills.SendOKAnswer([]fmodel.Forum{}, w)
		return
	}
	if err != nil {
		fmt.Println("err when forum get threads ", err)
		//todo
	}
	if len(threads) == 0 {
		fmt.Println("empty")
		utills.SendServerError("Can't find forum by slug: "+slug, http.StatusNotFound, w)
		return
	}
	utills.SendOKAnswer(threads, w)
}

func GetForumUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	forumSlug := mux.Vars(r)["slug"]
	isDesc := r.FormValue("desc")
	limit := r.FormValue("limit")
	since := r.FormValue("since")
	users, err := frepo.GetForumUsers(forumSlug, isDesc, limit, since)

	if err != nil {
		utills.SendServerError("forum not found", 404, w)
		return
	}
	if users == nil {
		empt := []swagger.User{}
		utills.SendOKAnswer(empt, w)
		return
	}
	utills.SendOKAnswer(users, w)
}
