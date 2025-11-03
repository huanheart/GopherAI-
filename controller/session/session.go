package session

import (
	"GopherAI/common/code"
	"GopherAI/controller"
	"GopherAI/model"
	"GopherAI/service/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	GetUserSessionsResponse struct {
		controller.Response
		Data map[string][]interface{} `json:"data,omitempty"`
	}
	CreateSessionAndSendMessageRequest struct {
		UserQuestion string `json:"question" binding:"required"`  // 用户问题;
		ModelType    string `json:"modelType" binding:"required"` // 模型类型;
	}

	CreateSessionAndSendMessageResponse struct {
		AiInformation string `json:"Information,omitempty"` // AI回答
		SessionID     string `json:"sessionId,omitempty"`   // 当前会话ID
		controller.Response
	}

	ChatSendRequest struct {
		UserQuestion string `json:"question" binding:"required"`            // 用户问题;
		ModelType    string `json:"modelType" binding:"required"`           // 模型类型;
		SessionID    string `json:"sessionId,omitempty" binding:"required"` // 当前会话ID
	}

	ChatSendResponse struct {
		AiInformation string `json:"Information,omitempty"` // AI回答
		controller.Response
	}

	ChatHistoryRequest struct {
		SessionID string `json:"sessionId,omitempty" binding:"required"` // 当前会话ID
	}
	ChatHistoryResponse struct {
		History []model.History `json:"history"`
		controller.Response
	}
)

// 这个写错了  todo：待更改，因为这边应该只是获取当前用户所有的sessionID
// 而这边的逻辑其实是应该放到初始化操作去的
func GetUserSessionsByUserID(c *gin.Context) {
	res := new(GetUserSessionsResponse)
	userID := c.GetInt64("user_id") // From JWT middleware

	userSessions, err := session.GetUserSessionsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	// Convert []model.Message to []interface{} for JSON serialization
	data := make(map[string][]interface{})
	for sessionID, msgs := range userSessions {
		msgsInterface := make([]interface{}, len(msgs))
		for i, msg := range msgs {
			msgsInterface[i] = map[string]interface{}{
				"id":         msg.ID,
				"session_id": msg.SessionID,
				"content":    msg.Content,
				"created_at": msg.CreatedAt,
			}
		}
		data[sessionID] = msgsInterface
	}

	res.Success()
	res.Data = data
	c.JSON(http.StatusOK, res)
}

func CreateSessionAndSendMessage(c *gin.Context) {
	req := new(CreateSessionAndSendMessageRequest)
	res := new(CreateSessionAndSendMessageResponse)
	userID := c.GetInt64("user_id") // From JWT middleware
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	//内部会创建会话并发送消息，并会将AI回答、当前会话返回
	session_id, aiInformation, code_ := session.CreateSessionAndSendMessage(userID, req.UserQuestion, req.ModelType)

	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return
	}

	res.Success()
	res.AiInformation = aiInformation
	res.SessionID = session_id
	c.JSON(http.StatusOK, res)
}

func ChatSend(c *gin.Context) {
	req := new(ChatSendRequest)
	res := new(ChatSendResponse)
	userID := c.GetInt64("user_id") // From JWT middleware
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	// 发送消息，并会将AI回答返回
	aiInformation, code_ := session.ChatSend(userID, req.SessionID, req.UserQuestion, req.ModelType)

	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return
	}

	res.Success()
	res.AiInformation = aiInformation
	c.JSON(http.StatusOK, res)
}

func ChatHistory(c *gin.Context) {
	req := new(ChatHistoryRequest)
	res := new(ChatHistoryResponse)
	userID := c.GetInt64("user_id") // From JWT middleware
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	history, code_ := session.GetChatHistory(userID, req.SessionID)
	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return
	}

	res.Success()
	res.History = history
	c.JSON(http.StatusOK, res)
}
