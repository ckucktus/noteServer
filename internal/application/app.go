package application

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"test_task/internal/application/server/notes"
	"test_task/internal/config"
	"test_task/internal/infrastructure/integration"
	"test_task/internal/infrastructure/persistence"
	"time"
)

type App struct {
	name           string
	version        string
	cfg            config.Config
	postgresClient *sql.DB
	logger         *slog.Logger
}

func (app *App) SetupConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config.Load, %w", err)
	}
	app.cfg = cfg
	return nil
}

func NewApp(version string) *App {
	return &App{name: "note-storage", version: version}
}
func (app *App) SetupLogger() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	logger.With()
	app.logger = logger
	return nil
}

func (app *App) Run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	textChecker := integration.NewClient(app.cfg.Speller.URL, &http.Client{})

	userStorage, err := persistence.NewUserStorage(app.cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("config.Load, %w", err)
	}
	noteStorage, err := persistence.NewNoteStorage(app.cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("config.Load, %w", err)
	}

	router := echo.New()
	router.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{Validator: func(login string, password string, e echo.Context) (bool, error) {
		_, err := userStorage.UserAuthentication(login, password) // todo add user into context
		if err != nil {
			return false, err
		}
		return true, nil
	}}))

	noteServer := notes.NewServer(noteStorage, userStorage, textChecker)

	router.POST("/note", noteServer.PostV1CreateNote)
	router.GET("/note/list", noteServer.GetV1ListNotes)

	srv := &http.Server{
		Addr:         app.cfg.HTTPServer.ListenAddress,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		ErrorLog:     &log.Logger{},
		BaseContext: func(listener net.Listener) context.Context {
			newCtx := context.WithValue(ctx, "addr", listener.Addr().String())
			return newCtx
		},
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("ListenAndServe: %v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// todo close storage
	// todo close logger

	if err := srv.Shutdown(shutdownCtx); err != nil {
		app.logger.Error(fmt.Sprintf("Shutdown: %w", err))

	}

	return nil

}
