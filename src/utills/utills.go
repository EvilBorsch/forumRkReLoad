package utills

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

var conn *sqlx.DB

func CreateConnection() {
	connStr := fmt.Sprintf("user=%s password=%s dbname=docker sslmode=disable port=%s",
		"docker",
		"docker",
		"5432")
	conn, _ = sqlx.Open("postgres", connStr)

}

func GetConnection() *sqlx.DB {
	return conn
}

type HttpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HttpResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []HttpError `json:"errors"`
}

type ModelError struct {

	// Текстовое описание ошибки. В процессе проверки API никаких проверок на содерижимое данного описание не делается.
	Message string `json:"message,omitempty"`
}

func SendServerError(errorMessage string, code int, w http.ResponseWriter) {
	log.Error().Msgf(errorMessage)
	w.WriteHeader(code)
	mes, _ := json.Marshal(ModelError{Message: errorMessage})
	w.Write(mes)
}

func SendOKAnswer(data interface{}, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	serializedData, err := json.Marshal(data)
	if err != nil {
		log.Error().Msgf(err.Error())
		return
	}

	_, err = w.Write(serializedData)
	if err != nil {
		message := fmt.Sprintf("HttpResponse while writing is socket: %s", err.Error())
		log.Error().Msgf(message)
		return
	}
	log.Info().Msgf("OK message sent")
}

func SendAnswerWithCode(data interface{}, code int, w http.ResponseWriter) {
	w.WriteHeader(code)
	serializedData, err := json.Marshal(data)
	if err != nil {
		log.Error().Msgf(err.Error())
		return
	}

	_, err = w.Write(serializedData)
	if err != nil {
		message := fmt.Sprintf("HttpResponse while writing is socket: %s", err.Error())
		log.Error().Msgf(message)
		return
	}
	log.Info().Msgf("Code message sent")
}

func StartTransaction() *sqlx.Tx {
	conn := GetConnection()
	tx := conn.MustBegin()
	return tx
}

func EndTransaction(tx *sqlx.Tx) error {
	err := tx.Commit()
	return err
}

func IsDigit(value string) (int, bool) {
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return valueInt, true
}
