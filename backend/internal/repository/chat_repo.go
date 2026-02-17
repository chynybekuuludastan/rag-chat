package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
)

type chatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) ChatRepository {
	return &chatRepo{pool: pool}
}

func (r *chatRepo) CreateSession(ctx context.Context, session *model.ChatSession) error {
	query := `INSERT INTO chat_sessions (id, user_id, title, created_at)
	          VALUES ($1, $2, $3, $4)`

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.Title, session.CreatedAt,
	)
	if err != nil {
		return model.WrapInternal(err)
	}
	return nil
}

func (r *chatRepo) GetSession(ctx context.Context, id uuid.UUID) (*model.ChatSession, error) {
	query := `SELECT id, user_id, title, created_at FROM chat_sessions WHERE id = $1`

	var session model.ChatSession
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.Title, &session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, model.WrapInternal(err)
	}
	return &session, nil
}

func (r *chatRepo) ListSessionsByUser(ctx context.Context, userID uuid.UUID) ([]model.ChatSession, error) {
	query := `SELECT id, user_id, title, created_at
	          FROM chat_sessions WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, model.WrapInternal(err)
	}
	defer rows.Close()

	var sessions []model.ChatSession
	for rows.Next() {
		var s model.ChatSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt); err != nil {
			return nil, model.WrapInternal(err)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, model.WrapInternal(err)
	}

	return sessions, nil
}
func (r *chatRepo) CreateMessage(ctx context.Context, msg *model.Message) error {
	query := `INSERT INTO messages (id, session_id, role, content, source_chunks, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
		msg.ID, msg.SessionID, msg.Role, msg.Content, msg.SourceChunks, msg.CreatedAt,
	)
	if err != nil {
		return model.WrapInternal(err)
	}
	return nil
}

func (r *chatRepo) GetMessagesBySession(ctx context.Context, sessionID uuid.UUID) ([]model.Message, error) {
	query := `SELECT id, session_id, role, content, source_chunks, created_at
	          FROM messages WHERE session_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, model.WrapInternal(err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(
			&m.ID, &m.SessionID, &m.Role, &m.Content, &m.SourceChunks, &m.CreatedAt,
		); err != nil {
			return nil, model.WrapInternal(err)
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, model.WrapInternal(err)
	}

	return messages, nil
}
