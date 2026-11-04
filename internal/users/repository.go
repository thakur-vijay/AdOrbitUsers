package users

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, dto UserDTO) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (id, full_name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`, dto.ID, dto.FullName, dto.Email, dto.Role, dto.CreatedAt, dto.UpdatedAt)
	return err
}

func (r *Repository) FetchByID(ctx context.Context, userID string) (*UserDTO, error) {
	var user UserDTO
	err := r.db.QueryRowContext(ctx, `SELECT id, full_name, email, role, created_at, updated_at FROM users WHERE id=$1`, userID).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
