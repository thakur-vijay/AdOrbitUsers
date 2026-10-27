package users

import (
	"net/http"

	"github.com/thakur-vijay/adorbit-go/middleware"
)

func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
	authMiddleware *middleware.Middleware,
) {
	mux.Handle("POST /users", authMiddleware.RequireAuth(http.HandlerFunc(handler.Create)))
	mux.Handle("GET /users", authMiddleware.RequireAuth(http.HandlerFunc(handler.Fetch)))
	mux.Handle("GET /users/{id}", authMiddleware.RequireAuth(http.HandlerFunc(handler.FetchByID)))
	mux.Handle("PUT /users/{id}", authMiddleware.RequireAuth(http.HandlerFunc(handler.Update)))
	mux.Handle("DELETE /users/{id}", authMiddleware.RequireAuth(http.HandlerFunc(handler.Delete)))
}
