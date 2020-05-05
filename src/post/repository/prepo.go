package post

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	pmodel "go-server-server-generated/src/post/models"
	tmodel "go-server-server-generated/src/thread/models"
	trepo "go-server-server-generated/src/thread/repository"
	"go-server-server-generated/src/user/repository"
	"go-server-server-generated/src/utills"
	"strconv"
	"time"
)

func IsDigit(v string) bool {
	if _, err := strconv.Atoi(v); err == nil {
		return true
	}
	return false
}

func UpdateForumPostsCountByThread(tx *sqlx.Tx, thread tmodel.Thread, incValue int) error {
	forumSlug := thread.Forum
	finalQuery := `UPDATE forum SET posts=posts+$1 where slug=$2`
	_, err := tx.Exec(finalQuery, incValue, forumSlug)
	return err
}

func CheckIfParentPostsInSameThread(tx *sqlx.Tx, post pmodel.Post) bool {
	if post.Parent == nil {
		return true
	}
	parentPost, err := GetPostById(tx, post.Parent)
	if err != nil {
		return false
	}
	return parentPost.Thread == post.Thread
}

func GetPostById(tx *sqlx.Tx, id *int) (pmodel.Post, error) {
	finalQuery := `Select * from posts where id=$1`
	var post pmodel.Post
	err := tx.Get(&post, finalQuery, id)
	return post, err
}

func AddNewPosts(posts []pmodel.Post, thr tmodel.Thread) ([]pmodel.Post, error) {
	timeCreated := time.Now().UTC()
	finalQuery := `INSERT INTO posts (author, created, forum, isedited, message, parent, thread) VALUES ($1,$2,$3,$4,$5,NULLIF($6,0),$7) returning *`
	conn := utills.GetConnection()
	tx := conn.MustBegin()
	defer tx.Commit()
	var postList []pmodel.Post
	var err error
	thread := thr

	for _, post := range posts {
		post.Created = timeCreated
		post.Forum = thread.Forum
		post.Thread = thread.Id
		post.IsEdited = false

		ok := checkIfAuthorExist(tx, post)
		if !ok {
			errMsg := "Can't find post author by nickname: " + post.Author
			return nil, errors.New(errMsg)
		}
		ok = CheckIfParentPostsInSameThread(tx, post)
		if !ok {
			return nil, errors.New("no parent")
		}

		var newPost pmodel.Post
		err := tx.Get(&newPost, finalQuery, post.Author, post.Created, post.Forum, post.IsEdited, post.Message, post.Parent, post.Thread)
		fmt.Println("err when add new posts:", err)
		newPost.Thread = post.Thread //COSTIL todo
		postList = append(postList, newPost)

	}

	err = UpdateForumPostsCountByThread(tx, thread, len(postList))

	return postList, err
}

func checkIfAuthorExist(tx *sqlx.Tx, post pmodel.Post) bool {
	_, err := repository.GetUserByNickname(post.Author)
	if err != nil {
		return false
	}
	return true
}

func GetPostsWithFlatSort(slug_or_id string, limit int, sinceID int, desc string) ([]pmodel.Post, error) {
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	id, isDig := utills.IsDigit(slug_or_id)
	if isDig {
		return GetPostsWithFlatSortById(tx, id, limit, sinceID, desc)
	} else {
		return GetPostsWithFlatSortBySlug(tx, slug_or_id, limit, sinceID, desc)
	}
}

func GetPostsWithFlatSortBySlug(tx *sqlx.Tx, slug string, limit int, sinceId int, desc string) ([]pmodel.Post, error) {
	thread, err := trepo.GetThreadBySlug(tx, slug)
	if err != nil {
		return nil, err
	}
	return GetPostsWithFlatSortById(tx, thread.Id, limit, sinceId, desc)

}

func GetPostsWithFlatSortById(tx *sqlx.Tx, id int, limit int, sinceID int, desc string) ([]pmodel.Post, error) {
	var finalQuery string
	var err error
	var posts []pmodel.Post
	_, err = trepo.GetThreadByID(tx, id)
	if err != nil {
		return nil, err
	}
	if desc == "true" {
		if sinceID != 0 {
			finalQuery = `Select * from posts where thread=$1 and id<$2 order by id DESC LIMIT $3`
			err = tx.Select(&posts, finalQuery, id, sinceID, limit)
		} else {
			finalQuery = `Select * from posts where thread=$1 order by id DESC LIMIT $2`
			err = tx.Select(&posts, finalQuery, id, limit)
		}
	} else {
		if sinceID != 0 {
			finalQuery = `Select * from posts where thread=$1 and id>$2 order by id LIMIT $3`
			err = tx.Select(&posts, finalQuery, id, sinceID, limit)
		} else {
			finalQuery = `Select * from posts where thread=$1 order by id LIMIT $2`
			err = tx.Select(&posts, finalQuery, id, limit)
		}
	}

	return posts, err
}

