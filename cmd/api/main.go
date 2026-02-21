package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/auth"
	property_ads "github.com/RuanHOliveira/estatehub_api/internal/domain/property_ads"
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
	propertyAdUC := property_ads.NewPropertyAdUsecase(txm)

	// Instância Handlers
	authHandler := auth.NewAuthHandler(authUC)
	viaCEPHandler := viacephandler.NewViaCEPHandler(viaCEPClient)
	propertyAdHandler := property_ads.NewPropertyAdHandler(propertyAdUC, "internal/domain/property_ads/uploads")

	// Criar router
	r := router.NewRouter(router.RouterConfig{
		JwtService:         &jwtService,
		AuthHandler:        authHandler,
		ViaCEPHandler:      viaCEPHandler,
		PropertyAdsHandler: propertyAdHandler,
	})

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.AppPort), r)
}
