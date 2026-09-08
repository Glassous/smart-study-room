package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

type AIHandler struct{ service *service.AIService }

func NewAIHandler(s *service.AIService) *AIHandler { return &AIHandler{service: s} }
func (h *AIHandler) ListConversations(c *gin.Context) {
	list, err := h.service.ListConversations(c, middleware.CurrentUID(c))
	if err != nil {
		fail(c, 500, "读取会话失败")
		return
	}
	ok(c, list)
}
func (h *AIHandler) CreateConversation(c *gin.Context) {
	item, err := h.service.CreateConversation(c, middleware.CurrentUID(c))
	if err != nil {
		fail(c, 500, "创建会话失败")
		return
	}
	ok(c, item)
}
func (h *AIHandler) ListMessages(c *gin.Context) {
	id, valid := parseAIID(c)
	if !valid {
		return
	}
	list, err := h.service.ListMessages(c, middleware.CurrentUID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		fail(c, 404, "会话不存在")
		return
	}
	if err != nil {
		fail(c, 500, "读取消息失败")
		return
	}
	ok(c, list)
}
func (h *AIHandler) DeleteConversation(c *gin.Context) {
	id, valid := parseAIID(c)
	if !valid {
		return
	}
	err := h.service.DeleteConversation(c, middleware.CurrentUID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		fail(c, 404, "会话不存在")
		return
	}
	if err != nil {
		fail(c, 500, "删除会话失败")
		return
	}
	ok(c, gin.H{"deleted": true})
}
func parseAIID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, 400, "会话 ID 无效")
		return 0, false
	}
	return id, true
}

func (h *AIHandler) Stream(c *gin.Context) {
	var req model.AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "消息不能为空或过长")
		return
	}
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	write := func(event string, payload any) error {
		b, _ := json.Marshal(payload)
		_, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, b)
		c.Writer.Flush()
		return err
	}
	conversationID, messageID, err := h.service.Chat(c.Request.Context(), middleware.CurrentUID(c), req.ConversationID, req.Message,
		func(conversationID, messageID int64) error {
			return write("meta", gin.H{"conversation_id": conversationID, "message_id": messageID})
		},
		func(delta string) error { return write("delta", gin.H{"content": delta}) }, func(cards []model.AIRecommendation) error { return write("card", gin.H{"recommendations": cards}) })
	if err != nil {
		_ = write("error", gin.H{"message": friendlyAIError(err)})
		return
	}
	_ = write("done", gin.H{"conversation_id": conversationID, "message_id": messageID, "status": "complete"})
}
func friendlyAIError(err error) string {
	if errors.Is(err, service.ErrAINotConfigured) {
		return "AI 助手尚未配置，请联系管理员"
	}
	if errors.Is(err, repository.ErrNotFound) {
		return "会话不存在"
	}
	return "AI 暂时不可用，请稍后重试"
}