func UpdatePost(tx *sqlx.Tx, postId int, UpdatedPost pmodel.Post) (pmodel.Post, error) {
	prevPost, _ := GetPostById(tx, &postId)
	if prevPost.Message == UpdatedPost.Message || UpdatedPost.Message == "" {
		return prevPost, nil
	}
	finalQuery := `UPDATE posts SET message=$1,isEdited=true where id=$2 returning *`
	var post pmodel.Post
	err := tx.Get(&post, finalQuery, UpdatedPost.Message, postId)
	return post, err

}

func GetPostsWithTreeSort(slug_or_id string, limit int, sinceID int, desc string) ([]pmodel.Post, error) {
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	id, isDig := utills.IsDigit(slug_or_id)
	if isDig {
		return GetPostsWithTreeSortById(tx, id, limit, sinceID, desc)
	} else {
		return GetPostsWithTreeSortBySlug(tx, slug_or_id, limit, sinceID, desc)
	}
}

func GetPostsWithParentTreeSort(slug_or_id string, limit int, sinceID int, desc string) ([]pmodel.Post, error) {
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	id, isDig := utills.IsDigit(slug_or_id)
	if isDig {
		return GetPostsWithParentTreeSortById(tx, id, limit, sinceID, desc)
	} else {
		return GetPostsWithParentTreeSortBySlug(tx, slug_or_id, limit, sinceID, desc)
	}
}

func GetPostsWithTreeSortBySlug(tx *sqlx.Tx, slug string, limit int, sinceId int, desc string) ([]pmodel.Post, error) {
	thread, err := trepo.GetThreadBySlug(tx, slug)
	if err != nil {
		return nil, err
	}
	return GetPostsWithTreeSortById(tx, thread.Id, limit, sinceId, desc)
}

func GetPostsWithParentTreeSortBySlug(tx *sqlx.Tx, slug string, limit int, sinceId int, desc string) ([]pmodel.Post, error) {
	thread, err := trepo.GetThreadBySlug(tx, slug)
	if err != nil {
		return nil, err
	}
	return GetPostsWithParentTreeSortById(tx, thread.Id, limit, sinceId, desc)
}

func GetPostsWithTreeSortById(tx *sqlx.Tx, threadID int, limit int, since int, desc string) ([]pmodel.Post, error) {
	_, err := trepo.GetThreadByID(tx, threadID)
	if err != nil {
		return nil, err
	}
	finalQuery := GenerateQueryToTreeSort(since, desc)

	var posts []pmodel.Post
	err = tx.Select(&posts, finalQuery, threadID, limit)
	return posts, err
}

func GenerateQueryToTreeSort(since int, desc string) string {
	var finalQuery string
	var newQuer = ""
	if since != 0 {
		if desc == "true" {
			newQuer = `AND parents < `
		} else {
			newQuer = `AND parents > `
		}
		newQuer += fmt.Sprintf(`(SELECT parents FROM posts WHERE id = %d)`, since)
	}
	if desc == "true" {
		finalQuery = fmt.Sprintf(
			`SELECT * FROM posts WHERE thread=$1 %s ORDER BY parents DESC, id DESC LIMIT NULLIF($2, 0);`, newQuer)
	} else {
		finalQuery = fmt.Sprintf(
			`SELECT * FROM posts WHERE thread=$1 %s ORDER BY parents, id LIMIT NULLIF($2, 0);`, newQuer)
	}
	return finalQuery
}

func GetPostsWithParentTreeSortById(tx *sqlx.Tx, threadID int, limit int, since int, desc string) ([]pmodel.Post, error) {
	_, err := trepo.GetThreadByID(tx, threadID)
	if err != nil {
		return nil, err
	}
	finalQuery := GenerateQueryToParentTreeSort(since, desc, limit)
	var posts []pmodel.Post
	err = tx.Select(&posts, finalQuery, threadID)
	return posts, err
}

func GenerateQueryToParentTreeSort(since int, desc string, limit int) string {
	var newQuer = ""
	var finalQuery string

	if since != 0 {
		if desc == "true" {
			newQuer = `AND parents[1] < `
		} else {
			newQuer = `AND parents[1] > `
		}
		newQuer += fmt.Sprintf(`(SELECT parents[1] FROM posts WHERE id = %d)`, since)
	}

	newQuer2 := fmt.Sprintf(
		`SELECT id FROM posts WHERE thread = $1 AND parent IS NULL %s`, newQuer)

	if desc == "true" {
		newQuer2 += `ORDER BY id DESC`
		if limit != 0 {
			newQuer2 += fmt.Sprintf(` LIMIT %d`, limit)
		}
		finalQuery = fmt.Sprintf(
			`SELECT * FROM posts WHERE parents[1] IN (%s) ORDER BY parents[1] DESC, parents, id;`, newQuer2)
	} else {
		newQuer2 += `ORDER BY id`
		if limit > 0 {
			newQuer2 += fmt.Sprintf(` LIMIT %d`, limit)
		}
		finalQuery = fmt.Sprintf(
			`SELECT * FROM posts WHERE parents[1] IN (%s) ORDER BY parents,id;`, newQuer2)
	}
	return finalQuery
}
