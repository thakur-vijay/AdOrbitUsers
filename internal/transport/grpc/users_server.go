package grpc

import (
	"context"
	"errors"

	usersv1 "github.com/adorbit/users/gen/users/v1"
	"github.com/adorbit/users/internal/users"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/thakur-vijay/adorbit-go/validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersServer struct {
	usersv1.UnimplementedUsersServiceServer
	service *users.Service
}

func NewUsersServer(service *users.Service) *UsersServer {
	return &UsersServer{
		service: service,
	}
}

func (s *UsersServer) CreateUser(
	ctx context.Context,
	req *usersv1.CreateUserRequest,
) (*usersv1.CreateUserResponse, error) {
	dto := users.UserDTO{
		ID:       req.GetId(),
		FullName: req.GetFullName(),
		Email:    req.GetEmail(),
		Role:     req.GetRole(),
	}

	if err := validation.Validator(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.service.Create(ctx, dto); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.Canceled, "request canceled")
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "request deadline exceeded")
		}

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case "23505":
				return nil, status.Error(codes.AlreadyExists, "user already exists")
			case "23503", "23514", "23502", "22P02":
				return nil, status.Error(codes.InvalidArgument, "invalid user data")
			}
		}

		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &usersv1.CreateUserResponse{
		Id: req.GetId(),
	}, nil
}
