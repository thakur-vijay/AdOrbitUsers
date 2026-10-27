package api

import (
	"net/http"

	"github.com/adorbit/users/internal/users"
	"github.com/thakur-vijay/adorbit-go/middleware"
)

const v1Prefix = "/api/v1"

type Dependencies struct {
	UserHandler   *users.Handler
	JWTMiddleware *middleware.JWTMiddleware
}

func NewRouter(dependencies Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

	v1Mux := http.NewServeMux()
	users.RegisterRoutes(v1Mux, dependencies.UserHandler, dependencies.JWTMiddleware)
	mux.Handle(v1Prefix+"/", http.StripPrefix(v1Prefix, v1Mux))

	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
