package background

import (
	"context"
	"gophermart/internal/app/clients"
	"gophermart/internal/app/storage"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type BGAccrual struct {
	ticker  *time.Ticker
	client  clients.AccrualClient
	storage storage.Storage
	Done    chan bool
}

func NewBGAccrual(accrualAddress string, storage storage.Storage) *BGAccrual {
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	result := BGAccrual{
		ticker:  ticker,
		client:  clients.NewAccrualClient(accrualAddress),
		storage: storage,
		Done:    done,
	}
	return &result
}

func (b BGAccrual) resetTicker(aErr clients.AccrualError) {
	var newPeriod int
	if aErr.MaxRequestsM < 1 {
		newPeriod = int(time.Minute)
	} else {
		newPeriod = int(time.Minute) / (aErr.MaxRequestsM - 1)
	}
	b.ticker.Reset(time.Duration(newPeriod))
	log.Info().Msgf("Ticker was reset, max requests per minute - %d", aErr.MaxRequestsM)
	time.Sleep(time.Second * time.Duration(aErr.RetryAfterS))
	return
}

func (b BGAccrual) processOrder(ctx context.Context, order storage.Order) error {
	resp, aErr := b.client.GetOrderInfo(order)
	if aErr.Err != nil {
		if aErr.Code == http.StatusTooManyRequests {
			b.resetTicker(aErr)
		} else {
			log.Error().Msg(aErr.Err.Error())
			time.Sleep(5 * time.Second)
			return nil
		}
	}

	order.Status = resp.Status
	order.Accrual = resp.Accrual
	// TODO: Тоже в одну транзакцию
	err := b.storage.UpdateOrder(ctx, order)
	if err != nil {
		return err
	}
	user, err := b.storage.GetUser(ctx, order.UserID)
	if err != nil {
		return err
	}
	user.Balance += order.Accrual
	err = b.storage.UpdateUser(ctx, user)
	return err
}

func (b BGAccrual) processOrders(ctx context.Context, orders []storage.Order) {
	for _, order := range orders {
		select {
		case <-b.Done:
			return
		case <-b.ticker.C:
			err := b.processOrder(ctx, order)
			if err != nil {
				log.Error().Msg(err.Error())
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (b BGAccrual) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		orders, err := b.storage.GetIncompleteOrders(ctx)
		if err != nil {
			log.Error().Msg(err.Error())
			time.Sleep(5 * time.Second)
			continue
		}

		b.processOrders(ctx, orders)
	}
}
