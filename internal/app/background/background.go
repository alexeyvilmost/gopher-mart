package background

import "gophermart/internal/app/storage"

type BGConfig struct {
	AccrualAddress string
	Storage        storage.Storage
}

type Background struct {
	accrual *BGAccrual
}

func NewBackground(cfg BGConfig) Background {
	accrual := NewBGAccrual(cfg.AccrualAddress, cfg.Storage)
	return Background{accrual: accrual}
}

func (b Background) Run() chan bool {
	go b.accrual.Run()
	return b.accrual.Done
}
