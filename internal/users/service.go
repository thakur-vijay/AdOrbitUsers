package users

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (s *Service) Create(ctx context.Context, dto UserDTO) (*UserDTO, error) {
	existingUser, err := s.repository.FetchByID(ctx, dto.ID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewInternal(
			"Failed to check existing user",
			err,
		)
	}

	if existingUser != nil {
		if existingUser.Email != dto.Email ||
			existingUser.FullName != dto.FullName ||
			existingUser.Role != dto.Role {
			return nil, apperrors.NewConflict(
				"User already exists with different details",
				nil,
			)
		}

		return existingUser, nil
	}

	if err := s.repository.Create(ctx, dto); err != nil {
		return nil, err
	}

	return &dto, nil
}

func (s *Service) FetchByID(ctx context.Context, userID string) (*UserDTO, error) {
	userCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := s.repository.FetchByID(userCtx, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NewNotFound("User not found", err)
		} else {
			return nil, err
		}
	} else {
		return user, nil
	}
}
