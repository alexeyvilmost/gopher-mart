package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophermart/internal/app/storage"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type AccrualClient struct {
	host string
}

type AccrualResponse struct {
	OrderID string              `json:"order_id"`
	Status  storage.OrderStatus `json:"status"`
	Accrual float64             `json:"accrual"`
}

type AccrualError struct {
	Err          error
	Code         int
	RetryAfterS  int
	MaxRequestsM int
}

const url = "/api/orders/"

func NewAccrualClient(host string) AccrualClient {
	return AccrualClient{host: host}
}

func (ac *AccrualClient) GetOrderInfo(order storage.Order) (AccrualResponse, AccrualError) {
	uri := ac.host + url + order.OrderID
	resp, err := http.Get(uri)
	if err != nil {
		return AccrualResponse{}, AccrualError{Err: err}
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// pass
	case http.StatusTooManyRequests:
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return AccrualResponse{}, AccrualError{Err: err}
		}

		retry := resp.Header.Get("Retry-After")
		retryInt, err := strconv.Atoi(retry)
		if err != nil {
			return AccrualResponse{}, AccrualError{Err: err}
		}
		response := strings.Split(string(content), " ")
		if len(response) < 4 {
			return AccrualResponse{}, AccrualError{Err: err}
		}
		maxReq := response[3]
		maxReqInt, err := strconv.Atoi(maxReq)
		if err != nil {
			return AccrualResponse{}, AccrualError{Err: err}
		}

		Err := AccrualError{
			Code:         http.StatusTooManyRequests,
			RetryAfterS:  retryInt,
			MaxRequestsM: maxReqInt,
			Err:          fmt.Errorf(""),
		}

		return AccrualResponse{}, Err
	default:
		stringCode := strconv.Itoa(resp.StatusCode)
		return AccrualResponse{}, AccrualError{Err: errors.New(stringCode), Code: resp.StatusCode}
	}

	response := AccrualResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return AccrualResponse{}, AccrualError{Err: err}
	}
	return response, AccrualError{}
}
