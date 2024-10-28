package handlers

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/app/auth"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/storage"
	"io"
	"net/http"

	"github.com/joeljunstrom/go-luhn"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Handlers struct {
	Storage storage.Storage
	Auth    auth.Auth
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Balance struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type CustomResponse struct {
	err  error
	msg  string
	code int
}

func H(f func(http.ResponseWriter, *http.Request) CustomResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := f(w, r)
		if result.err != nil {
			message := result.msg + ": " + result.err.Error()
			log.Error().Msg(message)
			http.Error(w, message, result.code)
		} else if result.msg != "" {
			log.Info().Msg(result.msg)
			http.Error(w, result.msg, result.code)
		}
	}
}

func (h Handlers) Register(res http.ResponseWriter, req *http.Request) CustomResponse {
	decoder := json.NewDecoder(req.Body)
	var reg AuthRequest
	err := decoder.Decode(&reg)
	if err != nil {
		return CustomResponse{err: err, msg: "Не удалось распарсить запрос", code: http.StatusBadRequest}
	}

	exists, err := h.Storage.CheckUser(req.Context(), reg.Login)
	if err != nil {
		return CustomResponse{err: err, msg: "Не удалось проверить наличие пользователя в системе", code: http.StatusInternalServerError}
	}
	if exists {
		return CustomResponse{msg: fmt.Sprintf("Логин %s уже занят", reg.Login), code: http.StatusConflict}
	}
	userID := uuid.NewString()
	user := domain.User{
		Login:    reg.Login,
		Password: auth.Hash(reg.Password),
		UserID:   userID,
		Balance:  0,
	}
	err = h.Storage.AddUser(req.Context(), user)
	if err != nil {
		return CustomResponse{err: err, msg: "Не удалось зарегистрировать пользователя", code: http.StatusInternalServerError}
	}

	h.Auth.AddAuth(res, userID)
	res.WriteHeader(http.StatusOK)
	return CustomResponse{}
}

func (h Handlers) Login(res http.ResponseWriter, req *http.Request) CustomResponse {
	decoder := json.NewDecoder(req.Body)
	var log AuthRequest
	err := decoder.Decode(&log)
	if err != nil {
		return CustomResponse{err: err, msg: "Не удалось распарсить запрос", code: http.StatusBadRequest}
	}

	userID, err := h.Storage.GetUserID(req.Context(), log.Login, auth.Hash(log.Password))
	switch err {
	case nil:
		// pass
	case storage.ErrEmpty:
		return CustomResponse{err: err, msg: "Неверная пара логин/пароль", code: http.StatusUnauthorized}
	default:
		return CustomResponse{err: err, msg: "Внутренняя ошибка сервера", code: http.StatusInternalServerError}
	}

	h.Auth.AddAuth(res, userID)
	res.WriteHeader(http.StatusOK)
	return CustomResponse{}
}

func (h Handlers) AddOrder(res http.ResponseWriter, req *http.Request) CustomResponse {
	userID := req.Header.Get("x-user-id")
	byteOrderID, err := io.ReadAll(req.Body)
	orderID := string(byteOrderID)
	if err != nil {
		return CustomResponse{err: err, msg: "Не удалось распарсить запрос", code: http.StatusBadRequest}
	}
	if !luhn.Valid(orderID) {
		return CustomResponse{msg: "Некорректный номер заказа", code: http.StatusUnprocessableEntity}
	}
	exists, err := h.Storage.CheckOrder(req.Context(), userID, orderID)
	if err != nil {
		return CustomResponse{err: err, msg: "Заказ загружен другим пользователем", code: http.StatusConflict}
	}
	if exists {
		return CustomResponse{err: err, msg: "Заказ уже был загружен", code: http.StatusOK}
	}

	order := domain.Order{
		UserID:  userID,
		OrderID: orderID,
		Status:  domain.NewOrderStatus,
	}
	err = h.Storage.AddOrder(req.Context(), order)
	if err != nil {
		return CustomResponse{err: err, msg: "Не удалось добавить заказ в систему", code: http.StatusInternalServerError}
	}

	res.WriteHeader(http.StatusAccepted)
	return CustomResponse{}
}

func (h Handlers) GetOrders(res http.ResponseWriter, req *http.Request) CustomResponse {
	userID := req.Header.Get("x-user-id")
	orders, err := h.Storage.GetOrders(req.Context(), userID)
	if err != nil {
		return CustomResponse{err: err, msg: "Ошибка при выполнении запроса", code: http.StatusInternalServerError}
	}
	if len(orders) == 0 {
		return CustomResponse{err: fmt.Errorf(""), msg: "Нет заказов", code: http.StatusNoContent}
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(orders)
	return CustomResponse{}
}

func (h Handlers) GetBalance(res http.ResponseWriter, req *http.Request) CustomResponse {
	userID := req.Header.Get("x-user-id")
	user, err := h.Storage.GetUser(req.Context(), userID)
	if err != nil {
		return CustomResponse{err: err, msg: "Ошибка при выполнении запроса", code: http.StatusInternalServerError}
	}

	result := Balance{
		Balance:   user.Balance,
		Withdrawn: user.Withdrawn,
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(result)
	return CustomResponse{}
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h Handlers) Withdraw(res http.ResponseWriter, req *http.Request) CustomResponse {
	decoder := json.NewDecoder(req.Body)
	var r WithdrawRequest
	err := decoder.Decode(&r)
	if err != nil {
		return CustomResponse{err: err, msg: "Не удалось распарсить запрос", code: http.StatusBadRequest}
	}

	userID := req.Header.Get("x-user-id")
	user, err := h.Storage.GetUser(req.Context(), userID)
	if err != nil {
		return CustomResponse{err: err, msg: "Ошибка при выполнении запроса", code: http.StatusInternalServerError}
	}
	if user.Balance < r.Sum {
		return CustomResponse{msg: "На счету недостаточно средств", code: http.StatusPaymentRequired}
	}
	if !luhn.Valid(r.Order) {
		return CustomResponse{msg: "Некорректный номер заказа", code: http.StatusUnprocessableEntity}
	}
	user.Balance -= r.Sum
	err = h.Storage.UpdateUser(req.Context(), user)
	if err != nil {
		return CustomResponse{err: err, msg: "Ошибка при выполнении запроса", code: http.StatusInternalServerError}
	}
	wd := domain.Withdrawal{
		OrderID: r.Order, // TOTHINK: должна ли эта ручка создавать новый заказ?
		UserID:  userID,
		Sum:     r.Sum,
	}
	err = h.Storage.AddWithdrawal(req.Context(), wd)
	if err != nil {
		return CustomResponse{err: err, msg: "Ошибка при выполнении запроса", code: http.StatusInternalServerError}
	}
	// TODO: Объединить два запроса в транзакцию
	res.WriteHeader(http.StatusOK)
	return CustomResponse{}
}

func (h Handlers) GetWithdrawals(res http.ResponseWriter, req *http.Request) CustomResponse {
	userID := req.Header.Get("x-user-id")
	wds, err := h.Storage.GetWithdrawals(req.Context(), userID)
	if err != nil {
		return CustomResponse{err: err, msg: "Ошибка при выполнении запроса", code: http.StatusInternalServerError}
	}
	if len(wds) == 0 {
		return CustomResponse{err: fmt.Errorf(""), msg: "Нет заказов", code: http.StatusNoContent}
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(wds)
	return CustomResponse{}
}
