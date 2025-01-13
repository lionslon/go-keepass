package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lionslon/go-keepass/internal/auth"
	"github.com/lionslon/go-keepass/internal/deadline"
	"github.com/lionslon/go-keepass/internal/logger"
	"github.com/lionslon/go-keepass/internal/server/config"
	"github.com/lionslon/go-keepass/internal/server/handlers"
	"github.com/lionslon/go-keepass/internal/storage"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const (
	shutdownTime = 5 * time.Second
)

type App struct {
	server     *http.Server
	notifyStop context.CancelFunc
}

func Create(cfg *config.Config, storage *storage.KeeperStorage) (*App, error) {

	// Инициализируем объект для создания/проверки jwt
	auth.Initialize(cfg)
	// Регистрируем хэндлеры в роутере
	router := chi.NewRouter()
	// Подключаем middleware логирования
	router.Use(logger.Middleware)
	// Подключаем middleware deadline context
	router.Use(deadline.Middleware)
	// Подключаем storage
	keeperHandler := handlers.NewKeeperHandler(storage)
	// Регистрируем роутер
	keeperHandler.Register(router)

	return &App{
		server: &http.Server{
			Addr:    cfg.Endpoint,
			Handler: router,
		},
	}, nil
}

func (m *App) Run() {
	if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("cannot listen: %s\n", err)
	}
}

func (m *App) ServerDone() <-chan struct{} {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	m.notifyStop = stop
	return ctx.Done()
}

func (m *App) Shutdown() error {
	defer m.notifyStop()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
