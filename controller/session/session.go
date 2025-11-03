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
		Sessions []model.SessionInfo `json:"sessions,omitempty"`
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

func GetUserSessionsByUserName(c *gin.Context) {
	res := new(GetUserSessionsResponse)
	userName := c.GetString("userName") // From JWT middleware

	userSessions, err := session.GetUserSessionsByUserName(userName)
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.Sessions = userSessions
	c.JSON(http.StatusOK, res)
}

func CreateSessionAndSendMessage(c *gin.Context) {
	req := new(CreateSessionAndSendMessageRequest)
	res := new(CreateSessionAndSendMessageResponse)
	userName := c.GetString("userName") // From JWT middleware
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	//内部会创建会话并发送消息，并会将AI回答、当前会话返回
	session_id, aiInformation, code_ := session.CreateSessionAndSendMessage(userName, req.UserQuestion, req.ModelType)

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
	userName := c.GetString("userName") // From JWT middleware
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	// 发送消息，并会将AI回答返回
	aiInformation, code_ := session.ChatSend(userName, req.SessionID, req.UserQuestion, req.ModelType)

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
	userName := c.GetString("userName") // From JWT middleware
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	history, code_ := session.GetChatHistory(userName, req.SessionID)
	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return
	}

	res.Success()
	res.History = history
	c.JSON(http.StatusOK, res)
}
