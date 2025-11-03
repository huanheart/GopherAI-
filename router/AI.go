package router

import (
	"GopherAI/controller/session"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup) {

	// 聊天相关接口
	{
		r.GET("/chat/sessions", session.GetUserSessionsByUserID)              // 获取当前用户的会话历史
		r.POST("/chat/send-new-session", session.CreateSessionAndSendMessage) // 创建新会话并发送消息
		// r.GET("/chat", AI.Chat)                             // ChatHandler
		r.POST("/chat/send", session.ChatSend)       // ChatSendHandler
		r.POST("/chat/history", session.ChatHistory) // ChatHistoryHandler
		// r.POST("/chat/tts", AI.ChatSpeech)                  // ChatSpeechHandler
	}
}
