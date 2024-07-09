package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	server "github.com/wopoczynski/playground/internal/http/echo"
	"github.com/wopoczynski/playground/internal/initialize"
)

type Config struct {
	*server.Config       `env:", prefix=HTTP_"`
	*initialize.DBConfig `env:", prefix=DB_"`
}

type ApplicationContainer struct {
	cfg    *Config
	db     *gorm.DB
	server server.Server
}

func New(cfg Config) (*ApplicationContainer, error) {
	db, err := initialize.DB(*cfg.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("db initialize error: %w", err)
	}

	handler := server.NewHandler(db)

	server := server.New(cfg.Config, handler)

	return &ApplicationContainer{
		cfg:    &cfg,
		db:     db,
		server: server,
	}, nil
}

func (s *ApplicationContainer) Init(ctx context.Context) {
	err := initialize.Automigrate(ctx, s.db)
	if err != nil {
		panic(err)
	}
}

func (s *ApplicationContainer) Start(ctx context.Context) {
	go func() {
		err := s.server.Start(":" + s.cfg.HTTPServerPort)
		if errors.Is(err, http.ErrServerClosed) {
			log.Err(err).Msg("Server shutdown")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit

	const shutdownTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("shutting down server...")
	}

	log.Info().Msg("server stopped gracefully")
}
