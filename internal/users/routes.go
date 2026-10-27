package users

import (
	"net/http"

	"github.com/thakur-vijay/adorbit-go/middleware"
)

func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
	jwtMiddleware *middleware.JWTMiddleware,
) {
	mux.Handle("POST /users", jwtMiddleware.RequireAuth(http.HandlerFunc(handler.Create)))
	mux.Handle("GET /users", jwtMiddleware.RequireAuth(http.HandlerFunc(handler.Fetch)))
	mux.Handle("GET /users/{id}", jwtMiddleware.RequireAuth(http.HandlerFunc(handler.FetchByID)))
	mux.Handle("PUT /users/{id}", jwtMiddleware.RequireAuth(http.HandlerFunc(handler.Update)))
	mux.Handle("DELETE /users/{id}", jwtMiddleware.RequireAuth(http.HandlerFunc(handler.Delete)))
}
