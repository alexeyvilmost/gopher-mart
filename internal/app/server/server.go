package server

import (
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/storage"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func StartServer() error {
	storage, err := storage.NewDBStorage("port=5432 user=app dbname=shortener password=app host=localhost")
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

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	err = http.ListenAndServe("localhost:8080", r)
	return err
}
