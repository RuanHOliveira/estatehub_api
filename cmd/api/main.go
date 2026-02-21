package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	"github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	conn := postgresql.Connect(ctx, cfg.PgConfig)
	defer conn.Close(ctx)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API Online"))
	})

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.AppPort), r)
}
