package session

import (
	"GopherAI/common/code"
	"GopherAI/dao/message"
	"GopherAI/dao/session"
	"GopherAI/model"
)

func GetUserSessionsByUserID(userID int64) (map[string][]model.Message, error) {
	//获取用户的所有会话
	sessions, err := session.GetSessionsByUserID(userID)
	if err != nil {
		return nil, err
	}

	sessionIDs := make([]string, len(sessions))
	for i, s := range sessions {
		sessionIDs[i] = s.ID
	}
	//获取这些会话的所有消息（存在于会话表中，且与会话ID相关联）
	//一次查询，而不是通过遍历sessionIDs多次查询mysql
	allMsgs, err := message.GetMessagesBySessionIDs(sessionIDs)
	if err != nil {
		return nil, err
	}

	msgMap := make(map[string][]model.Message)
	for _, msg := range allMsgs {
		msgMap[msg.SessionID] = append(msgMap[msg.SessionID], msg)
	}

	return msgMap, nil
}

func CreateSessionAndSendMessage(userID int64, userQuestion string, modelType string) (string, string, code.Code) {
	//1：创建一个新的会话
	newSession := &model.Session{
		UserID: userID,
		Title:  userQuestion, // 可以根据需求设置标题，这边暂时用用户第一次的问题作为标题
	}
	createdSession, err := session.CreateSession(newSession)
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	//2：将用户问题存储为消息
	userMessage := &model.Message{
		SessionID: createdSession.ID,
		Content:   userQuestion,
	}
	_, err = message.CreateMessage(userMessage)
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	//3：调用AI模型获取回答
	//todo：这个函数后续实现
	aiResponse, err := GetAIResponse(userID, userQuestion, modelType)
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	//4：将AI回答存储为消息
	aiMessage := &model.Message{
		SessionID: createdSession.ID,
		Content:   aiResponse,
	}
	_, err = message.CreateMessage(aiMessage)
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	return createdSession.ID, aiResponse, code.CodeSuccess
}

func ChatSend(userID int64, sessionID string, userQuestion string, modelType string) (string, code.Code) {
	var aiResponse string
	//1：将用户问题存储为消息
	userMessage := &model.Message{
		SessionID: sessionID,
		Content:   userQuestion,
	}
	_, err := message.CreateMessage(userMessage)
	if err != nil {
		return "", code.CodeServerBusy
	}

	//2：调用AI模型获取回答
	//todo：这个函数后续实现
	aiResponse, err = GetAIResponse(userID, userQuestion, modelType)
	if err != nil {
		return "", code.CodeServerBusy
	}

	//3：将AI回答存储为消息
	aiMessage := &model.Message{
		SessionID: sessionID,
		Content:   aiResponse,
	}
	_, err = message.CreateMessage(aiMessage)
	if err != nil {
		return "", code.CodeServerBusy
	}

	return aiResponse, code.CodeSuccess
}

func GetChatHistory(userID int64, sessionID string) ([]model.History, code.Code) {

}
