package delivery

import (
	"go-server-server-generated/src/service/repository"
	"go-server-server-generated/src/utills"
	"net/http"
)

func Count(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res := repository.CountNumStr()
	utills.SendOKAnswer(res, w)

}

func ClearAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	query := `TRUNCATE TABLE "user" CASCADE`
	tx := utills.StartTransaction()
	defer utills.EndTransaction(tx)
	tx.Exec(query)
	w.WriteHeader(http.StatusOK)

}
