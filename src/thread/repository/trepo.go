package repository

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	frepo "go-server-server-generated/src/forum/repository"
	tm "go-server-server-generated/src/thread/models"
	"go-server-server-generated/src/user/repository"
	"go-server-server-generated/src/utills"
	"strconv"
)

func AddNew(newThread tm.Thread) (tm.Thread, error) {
	conn := utills.GetConnection()
	query := `INSERT INTO threads (author, created, forum, message, title,slug) VALUES ($1,$2,$3,$4,$5,nullif($6,'')) returning *`
	var createdThred tm.Thread
	forum, err := frepo.GetForumBySlug(newThread.Forum) //kostil for tests
	if err != nil {
		return tm.Thread{}, errors.New("no forum")
	}
	newThread.Forum = forum.Slug

	err = conn.Get(&createdThred, query, newThread.Author, newThread.Created, newThread.Forum, newThread.Message, newThread.Title, newThread.Slug)
	createdThred.Title = newThread.Title
	if err == nil {
		err2 := frepo.IncrementFieldBySlug("threads", newThread.Forum)
		fmt.Println("new err: ", err2)
	}
	return createdThred, err

}

func GetThreadBySlug(tx *sqlx.Tx, threadSlug string) (tm.Thread, error) {
	var thread tm.Thread
	query := `SELECT * FROM threads where slug=$1`
	err := tx.Get(&thread, query, threadSlug)
	return thread, err
}

func GetThreadByID(tx *sqlx.Tx, id int) (tm.Thread, error) {
	var thread tm.Thread
	query := `SELECT * FROM threads where id=$1`
	err := tx.Get(&thread, query, id)
	return thread, err
}

func GetThreadBySlugWithoutTx(threadSlug string) (tm.Thread, error) {
	conn := utills.GetConnection()
	var thread tm.Thread
	query := `SELECT * FROM threads where slug=$1`
	err := conn.Get(&thread, query, threadSlug)
	return thread, err
}

func GetThreadByIDWithoutTx(id int) (tm.Thread, error) {
	conn := utills.GetConnection()
	var thread tm.Thread
	query := `SELECT * FROM threads where id=$1`
	err := conn.Get(&thread, query, id)
	return thread, err
}

func IsDigit(v string) bool {
	if _, err := strconv.Atoi(v); err == nil {
		return true
	}
	return false
}

func IncrementVoteBySlug(tx *sqlx.Tx, slug string, delta int) (tm.Thread, error) {
	query := `UPDATE threads SET votes =votes + $1 WHERE slug=$2 returning *`
	var thread tm.Thread
	err := tx.Get(&thread, query, delta, slug)
	return thread, err
}

func IncrementVoteByID(tx *sqlx.Tx, id string, delta int) (tm.Thread, error) {
	query := `UPDATE threads SET votes =votes + $1 WHERE id=$2 returning *`
	var thread tm.Thread
	err := tx.Get(&thread, query, delta, id)
	return thread, err
}

func UpdateVoteInVotes(tx *sqlx.Tx, newVote tm.Vote) error {
	query := `UPDATE votes SET voice =$1 WHERE nickname=$2`
	_, err := tx.Exec(query, newVote.Voice, newVote.Nickname)
	fmt.Println("upd: ", err)
	return err
}

func getOldVoteBySlugAndAuthor(tx *sqlx.Tx, slug string, author string) (tm.Vote, error) {
	query := `Select nickname,voice from votes where nickname=$1 and threadslug=$2`
	var oldVote tm.Vote
	err := tx.Get(&oldVote, query, author, slug)
	if err != nil {
		notFound := errors.New("no vote")
		return tm.Vote{}, notFound
	}
	return oldVote, nil
}

func getOldVoteByIdAndAuthor(tx *sqlx.Tx, id string, author string) (tm.Vote, error) {
	query := `Select nickname,voice from votes where nickname=$1 and threadID=$2 `
	var oldVote tm.Vote
	err := tx.Get(&oldVote, query, author, id)
	if err != nil {
		notFound := errors.New("no vote")
		return tm.Vote{}, notFound
	}
	return oldVote, nil
}

func InsertNewVoteWithThreadId(tx *sqlx.Tx, newVote tm.Vote, slug_or_id string) {
	query := `INSERT INTO votes (threadid, threadslug, nickname, voice) VALUES ($1,$2,$3,$4)`
	tx.Exec(query, slug_or_id, nil, newVote.Nickname, newVote.Voice)
}

func InsertNewVoteWithThreadSlug(tx *sqlx.Tx, newVote tm.Vote, slug_or_id string) {

	query := `INSERT INTO votes (threadid, threadslug, nickname, voice) VALUES ($1,$2,$3,$4)`
	_, err := tx.Exec(query, nil, slug_or_id, newVote.Nickname, newVote.Voice)
	fmt.Println(err)
}

