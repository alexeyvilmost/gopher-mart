package server

import (
	"fmt"
	"gophermart/internal/app/auth"
	"gophermart/internal/app/background"
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/storage"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func StartServer() error {
	cfg := NewConfig()

	storage, err := storage.NewDBStorage(cfg.DBConnection)
	if err != nil {
		log.Error().Err(err).Msg("Error while creating storage")
		return err
	}
	err = storage.Init()
	if err != nil {
		log.Error().Err(err).Msg("Error while init storage")
		return err
	}
	bg := background.NewBackground(background.BGConfig{Storage: storage, AccrualAddress: cfg.AccrualAddress})

	a := auth.Auth{Storage: storage, SecretKey: cfg.SecretKey}
	h := handlers.Handlers{Storage: storage, Auth: a}

	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.H(h.Register))
		r.Post("/login", handlers.H(h.Login))

		r.Group(func(r chi.Router) {
			r.Use(a.WithAuth)
			r.Post("/orders", handlers.H(h.AddOrder))
			r.Get("/orders", handlers.H(h.GetOrders))
			r.Get("/balance", handlers.H(h.GetBalance))
			r.Post("/withdraw", handlers.H(h.Withdraw))
			r.Get("/withdrawals", handlers.H(h.GetWithdrawals))
		})
	})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	bgDone := bg.Run()

	log.Info().Msg(fmt.Sprintf("Server listening on address %s", cfg.RunAddress))
	err = http.ListenAndServe(cfg.RunAddress, r)
	bgDone <- true
	return err
}
