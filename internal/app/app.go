package app

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/adorbit/users/internal/config"
	"github.com/adorbit/users/internal/database"
	"github.com/adorbit/users/internal/messaging"
	"github.com/adorbit/users/internal/users"
	"github.com/thakur-vijay/adorbit-go/apperrors"
)

type App struct {
	DB       *sql.DB
	Router   http.Handler
	GRPCPort string
	UsersSvc *users.Service
	RabbitMQ *messaging.RabbitMQ
	Consumer *messaging.Consumer
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, apperrors.NewConfigError("app configuration required", nil)
	}

	db, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(db, cfg.DatabaseURL); err != nil {
		_ = db.Close()
		return nil, err
	}

	router, userService, err := BuildRouter(db, cfg.JWTPublicKey)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	rabbitMQ, err := messaging.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	consumer := messaging.NewConsumer(rabbitMQ, userService)

	return &App{
		DB:       db,
		Router:   router,
		UsersSvc: userService,
		GRPCPort: cfg.GRPCPort,
		RabbitMQ: rabbitMQ,
		Consumer: consumer,
	}, nil
}
