package app

import (
	"database/sql"
	"net/http"

	"github.com/adorbit/users/internal/api"
	"github.com/adorbit/users/internal/users"
	"github.com/thakur-vijay/adorbit-go/authclient"
	"github.com/thakur-vijay/adorbit-go/middleware"
)

func BuildRouter(db *sql.DB) (http.Handler, func(), error) {
	conn, err := authclient.Dial("127.0.0.1:9090")
	if err != nil {
		return nil, nil, err
	}

	authValidator := authclient.New(conn)
	authMiddleware := middleware.NewAuthMiddleware(authValidator)

	userService := users.NewService()
	userHandler := users.NewHandler(userService)

	router := api.NewRouter(api.Dependencies{
		UserHandler:    userHandler,
		AuthMiddleware: authMiddleware,
	})

	cleanup := func() {
		_ = conn.Close()
	}

	return router, cleanup, nil
}
