package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	tmodel "go-server-server-generated/src/thread/models"
	trepo "go-server-server-generated/src/thread/repository"
	"go-server-server-generated/src/utills"
	"io/ioutil"
	"net/http"
	"strconv"
)

func fetchThread(r *http.Request) (tmodel.Thread, error) {
	var badStaff = errors.New("bad json data")
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return tmodel.Thread{}, badStaff
	}
	var thread tmodel.Thread
	err = json.Unmarshal(data, &thread)
	thread.Created = thread.Created.UTC()
	thread.Forum = mux.Vars(r)["slug"]
	return thread, err
}

func fetchVote(r *http.Request) (tmodel.Vote, error) {
	var badStaff = errors.New("bad json data")
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return tmodel.Vote{}, badStaff
	}
	var vote tmodel.Vote
	err = json.Unmarshal(data, &vote)
	return vote, err
}

func ThreadCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	thread, err := fetchThread(r)
	resultThread, err := trepo.AddNew(thread)

	if err != nil {
		if err.Error() == `pq: insert or update on table "threads" violates foreign key constraint "threads_author_fkey"` {
			utills.SendServerError("not found", http.StatusNotFound, w)
			return
		}
		if err.Error() == `pq: duplicate key value violates unique constraint "threads_slug_key"` {
			th, _ := trepo.GetThreadBySlugWithoutTx(*thread.Slug)
			utills.SendAnswerWithCode(th, http.StatusConflict, w)
			return
		}
		if err.Error() == "no forum" {
			utills.SendServerError("Can't find thread forum by slug: "+thread.Forum, http.StatusNotFound, w)
			return
		}

	}
	resultThread.Created = resultThread.Created.UTC()
	utills.SendAnswerWithCode(resultThread, http.StatusCreated, w)
}

func ThreadVote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	slug_or_id := mux.Vars(r)["slug_or_id"]
	vote, _ := fetchVote(r)
	thread, err := trepo.MakeVote(slug_or_id, vote)
	fmt.Println("err thread vote", err)
	if err != nil {
		if err.Error() == "author not found" {
			errMsg := "Can't find user by nickname: " + vote.Nickname
			utills.SendServerError(errMsg, 404, w)
			return
		}
		if err.Error() == "already voted" { // да да это тупо а что поделать
			utills.SendAnswerWithCode(thread, http.StatusOK, w)
			return
		}
		if err.Error() == "thread is not exist" {
			errMsg := "Can't find thread by slug or id: " + slug_or_id
			utills.SendServerError(errMsg, http.StatusNotFound, w)
			return
		}
	}
	utills.SendAnswerWithCode(thread, http.StatusOK, w)
}

func GetThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	slug_or_id := mux.Vars(r)["slug_or_id"]
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	if trepo.IsDigit(slug_or_id) {

		idInt, _ := strconv.Atoi(slug_or_id)
		thread, err := trepo.GetThreadByID(tx, idInt)
		if err != nil {
			utills.SendServerError("thread not found", 404, w)
			return
		}
		utills.SendOKAnswer(thread, w)
		return
	}
	thread, err := trepo.GetThreadBySlug(tx, slug_or_id)
	if err != nil {
		utills.SendServerError("thread not found", 404, w)
		return
	}
	utills.SendOKAnswer(thread, w)
	return
}

func ThreadUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	slug_or_id := mux.Vars(r)["slug_or_id"]
	newThread, _ := fetchThread(r)

	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	id, isId := utills.IsDigit(slug_or_id)
	if isId {
		updatedThread, err := trepo.UpdateThreadById(tx, id, newThread)
		if err != nil {
			utills.SendServerError("thread with this id not found", 404, w)
			return
		}
		utills.SendOKAnswer(updatedThread, w)
		return
	} else {
		updatedThread, err := trepo.UpdateThreadBySlug(tx, slug_or_id, newThread)
		if err != nil {
			utills.SendServerError("thread with this slug not found", 404, w)
			return
		}
		utills.SendOKAnswer(updatedThread, w)
		return

	}
}
