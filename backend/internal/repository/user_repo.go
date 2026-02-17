package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, email, password, created_at)
	          VALUES ($1, $2, $3, $4)`

	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.Password, user.CreatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return model.ErrConflict
		}
		return model.WrapInternal(err)
	}
	return nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password, created_at FROM users WHERE email = $1`

	var user model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, model.WrapInternal(err)
	}
	return &user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `SELECT id, email, password, created_at FROM users WHERE id = $1`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, model.WrapInternal(err)
	}
	return &user, nil
}
