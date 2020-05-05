package repository

import (
	"go-server-server-generated/src/service/models"
	"go-server-server-generated/src/utills"
)

func CountNumStr() models.ServiceAns {

	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)

	var numOfStrsUser int
	query := `SELECT COUNT(*) FROM "user"`
	tx.Get(&numOfStrsUser, query)

	var numOfStrsForum int
	query = `SELECT COUNT(*) FROM forum`
	tx.Get(&numOfStrsForum, query)

	var numOfStrsThread int
	query = `SELECT COUNT(*) FROM threads`
	tx.Get(&numOfStrsThread, query)

	var numOfStrsPosts int
	query = `SELECT COUNT(*) FROM posts`
	tx.Get(&numOfStrsPosts, query)

	return models.ServiceAns{
		User:   numOfStrsUser,
		Forum:  numOfStrsForum,
		Thread: numOfStrsThread,
		Post:   numOfStrsPosts,
	}

}
