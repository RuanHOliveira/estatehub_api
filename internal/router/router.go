package router

import (
	"net/http"
	"time"

	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/auth"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/exchange_rates"
	property_ads "github.com/RuanHOliveira/estatehub_api/internal/domain/property_ads"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/viacep"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type RouterConfig struct {
	JwtService           *security.JwtService
	AuthHandler          *auth.AuthHandler
	ViaCEPHandler        *viacep.ViaCEPHandler
	PropertyAdsHandler   *property_ads.PropertyAdHandler
	ExchangeRatesHandler *exchange_rates.ExchangeRateHandler
}

func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", healthHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", cfg.AuthHandler.Register)
			r.Post("/login", cfg.AuthHandler.Login)
		})

		// ViaCEP
		r.Route("/viacep", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(*cfg.JwtService))

			r.Get("/{cep}", cfg.ViaCEPHandler.FindAddress)
		})

		// PropertyAds
		r.Route("/property-ads", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(*cfg.JwtService))

			r.Get("/", cfg.PropertyAdsHandler.ListPropertyAds)
			r.Post("/", cfg.PropertyAdsHandler.CreatePropertyAd)
			r.Delete("/{id}", cfg.PropertyAdsHandler.DeletePropertyAd)
		})

		// ExchangeRates
		r.Route("/exchange-rates", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(*cfg.JwtService))

			r.Get("/", cfg.ExchangeRatesHandler.ListAllExchangeRates)
			r.Post("/", cfg.ExchangeRatesHandler.CreateExchangeRate)
		})
	})

	fileServer := http.FileServer(http.Dir("./internal/domain/property_ads/uploads"))
	r.Handle("/uploads/property_ads/*", http.StripPrefix("/uploads/property_ads", fileServer))

	return r
}

// healthHandler godoc
// @Summary      Health check
// @Description  Verifica se a API está online
// @Tags         health
// @Produce      plain
// @Success      200 {string} string "API Online!"
// @Router       /health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API Online!"))
}
