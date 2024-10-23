package handlers

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/app/auth"
	"gophermart/internal/app/storage"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Handlers struct {
	Storage storage.Storage
	BaseURL string
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Error struct {
	err  error
	msg  string
	code int
}

func H(f func(http.ResponseWriter, *http.Request) Error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := f(w, r)
		if result.err != nil {
			message := result.msg + ": " + result.err.Error()
			log.Error().Msg(message)
			http.Error(w, message, result.code)
		}
	}
}

func (h Handlers) Register(res http.ResponseWriter, req *http.Request) Error {
	decoder := json.NewDecoder(req.Body)
	var reg RegisterRequest
	err := decoder.Decode(&reg)
	if err != nil {
		return Error{err: err, msg: "Не удалось распарсить запрос", code: http.StatusBadRequest}
	}
	exists, err := h.Storage.CheckUser(req.Context(), reg.Login)
	if err != nil {
		return Error{err: err, msg: "Не удалось проверить наличие пользователя в системе", code: http.StatusInternalServerError}
	}
	if exists {
		return Error{err: err, msg: fmt.Sprintf("Логин %s уже занят", reg.Login), code: http.StatusConflict}
	}
	userID := uuid.NewString()
	user := storage.User{
		Login:    reg.Login,
		Password: reg.Password,
		UserID:   userID,
		Balance:  0,
	}
	ok, err := h.Storage.AddUser(req.Context(), user)
	if err != nil || !ok {
		return Error{err: err, msg: "Не удалось зарегистрировать пользователя", code: http.StatusInternalServerError}
	}

	auth.AddAuth(res, userID)
	res.WriteHeader(http.StatusOK)
	return Error{}
}
