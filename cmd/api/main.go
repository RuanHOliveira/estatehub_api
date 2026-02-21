package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/auth"
	viacephandler "github.com/RuanHOliveira/estatehub_api/internal/domain/viacep"
	"github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/RuanHOliveira/estatehub_api/internal/infra/viacep"
	"github.com/RuanHOliveira/estatehub_api/internal/router"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	conn := postgresql.Connect(ctx, cfg.PgConfig)
	defer conn.Close(ctx)

	// Instância JwtService
	jwtService := security.NewJwtService(cfg.AuthConfig)

	// Instância Clients
	viaCEPClient := viacep.NewViaCEPClient()

	// Instância Repositories
	queries := repo.New(conn)

	// Instância TxManager
	txm := postgresql.NewPgTxManager(queries, conn)

	// Instância Usecases
	authUC := auth.NewAuthUsecase(txm, jwtService)

	// Instância Handlers
	authHandler := auth.NewAuthHandler(authUC)
	viaCEPH := viacephandler.NewViaCEPHandler(viaCEPClient)

	// Criar router
	r := router.NewRouter(router.RouterConfig{
		JwtService:    &jwtService,
		AuthHandler:   authHandler,
		ViaCEPHandler: viaCEPH,
	})

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.AppPort), r)
}
