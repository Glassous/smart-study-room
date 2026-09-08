CREATE TABLE IF NOT EXISTS ai_conversations (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(80) NOT NULL DEFAULT '新对话',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ai_conversations_user_updated
    ON ai_conversations(user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS ai_messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL CHECK (role IN ('system', 'user', 'assistant', 'tool')),
    content         TEXT NOT NULL DEFAULT '',
    tool_calls      JSONB,
    tool_call_id    VARCHAR(120),
    status          VARCHAR(20) NOT NULL DEFAULT 'complete'
                    CHECK (status IN ('streaming', 'complete', 'failed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation_created
    ON ai_messages(conversation_id, created_at, id);

COMMENT ON TABLE ai_conversations IS '学生 AI 助手会话';
COMMENT ON TABLE ai_messages IS 'OpenAI role/content/tool_calls 兼容消息历史';
