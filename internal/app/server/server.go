package server

import (
	"fmt"
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
	h := handlers.Handlers{Storage: storage}

	r := chi.NewRouter()
	r.Use()
	r.Post("/register", handlers.H(h.Register))
	r.Post("/login", handlers.H(h.Login))

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg(fmt.Sprintf("Server listening on address %s", cfg.RunAddress))
	err = http.ListenAndServe(cfg.RunAddress, r)
	return err
}
