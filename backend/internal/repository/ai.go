package repository

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

type AIRepo struct{ pool *pgxpool.Pool }

func NewAIRepo(pool *pgxpool.Pool) *AIRepo { return &AIRepo{pool: pool} }

func (r *AIRepo) CreateConversation(ctx context.Context, userID int64, title string) (*model.AIConversation, error) {
	var c model.AIConversation
	err := r.pool.QueryRow(ctx, `INSERT INTO ai_conversations(user_id,title) VALUES($1,$2)
		RETURNING id,user_id,title,created_at,updated_at`, userID, title).
		Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *AIRepo) OwnsConversation(ctx context.Context, userID, id int64) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ai_conversations WHERE id=$1 AND user_id=$2)`, id, userID).Scan(&ok)
	return ok, err
}

func (r *AIRepo) ListConversations(ctx context.Context, userID int64) ([]*model.AIConversation, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,title,created_at,updated_at FROM ai_conversations
		WHERE user_id=$1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*model.AIConversation{}
	for rows.Next() {
		var c model.AIConversation
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (r *AIRepo) DeleteConversation(ctx context.Context, userID, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM ai_conversations WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AIRepo) AddMessage(ctx context.Context, m *model.AIMessage) error {
	var calls any
	if len(m.ToolCalls) > 0 {
		calls = string(m.ToolCalls)
	}
	return r.pool.QueryRow(ctx, `INSERT INTO ai_messages(conversation_id,role,content,tool_calls,tool_call_id,status)
		VALUES($1,$2,$3,$4::jsonb,$5,$6) RETURNING id,created_at`, m.ConversationID, m.Role, m.Content, calls, nullString(m.ToolCallID), m.Status).
		Scan(&m.ID, &m.CreatedAt)
}

func (r *AIRepo) UpdateMessage(ctx context.Context, id int64, content string, calls json.RawMessage, status string) error {
	var raw any
	if len(calls) > 0 {
		raw = string(calls)
	}
	_, err := r.pool.Exec(ctx, `UPDATE ai_messages SET content=$2,tool_calls=$3::jsonb,status=$4 WHERE id=$1`, id, content, raw, status)
	return err
}

func (r *AIRepo) TouchConversation(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE ai_conversations SET updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *AIRepo) ListMessages(ctx context.Context, userID, conversationID int64, limit int, completeOnly bool) ([]*model.AIMessage, error) {
	where := ""
	if completeOnly {
		where = " AND m.status='complete'"
	}
	rows, err := r.pool.Query(ctx, `SELECT id,conversation_id,role,content,tool_calls,COALESCE(tool_call_id,''),status,created_at FROM (
		SELECT m.* FROM ai_messages m JOIN ai_conversations c ON c.id=m.conversation_id
		WHERE c.user_id=$1 AND c.id=$2`+where+` ORDER BY m.created_at DESC,m.id DESC LIMIT $3) x
		ORDER BY created_at,id`, userID, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*model.AIMessage{}
	for rows.Next() {
		var m model.AIMessage
		var calls []byte
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &calls, &m.ToolCallID, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		if len(calls) > 0 {
			m.ToolCalls = json.RawMessage(calls)
		}
		list = append(list, &m)
	}
	if len(list) == 0 {
		ok, e := r.OwnsConversation(ctx, userID, conversationID)
		if e != nil {
			return nil, e
		}
		if !ok {
			return nil, ErrNotFound
		}
	}
	return list, rows.Err()
}

func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
