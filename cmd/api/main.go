package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/auth"
	"github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/RuanHOliveira/estatehub_api/internal/router"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	conn := postgresql.Connect(ctx, cfg.PgConfig)
	defer conn.Close(ctx)

	// Instância JwtService
	jwtService := security.NewJwtService(cfg.AuthConfig)

	// Instância Repositories
	repo := repo.New(conn)

	// Instância Usecases
	authUC := auth.NewAuthUsecase(repo, conn, cfg.AuthConfig, jwtService)

	// Instância Handlers
	authHandler := auth.NewAuthHandler(authUC)

	// Criar router
	r := router.NewRouter(router.RouterConfig{
		JwtService:  &jwtService,
		AuthHandler: authHandler,
	})

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.AppPort), r)
}
