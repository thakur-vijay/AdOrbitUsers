package app

import (
	"database/sql"
	"net/http"

	"github.com/adorbit/users/internal/api"
	"github.com/adorbit/users/internal/users"
	"github.com/thakur-vijay/adorbit-go/middleware"
)

func BuildRouter(db *sql.DB, publicKeyPEM string) (http.Handler, *users.Service, error) {
	jwtMiddleware, err := middleware.NewJWTMiddleware(
		publicKeyPEM,
		"adorbit-authentication",
	)
	if err != nil {
		return nil, nil, err
	}

	repository := users.NewRepository(db)
	userService := users.NewService(repository)
	userHandler := users.NewHandler(userService)

	router := api.NewRouter(api.Dependencies{
		UserHandler:   userHandler,
		JWTMiddleware: jwtMiddleware,
	})

	return router, userService, nil
}
