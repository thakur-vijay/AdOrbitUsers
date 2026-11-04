package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/thakur-vijay/adorbit-go/apperrors"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, dto UserDTO) error {
	existingUser, err := s.repository.FetchByID(ctx, dto.ID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return apperrors.NewInternal(
			"Failed to check existing user",
			err,
		)
	}

	if existingUser != nil {
		if existingUser.Email != dto.Email ||
			existingUser.FullName != dto.FullName ||
			existingUser.Role != dto.Role {
			return apperrors.NewConflict(
				"User already exists with different details",
				nil,
			)
		}

		// Same user already exists. Treat as successful retry.
		return nil
	}

	return s.repository.Create(ctx, dto)
}
