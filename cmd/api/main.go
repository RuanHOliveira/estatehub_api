// Package main is the entry point of the EstateHub API.
package main

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/RuanHOliveira/estatehub_api/docs"
	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/auth"
	exchange_rates "github.com/RuanHOliveira/estatehub_api/internal/domain/exchange_rates"
	property_ads "github.com/RuanHOliveira/estatehub_api/internal/domain/property_ads"
	viacephandler "github.com/RuanHOliveira/estatehub_api/internal/domain/viacep"
	"github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/RuanHOliveira/estatehub_api/internal/infra/viacep"
	"github.com/RuanHOliveira/estatehub_api/internal/router"
)

// @title           EstateHub API
// @version         1.0
// @description     API REST para gerenciamento de anúncios imobiliários com autenticação JWT, upload de imagens e integração com ViaCEP.
// @host            localhost:8080
// @BasePath        /v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Use o token retornado por /auth/login ou /auth/register. Formato: Bearer <token>

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
	exchangeRateUC := exchange_rates.NewExchangeRateUsecase(txm)

	// Instância Handlers
	authHandler := auth.NewAuthHandler(authUC)
	viaCEPHandler := viacephandler.NewViaCEPHandler(viaCEPClient)
	propertyAdHandler := property_ads.NewPropertyAdHandler(propertyAdUC, "internal/domain/property_ads/uploads")
	exchangeRateHandler := exchange_rates.NewExchangeRateHandler(exchangeRateUC)

	// Criar router
	r := router.NewRouter(router.RouterConfig{
		JwtService:           &jwtService,
		AuthHandler:          authHandler,
		ViaCEPHandler:        viaCEPHandler,
		PropertyAdsHandler:   propertyAdHandler,
		ExchangeRatesHandler: exchangeRateHandler,
	})

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.AppPort), r)
}
