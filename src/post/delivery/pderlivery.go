package post

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go-server-server-generated/src/forum/models"
	frepo "go-server-server-generated/src/forum/repository"
	pmodel "go-server-server-generated/src/post/models"
	prepo "go-server-server-generated/src/post/repository"
	tmodel "go-server-server-generated/src/thread/models"
	trepo "go-server-server-generated/src/thread/repository"
	swagger "go-server-server-generated/src/user/models"
	"go-server-server-generated/src/user/repository"
	"go-server-server-generated/src/utills"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func PostsCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	posts, err := fetchPost(r)
	fmt.Println(posts, err)
	threadSlug := mux.Vars(r)["slug_or_id"]
	threadId, isId := utills.IsDigit(threadSlug)
	var thread tmodel.Thread
	if isId {
		threadId, _ = strconv.Atoi(threadSlug)
		thread, err = trepo.GetThreadByIDWithoutTx(threadId)
		if err != nil {
			errMsg := "Can't find post thread by id: " + strconv.Itoa(threadId)
			utills.SendServerError(errMsg, 404, w)
			return
		}

	} else {
		thread, err = trepo.GetThreadBySlugWithoutTx(threadSlug)
		if err != nil {
			errMsg := "Can't find post thread by slug: " + threadSlug
			utills.SendServerError(errMsg, 404, w)
			return
		}
	}

	if len(posts) == 0 {
		utills.SendAnswerWithCode([]pmodel.Post{}, http.StatusCreated, w)
		return
	}

	newPosts, err := prepo.AddNewPosts(posts, thread)
	fmt.Println(err)
	if err != nil {
		if err.Error() == "no thread" {
			utills.SendServerError("no thread", http.StatusNotFound, w)
			return
		}
		if err.Error()[0] == 'C' {
			utills.SendServerError(err.Error(), 404, w)
			return
		} else {
			utills.SendServerError("no parent", http.StatusConflict, w)
			return
		}
	}

	utills.SendAnswerWithCode(newPosts, http.StatusCreated, w)

}

func fetchPost(r *http.Request) ([]pmodel.Post, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []pmodel.Post{}, err
	}
	var post []pmodel.Post
	err = json.Unmarshal(data, &post)
	return post, err
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var limit int
	limitStr := r.FormValue("limit")
	sinceStr := r.FormValue("since")
	desc := r.FormValue("desc")
	var sinceID int
	if sinceStr == "" {
		sinceID = 0
	} else {
		sinceID, _ = strconv.Atoi(sinceStr)
	}
	if desc == "" {
		desc = "false"
	}

	if limitStr == "" {
		limit = 100
	} else {
		limit, _ = strconv.Atoi(limitStr)
	}
	slug_or_id := mux.Vars(r)["slug_or_id"]

	sort := r.FormValue("sort")
	fmt.Println(limit, sort)
	var posts []pmodel.Post
	var err error
	if sort == "" || sort == "flat" {
		posts, err = prepo.GetPostsWithFlatSort(slug_or_id, limit, sinceID, desc)
	}
	if sort == "tree" {
		posts, err = prepo.GetPostsWithTreeSort(slug_or_id, limit, sinceID, desc)
	}
	if sort == "parent_tree" {
		posts, err = prepo.GetPostsWithParentTreeSort(slug_or_id, limit, sinceID, desc)
	}
	if err != nil {
		utills.SendServerError("posts not found", 404, w)
		return
	}

	if posts == nil {
		emptyPost := []pmodel.Post{}
		utills.SendOKAnswer(emptyPost, w)
		return
	}
	utills.SendOKAnswer(posts, w)
	return

}

func GetSinglePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	postId, _ := strconv.Atoi(mux.Vars(r)["id"])
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	post, err := prepo.GetPostById(tx, &postId)
	if err != nil {
		utills.SendServerError("post not found", 404, w)
		return
	}
	related := r.FormValue("related")
	relatedSplit := strings.Split(related, ",")
	var user *swagger.User = nil
	var thread *tmodel.Thread = nil
	var forum *models.Forum = nil
	for _, key := range relatedSplit {
		switch key {
		case "user":
			usertmp, err := repository.GetUserByNickname(post.Author)
			user = &usertmp
			if err != nil {
				utills.SendServerError("user not found", 404, w)
				return
			}
		case "forum":
			forumtmp, err := frepo.GetForumBySlug(post.Forum)
			forum = &forumtmp
			if err != nil {
				utills.SendServerError("forum not found", 404, w)
				return
			}
		case "thread":
			threadtmp, err := trepo.GetThreadByID(tx, post.Thread)
			thread = &threadtmp
			if err != nil {
				utills.SendServerError("forum not found", 404, w)
				return
			}
		}
	}
	fmt.Println(related, postId)

	retVal := pmodel.PostFull{
		Post:   post,
		Author: user,
		Forum:  forum,
		Thread: thread,
	}
	utills.SendOKAnswer(retVal, w)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	postId, _ := strconv.Atoi(mux.Vars(r)["id"])
	UpdatedPost := fetchPostUpdate(r)
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	EditedPost, err := prepo.UpdatePost(tx, postId, UpdatedPost)
	if err != nil {
		utills.SendServerError("post not found", 404, w)
		return
	}
	utills.SendOKAnswer(EditedPost, w)
}

func fetchPostUpdate(r *http.Request) pmodel.Post {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return pmodel.Post{}
	}
	type messageStruct struct {
		Message string
	}
	var message messageStruct
	err = json.Unmarshal(data, &message)
	return pmodel.Post{
		Id:       0,
		Parent:   nil,
		Author:   "",
		Message:  message.Message,
		IsEdited: false,
		Forum:    "",
		Thread:   0,
		Created:  time.Time{},
	}
}
