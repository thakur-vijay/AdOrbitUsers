package users

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, dto UserDTO) error {
	return s.repository.Create(ctx, dto)
}
