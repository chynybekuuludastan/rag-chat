package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
)

type chunkRepo struct {
	pool *pgxpool.Pool
}

func NewChunkRepository(pool *pgxpool.Pool) ChunkRepository {
	return &chunkRepo{pool: pool}
}

func (r *chunkRepo) CreateBatch(ctx context.Context, chunks []model.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.WrapInternal(err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO chunks (id, document_id, content, chunk_index, embedding, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`

	for _, chunk := range chunks {
		_, err := tx.Exec(ctx, query,
			chunk.ID, chunk.DocumentID, chunk.Content,
			chunk.ChunkIndex, chunk.Embedding, chunk.CreatedAt,
		)
		if err != nil {
			return model.WrapInternal(fmt.Errorf("insert chunk %d: %w", chunk.ChunkIndex, err))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.WrapInternal(err)
	}
	return nil
}

func (r *chunkRepo) SearchSimilar(ctx context.Context, embedding []float32, userID uuid.UUID, limit int, threshold float64) ([]model.ChunkWithDocument, error) {
	query := `SELECT c.id, c.document_id, c.content, c.chunk_index, c.created_at,
	                 d.filename,
	                 1 - (c.embedding <=> $1) AS similarity
	          FROM chunks c
	          JOIN documents d ON d.id = c.document_id
	          WHERE d.user_id = $2
	            AND 1 - (c.embedding <=> $1) > $3
	          ORDER BY c.embedding <=> $1
	          LIMIT $4`

	vec := pgvector.NewVector(embedding)
	rows, err := r.pool.Query(ctx, query, vec, userID, threshold, limit)
	if err != nil {
		return nil, model.WrapInternal(err)
	}
	defer rows.Close()

	var results []model.ChunkWithDocument
	for rows.Next() {
		var cwd model.ChunkWithDocument
		if err := rows.Scan(
			&cwd.ID, &cwd.DocumentID, &cwd.Content, &cwd.ChunkIndex,
			&cwd.CreatedAt, &cwd.DocumentFilename, &cwd.Similarity,
		); err != nil {
			return nil, model.WrapInternal(err)
		}
		results = append(results, cwd)
	}

	if err := rows.Err(); err != nil {
		return nil, model.WrapInternal(err)
	}

	return results, nil
}

func (r *chunkRepo) DeleteByDocument(ctx context.Context, documentID uuid.UUID) error {
	query := `DELETE FROM chunks WHERE document_id = $1`

	_, err := r.pool.Exec(ctx, query, documentID)
	if err != nil {
		return model.WrapInternal(err)
	}
	return nil
}
