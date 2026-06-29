// @title FinTrack API
// @version 1.0
// @description REST API de rastreamento de finanças pessoais
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite "Bearer" seguido do seu access token
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/muriloabranches/fintrack-api/internal/config"
	_ "github.com/muriloabranches/fintrack-api/docs"
	"github.com/muriloabranches/fintrack-api/internal/handler"
	"github.com/muriloabranches/fintrack-api/internal/repository"
	"github.com/muriloabranches/fintrack-api/internal/router"
	"github.com/muriloabranches/fintrack-api/internal/service"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Set up zerolog with configured log level
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(level)

	// 3. Connect to PostgreSQL
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to ping database")
	}
	logger.Info().Msg("connected to database")

	// 4. Apply migrations if AUTO_MIGRATE is enabled
	if cfg.Migration.AutoMigrate {
		logger.Info().Msg("running database migrations")
		m, err := migrate.New("file://migrations", cfg.Database.URL)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to create migrate instance")
		}
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.Fatal().Err(err).Msg("failed to run migrations")
		}
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			logger.Warn().Err(srcErr).Msg("failed to close migration source")
		}
		if dbErr != nil {
			logger.Warn().Err(dbErr).Msg("failed to close migration database")
		}
		logger.Info().Msg("migrations applied successfully")
	}

	// 5. Instantiate repositories
	userRepo := repository.NewUserRepository(pool)
	accountRepo := repository.NewAccountRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	transactionRepo := repository.NewTransactionRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)

	// 6. Instantiate services
	authService := service.NewAuthService(userRepo, tokenRepo, categoryRepo, cfg.JWT)
	accountService := service.NewAccountService(accountRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	transactionService := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo)
	summaryService := service.NewSummaryService(transactionRepo, categoryRepo)

	// 7. Instantiate handlers
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	summaryHandler := handler.NewSummaryHandler(summaryService)

	// 8. Set up router with middleware and routes
	r := router.Setup(router.Dependencies{
		Logger:             logger,
		JWTSecret:          cfg.JWT.Secret,
		AuthHandler:        authHandler,
		AccountHandler:     accountHandler,
		CategoryHandler:    categoryHandler,
		TransactionHandler: transactionHandler,
		SummaryHandler:     summaryHandler,
	})

	// 9. Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 10. Listen for shutdown signals in a goroutine
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info().Str("addr", addr).Msg("starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Block until signal is received
	sig := <-quit
	logger.Info().Str("signal", sig.String()).Msg("shutting down server")

	// 11. Graceful shutdown with configured timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("server shutdown error")
	}

	// 12. Close database pool (deferred above, but explicit log)
	logger.Info().Msg("closing database connections")
	pool.Close()

	logger.Info().Msg("server stopped gracefully")
}
