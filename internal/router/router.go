package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/muriloabranches/fintrack-api/internal/handler"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
)

// Dependencies holds all handler dependencies required to set up the router.
type Dependencies struct {
	Logger             zerolog.Logger
	JWTSecret          string
	AuthHandler        *handler.AuthHandler
	AccountHandler     *handler.AccountHandler
	CategoryHandler    *handler.CategoryHandler
	TransactionHandler *handler.TransactionHandler
	SummaryHandler     *handler.SummaryHandler
}

// Setup creates and configures the chi router with the middleware stack
// (Recovery → RequestID → Logger) and registers all routes.
func Setup(deps Dependencies) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares in order: Recovery → RequestID → Logger
	r.Use(middleware.Recovery)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(deps.Logger))

	// Public routes (no auth required)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", deps.AuthHandler.Register)
		r.Post("/login", deps.AuthHandler.Login)
		r.Post("/refresh", deps.AuthHandler.Refresh)
	})

	// Swagger (public)
	r.Get("/swagger/*", httpSwagger.Handler())

	// Protected routes (auth middleware applied)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(deps.JWTSecret))

		r.Route("/api/v1/accounts", func(r chi.Router) {
			r.Post("/", deps.AccountHandler.Create)
			r.Get("/", deps.AccountHandler.List)
			r.Put("/{id}", deps.AccountHandler.Update)
			r.Delete("/{id}", deps.AccountHandler.Delete)
		})

		r.Route("/api/v1/categories", func(r chi.Router) {
			r.Post("/", deps.CategoryHandler.Create)
			r.Get("/", deps.CategoryHandler.List)
			r.Put("/{id}", deps.CategoryHandler.Update)
			r.Delete("/{id}", deps.CategoryHandler.Delete)
		})

		r.Route("/api/v1/transactions", func(r chi.Router) {
			r.Post("/", deps.TransactionHandler.Create)
			r.Get("/", deps.TransactionHandler.List)
			r.Put("/{id}", deps.TransactionHandler.Update)
			r.Delete("/{id}", deps.TransactionHandler.Delete)
		})

		r.Route("/api/v1/summary", func(r chi.Router) {
			r.Get("/", deps.SummaryHandler.GetSummary)
			r.Get("/categories", deps.SummaryHandler.GetCategorySummary)
			r.Get("/trend", deps.SummaryHandler.GetMonthlyTrend)
		})
	})

	return r
}
