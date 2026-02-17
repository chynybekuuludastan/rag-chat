package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
)

type documentRepo struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) DocumentRepository {
	return &documentRepo{pool: pool}
}

func (r *documentRepo) Create(ctx context.Context, doc *model.Document) error {
	query := `INSERT INTO documents (id, user_id, filename, file_type, file_size, chunk_count, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		doc.ID, doc.UserID, doc.Filename, doc.FileType, doc.FileSize, doc.ChunkCount, doc.CreatedAt,
	)
	if err != nil {
		return model.WrapInternal(err)
	}
	return nil
}

func (r *documentRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Document, error) {
	query := `SELECT id, user_id, filename, file_type, file_size, chunk_count, created_at
	          FROM documents WHERE id = $1`

	var doc model.Document
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.UserID, &doc.Filename, &doc.FileType,
		&doc.FileSize, &doc.ChunkCount, &doc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, model.WrapInternal(err)
	}
	return &doc, nil
}

func (r *documentRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Document, error) {
	query := `SELECT id, user_id, filename, file_type, file_size, chunk_count, created_at
	          FROM documents WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, model.WrapInternal(err)
	}
	defer rows.Close()

	var docs []model.Document
	for rows.Next() {
		var doc model.Document
		if err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.Filename, &doc.FileType,
			&doc.FileSize, &doc.ChunkCount, &doc.CreatedAt,
		); err != nil {
			return nil, model.WrapInternal(err)
		}
		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, model.WrapInternal(err)
	}

	return docs, nil
}

func (r *documentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return model.WrapInternal(err)
	}

	if result.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
