package repository

import (
	"errors"
	"fmt"
	fmodel "go-server-server-generated/src/forum/models"
	tmodel "go-server-server-generated/src/thread/models"
	swagger "go-server-server-generated/src/user/models"
	urepo "go-server-server-generated/src/user/repository"
	"go-server-server-generated/src/utills"
)

func CreateForum(forum fmodel.Forum) (fmodel.Forum, error) {
	conn := utills.GetConnection()
	var newForum fmodel.Forum
	user, err := urepo.GetUserByNickname(forum.User_nickname)
	fmt.Println("err when creating forum", err)
	if err != nil {
		return fmodel.Forum{}, err
	}
	forum.User_nickname = user.Nickname
	query := `INSERT INTO forum (title, user_nickname, slug) VALUES ($1,$2,$3) returning *`
	err = conn.Get(&newForum, query, forum.Title, forum.User_nickname, forum.Slug)
	return newForum, err
}

func GetForumBySlug(slug string) (fmodel.Forum, error) {
	conn := utills.GetConnection()
	var forum fmodel.Forum
	query := `SELECT * from forum where slug=$1`
	err := conn.Get(&forum, query, slug)
	return forum, err
}

func GetThreadsByForumSlug(forumSlug string, isDesc string, limit string, since string) ([]tmodel.Thread, error, bool) {
	conn := utills.GetConnection()
	var thread []tmodel.Thread
	query := `SELECT * from threads where forum=$1 ORDER BY created`
	_, err := GetForumBySlug(forumSlug)
	if err != nil {
		return []tmodel.Thread{}, nil, false
	}
	if since != "" {
		tString := "'" + since + "'"

		query = `SELECT * from threads where forum=$1 and created<=` + tString + ` ORDER BY created`
		if isDesc == "false" || isDesc == "" {
			query = `SELECT * from threads where forum=$1 and created>=` + tString + ` ORDER BY created`
		}
	}
	if isDesc == "true" {
		query += " DESC"
	}
	if limit != "" {
		query = query + " LIMIT " + limit
	}
	err = conn.Select(&thread, query, forumSlug)
	return thread, err, true
}

func IncrementFieldBySlug(fieldName string, slug string) error {
	query := fmt.Sprintf(`UPDATE forum SET %s =%s + 1 WHERE slug=$1`, fieldName, fieldName)

	conn := utills.GetConnection()
	_, err := conn.Exec(query, slug)
	return err
}

func GetForumUsers(forumSlug string, isDesc string, limit string, since string) ([]swagger.User, error) {

	if limit == "" {
		limit = "100"
	}
	if isDesc == "" || isDesc == "false" {
		isDesc = "ASC"
	} else {
		isDesc = "DESC"
	}
	_, err := GetForumBySlug(forumSlug)
	if err != nil {
		return nil, errors.New("forum not found")
	}
	var queryFinal string
	queryFinal = `Select nickname, fullname, about, email from forum_user 
join "user" on lower(forum_user.user_nickname) = lower("user".nickname) 
where forum_slug=$1`
	strOrder := " order by nickname " + isDesc + " LIMIT " + limit
	if isDesc == "DESC" && since != "" {
		queryFinal = queryFinal + "and lower(nickname)<lower($2) "
	} else if isDesc == "ASC" && since != "" {
		queryFinal = queryFinal + "and lower(nickname)>lower($2) "
	}
	queryFinal = queryFinal + strOrder
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	var users []swagger.User

	if since != "" {
		err := tx.Select(&users, queryFinal, forumSlug, since)
		return users, err
	}
	err = tx.Select(&users, queryFinal, forumSlug)
	return users, err
}