func getAuthorVotesByThread(tx *sqlx.Tx, thread tm.Thread, nick string) (tm.Vote, error) {
	query := `Select nickname,voice from votes where nickname=$1 and (threadID=$2 or threadSlug=$3)`
	var vote tm.Vote
	err := tx.Get(&vote, query, nick, thread.Id, thread.Slug)
	return vote, err
}

func MakeVote(slug_or_id string, newVote tm.Vote) (tm.Thread, error) {
	AlreadyVotedErr := errors.New("already voted")
	ThreadIsNotExistErr := errors.New("thread is not exist")
	AuthorIsNotExistErr := errors.New("author not found")
	author := newVote.Nickname
	_, err := repository.GetUserByNickname(author)
	if err != nil {
		return tm.Thread{}, AuthorIsNotExistErr
	}
	conn := utills.GetConnection()
	tx := conn.MustBegin()
	defer tx.Commit()
	if IsDigit(slug_or_id) {
		idStr, _ := strconv.Atoi(slug_or_id)
		thread, err := GetThreadByID(tx, idStr)
		if err != nil {
			return tm.Thread{}, ThreadIsNotExistErr
		}
		oldVote, err := getAuthorVotesByThread(tx, thread, author)
		if err != nil { // не нашли оценки
			InsertNewVoteWithThreadId(tx, newVote, slug_or_id)
			incThread, err := IncrementVoteByID(tx, slug_or_id, newVote.Voice)
			fmt.Println(incThread, err)
			return incThread, err
		} //нашли старую оценку
		delta := newVote.Voice - oldVote.Voice
		if delta == 0 {
			return thread, AlreadyVotedErr
		}
		UpdateVoteInVotes(tx, newVote)
		incThread, err := IncrementVoteByID(tx, slug_or_id, delta)
		return incThread, err
	}

	thread, err := GetThreadBySlug(tx, slug_or_id)
	if err != nil {
		return tm.Thread{}, ThreadIsNotExistErr
	}
	oldVote, err := getAuthorVotesByThread(tx, thread, author)
	if err != nil { // не нашли оценки
		InsertNewVoteWithThreadSlug(tx, newVote, slug_or_id)
		incThread, err := IncrementVoteBySlug(tx, slug_or_id, newVote.Voice)
		fmt.Println(incThread, err)
		return incThread, err
	} //нашли старую оценку
	delta := newVote.Voice - oldVote.Voice
	if delta == 0 {
		return thread, AlreadyVotedErr
	}
	UpdateVoteInVotes(tx, newVote)
	incThread, err := IncrementVoteBySlug(tx, slug_or_id, delta)
	return incThread, err

}

func UpdateThreadBySlug(tx *sqlx.Tx, slug string, newThread tm.Thread) (tm.Thread, error) {

	var updatedThread tm.Thread
	query := `UPDATE threads SET
                author=COALESCE(NULLIF($1, ''), author),
                title=COALESCE(NULLIF($2, ''), title),
                message=COALESCE(NULLIF($3, ''), message)
			WHERE slug = $4 RETURNING *`
	err := tx.Get(&updatedThread, query, newThread.Author, newThread.Title, newThread.Message, slug)

	return updatedThread, err
}
func UpdateThreadById(tx *sqlx.Tx, id int, newThread tm.Thread) (tm.Thread, error) {

	var updatedThread tm.Thread
	query := `UPDATE threads SET
                author=COALESCE(NULLIF($1, ''), author),
                title=COALESCE(NULLIF($2, ''), title),
                message=COALESCE(NULLIF($3, ''), message)
			WHERE id = $4 RETURNING *`
	err := tx.Get(&updatedThread, query, newThread.Author, newThread.Title, newThread.Message, id)

	return updatedThread, err
}

//func checkIfAllOkSlugCase(tx *sqlx.Tx, slug string, author string) (tm.Thread, error) {
//
//	ThreadIsNotExistErr := errors.New("thread is not exist")
//	thread, err := GetThreadBySlug(tx, slug)
//	if err != nil {
//		return tm.Thread{}, ThreadIsNotExistErr
//	}
//	threadIdStr := strconv.Itoa(thread.Id)
//	_, err = getOldVoteByIdAndAuthor(tx, threadIdStr, author)
//	if err == nil {
//		return thread, errors.New("voted for another index")
//	}
//	return tm.Thread{}, nil
//}
//
//func checkIfAllOkIdCase(tx *sqlx.Tx, slug string, author string) (tm.Thread, error) {
//
//	ThreadIsNotExistErr := errors.New("thread is not exist")
//	thread, err := GetThreadBySlug(tx, slug)
//	if err != nil {
//		return tm.Thread{}, ThreadIsNotExistErr
//	}
//	_, err = getOldVoteBySlugAndAuthor(tx, thread.Slug, author)
//	if err == nil {
//		return thread, errors.New("voted for another index")
//	}
//	return tm.Thread{}, nil
//}
